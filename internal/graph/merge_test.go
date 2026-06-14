package graph

import "testing"

func TestUnionImpact(t *testing.T) {
	t.Parallel()

	results := map[string]*ImpactResult{
		"pkg/a/a.go": {
			Affected: []AffectedFile{
				{Path: "pkg/b/b.go", Depth: 1, Level: ImpactDirect},
				{Path: "pkg/c/c.go", Depth: 2, Level: ImpactIndirect},
			},
		},
		"pkg/x/x.go": {
			Affected: []AffectedFile{
				{Path: "pkg/b/b.go", Depth: 1, Level: ImpactDirect},
			},
		},
	}

	union := UnionImpact(results)
	if len(union) != 2 {
		t.Fatalf("UnionImpact() len = %d, want 2", len(union))
	}

	byPath := make(map[string]UnionEntry, len(union))
	for _, e := range union {
		byPath[e.Path] = e
	}
	b := byPath["pkg/b/b.go"]
	if len(b.TriggeredBy) != 2 {
		t.Fatalf("pkg/b triggers = %v, want 2 sources", b.TriggeredBy)
	}
}
