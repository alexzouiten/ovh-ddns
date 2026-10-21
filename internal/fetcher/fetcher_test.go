package fetcher

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchIp_Success(t *testing.T) {
	expected := net.ParseIP("203.22.13.56")

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expected.String()))
	}))
	defer testServer.Close()

	var actual net.IP
	client := newTransportSpecificClient("tcp4")

	err := client.fetchIp(&actual, t.Context(), testServer.URL)
	if err != nil {
		t.Fatalf("fetcherIp returned error: %v", err)
	}

	if !actual.Equal(expected) {
		t.Fatalf("expected IP %s, got %s", expected, actual)
	}
}

func TestFetchIp_InvalidIp(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-an-ip"))
	}))
	defer testServer.Close()

	var actual net.IP
	client := newTransportSpecificClient("tcp4")

	err := client.fetchIp(&actual, t.Context(), testServer.URL)
	if err == nil {
		t.Fatalf("expected error for invalid IP, got non")
	}
}

func TestFetchIp_HttpErrorStatus(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(""))
	}))
	defer testServer.Close()

	var actual net.IP
	client := newTransportSpecificClient("tcp4")

	err := client.fetchIp(&actual, t.Context(), testServer.URL)
	if err == nil {
		t.Fatalf("expected error for non-200 http response, got non")
	}
}
