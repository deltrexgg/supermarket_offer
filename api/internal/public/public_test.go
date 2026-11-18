package public

import (
	"net/url"
	"testing"
)

func TestQueryHashIgnoresParameterOrder(t *testing.T) {
	a := url.Values{"area": {"2"}, "sort": {"lowest_price"}, "page": {"1"}}
	b := url.Values{"page": {"1"}, "sort": {"lowest_price"}, "area": {"2"}}
	if queryHash(a) != queryHash(b) {
		t.Fatal("the same query in a different order must share a cache key")
	}
}

func TestQueryHashSeparatesDifferentQueries(t *testing.T) {
	a := url.Values{"area": {"2"}}
	b := url.Values{"area": {"3"}}
	c := url.Values{"category": {"2"}}
	if queryHash(a) == queryHash(b) || queryHash(a) == queryHash(c) {
		t.Fatal("different queries must not share a cache key")
	}
}

func TestPositiveInt(t *testing.T) {
	for _, s := range []string{"1", "42", "999999"} {
		if n, ok := positiveInt(s); !ok || n == 0 {
			t.Errorf("positiveInt(%q) = %d, %v; want a positive number", s, n, ok)
		}
	}
	for _, s := range []string{"", "0", "-3", "abc", "1.5", "7x"} {
		if _, ok := positiveInt(s); ok {
			t.Errorf("positiveInt(%q) accepted a bad value", s)
		}
	}
}
