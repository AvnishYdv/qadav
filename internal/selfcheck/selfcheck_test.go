package selfcheck

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestGoSHA3KnownVectors(t *testing.T) {
	emptyExpected := "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"
	abcExpected := "3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532"

	h1 := SHA3_256([]byte{})
	h2 := SHA3_256([]byte("abc"))
	emptyActual := hex.EncodeToString(h1[:])
	abcActual := hex.EncodeToString(h2[:])

	if emptyActual != emptyExpected {
		t.Errorf("SHA3-256(\"\") = %s, expected %s", emptyActual, emptyExpected)
	}
	if abcActual != abcExpected {
		t.Errorf("SHA3-256(\"abc\") = %s, expected %s", abcActual, abcExpected)
	}
}

func TestGoMerkleDeterministic(t *testing.T) {
	buf := make([]byte, 10000)
	for i := range buf {
		buf[i] = byte(i & 0xFF)
	}

	root1, leaves1 := ComputeMerkleRoot(buf, 4096)
	root2, leaves2 := ComputeMerkleRoot(buf, 4096)

	if root1 != root2 || leaves1 != leaves2 {
		t.Fatalf("Merkle root non-deterministic")
	}

	// Empty buffer
	emptyRoot, _ := ComputeMerkleRoot(nil, 4096)
	if hex.EncodeToString(emptyRoot[:]) != "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a" {
		t.Fatalf("unexpected empty Merkle root: %x", emptyRoot)
	}

	// Tampered buffer
	buf[5000] ^= 0xFF
	root3, _ := ComputeMerkleRoot(buf, 4096)
	if root1 == root3 {
		t.Fatalf("Merkle root did not change on tamper")
	}
}

func TestGoLamportRoundTrip(t *testing.T) {
	kp := GenerateLamportKeypair(12345)
	msgHash := SHA3_256([]byte("test message"))

	sig := LamportSign(kp, msgHash)

	if !LamportVerify(kp.PublicKey[:], msgHash, sig[:]) {
		t.Fatalf("valid Lamport signature failed verification")
	}

	// Tamper signature
	sig[0] ^= 0x01
	if LamportVerify(kp.PublicKey[:], msgHash, sig[:]) {
		t.Fatalf("tampered Lamport signature was accepted")
	}

	// Tamper message hash
	sig[0] ^= 0x01 // restore
	msgHash[0] ^= 0x01
	if LamportVerify(kp.PublicKey[:], msgHash, sig[:]) {
		t.Fatalf("tampered message hash was accepted")
	}

	// Invalid pubkey length
	if LamportVerify(kp.PublicKey[:100], msgHash, sig[:]) {
		t.Fatalf("truncated pubkey was accepted")
	}
}

func TestEmbedVerifyStrip(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "dummy.exe")
	dummyCode := []byte("FREESTANDING_TEST_BINARY_PAYLOAD_1234567890")
	if err := os.WriteFile(binPath, dummyCode, 0755); err != nil {
		t.Fatalf("writing dummy binary: %v", err)
	}

	// 1. Embed
	if err := Embed(binPath, 4096); err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	qscPath := SidecarPath(binPath)
	if _, err := os.Stat(qscPath); os.IsNotExist(err) {
		t.Fatalf("sidecar file not created: %s", qscPath)
	}

	// 2. Verify
	valid, root, err := Verify(binPath)
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if !valid {
		t.Fatalf("expected binary to verify successfully")
	}
	if hex.EncodeToString(root[:]) == "" {
		t.Fatalf("empty root returned")
	}

	// 3. Tamper binary
	tampered := append([]byte{}, dummyCode...)
	tampered[0] ^= 0xFF
	if err := os.WriteFile(binPath, tampered, 0755); err != nil {
		t.Fatalf("writing tampered binary: %v", err)
	}

	valid, _, err = Verify(binPath)
	if err != nil {
		t.Fatalf("Verify on tampered returned error: %v", err)
	}
	if valid {
		t.Fatalf("tampered binary was marked valid!")
	}

	// 4. Strip
	if err := Strip(binPath); err != nil {
		t.Fatalf("Strip failed: %v", err)
	}
	if _, err := os.Stat(qscPath); !os.IsNotExist(err) {
		t.Fatalf("sidecar file still exists after Strip")
	}
}
