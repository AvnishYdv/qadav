package selfcheck

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	keccakRndc = [24]uint64{
		0x0000000000000001, 0x0000000000008082, 0x800000000000808a, 0x8000000080008000,
		0x000000000000808b, 0x0000000080000001, 0x8000000080008081, 0x8000000000008009,
		0x000000000000008a, 0x0000000000000088, 0x0000000080008009, 0x000000008000000a,
		0x000000008000808b, 0x800000000000008b, 0x8000000000008089, 0x8000000000008003,
		0x8000000000008002, 0x8000000000000080, 0x000000000000800a, 0x800000008000000a,
		0x8000000080008081, 0x8000000000008080, 0x0000000080000001, 0x8000000080008008,
	}
)

func rol64(a uint64, n uint) uint64 {
	return (a << n) | (a >> (64 - n))
}

func keccakF1600(s *[25]uint64) {
	for r := 0; r < 24; r++ {
		// 1. Theta
		c0 := s[0] ^ s[5] ^ s[10] ^ s[15] ^ s[20]
		c1 := s[1] ^ s[6] ^ s[11] ^ s[16] ^ s[21]
		c2 := s[2] ^ s[7] ^ s[12] ^ s[17] ^ s[22]
		c3 := s[3] ^ s[8] ^ s[13] ^ s[18] ^ s[23]
		c4 := s[4] ^ s[9] ^ s[14] ^ s[19] ^ s[24]

		d0 := c4 ^ rol64(c1, 1)
		d1 := c0 ^ rol64(c2, 1)
		d2 := c1 ^ rol64(c3, 1)
		d3 := c2 ^ rol64(c4, 1)
		d4 := c3 ^ rol64(c0, 1)

		var a, b [25]uint64
		a[0] = s[0] ^ d0
		a[1] = s[1] ^ d1
		a[2] = s[2] ^ d2
		a[3] = s[3] ^ d3
		a[4] = s[4] ^ d4
		a[5] = s[5] ^ d0
		a[6] = s[6] ^ d1
		a[7] = s[7] ^ d2
		a[8] = s[8] ^ d3
		a[9] = s[9] ^ d4
		a[10] = s[10] ^ d0
		a[11] = s[11] ^ d1
		a[12] = s[12] ^ d2
		a[13] = s[13] ^ d3
		a[14] = s[14] ^ d4
		a[15] = s[15] ^ d0
		a[16] = s[16] ^ d1
		a[17] = s[17] ^ d2
		a[18] = s[18] ^ d3
		a[19] = s[19] ^ d4
		a[20] = s[20] ^ d0
		a[21] = s[21] ^ d1
		a[22] = s[22] ^ d2
		a[23] = s[23] ^ d3
		a[24] = s[24] ^ d4

		// 2. Rho & Pi
		b[0] = a[0]
		b[10] = rol64(a[1], 1)
		b[7] = rol64(a[10], 3)
		b[11] = rol64(a[7], 6)
		b[17] = rol64(a[11], 10)
		b[18] = rol64(a[17], 15)
		b[3] = rol64(a[18], 21)
		b[5] = rol64(a[3], 28)
		b[16] = rol64(a[5], 36)
		b[8] = rol64(a[16], 45)
		b[21] = rol64(a[8], 55)
		b[24] = rol64(a[21], 2)
		b[4] = rol64(a[24], 14)
		b[15] = rol64(a[4], 27)
		b[23] = rol64(a[15], 41)
		b[19] = rol64(a[23], 56)
		b[13] = rol64(a[19], 8)
		b[12] = rol64(a[13], 25)
		b[2] = rol64(a[12], 43)
		b[20] = rol64(a[2], 62)
		b[14] = rol64(a[20], 18)
		b[22] = rol64(a[14], 39)
		b[9] = rol64(a[22], 61)
		b[6] = rol64(a[9], 20)
		b[1] = rol64(a[6], 44)

		// 3. Chi
		s[0] = b[0] ^ ((^b[1]) & b[2])
		s[1] = b[1] ^ ((^b[2]) & b[3])
		s[2] = b[2] ^ ((^b[3]) & b[4])
		s[3] = b[3] ^ ((^b[4]) & b[0])
		s[4] = b[4] ^ ((^b[0]) & b[1])

		s[5] = b[5] ^ ((^b[6]) & b[7])
		s[6] = b[6] ^ ((^b[7]) & b[8])
		s[7] = b[7] ^ ((^b[8]) & b[9])
		s[8] = b[8] ^ ((^b[9]) & b[5])
		s[9] = b[9] ^ ((^b[5]) & b[6])

		s[10] = b[10] ^ ((^b[11]) & b[12])
		s[11] = b[11] ^ ((^b[12]) & b[13])
		s[12] = b[12] ^ ((^b[13]) & b[14])
		s[13] = b[13] ^ ((^b[14]) & b[10])
		s[14] = b[14] ^ ((^b[10]) & b[11])

		s[15] = b[15] ^ ((^b[16]) & b[17])
		s[16] = b[16] ^ ((^b[17]) & b[18])
		s[17] = b[17] ^ ((^b[18]) & b[19])
		s[18] = b[18] ^ ((^b[19]) & b[15])
		s[19] = b[19] ^ ((^b[15]) & b[16])

		s[20] = b[20] ^ ((^b[21]) & b[22])
		s[21] = b[21] ^ ((^b[22]) & b[23])
		s[22] = b[22] ^ ((^b[23]) & b[24])
		s[23] = b[23] ^ ((^b[24]) & b[20])
		s[24] = b[24] ^ ((^b[20]) & b[21])

		// 4. Iota
		s[0] ^= keccakRndc[r]
	}
}

// SHA3_256 computes FIPS 202 SHA3-256 hash using zero external dependencies.
func SHA3_256(data []byte) [32]byte {
	var s [25]uint64
	rate := 136
	pos := 0
	rem := len(data)

	for rem >= rate {
		for i := 0; i < rate/8; i++ {
			s[i] ^= binary.LittleEndian.Uint64(data[pos+i*8 : pos+(i+1)*8])
		}
		keccakF1600(&s)
		pos += rate
		rem -= rate
	}

	var block [136]byte
	copy(block[:], data[pos:])
	block[rem] = 0x06
	block[135] |= 0x80

	for i := 0; i < rate/8; i++ {
		s[i] ^= binary.LittleEndian.Uint64(block[i*8 : (i+1)*8])
	}
	keccakF1600(&s)

	var out [32]byte
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(out[i*8:(i+1)*8], s[i])
	}
	return out
}

// MerkleRoot builds a binary Merkle tree over chunked code bytes and returns root hash.
func MerkleRoot(code []byte, blockSize int) ([32]byte, int) {
	if blockSize <= 0 {
		blockSize = 4096
	}
	if len(code) == 0 {
		return SHA3_256(nil), 1
	}

	nLeaves := (len(code) + blockSize - 1) / blockSize
	leaves := make([][32]byte, nLeaves)
	for i := 0; i < nLeaves; i++ {
		offset := i * blockSize
		size := blockSize
		if offset+size > len(code) {
			size = len(code) - offset
		}
		leaves[i] = SHA3_256(code[offset : offset+size])
	}

	curr := leaves
	for len(curr) > 1 {
		nextLen := (len(curr) + 1) / 2
		nextLevel := make([][32]byte, nextLen)
		for i := 0; i < nextLen; i++ {
			left := curr[i*2]
			right := left
			if i*2+1 < len(curr) {
				right = curr[i*2+1]
			}
			var pair [64]byte
			copy(pair[:32], left[:])
			copy(pair[32:], right[:])
			nextLevel[i] = SHA3_256(pair[:])
		}
		curr = nextLevel
	}

	return curr[0], nLeaves
}

func ComputeMerkleRoot(code []byte, blockSize int) ([32]byte, int) {
	return MerkleRoot(code, blockSize)
}

const (
	WotsChains = 34
	KeySize    = WotsChains * 32 // 1088 bytes
	SigSize    = WotsChains * 32 // 1088 bytes
)

type LamportKeypair struct {
	PrivateKey [KeySize]byte
	PublicKey  [KeySize]byte
}

func GenerateLamportKeypair(seed uint64) *LamportKeypair {
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	kp := &LamportKeypair{}
	x := seed
	for i := 0; i < WotsChains; i++ {
		sk := kp.PrivateKey[i*32 : (i+1)*32]
		for w := 0; w < 4; w++ {
			x ^= x << 13
			x ^= x >> 7
			x ^= x << 17
			binary.LittleEndian.PutUint64(sk[w*8:(w+1)*8], x)
		}
		// Hash 255 times to compute public key
		curr := SHA3_256(sk)
		for step := 1; step < 255; step++ {
			curr = SHA3_256(curr[:])
		}
		copy(kp.PublicKey[i*32:(i+1)*32], curr[:])
	}
	return kp
}

func computeWotsDigits(msgHash [32]byte) [WotsChains]byte {
	var digits [WotsChains]byte
	var checksum uint32
	for i := 0; i < 32; i++ {
		digits[i] = msgHash[i]
		checksum += 255 - uint32(msgHash[i])
	}
	digits[32] = byte((checksum >> 8) & 0xFF)
	digits[33] = byte(checksum & 0xFF)
	return digits
}

func LamportSign(kp *LamportKeypair, msgHash [32]byte) [SigSize]byte {
	var sig [SigSize]byte
	digits := computeWotsDigits(msgHash)
	for i := 0; i < WotsChains; i++ {
		sk := kp.PrivateKey[i*32 : (i+1)*32]
		steps := int(digits[i])
		var curr [32]byte
		copy(curr[:], sk)
		for step := 0; step < steps; step++ {
			curr = SHA3_256(curr[:])
		}
		copy(sig[i*32:(i+1)*32], curr[:])
	}
	return sig
}

func LamportVerify(pubkey []byte, msgHash [32]byte, sig []byte) bool {
	if len(pubkey) < KeySize || len(sig) < SigSize {
		return false
	}
	digits := computeWotsDigits(msgHash)
	for i := 0; i < WotsChains; i++ {
		sigChain := sig[i*32 : (i+1)*32]
		steps := 255 - int(digits[i])
		var curr [32]byte
		copy(curr[:], sigChain)
		for step := 0; step < steps; step++ {
			curr = SHA3_256(curr[:])
		}
		expectedPk := pubkey[i*32 : (i+1)*32]
		if !bytes.Equal(curr[:], expectedPk) {
			return false
		}
	}
	return true
}

// SidecarPath returns the .qsc path corresponding to an executable.
func SidecarPath(binaryPath string) string {
	ext := filepath.Ext(binaryPath)
	if ext != "" {
		return strings.TrimSuffix(binaryPath, ext) + ".qsc"
	}
	return binaryPath + ".qsc"
}

// Embed generates and writes the .qsc sidecar file for a given executable binary.
func Embed(binaryPath string, blockSize int) error {
	code, err := os.ReadFile(binaryPath)
	if err != nil {
		return fmt.Errorf("reading binary %s: %w", binaryPath, err)
	}

	root, nLeaves := MerkleRoot(code, blockSize)
	kp := GenerateLamportKeypair(0)
	sig := LamportSign(kp, root)

	header := make([]byte, 8+32+SigSize+8+8+8+8+4+4)
	copy(header[0:8], []byte("QSC1\x00\x00\x00\x00"))
	copy(header[8:40], root[:])
	copy(header[40:40+SigSize], sig[:])
	offset := 40 + SigSize
	binary.LittleEndian.PutUint64(header[offset:offset+8], uint64(SigSize))
	offset += 8
	binary.LittleEndian.PutUint64(header[offset:offset+8], 0)
	offset += 8
	binary.LittleEndian.PutUint64(header[offset:offset+8], uint64(len(code)))
	offset += 8
	binary.LittleEndian.PutUint64(header[offset:offset+8], uint64(blockSize))
	offset += 8
	binary.LittleEndian.PutUint32(header[offset:offset+4], uint32(nLeaves))
	offset += 4
	binary.LittleEndian.PutUint32(header[offset:offset+4], 0)

	// Append public key (1088 bytes)
	payload := append(header, kp.PublicKey[:]...)

	qscPath := SidecarPath(binaryPath)
	if err := os.WriteFile(qscPath, payload, 0644); err != nil {
		return fmt.Errorf("writing sidecar %s: %w", qscPath, err)
	}

	return nil
}

// Verify validates binary integrity against its corresponding .qsc sidecar.
func Verify(binaryPath string) (bool, [32]byte, error) {
	code, err := os.ReadFile(binaryPath)
	if err != nil {
		return false, [32]byte{}, fmt.Errorf("reading binary %s: %w", binaryPath, err)
	}

	qscPath := SidecarPath(binaryPath)
	sidecar, err := os.ReadFile(qscPath)
	if err != nil {
		return false, [32]byte{}, fmt.Errorf("reading sidecar %s: %w", qscPath, err)
	}

	minHeaderSize := 8 + 32 + SigSize + 8 + 8 + 8 + 8 + 4 + 4
	if len(sidecar) < minHeaderSize+KeySize {
		return false, [32]byte{}, errors.New("sidecar file too short")
	}

	if string(sidecar[0:4]) != "QSC1" {
		return false, [32]byte{}, errors.New("invalid sidecar magic")
	}

	var expectedRoot [32]byte
	copy(expectedRoot[:], sidecar[8:40])
	sig := sidecar[40 : 40+SigSize]
	pubkey := sidecar[minHeaderSize : minHeaderSize+KeySize]

	offset := 40 + SigSize + 8 + 8 + 8
	blockSize := int(binary.LittleEndian.Uint64(sidecar[offset : offset+8]))
	if blockSize <= 0 {
		blockSize = 4096
	}

	computedRoot, _ := MerkleRoot(code, blockSize)
	if !bytes.Equal(computedRoot[:], expectedRoot[:]) {
		return false, computedRoot, nil
	}

	if !LamportVerify(pubkey, computedRoot, sig) {
		return false, computedRoot, nil
	}

	return true, computedRoot, nil
}

// Strip deletes the .qsc sidecar for the given binary.
func Strip(binaryPath string) error {
	qscPath := SidecarPath(binaryPath)
	if err := os.Remove(qscPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing sidecar %s: %w", qscPath, err)
	}
	return nil
}
