package runtime

import (
	"strings"
	"testing"
)

func TestVariantRecursivePointers(t *testing.T) {
	out := runCryptoTestHarness(t, "variant_recursive_pointers")
	if !strings.Contains(out, "VARIANT_RECURSIVE_POINTERS_OK") {
		t.Fatalf("variant_recursive_pointers failed: %s", out)
	}
	t.Logf("variant_recursive_pointers verified: %s", out)
}
