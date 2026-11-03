package runtime

import (
	"strings"
	"testing"
)

func TestVariantCreateAndMatch(t *testing.T) {
	out := runCryptoTestHarness(t, "variant_create_and_match")
	if !strings.Contains(out, "VARIANT_CREATE_AND_MATCH_OK") {
		t.Fatalf("variant_create_and_match failed: %s", out)
	}
	t.Logf("variant_create_and_match verified: %s", out)
}
