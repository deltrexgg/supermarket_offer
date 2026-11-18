package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeCounter struct {
	counts map[string]int64
	err    error
}

func (f *fakeCounter) Incr(_ context.Context, key string, _ time.Duration) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	f.counts[key]++
	return f.counts[key], nil
}

func newRouter(c Counter, limit int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.GET("/", Middleware(c, "test", limit, time.Minute), func(ctx *gin.Context) { ctx.String(http.StatusOK, "ok") })
	return r
}

func get(r *gin.Engine, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = ip + ":5000"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestBlocksAfterTheLimit(t *testing.T) {
	r := newRouter(&fakeCounter{counts: map[string]int64{}}, 3)
	for i := 1; i <= 3; i++ {
		if w := get(r, "10.0.0.1"); w.Code != http.StatusOK {
			t.Fatalf("request %d: status %d, want 200", i, w.Code)
		}
	}
	w := get(r, "10.0.0.1")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("4th request: status %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("429 response needs a Retry-After header")
	}
}

func TestEachIPHasItsOwnCounter(t *testing.T) {
	r := newRouter(&fakeCounter{counts: map[string]int64{}}, 1)
	if w := get(r, "10.0.0.1"); w.Code != http.StatusOK {
		t.Fatalf("first client: %d", w.Code)
	}
	if w := get(r, "10.0.0.2"); w.Code != http.StatusOK {
		t.Fatalf("second client was blocked by the first: %d", w.Code)
	}
}

func TestForgedForwardedForIsIgnored(t *testing.T) {
	r := newRouter(&fakeCounter{counts: map[string]int64{}}, 1)
	for i, fake := range []string{"1.1.1.1", "2.2.2.2"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.9:5000"
		req.Header.Set("X-Forwarded-For", fake)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if i == 1 && w.Code != http.StatusTooManyRequests {
			t.Fatalf("changing X-Forwarded-For must not reset the limit, got %d", w.Code)
		}
	}
}

func TestAllowsRequestsWhenTheCounterFails(t *testing.T) {
	r := newRouter(&fakeCounter{err: context.DeadlineExceeded}, 1)
	for i := 0; i < 3; i++ {
		if w := get(r, "10.0.0.1"); w.Code != http.StatusOK {
			t.Fatalf("a counter outage must not block requests, got %d", w.Code)
		}
	}
}

func TestZeroLimitTurnsItOff(t *testing.T) {
	r := newRouter(&fakeCounter{counts: map[string]int64{}}, 0)
	for i := 0; i < 5; i++ {
		if w := get(r, "10.0.0.1"); w.Code != http.StatusOK {
			t.Fatalf("limit 0 should not block, got %d", w.Code)
		}
	}
}

func TestWindowKeyAndRetryAfter(t *testing.T) {
	at := time.Unix(1_700_000_010, 0)
	k1 := windowKey("public", "1.2.3.4", at, time.Minute)
	k3 := windowKey("public", "1.2.3.4", at.Add(2*time.Minute), time.Minute)
	if k1 == k3 {
		t.Error("a later window needs a new key")
	}
	if windowKey("public", "1.2.3.4", at, time.Minute) == windowKey("admin", "1.2.3.4", at, time.Minute) {
		t.Error("different limits must not share a counter")
	}
	if n := retryAfter(at, time.Minute); n < 1 || n > 60 {
		t.Errorf("retryAfter = %d, want 1 to 60", n)
	}
}
