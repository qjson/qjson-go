package qjson

import (
	"path/filepath"
	"runtime"
	"testing"
)

func doesPanic(f func()) (res bool) {
	defer func() {
		if r := recover(); r == nil {
			res = false
		} else {
			res = true
		}
	}()
	f()
	return false
}

func TestToFloat64(t *testing.T) {
	x := toFloat64(10)
	if x != 10. {
		t.Fatal("invalid output value")
	}
	x = toFloat64(5.2)
	if x != 5.2 {
		t.Fatal("invalid output value")
	}
	if !doesPanic(func() { toFloat64(nil) }) {
		t.Fatal("invalid output value")
	}
	if !doesPanic(func() { toFloat64("test") }) {
		t.Fatal("invalid output value")
	}
}

func TestNumberEval(t *testing.T) {
	tests := []struct {
		in  string
		out float64
		pos int
		err error
	}{
		// 0
		{in: "", err: ErrEndOfInput},
		{in: " ", err: ErrEndOfInput, pos: 1},
		{in: "10", out: 10},
		{in: "a", err: ErrInvalidNumericExpression},
		{in: "2 + 5", out: 7},
		// 5
		{in: "2 + 5.3", out: 7.3},
		{in: "1 + 2 * 3", out: 7},
		{in: "1 + 2 * 3 / 2", out: 4},
		{in: "1 + 2 * (3 / 2)", out: 3},
		{in: "1 + -2", out: -1},
		// 10
		{in: "1 + +2", out: 3},
		{in: "1 + -(10 + 2)", out: -11},
		{in: "1 + -(10 + 2", err: ErrUnclosedParenthesis, pos: 5},
		{in: "1 + -(10 - (2 + 3))", out: -4},
		{in: "1 + ~7.3", err: ErrOperandMustBeInteger, pos: 4},
		// 15
		{in: "10 % 3", out: 1},
		{in: "10. % 3", err: ErrOperandsMustBeInteger, pos: 4},
		{in: "3*1024*1024", out: 3 * 1024 * 1024},
		{in: "2. + (0x3 | 0x4)", out: 9},
		{in: "2. + (0x7 & ~0x2)", out: 7},
		// 20
		{in: "5 a", err: ErrInvalidNumericExpression, pos: 2},
		{in: "10. * 3.", out: 30},
		{in: "10. & 3", err: ErrOperandsMustBeInteger, pos: 4},
		{in: "10. | 3", err: ErrOperandsMustBeInteger, pos: 4},
		{in: "10. ^ 3", err: ErrOperandsMustBeInteger, pos: 4},
		// 25
		{in: "10. & ", err: ErrInvalidNumericExpression, pos: 4},
		{in: "10. | ", err: ErrInvalidNumericExpression, pos: 4},
		{in: "10. ^ ", err: ErrInvalidNumericExpression, pos: 4},
		{in: "& ", err: ErrInvalidNumericExpression, pos: 0},
		{in: "10 ** ", err: ErrInvalidNumericExpression, pos: 4},
		// 30
		{in: "10 + ", err: ErrInvalidNumericExpression, pos: 3},
		{in: "10 * ", err: ErrInvalidNumericExpression, pos: 3},
		{in: "10 - ", err: ErrInvalidNumericExpression, pos: 3},
		{in: "10 / ", err: ErrInvalidNumericExpression, pos: 3},
		{in: "10 % ", err: ErrInvalidNumericExpression, pos: 3},
		// 35
		{in: "10 & 0b_1_ ", err: ErrInvalidBinaryNumber, pos: 5},
		{in: "10 ^ 0b_1_ ", err: ErrInvalidBinaryNumber, pos: 5},
		{in: "10 | 0b_1_ ", err: ErrInvalidBinaryNumber, pos: 5},
		{in: "10 ^ 3", out: 9},
		{in: "10 % 0b_1_ ", err: ErrInvalidBinaryNumber, pos: 5},
		// 40
		{in: "~0b_1_ ", err: ErrInvalidBinaryNumber, pos: 1},
		{in: "~", err: ErrInvalidNumericExpression, pos: 0},
		{in: "(0b_1_) ", err: ErrInvalidBinaryNumber, pos: 1},
		{in: "( ", err: ErrInvalidNumericExpression, pos: 0},
		{in: ") ", err: ErrUnopenedParenthesis, pos: 0},
		// 45
		{in: "10 +) ", err: ErrUnopenedParenthesis, pos: 4},
		{in: "(10 + 3)) ", err: ErrUnopenedParenthesis, pos: 8},
		{in: "10) ", err: ErrUnopenedParenthesis, pos: 2},
		{in: "10. / 5__2. ", err: ErrInvalidIntegerNumber, pos: 6},
		{in: "10. / 2. ", out: 5},
		// 50
		{in: "10. - 2. ", out: 8},
		{in: "10. + -", err: ErrInvalidNumericExpression, pos: 6},
		{in: "10. + - 2. ", out: 8},
		{in: "10. ~ ", err: ErrInvalidNumericExpression, pos: 4},
		{in: "1 + 1 2_ ", err: ErrInvalidIntegerNumber, pos: 6},
		// 55
		{in: "10. / 0", err: ErrDivisionByZero, pos: 4}, // go-fuzz
		{in: "10 % 0", err: ErrDivisionByZero, pos: 3},  // go-fuzz
		{in: "10 / 0", err: ErrDivisionByZero, pos: 3},  // go-fuzz
		{in: "+", err: ErrInvalidNumericExpression},
		{in: "10 - +5", out: 5},
		// 60
		{in: "2m", out: 120},
		{in: "2m 15", out: 135},
		{in: "2m a", err: ErrInvalidNumericExpression, pos: 3},
		{in: "1h", out: 3600},
		{in: "1h 10", out: 3610},
		// 65
		{in: "2h a", err: ErrInvalidNumericExpression, pos: 3},
		{in: "1d", out: 86400},
		{in: "1d 10", out: 86410},
		{in: "2d a", err: ErrInvalidNumericExpression, pos: 3},
		{in: "1w", out: 604800},
		// 70
		{in: "1w 10", out: 604810},
		{in: "2w a", err: ErrInvalidNumericExpression, pos: 3},
		{in: "1 s", out: 1},
		{in: "1s 10", out: 11},
		{in: "2s a", err: ErrInvalidNumericExpression, pos: 3},
		// 75
		{in: "6 7", err: ErrInvalidNumericExpression, pos: 2},
		{in: "1.3 5h", err: ErrInvalidNumericExpression, pos: 4},
		{in: "(2 + 3)*2", out: 10},
		{in: "1.3 + 1h", out: 3601.3},
		{in: "1h 2m 2s", out: 3722},
		// 80
		{in: "1h 2m 2s + 4", out: 3726},
		{in: "(1m) * 2", out: 120},
		{in: "(1h 2m 2s) * 2", out: 7444},
		{in: "1h 2m 2s - 2", out: 3720},
		{in: "-(1h 2m 2s)", out: -3722},
		// 85
		{in: "-1h 2m 2s", out: -3478},
		{in: "(1w) * 2", out: 1209600},
		{in: "(1d) * 2", out: 172800},
		{in: "(1h) * 2", out: 7200},
		{in: "1h 2m 2s * 3", err: ErrInvalidNumericExpression, pos: 9},
		// 90
		{in: "2020-12-23T15:40:05", out: 1608738005},
		{in: "2020-12-23T15:40:05 + 2m", out: 1608738125},
		{in: "2020-12-23T25:40:05", err: ErrInvalidISODateTime, pos: 0},
		{in: "2020-12-23T15:40:60", err: ErrInvalidISODateTime, pos: 0},
	}
	for i, test := range tests {
		out, pos, err := evalNumberExpression([]byte(test.in))
		var hasErrors bool
		if out != test.out {
			hasErrors = true
			t.Errorf("expected out %g, got %g", test.out, out)
		}
		if exp, outErr := errStr(test.err), errStr(err); exp != outErr || test.pos != pos {
			hasErrors = true
			t.Errorf("expected err: %s pos: %d, got err: %s pos: %d", exp, test.pos, outErr, pos)
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
