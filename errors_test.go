package qjson

import (
	"io"
	"testing"
)

func TestErrors(t *testing.T) {

	if exp, out := "syntax error", ErrSyntaxError.Error(); exp != out {
		t.Fatalf("expected %q, got %q", exp, out)
	}
	if exp, out := "nil", errStr(nil); exp != out {
		t.Fatalf("expected %q, got %q", exp, out)
	}
	if exp, out := "error: EOF", errStr(io.EOF); exp != out {
		t.Fatalf("expected %q, got %q", exp, out)
	}
	if exp, out := "Err??? (dummy error))", errStr(Error("dummy error")); exp != out {
		t.Fatalf("expected %q, got %q", exp, out)
	}
	if exp, out := "syntax error pos{b: 10, s: 5, l:1}", (atError{err: ErrSyntaxError, pos: pos{b: 10, s: 5, l: 1}}).Error(); exp != out {
		t.Fatalf("expected %q, got %q", exp, out)
	}
}
