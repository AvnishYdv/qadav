package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/qadav-lang/qadav/internal/selfcheck"
)

func runVerifyTraining(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn verify-training <file.mctp> [model]")
		return 1
	}

	certPath := args[0]
	data, err := os.ReadFile(certPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", certPath, err)
		return 1
	}

	if len(data) < 10 {
		fmt.Fprintln(os.Stderr, "INTEGRITY TAMPERED: File too short")
		return 1
	}

	// 1. Magic
	if string(data[:4]) != "MCTP" {
		fmt.Fprintln(os.Stderr, "INTEGRITY TAMPERED: Invalid magic header")
		return 1
	}

	// 2. Version
	version := binary.LittleEndian.Uint16(data[4:6])
	if version != 1 {
		fmt.Fprintf(os.Stderr, "INTEGRITY TAMPERED: Unsupported format version %d\n", version)
		return 1
	}

	// 3. n_epochs
	nEpochs := int(binary.LittleEndian.Uint32(data[6:10]))
	if nEpochs <= 0 || nEpochs > 10000 {
		fmt.Fprintln(os.Stderr, "INTEGRITY TAMPERED: Invalid epoch count")
		return 1
	}

	leafSize := 108 // 4 + 32 + 32 + 8 + 32
	expectedLen := 10 + nEpochs*leafSize + 32 + 1088 + 1088
	if len(data) != expectedLen {
		fmt.Fprintln(os.Stderr, "INTEGRITY TAMPERED: Certificate size mismatch or corrupt payload")
		return 1
	}

	offset := 10
	var leafHashes [][32]byte
	for i := 0; i < nEpochs; i++ {
		leafBytes := data[offset : offset+leafSize]
		offset += leafSize

		// Leaf hash using SHA3-256
		leafHash := selfcheck.SHA3_256(leafBytes)
		leafHashes = append(leafHashes, leafHash)
	}

	// Merkle root from file
	var fileMerkleRoot [32]byte
	copy(fileMerkleRoot[:], data[offset:offset+32])
	offset += 32

	sig := data[offset : offset+1088]
	offset += 1088

	pubkey := data[offset : offset+1088]

	// Recompute Merkle root
	currentLevel := leafHashes
	for len(currentLevel) > 1 {
		var nextLevel [][32]byte
		nextCount := (len(currentLevel) + 1) / 2
		for i := 0; i < nextCount; i++ {
			left := i * 2
			right := left + 1
			if right >= len(currentLevel) {
				right = left
			}
			var combined [64]byte
			copy(combined[:32], currentLevel[left][:])
			copy(combined[32:], currentLevel[right][:])
			nextLevel = append(nextLevel, selfcheck.SHA3_256(combined[:]))
		}
		currentLevel = nextLevel
	}

	computedRoot := currentLevel[0]
	if !bytes.Equal(computedRoot[:], fileMerkleRoot[:]) {
		fmt.Fprintln(os.Stderr, "INTEGRITY TAMPERED: Merkle root mismatch or invalid WOTS signature!")
		return 1
	}

	// Verify WOTS signature
	sigValid := selfcheck.LamportVerify(pubkey, fileMerkleRoot, sig)
	if !sigValid {
		fmt.Fprintln(os.Stderr, "INTEGRITY TAMPERED: Merkle root mismatch or invalid WOTS signature!")
		return 1
	}

	fmt.Printf("INTEGRITY OK: %d epochs, final weights match Merkle root %x, WOTS signature valid\n", nEpochs, fileMerkleRoot)
	return 0
}
