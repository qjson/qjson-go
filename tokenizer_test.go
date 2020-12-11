package qjson

import (
	"fmt"
	"reflect"
	"testing"
)

func atErrorStr(a *atError) string {
	if a == nil {
		return "nil"
	}
	if a.pos.b != 0 || a.pos.s != 0 || a.pos.l != 0 {
		return fmt.Sprintf("&atError{pos: %s, err: %s}", a.pos, errStr(a.err))
	}
	return fmt.Sprintf("&atError{err: %s}", errStr(a.err))
}

func matchErrorFail(exp, got *atError, t *testing.T) bool {
	if exp == nil && got == nil {
		return false
	}
	if exp == nil {
		t.Errorf("expected error nil, got %s", atErrorStr(got))
		return true
	}
	if got == nil {
		t.Errorf("expected error %s, got nil", atErrorStr(exp))
		return true
	}
	if exp.pos != got.pos || exp.err.Error() != got.err.Error() {
		t.Errorf("expected error %s, got %s", atErrorStr(exp), atErrorStr(got))
		return true
	}
	return false
}

func TestBasics(t *testing.T) {
	var tk tokenizer
	tk.init([]byte("test"))
	tk.nextToken()
	if len(tk.p) != 0 {
		t.Fatal("expect same length")
	}
	if exp, str := "{tag: tagQuotelessString, val: \"test\"}", tk.token().String(); exp != str {
		t.Fatalf("expect token %q, got %q", exp, str)
	}
	if exp, str := "pos{b: 4}", tk.pos.String(); exp != str {
		t.Fatalf("expect pos %q, got %q", exp, str)
	}

	tk.init([]byte("#...\n test"))
	tk.nextToken()
	if len(tk.p) != 0 {
		t.Fatal("expect same length")
	}
	if exp, str := "{tag: tagQuotelessString, pos: pos{b: 6, s: 5, l:1}, val: \"test\"}", tk.token().String(); exp != str {
		t.Fatalf("expect token %q, got %q", exp, str)
	}
	if exp, str := "pos{b: 10, s: 5, l:1}", tk.pos.String(); exp != str {
		t.Fatalf("expect pos %q, got %q", exp, str)
	}

	if exp, str := "{tag: 255}", (token{tag: tokenTag(255)}).String(); exp != str {
		t.Fatalf("expect token %q, got %q", exp, str)
	}
	if exp, str := "{tag: tagError, val: ErrSyntaxError}", (token{tag: tagError, val: ErrSyntaxError}).String(); exp != str {
		t.Fatalf("expect token %q, got %q", exp, str)
	}

	if exp, str := "{tag: tagError, val: 12345}", (token{tag: tagError, val: 12345}).String(); exp != str {
		t.Fatalf("expect token %q, got %q", exp, str)
	}

	tk.init([]byte("\u00A0")) // non-breaking space
	tk.nextToken()
	if len(tk.p) != 0 {
		t.Fatal("expect same length")
	}
	if exp, str := "{tag: tagError, pos: pos{b: 2}, val: ErrEndOfInput}", tk.token().String(); exp != str {
		t.Fatalf("expect token %q, got %q", exp, str)
	}
	if exp, str := "pos{b: 2}", tk.pos.String(); exp != str {
		t.Fatalf("expect pos %q, got %q", exp, str)
	}

	tk.init([]byte("//...\r\n"))
	tk.nextToken()
	if len(tk.p) != 0 {
		t.Fatal("expect same length")
	}
	if exp, str := "{tag: tagError, pos: pos{b: 7, s: 7, l:1}, val: ErrEndOfInput}", tk.token().String(); exp != str {
		t.Fatalf("expect token %q, got %q", exp, str)
	}
	if exp, str := "pos{b: 7, s: 7, l:1}", tk.pos.String(); exp != str {
		t.Fatalf("expect pos %q, got %q", exp, str)
	}
	if exp, out := false, tk.popNewline(); exp != out {
		t.Fatalf("expect out %v, got %v", exp, out)
	}

	tk.init([]byte(""))
	tk.nextToken()
	if len(tk.p) != 0 {
		t.Fatal("expect same length")
	}
	if exp, str := "{tag: tagError, val: ErrEndOfInput}", tk.token().String(); exp != str {
		t.Fatalf("expect token %q, got %q", exp, str)
	}
	if exp, str := "pos{b: 0}", tk.pos.String(); exp != str {
		t.Fatalf("expect pos %q, got %q", exp, str)
	}

	if out := b2q(nil); out != "nil" {
		t.Fatalf("expected \"nil\", got %q", out)
	}

	if out := whitespace([]byte{0xC2, 0xA0}); out != 2 {
		t.Fatalf("expected 2, got %d", out)
	}
	exp, out := "{tag: tagIntegerVal, val: 10 type: float32}", token{tag: tagIntegerVal, val: float32(10)}.String()
	if exp != out {
		t.Fatalf("expected %q, got %q", exp, out)
	}

	tk.init([]byte(""))
	tk.tk.tag = tagError
	tk.nextToken()

}

func TestParseChar(t *testing.T) {
	var tk tokenizer
	tests := []struct {
		in  string
		out int
		err *atError
		p   pos
	}{
		// 0
		{in: ""},
		{in: " ", out: 1},
		{in: "\t", out: 1},
		{in: "a ", out: 1},
		{in: "a", out: 1},
		// 5
		{in: "\b", err: &atError{err: ErrInvalidChar}},
		{in: "a \n", out: 1},
		{in: "é ", out: 2},
		{in: "\xA0 ", err: &atError{err: ErrInvalidChar}},
		{in: "\u2000 ", out: 3},
		// 10
		{in: "\xA0", err: &atError{err: ErrInvalidChar}},
		{in: "\xC2", err: &atError{err: ErrTruncatedChar}},
		{in: "\xF1\x80\x10  ", err: &atError{err: ErrInvalidChar}},
		{in: "\xF1\x80\x80  ", err: &atError{err: ErrInvalidChar}},
		{in: "\xF1\x80\x80", err: &atError{err: ErrTruncatedChar}},
		// 15
		{in: "\xF1\x70\x80\x80  ", err: &atError{err: ErrInvalidChar}},
		{in: "\xE0\x80\x80\x80", err: &atError{err: ErrInvalidChar}},
		{in: "\xE1\x80\x80\x20", out: 3},
		{in: "\xED\xA0\x80\x20", err: &atError{err: ErrInvalidChar}},
		{in: "\xF0\x40\x80\x20", err: &atError{err: ErrInvalidChar}},
		// 20
		{in: "\xF0\x90\x80\x80\x20", out: 4},
		{in: "\xF0\x90\xF0\x80\x20", err: &atError{err: ErrInvalidChar}},
		{in: "\xF0\x90\x80\x50\x20", err: &atError{err: ErrInvalidChar}},
		{in: "\xF0\x90\x80", err: &atError{err: ErrTruncatedChar}},
	}
	for i, test := range tests {
		tk.init([]byte(test.in))
		var hasErrors bool
		out, err := tk.char()
		if out != test.out {
			hasErrors = true
			t.Errorf("expected out %d, got %d", test.out, out)
		}
		if matchErrorFail(test.err, err, t) {
			hasErrors = true
		}
		if tk.pos != test.p {
			hasErrors = true
			t.Errorf("expected pos %s, got %s", test.p, tk.pos)
		}
		if hasErrors {
			t.Fatalf("test %d failed", i)
		}
	}
}
func TestSkipMultilineComment(t *testing.T) {
	var tk tokenizer
	tests := []struct {
		in  string
		out bool
		err *atError
		p   pos
	}{
		// 0
		{in: ""},
		{in: " "},
		{in: "/* auieu * /* \n \r\n */ ", out: true, p: pos{b: 21, s: 18, l: 2}},
		{in: "/*", p: pos{b: 2}, err: &atError{err: ErrUnclosedSlashStarComment}},
		{in: "/* ", p: pos{b: 3}, err: &atError{err: ErrUnclosedSlashStarComment}},
		// 5
		{in: "/* \xA0", p: pos{b: 3}, err: &atError{pos: pos{b: 3}, err: ErrInvalidChar}},
		{in: "/* \b */", p: pos{b: 7}, out: true},
	}
	for i, test := range tests {
		tk.init([]byte(test.in))
		var hasErrors bool
		out, err := tk.skipMultilineComment()
		if out != test.out {
			hasErrors = true
			t.Errorf("expected out %v, got %v", test.out, out)
		}
		if matchErrorFail(test.err, err, t) {
			hasErrors = true
		}
		if tk.pos != test.p {
			hasErrors = true
			t.Errorf("expected pos %s, got %s", test.p, tk.pos)
		}
		if hasErrors {
			t.Fatalf("test %d failed", i)
		}
	}
}

func TestDoubleQuoteString(t *testing.T) {
	var tk tokenizer
	tests := []struct {
		in  string
		out []byte
		err *atError
		p   pos
	}{
		// 0
		{in: ""},
		{in: " "},
		{in: "\"...\" ", out: []byte("\"...\""), p: pos{b: 5}},
		{in: "\"...\"", out: []byte("\"...\""), err: nil, p: pos{b: 5}},
		{in: "\".\\\"..\"", out: []byte("\".\\\"..\""), err: nil, p: pos{b: 7}},
		// 5
		{in: "\".\\\"..\" ", out: []byte("\".\\\"..\""), p: pos{b: 7}},
		{in: "\" \xA0 ", p: pos{b: 2}, err: &atError{pos: pos{b: 2}, err: ErrInvalidChar}},
		{in: "\" \\\"", p: pos{b: 4}, err: &atError{err: ErrUnclosedDoubleQuoteString}},
		{in: "\" ", p: pos{b: 2}, err: &atError{err: ErrUnclosedDoubleQuoteString}},
		{in: "\" \r\n ", p: pos{b: 2}, err: &atError{err: ErrNewlineInDoubleQuoteString}},
	}
	for i, test := range tests {
		tk.init([]byte(test.in))
		var hasErrors bool
		out, err := tk.doubleQuotedString()
		if !reflect.DeepEqual(out, test.out) {
			hasErrors = true
			t.Errorf("expected out %v, got %v", b2q(test.out), b2q(out))
		}
		if matchErrorFail(test.err, err, t) {
			hasErrors = true
		}
		if tk.pos != test.p {
			hasErrors = true
			t.Errorf("expected pos %s, got %s", test.p, tk.pos)
		}
		if hasErrors {
			t.Fatalf("test %d failed", i)
		}
	}
}

func TestSingleQuoteString(t *testing.T) {
	var tk tokenizer
	tests := []struct {
		in  string
		out []byte
		err *atError
		p   pos
	}{
		// 0
		{in: " "},
		{in: "'...' ", out: []byte("'...'"), p: pos{b: 5}},
		{in: "'...'", out: []byte("'...'"), err: nil, p: pos{b: 5}},
		{in: "'.\\'..'", out: []byte("'.\\'..'"), err: nil, p: pos{b: 7}},
		{in: "'.\\'..' ", out: []byte("'.\\'..'"), p: pos{b: 7}},
		// 5
		{in: "' \xA0 ", p: pos{b: 2}, err: &atError{pos: pos{b: 2}, err: ErrInvalidChar}},
		{in: "' \\'", p: pos{b: 4}, err: &atError{err: ErrUnclosedSingleQuoteString}},
		{in: "' ", p: pos{b: 2}, err: &atError{err: ErrUnclosedSingleQuoteString}},
		{in: "' \r\n ", p: pos{b: 2}, err: &atError{err: ErrNewlineInSingleQuoteString}},
	}
	for i, test := range tests {
		tk.init([]byte(test.in))
		var hasErrors bool
		out, err := tk.singleQuotedString()
		if !reflect.DeepEqual(out, test.out) {
			hasErrors = true
			t.Errorf("expected out %v, got %v", b2q(test.out), b2q(out))
		}
		if matchErrorFail(test.err, err, t) {
			hasErrors = true
		}
		if tk.pos != test.p {
			hasErrors = true
			t.Errorf("expected pos %s, got %s", test.p, tk.pos)
		}
		if hasErrors {
			t.Fatalf("test %d failed", i)
		}
	}
}

func TestQuotelessString(t *testing.T) {
	var tk tokenizer
	tests := []struct {
		in  string
		out []byte
		err *atError
		p   pos
	}{
		// 0
		{in: ""},
		{in: "test 1", out: []byte("test 1"), err: nil, p: pos{b: 6}},
		{in: "'...' ", out: []byte("'...'"), err: nil, p: pos{b: 6}},
		{in: "'...'", out: []byte("'...'"), err: nil, p: pos{b: 5}},
		{in: "test 1,", out: []byte("test 1"), p: pos{b: 6}},
		// 5
		{in: "a b {", out: []byte("a b"), p: pos{b: 4}},
		{in: "z \xA0 ", out: nil, p: pos{b: 2}, err: &atError{pos: pos{b: 2}, err: ErrInvalidChar}},
		{in: "a b   \r\n", out: []byte("a b"), p: pos{b: 6}},
		{in: "a b   \r\n ", out: []byte("a b"), p: pos{b: 6}},
		{in: "a b  /* ", out: []byte("a b"), p: pos{b: 5}},
	}
	for i, test := range tests {
		tk.init([]byte(test.in))
		var hasErrors bool
		out, err := tk.quotelessString()
		if !reflect.DeepEqual(out, test.out) {
			hasErrors = true
			t.Errorf("expected out %v, got %v", b2q(test.out), b2q(out))
		}
		if matchErrorFail(test.err, err, t) {
			hasErrors = true
		}
		if tk.pos != test.p {
			hasErrors = true
			t.Errorf("expected pos %s, got %s", test.p, tk.pos)
		}
		if hasErrors {
			t.Fatalf("test %d failed", i)
		}
	}
}

func TestMultilineString(t *testing.T) {
	var tk tokenizer
	tests := []struct {
		in  string
		p0  pos
		out []byte
		err *atError
		p   pos
	}{
		// 0
		{in: "`\\n\n`", out: []byte("`\\n\n`"), err: nil, p: pos{b: 5, s: 4, l: 1}},
		{in: "`\\n\na\n`\n\n", out: []byte("`\\n\na\n`"), p: pos{b: 7, s: 6, l: 2}},
		{in: "  `\\n#...\n  a\n  `\n\n", p0: pos{b: 2}, out: []byte("  `\\n#...\n  a\n  `"), p: pos{b: 17, s: 14, l: 2}},
		{in: "  `\\n#...\n  a\n  `\n\n", p0: pos{b: 2}, out: []byte("  `\\n#...\n  a\n  `"), p: pos{b: 17, s: 14, l: 2}},
		{in: "\n  `\\n//..\n  a\n  `\n\n", p0: pos{b: 3, s: 1, l: 1}, out: []byte("  `\\n//..\n  a\n  `"), p: pos{b: 18, s: 15, l: 3}},
		// 5
		{in: " \t `\\r\\n\n  \n \t `\n\n", p0: pos{b: 3}, err: &atError{pos: pos{b: 10, s: 9, l: 1}, err: ErrInvalidMarginChar}, p: pos{b: 9, s: 9, l: 1}},
		{in: " \t `\\r\\n\n `", p0: pos{b: 3}, err: &atError{pos: pos{b: 10, s: 9, l: 1}, err: ErrInvalidMarginChar}, p: pos{b: 9, s: 9, l: 1}},
		{in: " `\n `", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrInvalidNewlineSpecifier}, p: pos{b: 2}},
		{in: " a`\n `", p0: pos{b: 2}, err: &atError{pos: pos{b: 1}, err: ErrMarginMustBeWhitespaceOnly}, p: pos{b: 2}},
		{in: " `", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrMissingNewlineSpecifier}, p: pos{b: 2}},
		// 10
		{in: " `  ", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrMissingNewlineSpecifier}, p: pos{b: 4}},
		{in: " `\\n", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrInvalidMultilineStart}, p: pos{b: 4}},
		{in: " `\\n  ", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrInvalidMultilineStart}, p: pos{b: 6}},
		{in: " `\\n  a", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrInvalidMultilineStart}, p: pos{b: 6}},
		{in: " `\\n\n ", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrUnclosedMultiline}, p: pos{b: 6, s: 5, l: 1}},
		// 15
		{in: " `\\n\n \n", p0: pos{b: 1}, err: &atError{pos: pos{b: 7, s: 7, l: 2}, err: ErrInvalidMarginChar}, p: pos{b: 7, s: 7, l: 2}},
		{in: " `\\n\n \na", p0: pos{b: 1}, err: &atError{pos: pos{b: 7, s: 7, l: 2}, err: ErrInvalidMarginChar}, p: pos{b: 7, s: 7, l: 2}},
		{in: " `\\n\n \n ", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrUnclosedMultiline}, p: pos{b: 8, s: 7, l: 2}},
		{in: " `\\n\n \n \b", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrUnclosedMultiline}, p: pos{b: 9, s: 7, l: 2}},
		{in: " `\\n\n \n `\\", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrUnclosedMultiline}, p: pos{b: 10, s: 7, l: 2}},
		// 20
		{in: " `\\n\n \n `", p0: pos{b: 1}, out: []byte(" `\\n\n \n `"), p: pos{b: 9, s: 7, l: 2}},
		{in: " `\\n\n \n \xA0`", p0: pos{b: 1}, err: &atError{pos: pos{b: 8, s: 7, l: 2}, err: ErrInvalidChar}, p: pos{b: 8, s: 7, l: 2}},
		{in: " `\\n#\xA0  ", p0: pos{b: 1}, err: &atError{pos: pos{b: 5}, err: ErrInvalidChar}, p: pos{b: 5}},
		{in: " `\\n\n", p0: pos{b: 1}, err: &atError{pos: pos{b: 1}, err: ErrUnclosedMultiline}, p: pos{b: 5, s: 5, l: 1}},
	}
	for i, test := range tests {
		tk.init([]byte(test.in))
		tk.pos = test.p0
		tk.p = tk.p[test.p0.b:]
		var hasErrors bool
		out, err := tk.multilineString()
		if !reflect.DeepEqual(out, test.out) {
			hasErrors = true
			t.Errorf("expected out %v, got %v", b2q(test.out), b2q(out))
		}
		if matchErrorFail(test.err, err, t) {
			hasErrors = true
		}
		if tk.pos != test.p {
			hasErrors = true
			t.Errorf("expected pos %s, got %s", test.p, tk.pos)
		}
		if hasErrors {
			t.Fatalf("test %d failed", i)
		}
	}
}

func TestTokenizer(t *testing.T) {
	tests := []struct {
		in  string
		out []token
	}{
		// 0
		{in: ""},
		{in: "  "},
		{in: "\xA0", out: []token{{tag: tagError, val: ErrInvalidChar}}},
		{in: "\xC2", out: []token{{tag: tagError, val: ErrTruncatedChar}}},
		{in: "\xF1\x80\x10  ", out: []token{{tag: tagError, val: ErrInvalidChar}}},
		// 5
		{in: "\xF1\x80\x80  ", out: []token{{tag: tagError, val: ErrInvalidChar}}},
		{in: "\xF1\x80\x80", out: []token{{tag: tagError, val: ErrTruncatedChar}}},
		{in: " #..."},
		{in: " //..."},
		{in: " "},
		// 10
		{in: " /*...", out: []token{{tag: tagError, pos: pos{b: 1}, val: ErrUnclosedSlashStarComment}}},
		{in: "#\xA0", out: []token{{tag: tagError, pos: pos{b: 1}, val: ErrInvalidChar}}},
		{in: "#\xC2", out: []token{{tag: tagError, pos: pos{b: 1}, val: ErrTruncatedChar}}},
		{in: "#\xF1\x80\x10  ", out: []token{{tag: tagError, pos: pos{b: 1}, val: ErrInvalidChar}}},
		{in: "#\xF1\x80\x80  ", out: []token{{tag: tagError, pos: pos{b: 1}, val: ErrInvalidChar}}},
		// 15
		{in: "#\xF1\x80\x80", out: []token{{tag: tagError, pos: pos{b: 1}, val: ErrTruncatedChar}}},
		{in: "a b  , c\nd e  ", out: []token{
			{tag: tagQuotelessString, val: []byte("a b")},
			{tag: tagComma, pos: pos{b: 5}},
			{tag: tagQuotelessString, pos: pos{b: 7}, val: []byte("c")},
			{tag: tagQuotelessString, pos: pos{b: 9, s: 9, l: 1}, val: []byte("d e")}}},
		{in: "a : \"...\"", out: []token{
			{tag: tagQuotelessString, val: []byte("a")},
			{tag: tagColon, pos: pos{b: 2}},
			{tag: tagDoubleQuotedString, pos: pos{b: 4}, val: []byte("\"...\"")}}},
		{in: "a : { 'abc': d}", out: []token{
			{tag: tagQuotelessString, val: []byte("a")},
			{tag: tagColon, pos: pos{b: 2}},
			{tag: tagOpenBrace, pos: pos{b: 4}},
			{tag: tagSingleQuotedString, pos: pos{b: 6}, val: []byte("'abc'")},
			{tag: tagColon, pos: pos{b: 11}},
			{tag: tagQuotelessString, pos: pos{b: 13}, val: []byte("d")},
			{tag: tagCloseBrace, pos: pos{b: 14}}}},
		{in: "`\\n\na` ", out: []token{{tag: tagMultilineString, val: []byte("`\\n\na`")}}},
		// 20
		{in: "/*...*/a", out: []token{{tag: tagQuotelessString, pos: pos{b: 7}, val: []byte("a")}}},
		{in: "\"...", out: []token{{tag: tagError, val: ErrUnclosedDoubleQuoteString}}},
		{in: "'...", out: []token{{tag: tagError, val: ErrUnclosedSingleQuoteString}}},
		{in: "`...", out: []token{{tag: tagError, val: ErrInvalidNewlineSpecifier}}},
	}

	for i, test := range tests {
		var tk tokenizer
		tk.init([]byte(test.in))
		var out []token
		for tk.token().tag != tagError {
			tk.nextToken()
			out = append(out, tk.token())
		}
		if len(out) > 0 && out[len(out)-1].tag == tagError && out[len(out)-1].val == ErrEndOfInput {
			out = out[:len(out)-1]
			if len(out) == 0 {
				out = nil
			}
		}
		var hasErrors bool
		if !reflect.DeepEqual(out, test.out) {
			hasErrors = true
			n := len(test.out)
			if len(out) < n {
				n = len(out)
			}
			for j := 0; j < n; j++ {
				if !reflect.DeepEqual(out[j], test.out[j]) {
					t.Errorf("token %d: expect %v, got %v", j, test.out[j], out[j])
				}
			}
			if len(test.out) > n {
				for j := n; j < len(test.out); j++ {
					t.Errorf("token %d: expect %v, got nothing", j, test.out[j])
				}
			} else if len(out) > n {
				for j := n; j < len(out); j++ {
					t.Errorf("token %d: expect nothing, got %v", j, out[j])
				}
			}
		}
		if hasErrors {
			t.Fatalf("test %d failed", i)
		}
	}
}

func TestColumn(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: ""},
		{in: "é", out: 1},
		{in: "é\xC2", out: 1},
	}

	for i, test := range tests {
		if exp, out := test.out, column([]byte(test.in)); exp != out {
			t.Fatalf("%d expected %d, got %d", i, exp, out)
		}
	}
}
