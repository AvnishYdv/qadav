package gpu

import (
	"crypto/sha256"
	"errors"
	"math"
	"sync"
)

// PSUpdate captures a single worker's pushed gradient update and hash commitment.
type PSUpdate struct {
	WorkerID     int
	GradientHash [32]byte
	Update       []byte
}

// ParameterServer maintains the global authoritative model weights in shared memory.
type ParameterServer struct {
	GlobalModel []byte
	ModelSize   uint64
	Workers     []*WorkerState
	Version     uint64
	pending     []*PSUpdate
	mu          sync.Mutex
}

// NewParameterServer creates an in-process parameter server for nWorkers and modelSize.
func NewParameterServer(nWorkers int, modelSize uint64) (*ParameterServer, error) {
	if nWorkers < 1 {
		return nil, errors.New("nWorkers must be at least 1")
	}
	if modelSize == 0 {
		return nil, errors.New("modelSize must be greater than 0")
	}

	workers := make([]*WorkerState, nWorkers)
	for i := 0; i < nWorkers; i++ {
		workers[i] = &WorkerState{
			ID:          i,
			GradientBuf: make([]byte, modelSize),
			BufSize:     modelSize,
			Active:      true,
		}
	}

	return &ParameterServer{
		GlobalModel: make([]byte, modelSize),
		ModelSize:   modelSize,
		Workers:     workers,
		Version:     1,
		pending:     make([]*PSUpdate, 0, nWorkers),
	}, nil
}

// Push receives a gradient update from an active worker.
func (p *ParameterServer) Push(w *WorkerState) error {
	if w == nil {
		return errors.New("worker is nil")
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(w.GradientBuf) != int(p.ModelSize) {
		return errors.New("worker gradient buffer size mismatch")
	}

	hash := sha256.Sum256(w.GradientBuf)
	updateBuf := make([]byte, p.ModelSize)
	copy(updateBuf, w.GradientBuf)

	p.pending = append(p.pending, &PSUpdate{
		WorkerID:     w.ID,
		GradientHash: hash,
		Update:       updateBuf,
	})
	return nil
}

// Pull copies the current authoritative global model into worker buffer.
func (p *ParameterServer) Pull(w *WorkerState) error {
	if w == nil {
		return errors.New("worker is nil")
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(w.GradientBuf) != int(p.ModelSize) {
		w.GradientBuf = make([]byte, p.ModelSize)
		w.BufSize = p.ModelSize
	}

	copy(w.GradientBuf, p.GlobalModel)
	return nil
}

// Aggregate performs weighted summation over pending worker gradient updates and bumps the model version.
func (p *ParameterServer) Aggregate() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	count := len(p.pending)
	if count == 0 {
		return nil
	}

	// If model contains float64 numbers (8-byte aligned)
	if p.ModelSize%8 == 0 {
		numFloats := p.ModelSize / 8
		weight := 1.0 / float64(count)

		for f := uint64(0); f < numFloats; f++ {
			fStart := f * 8
			current := math.Float64frombits(byteSliceToUint64(p.GlobalModel[fStart : fStart+8]))
			var gradSum float64 = 0.0

			for _, upd := range p.pending {
				gVal := math.Float64frombits(byteSliceToUint64(upd.Update[fStart : fStart+8]))
				gradSum += gVal
			}

			// Apply aggregated delta (gradient step with learning rate 1.0 or direct addition)
			updated := current + (gradSum * weight)
			uint64ToByteSlice(math.Float64bits(updated), p.GlobalModel[fStart:fStart+8])
		}
	} else {
		for i := uint64(0); i < p.ModelSize; i++ {
			var sum uint32 = 0
			for _, upd := range p.pending {
				sum += uint32(upd.Update[i])
			}
			avg := byte(sum / uint32(count))
			p.GlobalModel[i] = byte((uint16(p.GlobalModel[i]) + uint16(avg)) & 0xFF)
		}
	}

	p.pending = p.pending[:0]
	p.Version++
	return nil
}

// GetVersion returns the monotonic version counter of the parameter server.
func (p *ParameterServer) GetVersion() uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Version
}
