package ovh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alexzouiten/ovh-ddns/internal/model"
	"github.com/alexzouiten/ovh-ddns/internal/ovh/dto"
	ovhapi "github.com/ovh/go-ovh/ovh"
)

// ---------------------------------------------------------------------------
// Test infrastructure: fake the HTTP layer of *ovhapi.Client
// ---------------------------------------------------------------------------

type recordedRequest struct {
	Method string
	Path   string // path + query, without the "/1.0" version prefix
	Body   []byte
}

type fakeTransport struct {
	mu       sync.Mutex
	requests []recordedRequest
	handler  func(r recordedRequest) (status int, body any)
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
	}
	rec := recordedRequest{
		Method: req.Method,
		// The endpoint is https://eu.api.ovh.com/1.0, so strip the version prefix.
		Path: strings.TrimPrefix(req.URL.RequestURI(), "/1.0"),
		Body: body,
	}

	// The client calls GET /auth/time before its first authenticated request.
	if rec.Method == http.MethodGet && rec.Path == "/auth/time" {
		return jsonResponse(http.StatusOK, time.Now().Unix()), nil
	}

	f.mu.Lock()
	f.requests = append(f.requests, rec)
	f.mu.Unlock()

	status, resp := f.handler(rec)
	return jsonResponse(status, resp), nil
}

func jsonResponse(status int, body any) *http.Response {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(b)),
	}
}

func newTestOvh(t *testing.T, dryRun bool, h func(recordedRequest) (int, any)) (*Ovh, *fakeTransport) {
	t.Helper()

	c, err := ovhapi.NewClient("ovh-eu", "key", "secret", "consumer")
	if err != nil {
		t.Fatal(err)
	}

	ft := &fakeTransport{handler: h}
	c.Client = &http.Client{Transport: ft}

	return NewFromClient(c, dryRun), ft
}

// routes builds a handler that answers from a static "METHOD /path?query" map
// and returns 404 for anything unexpected.
func routes(m map[string]any) func(recordedRequest) (int, any) {
	return func(r recordedRequest) (int, any) {
		key := r.Method + " " + r.Path
		if v, ok := m[key]; ok {
			return http.StatusOK, v
		}
		return http.StatusNotFound, map[string]string{"message": "unexpected request: " + key}
	}
}

// ---------------------------------------------------------------------------
// GetRecord
// ---------------------------------------------------------------------------

func TestGetRecord(t *testing.T) {
	ovh, _ := newTestOvh(t, false, routes(map[string]any{
		"GET /domain/zone/example.com/record?fieldType=A&subDomain=":    []int{1},
		"GET /domain/zone/example.com/record/1":                         newARecordDto("1.2.3.4"),
		"GET /domain/zone/example.com/record?fieldType=AAAA&subDomain=": []int{2},
		"GET /domain/zone/example.com/record/2":                         newAAAARecordDto("2001:db8::1"),
	}))

	a, err := ovh.GetRecord(context.Background(), "example.com", model.RecordA, "")
	if err != nil {
		t.Fatalf("GetRecord(A) returned error: %v", err)
	}
	if !a.Target.Equal(net.ParseIP("1.2.3.4")) {
		t.Errorf("unexpected ipv4: %v", a.Target)
	}

	aaaa, err := ovh.GetRecord(context.Background(), "example.com", model.RecordAAAA, "")
	if err != nil {
		t.Fatalf("GetRecord(AAAA) returned error: %v", err)
	}
	if !aaaa.Target.Equal(net.ParseIP("2001:db8::1")) {
		t.Errorf("unexpected ipv6: %v", aaaa.Target)
	}
}

func TestGetRecordNoRecord(t *testing.T) {
	ovh, _ := newTestOvh(t, false, routes(map[string]any{
		"GET /domain/zone/example.com/record?fieldType=A&subDomain=": []int{},
	}))

	record, err := ovh.GetRecord(context.Background(), "example.com", model.RecordA, "")
	if record != nil {
		t.Errorf("expected returned record to be nil")
	}
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGetRecordMultipleRecords(t *testing.T) {
	ovh, _ := newTestOvh(t, false, routes(map[string]any{
		"GET /domain/zone/example.com/record?fieldType=A&subDomain=": []int{1, 2},
	}))

	_, err := ovh.GetRecord(context.Background(), "example.com", model.RecordA, "")
	if !errors.Is(err, ErrMultipleRecords) {
		t.Fatalf("expected ErrMultipleRecords, got %v", err)
	}
}

func TestGetRecordAPIError(t *testing.T) {
	ovh, _ := newTestOvh(t, false, func(recordedRequest) (int, any) {
		return http.StatusInternalServerError, map[string]string{"message": "boom"}
	})

	_, err := ovh.GetRecord(context.Background(), "example.com", model.RecordA, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// PostRecord
// ---------------------------------------------------------------------------

func TestPostRecord(t *testing.T) {
	ovh, ft := newTestOvh(t, false, func(recordedRequest) (int, any) {
		return http.StatusOK, dto.Record{
			Id:        42,
			FieldType: model.RecordA,
			Target:    net.ParseIP("1.2.3.4"),
		}
	})

	got, err := ovh.PostRecord(context.Background(), model.Record{
		Zone:   "example.com",
		Type:   model.RecordA,
		Target: net.ParseIP("1.2.3.4"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != 42 {
		t.Errorf("expected id 42, got %d", got.Id)
	}

	if len(ft.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(ft.requests))
	}
	req := ft.requests[0]
	if req.Method != http.MethodPost || req.Path != "/domain/zone/example.com/record" {
		t.Errorf("unexpected request: %s %s", req.Method, req.Path)
	}

	// Keys are OVH's API field names; they must match the json tags on dto.RecordPost.
	var sent map[string]any
	if err := json.Unmarshal(req.Body, &sent); err != nil {
		t.Fatal(err)
	}
	if sent["fieldType"] != "A" || sent["target"] != "1.2.3.4" {
		t.Errorf("unexpected body: %s", req.Body)
	}
}

func TestPostRecordAPIError(t *testing.T) {
	ovh, _ := newTestOvh(t, false, func(recordedRequest) (int, any) {
		return http.StatusInternalServerError, map[string]string{"message": "boom"}
	})

	_, err := ovh.PostRecord(context.Background(), model.Record{
		Zone: "example.com",
		Type: model.RecordA,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// PutRecord
// ---------------------------------------------------------------------------

func TestPutRecord(t *testing.T) {
	ovh, ft := newTestOvh(t, false, func(recordedRequest) (int, any) {
		return http.StatusNoContent, nil // OVH returns no body on PUT
	})

	err := ovh.PutRecord(context.Background(), model.Record{
		Id:     7,
		Zone:   "example.com",
		Type:   model.RecordAAAA,
		Target: net.ParseIP("2001:db8::1"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ft.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(ft.requests))
	}
	req := ft.requests[0]
	if req.Method != http.MethodPut || req.Path != "/domain/zone/example.com/record/7" {
		t.Errorf("unexpected request: %s %s", req.Method, req.Path)
	}

	var sent map[string]any
	if err := json.Unmarshal(req.Body, &sent); err != nil {
		t.Fatal(err)
	}
	if sent["target"] != "2001:db8::1" {
		t.Errorf("unexpected body: %s", req.Body)
	}
}

func TestPutRecordAPIError(t *testing.T) {
	ovh, _ := newTestOvh(t, false, func(recordedRequest) (int, any) {
		return http.StatusNotFound, map[string]string{"message": "not found"}
	})

	err := ovh.PutRecord(context.Background(), model.Record{
		Id:   7,
		Zone: "example.com",
		Type: model.RecordA,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// Refresh
// ---------------------------------------------------------------------------

func TestRefresh(t *testing.T) {
	ovh, ft := newTestOvh(t, false, func(recordedRequest) (int, any) {
		return http.StatusOK, nil
	})

	if err := ovh.Refresh(context.Background(), "example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ft.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(ft.requests))
	}
	if got := ft.requests[0]; got.Method != http.MethodPost || got.Path != "/domain/zone/example.com/refresh" {
		t.Errorf("unexpected request: %s %s", got.Method, got.Path)
	}
}

// ---------------------------------------------------------------------------
// Dry run
// ---------------------------------------------------------------------------

func TestDryRunMakesNoRequests(t *testing.T) {
	ovh, ft := newTestOvh(t, true, func(recordedRequest) (int, any) {
		t.Error("handler must not be called in dry run")
		return http.StatusOK, nil
	})

	ctx := context.Background()
	r := model.Record{
		Id:     1,
		Zone:   "example.com",
		Type:   model.RecordA,
		Target: net.ParseIP("1.2.3.4"),
	}

	if _, err := ovh.PostRecord(ctx, r); err != nil {
		t.Fatal(err)
	}
	if err := ovh.PutRecord(ctx, r); err != nil {
		t.Fatal(err)
	}
	if err := ovh.Refresh(ctx, "example.com"); err != nil {
		t.Fatal(err)
	}

	if len(ft.requests) != 0 {
		t.Errorf("expected no requests, got %d", len(ft.requests))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newARecordDto(target string) dto.Record {
	return dto.Record{
		Target:    net.ParseIP(target),
		FieldType: model.RecordA,
	}
}

func newAAAARecordDto(target string) dto.Record {
	return dto.Record{
		Target:    net.ParseIP(target),
		FieldType: model.RecordAAAA,
	}
}
