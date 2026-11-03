package testutil

import (
	"encoding/binary"
	"io"
	"os"
)

// TensorSpec describes a tensor to encode into a QDS file.
type TensorSpec struct {
	Name     string
	Shape    []int
	DType    string // "float64" or "int64"
	Float64s []float64
	Int64s   []int64
}

// WriteQDS writes a valid QDS file containing the specified tensors.
func WriteQDS(path string, tensors []TensorSpec) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return WriteQDSStream(f, tensors)
}

// WriteQDSStream writes a QDS binary stream to w.
func WriteQDSStream(w io.Writer, tensors []TensorSpec) error {
	// First calculate total payload bytes
	var payloadBytes uint64
	for _, t := range tensors {
		// name_len(4) + name + rank(4) + dims(4*len) + elem_kind(4) + data_len(8) + data
		payloadBytes += 4 + uint64(len(t.Name)) + 4 + uint64(len(t.Shape)*4) + 4 + 8
		if t.DType == "float64" {
			payloadBytes += uint64(len(t.Float64s) * 8)
		} else {
			payloadBytes += uint64(len(t.Int64s) * 8)
		}
	}

	// 32-byte Header
	// Magic "QDS1"
	if _, err := w.Write([]byte("QDS1")); err != nil {
		return err
	}
	// n_tensors (uint32)
	if err := binary.Write(w, binary.LittleEndian, uint32(len(tensors))); err != nil {
		return err
	}
	// total_bytes (uint64)
	if err := binary.Write(w, binary.LittleEndian, payloadBytes); err != nil {
		return err
	}
	// reserved (16 bytes)
	reserved := make([]byte, 16)
	if _, err := w.Write(reserved); err != nil {
		return err
	}

	// Write each tensor
	for _, t := range tensors {
		// name_len (uint32)
		nameBytes := []byte(t.Name)
		if err := binary.Write(w, binary.LittleEndian, uint32(len(nameBytes))); err != nil {
			return err
		}
		// name
		if _, err := w.Write(nameBytes); err != nil {
			return err
		}
		// rank (uint32)
		if err := binary.Write(w, binary.LittleEndian, uint32(len(t.Shape))); err != nil {
			return err
		}
		// dims (uint32 each)
		for _, d := range t.Shape {
			if err := binary.Write(w, binary.LittleEndian, uint32(d)); err != nil {
				return err
			}
		}
		// elem_kind (uint32: 1 for float64, 0 for int64)
		var elemKind uint32
		var dataLen uint64
		if t.DType == "float64" {
			elemKind = 1
			dataLen = uint64(len(t.Float64s) * 8)
		} else {
			elemKind = 0
			dataLen = uint64(len(t.Int64s) * 8)
		}
		if err := binary.Write(w, binary.LittleEndian, elemKind); err != nil {
			return err
		}
		// data_len (uint64)
		if err := binary.Write(w, binary.LittleEndian, dataLen); err != nil {
			return err
		}
		// data
		if t.DType == "float64" {
			for _, v := range t.Float64s {
				if err := binary.Write(w, binary.LittleEndian, v); err != nil {
					return err
				}
			}
		} else {
			for _, v := range t.Int64s {
				if err := binary.Write(w, binary.LittleEndian, v); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
