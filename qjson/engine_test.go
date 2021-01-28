package qjson

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestOutputDoubleQuotedString(t *testing.T) {
	tests := []struct {
		in  string
		out string
		err error
	}{
		// 0
		{in: "\"\"", out: "\"\""},
		{in: "\"a é\"", out: "\"a é\""},
		{in: "\"a </ é\"", out: "\"a <\\/ é\""},
		{in: "\"a ' é\"", out: "\"a ' é\""},
		{in: "\"a \t é\"", out: "\"a \\t é\""},
		// 5
		{in: "\"a \\\" é\"", out: "\"a \\\" é\""},
		{in: "\"\\0\"", err: ErrInvalidEscapeSequence},
		{in: "\"\\uz0az\"", err: ErrInvalidEscapeSequence},
		{in: "\"\\u00az\"", err: ErrInvalidEscapeSequence},
		{in: "\"\\u00aB\"", out: "\"\\u00aB\""},
	}
	var e engine
	for i, test := range tests {
		e.init([]byte(""))
		e.tk = token{val: []byte(test.in)}
		e.outputDoubleQuotedString()
		if e.done() && e.tk.val.(error) != test.err {
			t.Fatalf("%d expected err %q, got %q", i, test.err, errStr(e.tk.val.(error)))
		}
		if exp, out := test.out, e.out.String(); !e.done() && exp != out {
			t.Fatalf("%d expected out %q, got %q", i, exp, out)
		}
	}
}

func TestOutputSingleQuotedString(t *testing.T) {
	tests := []struct {
		in  string
		out string
		err error
	}{
		// 0
		{in: "''", out: "\"\""},
		{in: "'a é'", out: "\"a é\""},
		{in: "'a </ é'", out: "\"a <\\/ é\""},
		{in: "'a \t é'", out: "\"a \\t é\""},
		{in: "'a \" é'", out: "\"a \\\" é\""},
		// 5
		{in: "'a \\' é'", out: "\"a ' é\""},
		{in: "'\\0\"", err: ErrInvalidEscapeSequence},
		{in: "'\\uz0az'", err: ErrInvalidEscapeSequence},
		{in: "'\\u00az'", err: ErrInvalidEscapeSequence},
		{in: "'\\u00aB'", out: "\"\\u00aB\""},
	}
	var e engine
	for i, test := range tests {
		e.init([]byte(""))
		e.tk = token{val: []byte(test.in)}
		e.outputSingleQuotedString()
		if e.done() && e.tk.val.(error) != test.err {
			t.Fatalf("%d expected err %q, got %q", i, test.err, e.tk.val.(error))
		}
		if exp, out := test.out, e.out.String(); !e.done() && exp != out {
			t.Fatalf("%d expected %q, got %q", i, exp, out)
		}
	}
}

func TestOutputQuotelessString(t *testing.T) {
	tests := []struct {
		in  string
		out string
		err error
	}{
		// 0
		{in: "", out: "\"\""},
		{in: "a é", out: "\"a é\""},
		{in: "a </ é", out: "\"a <\\/ é\""},
		{in: "a ' é", out: "\"a ' é\""},
		{in: "a \t é", out: "\"a \\t é\""},
		// 5
		{in: "a \" é", out: "\"a \\\" é\""},
		{in: "a \\ é", out: "\"a \\\\ é\""},
		{in: "a \\r é", out: "\"a \\\\r é\""},
	}
	var e engine
	for i, test := range tests {
		e.init([]byte(""))
		e.tk = token{val: []byte(test.in)}
		e.outputQuotelessString()
		if e.done() && e.tk.val.(error) != test.err {
			t.Fatalf("%d expected err %q, got %q", i, test.err, e.tk.val.(error))
		}
		if exp, out := test.out, e.out.String(); !e.done() && exp != out {
			t.Fatalf("%d expected %q, got %q", i, exp, out)
		}
	}
}

func TestOutputMultilineString(t *testing.T) {
	tests := []struct {
		in  string
		out string
		err error
	}{
		// 0
		{in: "`\\n\na`", out: "\"a\""},
		{in: "`\\n\n\b`", out: "\"\\b\""},
		{in: "  `\\n\n  `", out: "\"\""},
		{in: "` \\r\\n \n`", out: "\"\""},
		{in: "`\\r\\n\n\n`", out: "\"\r\n\""},
		// 5
		{in: "`\\r\\n\n\t\r\f\\\"</ /`", out: "\"\\t\\r\\f\\\\\\\"<\\/ /\""},
		{in: "`\\n\n\x17`", out: "\"\\u0017\""},
	}
	var e engine
	for i, test := range tests {
		e.init([]byte(""))
		e.tk = token{val: []byte(test.in)}
		e.outputMultilineString()
		if e.done() && e.tk.val.(error) != test.err {
			t.Fatalf("%d expected err %q, got %q", i, test.err, e.tk.val.(error))
		}
		if exp, out := test.out, e.out.String(); !e.done() && exp != out {
			t.Fatalf("%d expected %q, got %q", i, exp, out)
		}
	}
}

func TestIsLiteralValue(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		// 0
		{in: "true", out: "true"},
		{in: "TRUE", out: "true"},
		{in: "Yes", out: "true"},
		{in: "On", out: "true"},
		{in: "False", out: "false"},
		// 5
		{in: "off", out: "false"},
		{in: "no", out: "false"},
		{in: "NULL", out: "null"},
		{in: "out", out: ""},
	}
	for i, test := range tests {
		if exp, out := test.out, isLiteralValue([]byte(test.in)); exp != out {
			t.Fatalf("%d expected %q, got %q", i, exp, out)
		}
	}
}

func b2s(b []byte) string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("%s", string(b))
}

func e2s(e error) string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s", e)
}

func TestDecode(t *testing.T) {
	tests := []struct {
		in  string
		out string
		err string
	}{
		// 0
		{in: "", out: "{}"},
		{in: "  #...\n", out: "{}"},
		{in: "a:b", out: "{\"a\":\"b\"}"},
		{in: "a:", err: "unexpected end of input at line 1 col 3"},
		{in: "a", err: "unexpected end of input at line 1 col 2"},
		// 5
		{in: "{abcd:b}", err: "expect string identifier at line 1 col 1"},
		{in: "'abcd' : b ", out: "{\"abcd\":\"b\"}"},
		{in: "a:{}", out: "{\"a\":{}}"},
		{in: "a:{b:[]}", out: "{\"a\":{\"b\":[]}}"},
		{in: "a:[],b:c", out: "{\"a\":[],\"b\":\"c\"}"},
		// 10
		{in: "a:[]b:c", out: "{\"a\":[],\"b\":\"c\"}"},
		{in: "a:[]\nb:c", out: "{\"a\":[],\"b\":\"c\"}"},
		{in: ",a:{}", err: "expect string identifier at line 1 col 1"},
		{in: "a:{},", err: "expect identifier after comma at line 1 col 6"},
		{in: "a:b,", err: "expect identifier after comma at line 1 col 5"},
		// 15
		{in: "a:{", err: "unclosed object at line 1 col 3"},
		{in: "a:[", err: "unclosed array at line 1 col 4"},
		{in: "a:}", err: "unexpected } at line 1 col 3"},
		{in: "a:{},}", err: "expect identifier after comma at line 1 col 6"},
		{in: "a:[b,c]", out: "{\"a\":[\"b\",\"c\"]}"},
		// 20
		{in: "a:{b:[c,d]}", out: "{\"a\":{\"b\":[\"c\",\"d\"]}}"},
		{in: "tête\f:{b:[c,d]}", err: "invalid character at line 1 col 5"},
		{in: "\"\\0\":0", err: "invalid escape squence at line 1 col 2"}, // go-fuzz
		{in: "a:true, b:OFF, c:Null", out: "{\"a\":true,\"b\":false,\"c\":null}"},
		{in: "a:10+3.2", out: "{\"a\":13.2}"},
		// 25
		{in: "a,b", err: "expect a colon at line 1 col 2"},
		{in: "a:[a,", err: "expect value after comma at line 1 col 6"},
		{in: "a:[a,b", err: "unclosed array at line 1 col 4"},
		{in: "a:[\"b\"]", out: "{\"a\":[\"b\"]}"},
		{in: "a:['b']", out: "{\"a\":[\"b\"]}"},
		// 30
		{in: "a:\n`\\n\n`", out: "{\"a\":\"\"}"},
		{in: "a:2.3 | 5", err: "operands must be integer at line 1 col 7"},
		{in: "a:{b:{}", err: "unclosed object at line 1 col 3"},
		{in: "a:[]", out: "{\"a\":[]}"},
		{in: "a:[[]]", out: "{\"a\":[[]]}"},
		// 35
		{in: "a:[b]", out: "{\"a\":[\"b\"]}"},
		{in: "a:[b[]]", out: "{\"a\":[\"b\",[]]}"},
		{in: "a:[b[]]", out: "{\"a\":[\"b\",[]]}"},
		{in: "a:{b:[]}", out: "{\"a\":{\"b\":[]}}"},
		{in: "a:{b:]}", err: "unexpected ] at line 1 col 6"},
		// 40
		{in: "a:{]}", err: "unexpected ] at line 1 col 4"},
		{in: "a:,", err: "syntax error at line 1 col 3"},
		{in: "0:\n`\\n#\x04`", err: "invalid character at line 2 col 5"}, // go-fuzz
		{in: "a:[b,}", err: "expect value after comma at line 1 col 6"},
		{in: "a:{b:c,}", err: "expect identifier after comma at line 1 col 8"},
		// 45
		{in: "a:{b:}", err: "unexpected } at line 1 col 6"},
		{in: "a:b}", err: "unexpected } at line 1 col 4"},
		{in: "a:\n`\\n\nthe `\\example`\\\n`", out: "{\"a\":\"the `example`\\n\"}"},
	}

	for i, test := range tests {
		out, err := Decode([]byte(test.in))
		if tout, sout, terr, serr := test.out, b2s(out), test.err, e2s(err); tout != sout || terr != serr {
			t.Fatalf("%d in %q: expected out: %q err: %q, got out: %q err: %q", i, test.in, tout, terr, sout, serr)
		}
		if err == nil {
			var data map[string]interface{}
			err = json.NewDecoder(strings.NewReader(b2s(out))).Decode(&data)
			if err != nil {
				t.Fatalf("%d in: %q out: %q invalid JSON: %s", i, test.in, b2s(out), err)
			}
		}
	}

	if out, err := Decode(nil); b2q(out) != "\"{}\"" || err != nil {
		t.Fatalf("expect out: %s err: %s, got out: %s err: %s", "\"{}\"", "nil", b2q(out), errStr(err))
	}

	maxDepth = 3
	if out, err := Decode([]byte("a:[[[[]]]]")); b2q(out) != "nil" || errStr(err) != "error: too many object or array encapsulations at line 1 col 7" {
		t.Fatalf("expect out: %s err: %s, got out: %s err: %s", "\"{}\"", "nil", b2q(out), errStr(err))
	}
	if out, err := Decode([]byte("a:{b:{c:{d:{}}}}}")); b2q(out) != "nil" || errStr(err) != "error: too many object or array encapsulations at line 1 col 13" {
		t.Fatalf("expect out: %s err: %s, got out: %s err: %s", "\"{}\"", "nil", b2q(out), errStr(err))
	}
	maxDepth = 200

}
