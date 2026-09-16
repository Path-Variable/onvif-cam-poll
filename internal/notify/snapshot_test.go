package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serve(t *testing.T, status int, contentType, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchSnapshotAcceptsImages(t *testing.T) {
	srv := serve(t, http.StatusOK, "image/jpeg", "\xff\xd8\xff\xe0jpeg-bytes")
	got, err := FetchSnapshot(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "\xff\xd8\xff\xe0jpeg-bytes" {
		t.Fatalf("body mismatch: %q", got)
	}
}

func TestFetchSnapshotRejectsNonImages(t *testing.T) {
	cases := map[string]*httptest.Server{
		"unauthorized":  serve(t, http.StatusUnauthorized, "text/html", "<html>login</html>"),
		"html with 200": serve(t, http.StatusOK, "text/html; charset=utf-8", "<html>oops</html>"),
		"empty image":   serve(t, http.StatusOK, "image/jpeg", ""),
	}
	for name, srv := range cases {
		if _, err := FetchSnapshot(context.Background(), srv.Client(), srv.URL); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestFetchSnapshotRejectsOversize(t *testing.T) {
	srv := serve(t, http.StatusOK, "image/jpeg", strings.Repeat("x", MaxSnapshotBytes+1))
	if _, err := FetchSnapshot(context.Background(), srv.Client(), srv.URL); err == nil {
		t.Fatal("expected an error for an oversize body")
	}
}

func TestFetchSnapshotHonoursContext(t *testing.T) {
	srv := serve(t, http.StatusOK, "image/jpeg", "x")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := FetchSnapshot(ctx, srv.Client(), srv.URL); err == nil {
		t.Fatal("expected an error from a cancelled context")
	}
}
