package main

import "testing"

func TestProcessContent(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"  hello\nworld  ", "hello world"}, {"", ""}} {
		t.Run(tc.want, func(t *testing.T) {
			if g := processContent(tc.in); g != tc.want {
				t.Fatalf("got %q, want %q", g, tc.want)
			}
		})
	}
}
