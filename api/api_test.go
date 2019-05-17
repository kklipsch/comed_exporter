package api

import (
	"net/http"
	"testing"
	"time"
)

func TestGetPrice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	c := &http.Client{Timeout: time.Second * 10}

	p, err := GetLastPrice(c, Address)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if p.CentsPerKWh <= 0 || p.CentsPerKWh > 100 {
		t.Fatalf("bad price parse %v", p)
	}

	if p.AsOf.Before(time.Now().Add(-48 * time.Hour)) {
		t.Fatalf("bad time parse %v", p)
	}
}
