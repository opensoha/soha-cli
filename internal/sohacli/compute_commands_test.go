package sohacli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestComputeCommandsCoverReadAndMutationSurfaces(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		switch requestCount {
		case 1:
			if r.Method != http.MethodGet || r.URL.Path != "/api/v1/compute/overview" {
				t.Fatalf("overview request = %s %s", r.Method, r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"data":{"attention":[],"providerHealth":[],"partial":false,"warnings":[]}}`))
		case 2:
			if r.Method != http.MethodPost || r.URL.Path != "/api/v1/compute/provider-instances/virtualization/pve/connection-1/health-checks" {
				t.Fatalf("health request = %s %s", r.Method, r.URL.Path)
			}
			if key := r.Header.Get("Idempotency-Key"); key != "health-key-1" {
				t.Fatalf("health key = %q", key)
			}
			_, _ = w.Write(computeTaskResponse(t, "health-1"))
		case 3:
			if r.Method != http.MethodGet || r.URL.Path != "/api/v1/compute/resources/virtualization/vm/vm-1/relations" {
				t.Fatalf("relations request = %s %s", r.Method, r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"data":{"resource":{"domain":"virtualization","kind":"vm","id":"vm-1","displayName":"vm-1"},"relations":[]}}`))
		case 4:
			if r.Method != http.MethodGet || r.URL.Path != "/api/v1/compute/tasks" || r.URL.Query().Get("sortBy") != "status" || r.URL.Query().Get("sortOrder") != "desc" {
				t.Fatalf("tasks request = %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"items":[]}`))
		case 5:
			if r.Method != http.MethodPost || r.URL.Path != "/api/v1/compute/tasks/virtualization/task-1/retry" {
				t.Fatalf("retry request = %s %s", r.Method, r.URL.Path)
			}
			if key := r.Header.Get("Idempotency-Key"); key != "retry-key-1" {
				t.Fatalf("retry key = %q", key)
			}
			_, _ = w.Write(computeTaskResponse(t, "task-1"))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	configPath := writeTestConfig(t, server.URL)
	commands := [][]string{
		{"compute", "overview", "--profile", "dev"},
		{"compute", "provider-instances", "health", "virtualization", "pve", "connection-1", "--generation", "1", "--idempotency-key", "health-key-1", "--profile", "dev"},
		{"compute", "resources", "relations", "virtualization", "vm", "vm-1", "--profile", "dev"},
		{"compute", "tasks", "list", "--sort-by", "status", "--sort-order", "desc", "--profile", "dev"},
		{"compute", "tasks", "retry", "virtualization", "task-1", "--yes", "--idempotency-key", "retry-key-1", "--profile", "dev"},
	}
	for _, command := range commands {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), command, Runtime{Out: &stdout, Err: &stderr, ConfigPath: configPath}); code != 0 {
			t.Fatalf("Run(%v) code=%d stderr=%s", command, code, stderr.String())
		}
		if stdout.Len() == 0 {
			t.Fatalf("Run(%v) produced no output", command)
		}
	}
}

func computeTaskResponse(t *testing.T, id string) []byte {
	t.Helper()
	value := map[string]any{"data": map[string]any{
		"id": id, "domain": "virtualization", "sourceType": "virtualization_task", "sourceId": id,
		"kind": "vm_action", "category": "lifecycle", "normalizedStatus": "queued", "rawStatus": "queued",
		"resources": []any{}, "availableActions": []any{}, "createdAt": "2026-08-24T12:00:00Z",
	}}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
