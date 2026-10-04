package retrieve

import (
	"testing"

	"github.com/raj-khan/rag-xray/internal/store"
)

func TestRRFRewardsAgreement(t *testing.T) {
	a, b, c := &store.Chunk{ID: 0}, &store.Chunk{ID: 1}, &store.Chunk{ID: 2}
	vector := []Hit{{a, 0.9}, {b, 0.8}, {c, 0.7}}
	keyword := []Hit{{b, 12}, {c, 3}}
	got := top(RRF(vector, keyword), 3)
	if got[0].Chunk != b {
		t.Errorf("chunk ranked high in both lists should win, got ID %d", got[0].Chunk.ID)
	}
	if len(got) != 3 {
		t.Errorf("fusion should keep every chunk, got %d", len(got))
	}
}
