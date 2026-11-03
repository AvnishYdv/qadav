package gpu

import (
	"errors"
	"math"
	"sync"
)

// WorkerState represents an in-process virtual worker with dedicated shared-memory gradient buffers.
type WorkerState struct {
	ID          int
	GradientBuf []byte
	BufSize     uint64
	Active      bool
	mu          sync.Mutex
}

// RingAllReduce orchestrates a ring-topology all-reduce algorithm over in-process workers.
type RingAllReduce struct {
	Workers  []*WorkerState
	RingSize int
	Compress bool // bfloat16 compression
	mu       sync.Mutex
}

// NewRingAllReduce initializes an in-process ring topology across nWorkers.
func NewRingAllReduce(nWorkers int, bufSize uint64) (*RingAllReduce, error) {
	if nWorkers < 1 {
		return nil, errors.New("nWorkers must be at least 1")
	}
	if bufSize == 0 {
		return nil, errors.New("bufSize must be greater than 0")
	}

	workers := make([]*WorkerState, nWorkers)
	for i := 0; i < nWorkers; i++ {
		workers[i] = &WorkerState{
			ID:          i,
			GradientBuf: make([]byte, bufSize),
			BufSize:     bufSize,
			Active:      true,
		}
	}

	return &RingAllReduce{
		Workers:  workers,
		RingSize: nWorkers,
		Compress: false,
	}, nil
}

// NumActive returns the number of active workers participating in the ring.
func (r *RingAllReduce) NumActive() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, w := range r.Workers {
		if w.Active {
			count++
		}
	}
	return count
}

// float64ToBfloat16 truncates float64 to 16-bit representation
func float64ToBfloat16(f float64) uint16 {
	bits := math.Float64bits(f)
	// Extract top 16 bits of IEEE 754 float64 representation
	return uint16(bits >> 48)
}

// bfloat16ToFloat64 reconstructs float64 from top 16 bits
func bfloat16ToFloat64(bf uint16) float64 {
	bits := uint64(bf) << 48
	return math.Float64frombits(bits)
}

// ScatterReduce executes Phase 1 of Ring AllReduce (scatter-reduce).
// Over N-1 steps, each worker sends a chunk of size BufSize/N to its successor
// and accumulates the chunk received from its predecessor.
func (r *RingAllReduce) ScatterReduce() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := len(r.Workers)
	if n <= 1 {
		return nil // Single worker is a trivial no-op
	}

	chunkSize := r.Workers[0].BufSize / uint64(n)
	if chunkSize == 0 {
		chunkSize = 1
	}

	// Perform n-1 scatter-reduce steps
	for step := 0; step < n-1; step++ {
		// Temporary buffer to avoid race conditions during concurrent chunk exchange
		transfers := make([][]byte, n)
		for i := 0; i < n; i++ {
			sendChunk := (i - step + n) % n
			start := uint64(sendChunk) * chunkSize
			end := start + chunkSize
			if end > r.Workers[i].BufSize {
				end = r.Workers[i].BufSize
			}
			data := make([]byte, end-start)
			copy(data, r.Workers[i].GradientBuf[start:end])
			transfers[i] = data
		}

		// Accumulate into next worker in ring (i+1)%n
		for i := 0; i < n; i++ {
			recvWorker := (i + 1) % n
			sendChunk := (i - step + n) % n
			start := uint64(sendChunk) * chunkSize
			end := start + chunkSize
			if end > r.Workers[recvWorker].BufSize {
				end = r.Workers[recvWorker].BufSize
			}

			src := transfers[i]
			dst := r.Workers[recvWorker].GradientBuf[start:end]

			// Accumulate element-wise (interpreting as float64 if 8-byte aligned, else byte addition)
			if (end-start)%8 == 0 {
				numFloats := (end - start) / 8
				for f := uint64(0); f < numFloats; f++ {
					fStart := f * 8
					v1 := math.Float64frombits(byteSliceToUint64(dst[fStart : fStart+8]))
					v2 := math.Float64frombits(byteSliceToUint64(src[fStart : fStart+8]))
					sum := v1 + v2
					if r.Compress {
						bf := float64ToBfloat16(sum)
						sum = bfloat16ToFloat64(bf)
					}
					uint64ToByteSlice(math.Float64bits(sum), dst[fStart:fStart+8])
				}
			} else {
				for b := 0; b < len(src) && b < len(dst); b++ {
					dst[b] = byte((uint16(dst[b]) + uint16(src[b])) & 0xFF)
				}
			}
		}
	}

	return nil
}

// AllGather executes Phase 2 of Ring AllReduce (all-gather).
// Over N-1 steps, each worker shares its fully reduced chunk around the ring.
func (r *RingAllReduce) AllGather() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := len(r.Workers)
	if n <= 1 {
		return nil
	}

	chunkSize := r.Workers[0].BufSize / uint64(n)
	if chunkSize == 0 {
		chunkSize = 1
	}

	for step := 0; step < n-1; step++ {
		transfers := make([][]byte, n)
		for i := 0; i < n; i++ {
			sendChunk := (i - step + 1 + n) % n
			start := uint64(sendChunk) * chunkSize
			end := start + chunkSize
			if end > r.Workers[i].BufSize {
				end = r.Workers[i].BufSize
			}
			data := make([]byte, end-start)
			copy(data, r.Workers[i].GradientBuf[start:end])
			transfers[i] = data
		}

		for i := 0; i < n; i++ {
			recvWorker := (i + 1) % n
			sendChunk := (i - step + 1 + n) % n
			start := uint64(sendChunk) * chunkSize
			end := start + chunkSize
			if end > r.Workers[recvWorker].BufSize {
				end = r.Workers[recvWorker].BufSize
			}
			copy(r.Workers[recvWorker].GradientBuf[start:end], transfers[i])
		}
	}

	// Finally, normalize by dividing by N if floats
	for i := 0; i < n; i++ {
		buf := r.Workers[i].GradientBuf
		if len(buf)%8 == 0 {
			numFloats := len(buf) / 8
			for f := 0; f < numFloats; f++ {
				fStart := f * 8
				v := math.Float64frombits(byteSliceToUint64(buf[fStart : fStart+8]))
				avg := v / float64(n)
				uint64ToByteSlice(math.Float64bits(avg), buf[fStart:fStart+8])
			}
		}
	}

	return nil
}

// Execute performs both Phase 1 (ScatterReduce) and Phase 2 (AllGather).
func (r *RingAllReduce) Execute() error {
	if err := r.ScatterReduce(); err != nil {
		return err
	}
	return r.AllGather()
}

func byteSliceToUint64(b []byte) uint64 {
	_ = b[7]
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

func uint64ToByteSlice(v uint64, b []byte) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}
