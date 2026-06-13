package scorer

func BandForScore(score int) RiskBand {
	switch {
	case score >= 75:
		return BandHigh
	case score >= 50:
		return BandMedium
	case score >= 25:
		return BandLow
	default:
		return BandMinimal
	}
}
