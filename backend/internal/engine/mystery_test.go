package engine

import (
	"testing"

	"github.com/gotosky/gotosky/internal/domain"
)

func TestRecommendNotRandom(t *testing.T) {
	low := Candidate{Target: domain.Target{CatalogID: "M1"}, MeanScore: 55, PeakScore: 60, FOVFit: 0.4}
	high := Candidate{Target: domain.Target{CatalogID: "M31"}, MeanScore: 88, PeakScore: 92, FOVFit: 1.0}
	out := Recommend([]Candidate{low, high}, 1)
	if len(out) != 1 || out[0].Target.CatalogID != "M31" {
		t.Fatalf("%+v", out)
	}
}

func TestFOVFitComfort(t *testing.T) {
	if FOVFit(40, 80) < 0.9 {
		t.Fatal("comfort zone")
	}
}
