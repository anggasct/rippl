package scorer

import "testing"

func TestBandForScore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		score int
		want  RiskBand
	}{
		{100, BandHigh},
		{75, BandHigh},
		{74, BandMedium},
		{50, BandMedium},
		{49, BandLow},
		{25, BandLow},
		{24, BandMinimal},
		{0, BandMinimal},
	}
	for _, tc := range tests {
		if got := BandForScore(tc.score); got != tc.want {
			t.Errorf("BandForScore(%d) = %q, want %q", tc.score, got, tc.want)
		}
	}
}
