package runtime

import (
	"strings"
	"testing"
)

func TestDistInitSingleWorker(t *testing.T) {
	out := runCryptoTestHarness(t, "dist_init_single_worker")
	if !strings.Contains(out, "DIST_INIT_SINGLE_WORKER_OK") {
		t.Fatalf("dist_init_single_worker failed: %s", out)
	}
	t.Logf("dist_init_single_worker verified: %s", out)
}

func TestDistRingAllReduce2Workers(t *testing.T) {
	out := runCryptoTestHarness(t, "dist_ring_allreduce_2workers")
	if !strings.Contains(out, "DIST_RING_ALLREDUCE_2WORKERS_OK") {
		t.Fatalf("dist_ring_allreduce_2workers failed: %s", out)
	}
	t.Logf("dist_ring_allreduce_2workers verified: %s", out)
}

func TestDistRingAllReduce4Workers(t *testing.T) {
	out := runCryptoTestHarness(t, "dist_ring_allreduce_4workers")
	if !strings.Contains(out, "DIST_RING_ALLREDUCE_4WORKERS_OK") {
		t.Fatalf("dist_ring_allreduce_4workers failed: %s", out)
	}
	t.Logf("dist_ring_allreduce_4workers verified: %s", out)
}

func TestDistPserverAggregate(t *testing.T) {
	out := runCryptoTestHarness(t, "dist_pserver_aggregate")
	if !strings.Contains(out, "DIST_PSERVER_AGGREGATE_OK") {
		t.Fatalf("dist_pserver_aggregate failed: %s", out)
	}
	t.Logf("dist_pserver_aggregate verified: %s", out)
}

func TestDistFallbackToSingleWorker(t *testing.T) {
	out := runCryptoTestHarness(t, "dist_fallback_single_worker")
	if !strings.Contains(out, "DIST_FALLBACK_SINGLE_WORKER_OK") {
		t.Fatalf("dist_fallback_single_worker failed: %s", out)
	}
	t.Logf("dist_fallback_single_worker verified: %s", out)
}
