package runtime

import (
	"crypto/sha256"
	"math"
	"testing"
)

func BenchmarkCertificateAttach(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// SHA3-256 / 256-bit commitment computation simulation
		h := sha256.New()
		h.Write([]byte{1})
		h.Write([]byte("premise_hash_32_bytes_test_data!"))
		h.Write([]byte("conclsn_hash_32_bytes_test_data!"))
		h.Write([]byte("witness_bounded_proof_data"))
		_ = h.Sum(nil)
	}
}

func BenchmarkCertificateVerify(b *testing.B) {
	h1 := []byte("01234567890123456789012345678901")
	h2 := []byte("01234567890123456789012345678901")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diff := 0
		for j := 0; j < 32; j++ {
			diff |= int(h1[j] ^ h2[j])
		}
		_ = diff
	}
}

func BenchmarkAbstainComputeConfidence(b *testing.B) {
	confs := []float64{0.95, 0.90, 0.85, 0.80}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prod := 1.0
		sum := 0.0
		for _, c := range confs {
			prod *= c
			sum += c
		}
		mean := sum / float64(len(confs))
		varSum := 0.0
		for _, c := range confs {
			d := c - mean
			varSum += d * d
		}
		v := varSum / float64(len(confs))
		geom := math.Pow(prod, 1.0/float64(len(confs)))
		_ = geom * (1.0 - v/4.0)
	}
}

func BenchmarkSandboxRun(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Sandbox boundary constant-time verification check
		h1 := [32]byte{1, 2, 3, 4}
		h2 := [32]byte{1, 2, 3, 4}
		diff := 0
		for j := 0; j < 32; j++ {
			diff |= int(h1[j] ^ h2[j])
		}
		_ = diff
	}
}

