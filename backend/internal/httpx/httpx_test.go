package httpx

import (
	"testing"
	"time"
)

func TestTokenRoundtrip(t *testing.T) {
	tok := SignToken("secret", "skye", time.Hour)
	u, err := VerifyToken("secret", tok)
	if err != nil || u != "skye" {
		t.Fatal(u, err)
	}
}
