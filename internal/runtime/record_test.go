package runtime

import (
	"strings"
	"testing"
)

func TestRecordAllocAndFields(t *testing.T) {
	out := runCryptoTestHarness(t, "record_alloc_and_fields")
	if !strings.Contains(out, "RECORD_ALLOC_AND_FIELDS_OK") {
		t.Fatalf("record_alloc_and_fields failed: %s", out)
	}
	t.Logf("record_alloc_and_fields verified: %s", out)
}
