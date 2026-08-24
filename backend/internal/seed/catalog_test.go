package seed

import "testing"

func TestCatalogSize(t *testing.T) {
	c := Catalog()
	if len(c) < 120 {
		t.Fatalf("%d", len(c))
	}
	seen := map[string]bool{}
	for _, x := range c {
		if seen[x.CatalogID] {
			t.Fatalf("dup %s", x.CatalogID)
		}
		seen[x.CatalogID] = true
	}
	if !seen["M31"] {
		t.Fatal("missing M31")
	}
}
