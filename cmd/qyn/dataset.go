package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func runDataset(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: qyn dataset inspect <path.qds>")
		return 1
	}

	switch args[0] {
	case "inspect":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: missing dataset file path")
			fmt.Fprintln(os.Stderr, "Usage: qyn dataset inspect <path.qds>")
			return 1
		}
		return inspectDataset(args[1])
	default:
		fmt.Fprintf(os.Stderr, "Unknown dataset subcommand: %s\n", args[0])
		return 1
	}
}

func inspectDataset(path string) int {
	cleanPath := filepath.ToSlash(filepath.Clean(path))
	f, err := os.Open(cleanPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		return 1
	}
	defer func() { _ = f.Close() }()

	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading header: %v\n", err)
		return 1
	}
	if string(magic[:]) != "QDS1" {
		fmt.Fprintf(os.Stderr, "Invalid QDS magic: %q\n", string(magic[:]))
		return 1
	}

	var nTensors uint32
	if err := binary.Read(f, binary.LittleEndian, &nTensors); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading tensor count: %v\n", err)
		return 1
	}

	var totalBytes uint64
	if err := binary.Read(f, binary.LittleEndian, &totalBytes); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading payload byte count: %v\n", err)
		return 1
	}

	var reserved [16]byte
	if _, err := io.ReadFull(f, reserved[:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading reserved header bytes: %v\n", err)
		return 1
	}

	fmt.Printf("QDS Dataset: %s\n", cleanPath)
	fmt.Printf("Version: 1\n")
	fmt.Printf("Tensors: %d\n", nTensors)

	var payloadDataBytes uint64
	for i := uint32(0); i < nTensors; i++ {
		var nameLen uint32
		if err := binary.Read(f, binary.LittleEndian, &nameLen); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading tensor %d name length: %v\n", i, err)
			return 1
		}
		nameBuf := make([]byte, nameLen)
		if _, err := io.ReadFull(f, nameBuf); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading tensor %d name: %v\n", i, err)
			return 1
		}

		var rank uint32
		if err := binary.Read(f, binary.LittleEndian, &rank); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading tensor %d rank: %v\n", i, err)
			return 1
		}

		dims := make([]uint32, rank)
		for r := uint32(0); r < rank; r++ {
			if err := binary.Read(f, binary.LittleEndian, &dims[r]); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading tensor %d dim %d: %v\n", i, r, err)
				return 1
			}
		}

		var elemKind uint32
		if err := binary.Read(f, binary.LittleEndian, &elemKind); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading tensor %d elem kind: %v\n", i, err)
			return 1
		}

		var dataLen uint64
		if err := binary.Read(f, binary.LittleEndian, &dataLen); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading tensor %d data len: %v\n", i, err)
			return 1
		}

		payloadDataBytes += dataLen

		// Skip data
		if _, err := f.Seek(int64(dataLen), io.SeekCurrent); err != nil {
			fmt.Fprintf(os.Stderr, "Error seeking tensor %d data: %v\n", i, err)
			return 1
		}

		dtypeStr := "int64"
		if elemKind == 1 {
			dtypeStr = "float64"
		}

		shapeStr := "["
		for dIdx, d := range dims {
			if dIdx > 0 {
				shapeStr += ", "
			}
			shapeStr += fmt.Sprintf("%d", d)
		}
		shapeStr += "]"

		fmt.Printf("Tensor %d: shape=%s, dtype=%s, bytes=%d\n", i, shapeStr, dtypeStr, dataLen)
	}

	fmt.Printf("Total Size: %d bytes (header: 32 bytes, payload: %d bytes)\n", payloadDataBytes, payloadDataBytes)
	fmt.Printf("Status: VALID\n")
	return 0
}
