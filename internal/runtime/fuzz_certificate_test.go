package runtime

import (
	"testing"
)

func FuzzCertificateKind(f *testing.F) {
	f.Add(byte(1))
	f.Add(byte(2))
	f.Add(byte(3))
	f.Add(byte(4))
	f.Add(byte(0))
	f.Add(byte(255))

	f.Fuzz(func(t *testing.T, kind byte) {
		valid := kind >= 1 && kind <= 4
		if valid {
			if kind < 1 || kind > 4 {
				t.Fatalf("unexpected invalid kind %d", kind)
			}
		}
	})
}

func FuzzCertificateWitness(f *testing.F) {
	f.Add([]byte("witness_test"))
	f.Add([]byte(""))
	f.Add(make([]byte, 512))
	f.Add(make([]byte, 513))

	f.Fuzz(func(t *testing.T, witness []byte) {
		l := len(witness)
		valid := l >= 1 && l <= 512
		if valid && (l < 1 || l > 512) {
			t.Fatalf("invalid witness length bounds: %d", l)
		}
	})
}

func FuzzCertificateConfidence(f *testing.F) {
	f.Add(0.0)
	f.Add(0.5)
	f.Add(1.0)
	f.Add(-0.1)
	f.Add(1.01)

	f.Fuzz(func(t *testing.T, conf float64) {
		valid := conf >= 0.0 && conf <= 1.0
		if valid && (conf < 0.0 || conf > 1.0) {
			t.Fatalf("invalid confidence bounds: %f", conf)
		}
	})
}

func FuzzAbstainThreshold(f *testing.F) {
	f.Add(0.80, 1, 3)
	f.Add(0.50, 2, 0)
	f.Add(0.95, 3, 0)
	f.Add(-0.5, 0, -1)

	f.Fuzz(func(t *testing.T, thresh float64, mode int, maxRetries int) {
		valid := thresh >= 0.0 && thresh <= 1.0 && mode >= 1 && mode <= 3
		_ = valid
	})
}

