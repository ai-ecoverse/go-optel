package integration

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	optel "github.com/ai-ecoverse/go-optel"
)

type collector struct {
	mu     sync.Mutex
	paths  []string
	events []optel.Event
}

func (c *collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var event optel.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.mu.Lock()
	c.paths = append(c.paths, r.URL.Path)
	c.events = append(c.events, event)
	c.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func TestBeaconsReachTheCollectorSanitized(t *testing.T) {
	sink := &collector{}
	server := httptest.NewServer(sink)
	defer server.Close()

	client := optel.Configure("slicc-cli", optel.Options{
		Rate:           "on",
		CollectBaseURL: server.URL + "/",
		Environment:    map[string]string{},
	})
	client.Sample(optel.Enter, "follow", "")
	client.ReportError("dial", errors.New(`Post "https://www.sliccy.ai/join/secret-token?x=1": dial tcp: refused`))
	client.Flush(5 * time.Second)

	sink.mu.Lock()
	defer sink.mu.Unlock()
	if len(sink.events) != 3 {
		t.Fatalf("got %d beacons, want top, enter and error", len(sink.events))
	}
	seen := map[optel.Checkpoint]optel.Event{}
	for i, event := range sink.events {
		if sink.paths[i] != "/.rum/1" {
			t.Fatalf("beacon path %q, want /.rum/1", sink.paths[i])
		}
		if event.Referer != "https://slicc-cli/" {
			t.Fatalf("referer %q", event.Referer)
		}
		seen[event.Checkpoint] = event
	}
	if seen[optel.Enter].Source != "follow" {
		t.Fatalf("enter source %q", seen[optel.Enter].Source)
	}
	failure := seen[optel.Error]
	if failure.Source != "dial" || strings.Contains(failure.Target, "secret-token") || !strings.Contains(failure.Target, "https://www.sliccy.ai/...") {
		t.Fatalf("error beacon not sanitized: %+v", failure)
	}
}

func TestRateOffSendsNothing(t *testing.T) {
	sink := &collector{}
	server := httptest.NewServer(sink)
	defer server.Close()

	client := optel.Configure("slicc-cli", optel.Options{Rate: "off", CollectBaseURL: server.URL + "/", Environment: map[string]string{}})
	client.Sample(optel.Enter, "follow", "")
	client.Flush(time.Second)

	sink.mu.Lock()
	defer sink.mu.Unlock()
	if len(sink.events) != 0 {
		t.Fatalf("got %d beacons with OPTEL_RATE off", len(sink.events))
	}
}
