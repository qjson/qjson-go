package qjson

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestIsNumberExpr(t *testing.T) {
	tests := []struct {
		in  string
		out bool
	}{
		{in: ""},
		{in: "a"},
		{in: " "},
		{in: "-( (0", out: true},
	}
	for i, test := range tests {
		if exp, out := test.out, isNumberExpr([]byte(test.in)); exp != out {
			t.Fatalf("%d expected %v, got %v", i, exp, out)
		}
	}
}

func checkToken(exp, out numToken, t *testing.T) bool {
	if reflect.DeepEqual(exp, out) {
		return false
	}
	t.Errorf("expect numToken %v, got %v", exp, out)
	return true
}

func TestNumberBasics(t *testing.T) {
	if res := inRange(5, 10, 20); res {
		t.Fatal("expect false, got true")
	}
	if res := inRange(30, 10, 20); res {
		t.Fatal("expect false, got true")
	}
	if res := inRange(15, 10, 20); !res {
		t.Fatal("expect true, got false")
	}
	if res := inRange(10, 10, 20); !res {
		t.Fatal("expect true, got false")
	}
	if res := inRange(20, 10, 20); !res {
		t.Fatal("expect true, got false")
	}
	if res := isBinDigit('0'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isBinDigit('1'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isBinDigit(' '); res {
		t.Fatal("expect false, got true")
	}
	if res := isIntDigit('0'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isIntDigit('9'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isIntDigit(' '); res {
		t.Fatal("expect false, got true")
	}
	if res := isHexDigit('0'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isHexDigit('9'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isHexDigit('a'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isHexDigit('F'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isHexDigit(' '); res {
		t.Fatal("expect false, got true")
	}
	if res := isOctDigit('0'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isOctDigit('7'); !res {
		t.Fatal("expect true, got false")
	}
	if res := isOctDigit('8'); res {
		t.Fatal("expect false, got true")
	}
	if res := isOctDigit(' '); res {
		t.Fatal("expect false, got true")
	}
	exp, out := "numToken{tag: tagError, pos: 5, val: ErrSyntaxError}", numToken{tag: tagError, pos: 5, val: ErrSyntaxError}.String()
	if exp != out {
		t.Fatalf("expect %q, got %q", exp, out)
	}
	exp, out = "numToken{tag: tagError, val: 5}", numToken{tag: tagError, val: 5}.String()
	if exp != out {
		t.Fatalf("expect %q, got %q", exp, out)
	}
	exp, out = "numToken{tag: tagError, val: true of type bool}", numToken{tag: tagError, val: true}.String()
	if exp != out {
		t.Fatalf("expect %q, got %q", exp, out)
	}
	exp, out = "numToken{tag: 255, val: 5.2}", numToken{tag: 255, val: 5.2}.String()
	if exp != out {
		t.Fatalf("expect %q, got %q", exp, out)
	}
	exp, out = "numToken{tag: tagPlus}", numToken{tag: tagPlus}.String()
	if exp != out {
		t.Fatalf("expect %q, got %q", exp, out)
	}

	var tk numTokenizer
	tk.init(make([]byte, 3))
	tk.popBytes(10)
}

func TestTokenizerBasics(t *testing.T) {
	var tk numTokenizer
	tk.init([]byte("test"))

	tk.popBytes(1)
	if tk.pos != 1 || len(tk.p)+tk.pos != len(tk.in) {
		t.Fatalf("expect pos=%d len(tk.p)=%d len(tk.in)=%d, pos=%d len(tk.p)=%d len(tk.in)=%d", 1, 3, 4, tk.pos, len(tk.p), len(tk.in))
	}

	tk.init([]byte(" +"))
	if tk.nextOperator() {
		t.Fatal("expect  false, got true")
	}
	tk.popBytes(1)
	if !tk.nextOperator() {
		t.Fatal("expect true, got false")
	}
}

func TestParseBinDigits(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: " ", out: 0},
		{in: "0", out: 1},
		{in: "1", out: 1},
		{in: "_0", out: 0},
		{in: "0_", out: -1},
		{in: "0_ ", out: -1},
		{in: "0_10_1", out: 6},
		{in: "0_1__0", out: -1},
		{in: "0_1a", out: 3},
	}
	for i, test := range tests {
		if out := parseBinDigits([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestParseOctDigits(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: " ", out: 0},
		{in: "0", out: 1},
		{in: "7", out: 1},
		{in: "_0", out: 0},
		{in: "0_", out: -1},
		{in: "0_8", out: -1},
		{in: "0_10_1", out: 6},
		{in: "0_1__0", out: -1},
		{in: "0_18", out: 3},
	}
	for i, test := range tests {
		if out := parseOctDigits([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestParseHexDigits(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: " ", out: 0},
		{in: "0", out: 1},
		{in: "7", out: 1},
		{in: "_a", out: 0},
		{in: "F_", out: -1},
		{in: "0_G", out: -1},
		{in: "0_9a_d", out: 6},
		{in: "0_8__0", out: -1},
		{in: "2_1 ", out: 3},
	}
	for i, test := range tests {
		if out := parseHexDigits([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestParseDecDigits(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: " ", out: 0},
		{in: "0", out: 1},
		{in: "7", out: 1},
		{in: "_9", out: 0},
		{in: "3_", out: -1},
		{in: "0_z", out: -1},
		{in: "0_91_7", out: 6},
		{in: "0_8__0", out: -1},
		{in: "2_1 ", out: 3},
	}
	for i, test := range tests {
		if out := parseIntDigits([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestParseBinLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: " ", out: 0},
		{in: "0b0", out: 3},
		{in: "0B1", out: 3},
		{in: "0b_0", out: 4},
		{in: "0b0_", out: -1},
		// 5
		{in: "0B0_ ", out: -1},
		{in: "0b0_10_1", out: 8},
		{in: "0B0_1__0", out: -1},
		{in: "0b0_1a", out: 5},
		{in: "0b   ", out: -1},
		// 10
		{in: "0b_", out: -1},
		{in: "0b", out: -1},
		{in: "0b_ ", out: -1},
		{in: "0b ", out: -1},
	}
	for i, test := range tests {
		if out := parseBinLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestDecodeBinLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: "0b0", out: 0},
		{in: "0B1", out: 1},
		{in: "0b_0", out: 0},
		{in: "0b0", out: 0},
		{in: "0B0", out: 0},
		// 5
		{in: "0b0_10_1", out: 5},
		{in: "0B0_1", out: 1},
		{in: "0b0_1", out: 1},
		{in: "0b11111111_11111111_11111111_11111111_11111111_11111111_11111111_11111111", out: -1},
		{in: "0b0000000000000001111111_11111111_11111111_11111111_11111111_11111111_11111111_11111111", out: 0x7FFFFFFF_FFFFFFFF},
	}
	for i, test := range tests {
		if out := decodeBinLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestNextBinValue(t *testing.T) {
	tests := []struct {
		in  string
		out bool
		tk  numToken
		pos int
	}{
		// 0
		{in: " "},
		{in: "0b0", out: true, pos: 3, tk: numToken{tag: tagIntegerVal, val: 0}},
		{in: "0B1", out: true, pos: 3, tk: numToken{tag: tagIntegerVal, val: 1}},
		{in: "0b_0", out: true, pos: 4, tk: numToken{tag: tagIntegerVal, val: 0}},
		{in: "0b11111111_11111111_11111111_11111111_11111111_11111111_11111111_11111111", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrNumberOverflow}},
		// 5
		{in: "0b0000000000000001111111_11111111_11111111_11111111_11111111_11111111_11111111_11111111", out: true, pos: 87, tk: numToken{tag: tagIntegerVal, val: 9223372036854775807}},
		{in: "0b11111111_", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrInvalidBinaryNumber}},
		{in: "0b111__11111_", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrInvalidBinaryNumber}},
		{in: "0b1_11111111_11111111_11111111_11111111_11111111_11111111_11111111_11111111", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrNumberOverflow}},
		{in: "0b", out: true, tk: numToken{tag: tagError, val: ErrInvalidBinaryNumber}},
		// 10
		{in: "0b_", out: true, tk: numToken{tag: tagError, val: ErrInvalidBinaryNumber}},
		{in: "0b ", out: true, tk: numToken{tag: tagError, val: ErrInvalidBinaryNumber}},
		{in: "0b_ ", out: true, tk: numToken{tag: tagError, val: ErrInvalidBinaryNumber}},
	}
	for i, test := range tests {
		var hasErrors bool
		var tk numTokenizer
		tk.init([]byte(test.in))
		if exp, out := test.out, tk.nextBinValue(); out != exp {
			hasErrors = true
			t.Errorf("expect %v, got %v", exp, out)
		}
		b := checkToken(test.tk, tk.tk, t)
		hasErrors = hasErrors || b
		if exp, out := test.pos, tk.pos; exp != out {
			hasErrors = true
			t.Errorf("expect pos %d, got %d", exp, out)
		}
		if hasErrors {
			pc := make([]uintptr, 15)
			n := runtime.Callers(1, pc)
			frames := runtime.CallersFrames(pc[:n])
			frame, _ := frames.Next()
			// fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
			t.Fatalf("Errors with test %d in %s (%s:%d)", i, t.Name(), filepath.Base(frame.File), frame.Line)
		}
	}
}

func TestParseHexLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: " ", out: 0},
		{in: "0x0", out: 3},
		{in: "0X1", out: 3},
		{in: "0x_0", out: 4},
		{in: "0x0_", out: -1},
		// 5
		{in: "0X0_ ", out: -1},
		{in: "0x0_10_1", out: 8},
		{in: "0X0_1__0", out: -1},
		{in: "0x0_1z", out: 5},
		{in: "0x   ", out: -1},
		// 10
		{in: "0    ", out: 0},
		{in: "0x", out: -1},
		{in: "0x_", out: -1},
		{in: "0x ", out: -1},
		{in: "0x_ ", out: -1},
	}
	for i, test := range tests {
		if out := parseHexLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestDecodeHexLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: "0x0", out: 0},
		{in: "0X1", out: 1},
		{in: "0x_0", out: 0},
		{in: "0x_0_A_B", out: 0xAB},
		{in: "0xFFFFFFFF_FFFFFFFF", out: -1},
		// 5
		{in: "0xF_FFFFFFFF_FFFFFFFF", out: -1},
		{in: "0x7FFFFFFF_FFFFFFFF", out: 0x7FFFFFFF_FFFFFFFF},
		{in: "0x00000000_7FFFFFFF_FFFFFFFF", out: 0x7FFFFFFF_FFFFFFFF},
	}
	for i, test := range tests {
		if out := decodeHexLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestNextHexValue(t *testing.T) {
	tests := []struct {
		in  string
		out bool
		tk  numToken
		pos int
	}{
		{in: " "},
		{in: "0x0", out: true, pos: 3, tk: numToken{tag: tagIntegerVal, val: 0}},
		{in: "0X1", out: true, pos: 3, tk: numToken{tag: tagIntegerVal, val: 1}},
		{in: "0x_0", out: true, pos: 4, tk: numToken{tag: tagIntegerVal, val: 0}},
		{in: "0x_0_A_B", out: true, pos: 8, tk: numToken{tag: tagIntegerVal, val: 0xAB}},
		// 5
		{in: "0xFFFFFFFF_FFFFFFFF", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrNumberOverflow}},
		{in: "0x7FFFFFFF_FFFFFFFF", out: true, pos: 19, tk: numToken{tag: tagIntegerVal, val: 0x7FFFFFFF_FFFFFFFF}},
		{in: "0x00000000_7FFFFFFF_FFFFFFFF", out: true, pos: 28, tk: numToken{tag: tagIntegerVal, val: 0x7FFFFFFF_FFFFFFFF}},
		{in: "0x0af_", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrInvalidHexadecimalNumber}},
		{in: "0x", out: true, tk: numToken{tag: tagError, val: ErrInvalidHexadecimalNumber}},
		// 10
		{in: "0x_", out: true, tk: numToken{tag: tagError, val: ErrInvalidHexadecimalNumber}},
		{in: "0x ", out: true, tk: numToken{tag: tagError, val: ErrInvalidHexadecimalNumber}},
		{in: "0x_ ", out: true, tk: numToken{tag: tagError, val: ErrInvalidHexadecimalNumber}},
	}
	for i, test := range tests {
		var hasErrors bool
		var tk numTokenizer
		tk.init([]byte(test.in))
		if exp, out := test.out, tk.nextHexValue(); out != exp {
			hasErrors = true
			t.Errorf("expect %v, got %v", exp, out)
		}
		b := checkToken(test.tk, tk.tk, t)
		hasErrors = hasErrors || b
		if exp, out := test.pos, tk.pos; exp != out {
			hasErrors = true
			t.Errorf("expect pos %d, got %d", exp, out)
		}
		if hasErrors {
			pc := make([]uintptr, 15)
			n := runtime.Callers(1, pc)
			frames := runtime.CallersFrames(pc[:n])
			frame, _ := frames.Next()
			// fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
			t.Fatalf("Errors with test %d in %s (%s:%d)", i, t.Name(), filepath.Base(frame.File), frame.Line)
		}
	}
}

func TestParseOctLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: " ", out: 0},
		{in: "0o0", out: 3},
		{in: "0O1", out: 3},
		{in: "0O_0", out: 4},
		{in: "0o0_", out: -1},
		// 5
		{in: "0O0_ ", out: -1},
		{in: "0o0_10_1", out: 8},
		{in: "0O0_1__0", out: -1},
		{in: "0o0_1z", out: 5},
		{in: "0o   ", out: -1},
		// 10
		{in: "0o_   ", out: -1},
		{in: "0o", out: -1},
		{in: "0o_", out: -1},
		{in: "0    ", out: 0},
		{in: "00   ", out: 2},
		// 15
		{in: "0750 ", out: 4},
		{in: "0_750 ", out: 5},
		{in: "0_75_0 ", out: 6},
		{in: "0_75__0 ", out: -1},
		{in: "0_750_ ", out: -1},
		// 20
		{in: "0_750_ ", out: -1},
	}
	for i, test := range tests {
		if out := parseOctLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestDecodeOctLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: "0o0", out: 0},
		{in: "0O1", out: 1},
		{in: "0O_0", out: 0},
		{in: "0o0_10_1", out: 65},
		{in: "0o0_1", out: 1},
		// 5
		{in: "00", out: 0},
		{in: "0750", out: 488},
		{in: "0_750", out: 488},
		{in: "01777777777777777777777", out: -1},
		{in: "0777777777777777777777", out: 9223372036854775807},
	}
	for i, test := range tests {
		if out := decodeOctLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestNextOctValue(t *testing.T) {
	tests := []struct {
		in  []byte
		out bool
		tk  numToken
		pos int
	}{
		{in: []byte(" "), out: false},
		{in: []byte("0o0"), out: true, pos: 3, tk: numToken{tag: tagIntegerVal, val: 0}},
		{in: []byte("0O1"), out: true, pos: 3, tk: numToken{tag: tagIntegerVal, val: 1}},
		{in: []byte("0o_0"), out: true, pos: 4, tk: numToken{tag: tagIntegerVal, val: 0}},
		{in: []byte("0_0_2_3"), out: true, pos: 7, tk: numToken{tag: tagIntegerVal, val: 19}},
		// 5
		{in: []byte("01777777777777777777777"), out: true, pos: 0, tk: numToken{tag: tagError, val: ErrNumberOverflow}},
		{in: []byte("0777777777777777777777"), out: true, pos: 22, tk: numToken{tag: tagIntegerVal, val: 0x7FFFFFFF_FFFFFFFF}},
		{in: []byte("00000000000777777777777777777777"), out: true, pos: 32, tk: numToken{tag: tagIntegerVal, val: 0x7FFFFFFF_FFFFFFFF}},
		{in: []byte("0o750_"), out: true, tk: numToken{tag: tagError, val: ErrInvalidOctalNumber}},
		{in: []byte("0"), out: false},
		// 10
		{in: []byte("0 "), out: false},
		{in: []byte("0_"), out: true, tk: numToken{tag: tagError, val: ErrInvalidOctalNumber}},
		{in: []byte("0_ "), out: true, tk: numToken{tag: tagError, val: ErrInvalidOctalNumber}},
		{in: []byte("0o"), out: true, tk: numToken{tag: tagError, val: ErrInvalidOctalNumber}},
		{in: []byte("0o "), out: true, tk: numToken{tag: tagError, val: ErrInvalidOctalNumber}},
		// 15
		{in: []byte("0o_"), out: true, tk: numToken{tag: tagError, val: ErrInvalidOctalNumber}},
		{in: []byte("0o_ "), out: true, tk: numToken{tag: tagError, val: ErrInvalidOctalNumber}},
	}
	for i, test := range tests {
		var hasErrors bool
		var tk numTokenizer
		tk.init(test.in)
		if exp, out := test.out, tk.nextOctValue(); out != exp {
			hasErrors = true
			t.Errorf("expect %v, got %v", exp, out)
		}
		b := checkToken(test.tk, tk.tk, t)
		hasErrors = hasErrors || b
		if exp, out := test.pos, tk.pos; exp != out {
			hasErrors = true
			t.Errorf("expect pos %d, got %d", exp, out)
		}
		if hasErrors {
			pc := make([]uintptr, 15)
			n := runtime.Callers(1, pc)
			frames := runtime.CallersFrames(pc[:n])
			frame, _ := frames.Next()
			// fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
			t.Fatalf("Errors with test %d in %s (%s:%d)", i, t.Name(), filepath.Base(frame.File), frame.Line)
		}
	}
}
func TestParseIntLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: " ", out: 0},
		{in: "0x0", out: 1},
		{in: "123", out: 3},
		{in: "12_0", out: 4},
		{in: "456_", out: -1},
		{in: "678_ ", out: -1},
		{in: "1_0_2_40", out: 8},
		{in: "12345__0", out: -1},
		{in: "54321z", out: 5},
	}
	for i, test := range tests {
		if out := parseIntLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestDecodeIntLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: "0", out: 0},
		{in: "10", out: 10},
		{in: "1_0_0", out: 100},
		{in: "18446744073709551615", out: -1}, // 0xFFFFFFFF_FFFFFFFF
		{in: "9223372036854775807", out: 0x7FFFFFFFFFFFFFFF},
	}
	for i, test := range tests {
		if out := decodeIntLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestNextIntValue(t *testing.T) {
	tests := []struct {
		in  string
		out bool
		tk  numToken
		pos int
	}{
		{in: " ", out: false, pos: 0, tk: numToken{tag: tagUnknown}},
		{in: "0", out: true, pos: 1, tk: numToken{tag: tagIntegerVal, val: 0}},
		{in: "10", out: true, pos: 2, tk: numToken{tag: tagIntegerVal, val: 10}},
		{in: "1_0_0", out: true, pos: 5, tk: numToken{tag: tagIntegerVal, val: 100}},
		{in: "18446744073709551615", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrNumberOverflow}},
		// 5
		{in: "9223372036854775807", out: true, pos: 19, tk: numToken{tag: tagIntegerVal, val: 0x7FFFFFFF_FFFFFFFF}},
		{in: "750_", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrInvalidIntegerNumber}},
		{in: "00", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrInvalidIntegerNumber}},
		{in: "184467440737095516150", out: true, pos: 0, tk: numToken{tag: tagError, val: ErrNumberOverflow}},
	}
	for i, test := range tests {
		var hasErrors bool
		var tk numTokenizer
		tk.init([]byte(test.in))
		if exp, out := test.out, tk.nextIntValue(); out != exp {
			hasErrors = true
			t.Errorf("expect %v, got %v", exp, out)
		}
		b := checkToken(test.tk, tk.tk, t)
		hasErrors = hasErrors || b
		if exp, out := test.pos, tk.pos; exp != out {
			hasErrors = true
			t.Errorf("expect pos %d, got %d", exp, out)
		}
		if hasErrors {
			pc := make([]uintptr, 15)
			n := runtime.Callers(1, pc)
			frames := runtime.CallersFrames(pc[:n])
			frame, _ := frames.Next()
			// fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
			t.Fatalf("Errors with test %d in %s (%s:%d)", i, t.Name(), filepath.Base(frame.File), frame.Line)
		}
	}
}
func TestParseExponent(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{in: "", out: 0},
		{in: " ", out: 0},
		{in: "a+123", out: 0},
		{in: "123", out: 0},
		{in: "e123", out: 4},
		// 5
		{in: "e+12", out: 4},
		{in: "e-12 ", out: 4},
		{in: "E121_0", out: 6},
		{in: "e123__0", out: -1},
		{in: "e4321z", out: 5},
		// 10
		{in: "e+", out: -1},
		{in: "e+0", out: 3},
		{in: "E", out: -1},
		{in: "e", out: -1},
		{in: "E+", out: -1},
		// 15
		{in: "e-", out: -1},
		{in: "E ", out: -1},
		{in: "e ", out: -1},
		{in: "E+ ", out: -1},
		{in: "e- ", out: -1},
	}
	for i, test := range tests {
		if out := parseExponent([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestParseDecLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		// 0
		{in: " ", out: 0},
		{in: "a+123", out: 0},
		{in: "123", out: 0},
		{in: "e123", out: 0},
		{in: "0", out: 0},
		// 5
		{in: "0 ", out: 0},
		{in: "1e-12 ", out: 5},
		{in: "1e+ ", out: -1},
		{in: "1_2_e+ ", out: 0},
		{in: "1_2_4e-1 ", out: 8},
		// 10
		{in: ".0", out: 2},
		{in: ".0 ", out: 2},
		{in: ".123e+12", out: 8},
		{in: ".12_3e-12 ", out: 9},
		{in: ".123_E121_0", out: -1},
		// 15
		{in: ".e123__0", out: -1},
		{in: ".9e4321z", out: 7},
		{in: ".0e+", out: -1},
		{in: ". e+0", out: 0},
		{in: "0.", out: 2},
		// 20
		{in: "0. ", out: 2},
		{in: "12.0000e+12", out: 11},
		{in: "12.0000e-12 ", out: 11},
		{in: "1.E121_0", out: 8},
		{in: "5.e123__0", out: -1},
		// 25
		{in: "5.12_33", out: 7},
		{in: "5.12_33 ", out: 7},
		{in: "5.12_33__ ", out: -1},
		{in: ".e123", out: -1},
		{in: "23e", out: -1},
		// 30
		{in: "23.5e", out: -1},
		{in: "23.e12", out: 6},
	}
	for i, test := range tests {
		if out := parseDecLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %d, got %d", i, test.out, out)
		}
	}
}

func TestDecodeDecLiteral(t *testing.T) {
	tests := []struct {
		in  string
		out float64
	}{
		{in: "0.", out: 0},
		{in: ".0", out: 0},
		{in: "10.05", out: 10.05},
		{in: "45e1345", out: -1},
		{in: ".3", out: .3},
	}
	for i, test := range tests {
		if out := decodeDecLiteral([]byte(test.in)); out != test.out {
			t.Fatalf("%d expect %g, got %g", i, test.out, out)
		}
	}
}

func TestNextDecValue(t *testing.T) {
	tests := []struct {
		in  string
		out bool
		tk  numToken
		pos int
	}{
		// 0
		{in: " ", out: false, pos: 0, tk: numToken{tag: tagUnknown}},
		{in: ".0", out: true, pos: 2, tk: numToken{tag: tagDecimalVal, val: 0.}},
		{in: "10.", out: true, pos: 3, tk: numToken{tag: tagDecimalVal, val: 10.}},
		{in: ".1_0_0", out: true, pos: 6, tk: numToken{tag: tagDecimalVal, val: .100}},
		{in: "184467440.73709551615e10", out: true, pos: 24, tk: numToken{tag: tagDecimalVal, val: 1.8446744073709553e+18}},
		// 5
		{in: "750._", out: true, tk: numToken{tag: tagError, val: ErrInvalidDecimalNumber}},
		{in: "45e1345", out: true, tk: numToken{tag: tagError, val: ErrInvalidDecimalNumber}},
		{in: "10.e12__0", out: true, tk: numToken{tag: tagError, val: ErrInvalidDecimalNumber}},
		{in: "10.e", out: true, tk: numToken{tag: tagError, val: ErrInvalidDecimalNumber}},
		{in: "._123", out: true, tk: numToken{tag: tagError, val: ErrInvalidDecimalNumber}},
	}
	for i, test := range tests {
		var hasErrors bool
		var tk numTokenizer
		tk.init([]byte(test.in))
		if exp, out := test.out, tk.nextDecValue(); out != exp {
			hasErrors = true
			t.Errorf("expect %v, got %v", exp, out)
		}
		b := checkToken(test.tk, tk.tk, t)
		hasErrors = hasErrors || b
		if exp, out := test.pos, tk.pos; exp != out {
			hasErrors = true
			t.Errorf("expect pos %d, got %d", exp, out)
		}
		if hasErrors {
			pc := make([]uintptr, 15)
			n := runtime.Callers(1, pc)
			frames := runtime.CallersFrames(pc[:n])
			frame, _ := frames.Next()
			// fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
			t.Fatalf("Errors with test %d in %s (%s:%d)", i, t.Name(), filepath.Base(frame.File), frame.Line)
		}
	}
}

func TestNumTokenizer(t *testing.T) {
	var tk numTokenizer
	tests := []struct {
		in  string
		out []numToken
	}{
		// 0
		{in: "", out: []numToken{{tag: tagError, val: ErrEndOfInput}}},
		{in: "( 10", out: []numToken{{tag: tagOpenParen}, {tag: tagIntegerVal, pos: 2, val: 10}, {tag: tagError, val: ErrEndOfInput, pos: 4}}},
		{in: "-0x1A", out: []numToken{{tag: tagMinus}, {tag: tagIntegerVal, pos: 1, val: 26}, {tag: tagError, val: ErrEndOfInput, pos: 5}}},
		{in: "2.3 x", out: []numToken{{tag: tagDecimalVal, val: 2.3}, {tag: tagError, pos: 4, val: ErrInvalidNumericExpression}}},
		{in: "0b10 23_", out: []numToken{{tag: tagIntegerVal, val: 2}, {tag: tagError, pos: 5, val: ErrInvalidIntegerNumber}}},
		// 5
		{in: "0o750 ", out: []numToken{{tag: tagIntegerVal, val: 488}, {tag: tagError, val: ErrEndOfInput, pos: 6}}},
		{in: "1 + 2 * 3", out: []numToken{{tag: tagIntegerVal, val: 1}, {tag: tagPlus, pos: 2}, {tag: tagIntegerVal, pos: 4, val: 2},
			{tag: tagMultiplication, pos: 6}, {tag: tagIntegerVal, pos: 8, val: 3}, {tag: tagError, val: ErrEndOfInput, pos: 9}}},
		{in: "0b_", out: []numToken{{tag: tagError, val: ErrInvalidBinaryNumber}}},
		{in: "0", out: []numToken{{tag: tagIntegerVal, val: 0}, {tag: tagError, val: ErrEndOfInput, pos: 1}}},
		{in: "0 ", out: []numToken{{tag: tagIntegerVal, val: 0}, {tag: tagError, val: ErrEndOfInput, pos: 2}}},
	}

	for i, test := range tests {
		tk.init([]byte(test.in))
		var out []numToken
		for {
			tk.nextToken()
			out = append(out, tk.token())
			if tk.tk.tag == tagError {
				break
			}
		}
		var hasErrors bool
		if len(out) != len(test.out) {
			hasErrors = true
			t.Errorf("expect nbr numToken %d, got %d", len(test.out), len(out))
		}
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
			pc := make([]uintptr, 15)
			n := runtime.Callers(1, pc)
			frames := runtime.CallersFrames(pc[:n])
			frame, _ := frames.Next()
			// fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
			t.Fatalf("Errors with test %d in %s (%s:%d)", i, t.Name(), filepath.Base(frame.File), frame.Line)
		}
	}

	tk.init([]byte("test"))
	tmp := numToken(numToken{tag: tagError, val: ErrEndOfInput})
	tk.tk = tmp
	tk.nextToken()
	tk.nextToken()
	if !reflect.DeepEqual(tk.token(), tmp) {
		t.Fatalf("expect numToken %v, got %v", tmp, tk.token())
	}

}
