package fetcher

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
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
		t.Fatalf("expected error for invalid IP, got none")
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
		t.Fatalf("expected error for non-200 http response, got none")
	}
}

// tests about lower level network stuff

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newFakeClient(transport string, rt roundTripFunc) *client {
	return &client{Client: http.Client{Transport: rt}, transport: transport}
}

func TestFetchIp_NetworkUnreachable(t *testing.T) {
	c := newFakeClient("tcp6", func(*http.Request) (*http.Response, error) {
		return nil, &net.OpError{Op: "dial", Net: "tcp6", Err: os.NewSyscallError("connect", syscall.ENETUNREACH)}
	})

	var ip net.IP
	if err := c.fetchIp(&ip, t.Context(), "http://example.invalid"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ip != nil {
		t.Fatalf("expected nil ip, got %v", ip)
	}
}

func TestFetchIp_OtherNetworkError(t *testing.T) {
	c := newFakeClient("tcp4", func(*http.Request) (*http.Response, error) {
		return nil, &net.OpError{Op: "dial", Net: "tcp4", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}
	})

	var ip net.IP
	if err := c.fetchIp(&ip, t.Context(), "http://example.invalid"); err == nil {
		t.Fatal("expected an error")
	}
}
