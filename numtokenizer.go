package qjson

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// The tokenizer is used only for numbers and arithmetic operations.
// The tokenizer input is a quoteless string. The output are the operators
// "()+-*/%^|&~" and int and float values, or an error. The binary and
// hexadecimal numbers are converted into int by the tokenizer.
//
//
// Expression evaluation:
// int are implicitely cast into float when combined with a float in a
// the operations "+-*/". The operations "^~|&%" require two ints and
// output an int.

// isNumberExpr return true if p is a number expression. It looks for the
// first digit that must be in the range '0' to '9'.
func isNumberExpr(p []byte) bool {
	for i := 0; i < len(p); i++ {
		if p[i] == '+' || p[i] == '-' || p[i] == ' ' || p[i] == '\t' || p[i] == '(' {
			continue
		}
		return isIntDigit(p[i]) || (p[i] == '.' && i+1 < len(p) && isIntDigit(p[i+1]))
	}
	return false
}

// numToken is a token produced by the number expression tokenizer.
// The field val may be nil, an error, a float64 or nil. Nothing else.
// val is nil in case of error or when the end of input is reached.
// In this case, the error, if any, and its position are stored in
// the err and errPos field of the tokenizer.
type numToken struct {
	val interface{}
	pos int
	tag tokenTag
}

func (t numToken) String() string {
	var buf strings.Builder
	if str, ok := tagStr[t.tag]; ok {
		buf.WriteString(fmt.Sprintf("numToken{tag: %s", str))
	} else {
		buf.WriteString(fmt.Sprintf("numToken{tag: %d", t.tag))
	}
	if t.pos != 0 {
		buf.WriteString(fmt.Sprintf(", pos: %v", t.pos))
	}
	switch x := t.val.(type) {
	case nil:
	case error:
		buf.WriteString(fmt.Sprintf(", val: %v", errStr(x)))
	case int, float64:
		buf.WriteString(fmt.Sprintf(", val: %v", x))
	default:
		buf.WriteString(fmt.Sprintf(", val: %v of type %s", x, reflect.TypeOf(x)))
	}
	buf.WriteString("}")
	return buf.String()
}

type numTokenizer struct {
	in     []byte   // input expression to parse
	p      []byte   // expression left to parse
	pos    int      // position in b of the first byte of p
	err    error    // the last error or nil if none
	errPos int      // the index of the error
	tk     numToken // the last token
}

func (tk *numTokenizer) init(input []byte) {
	tk.in = input
	tk.p = input
	tk.pos = 0
	tk.tk = numToken{tag: tagUnknown}
}

func (tk *numTokenizer) token() numToken {
	return tk.tk
}

func (tk *numTokenizer) done() bool {
	return tk.tk.tag == tagError
}

func (tk *numTokenizer) setToken(tag tokenTag, val interface{}) {
	tk.tk = numToken{tag: tag, pos: tk.pos, val: val}
}

func (tk *numTokenizer) setError(err error) {
	tk.setToken(tagError, err)
}

func (tk *numTokenizer) setErrorAndPos(err error, pos int) {
	tk.tk = numToken{tag: tagError, pos: pos, val: err}
}

// popBytes skips n bytes of input. It has no effect if at end of input or
// the end of input is reached.
func (tk *numTokenizer) popBytes(n int) {
	if n > len(tk.p) {
		n = len(tk.p)
	}
	tk.p = tk.p[n:]
	tk.pos += n
}

func (tk *numTokenizer) nextToken() {
	if tk.done() {
		return
	}
	for len(tk.p) > 0 {
		if n := whitespace(tk.p); n != 0 {
			tk.popBytes(n)
		} else {
			break
		}
	}
	if len(tk.p) == 0 {
		tk.setError(ErrEndOfInput)
		return
	}

	if !tk.nextISODateTimeValue() && !tk.nextOperator() && !tk.nextBinValue() && !tk.nextHexValue() &&
		!tk.nextDecValue() && !tk.nextOctValue() && !tk.nextIntValue() {
		tk.setError(ErrInvalidNumericExpression)
	}
}

var tkOpTable = [256]tokenTag{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 00
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 10
	0, 0, 0, 0, 0, tagModulo, tagAnd, 0, tagOpenParen, // 20
	tagCloseParen, tagMultiplication, tagPlus, 0, tagMinus, 0, tagDivision, // 29
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 30
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 40
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, tagXor, 0, // 50
	0, 0, 0, 0, tagDays, 0, 0, 0, tagHours, 0, 0, 0, 0, tagMinutes, 0, 0, // 60
	0, 0, 0, tagSeconds, 0, 0, 0, tagWeeks, 0, 0, 0, 0, tagOr, 0, tagInverse, 0, // 70
}

// nextOperator returns true and pops the operator if tk.p start with
// an operatore. Otherwise return false.
func (tk *numTokenizer) nextOperator() bool {
	x := tkOpTable[tk.p[0]]
	if x == tagUnknown {
		return false
	}
	tk.setToken(x, nil)
	tk.popBytes(1)
	return true
}

// inRange return true if lo <= v <= hi.
func inRange(v, lo, hi byte) bool {
	return v-lo <= hi-lo
}

// skips n bytes in front of v and the optional underscore.
// returns the number of bytes skipped and v trimmed by this number,
// or -1 and nil if the of v is reached.
func skipHeaderAndOptionalUnderscore(n int, v []byte) (int, []byte) {
	if n >= len(v) {
		return -1, nil
	}
	v = v[n:]
	if v[0] == '_' {
		n++
		v = v[1:]
		if len(v) == 0 {
			return -1, nil
		}
	}
	return n, v
}

func isBinDigit(v byte) bool {
	return v == '0' || v == '1'
}

// return the number of bytes parsed or 0 if not valid digits
func parseBinDigits(v []byte) int {
	if !isBinDigit(v[0]) {
		return 0
	}
	for p := 1; p < len(v); p++ {
		if v[p] == '_' {
			p++
			if p == len(v) {
				return -1
			}
		}
		if !isBinDigit(v[p]) {
			if v[p-1] == '_' {
				return -1
			}
			return p
		}
	}
	return len(v)
}

// return 0 if not literal, -1 if invalid literal, n when valid and length is n bytes.
func parseBinLiteral(v []byte) int {
	if len(v) < 2 || v[0] != '0' || v[1]&0b11011111 != 'B' {
		return 0
	}
	var n int
	if n, v = skipHeaderAndOptionalUnderscore(2, v); n >= 0 {
		if p := parseBinDigits(v); p > 0 {
			return n + p
		}
	}
	return -1
}

// return the value which is in the range 0 to MAX_INT, or -1 if overflows.
func decodeBinLiteral(v []byte) int {
	var val uint64
	v = v[2:]
	for p := 0; p < len(v); p++ {
		if v[p] == '_' {
			continue
		}
		if val&0x80000000_00000000 != 0 {
			return -1
		}
		val = val << 1
		if v[p] == '1' {
			val |= 1
		}
	}
	if val&0x80000000_00000000 != 0 {
		return -1
	}
	return int(val)
}

func (tk *numTokenizer) nextBinValue() bool {
	n := parseBinLiteral(tk.p)
	if n == 0 {
		return false
	}
	if n < 0 {
		tk.setError(ErrInvalidBinaryNumber)
		return true
	}
	val := decodeBinLiteral(tk.p[:n])
	if val < 0 {
		tk.setError(ErrNumberOverflow)
		return true
	}
	tk.setToken(tagIntegerVal, val)
	tk.popBytes(n)
	return true
}

func isHexDigit(v byte) bool {
	return isIntDigit(v) || inRange(v&0b11011111, 'A', 'F')
}

// return the number of bytes parsed or 0 if not valid digits
func parseHexDigits(v []byte) int {
	l := len(v)
	if l == 0 || !isHexDigit(v[0]) {
		return 0
	}
	for p := 1; p < l; p++ {
		if v[p] == '_' {
			p++
			if p == l {
				return -1
			}
		}
		if !isHexDigit(v[p]) {
			if v[p-1] == '_' {
				return -1
			}
			return p
		}
	}
	return l
}

// return 0 if not literal, -1 if invalid literal, n when valid and length is n bytes.
func parseHexLiteral(v []byte) int {
	if len(v) < 2 || v[0] != '0' || v[1]&0b11011111 != 'X' {
		return 0
	}
	var n int
	if n, v = skipHeaderAndOptionalUnderscore(2, v); n >= 0 {
		if p := parseHexDigits(v); p > 0 {
			return n + p
		}
	}
	return -1
}

// return the value which is in the range 0 to MAX_INT, or -1 if overflows.
func decodeHexLiteral(v []byte) int {
	var val uint64
	v = v[2:]
	for p := 0; p < len(v); p++ {
		if v[p] == '_' {
			continue
		}
		if val&0xF0000000_00000000 != 0 {
			return -1
		}
		if inRange(v[p], '0', '9') {
			val = val<<4 | uint64(v[p]-'0')
			continue
		}
		x := v[p] & 0b11011111
		val = val<<4 | uint64(x-'A') + 10
	}
	if val&0x80000000_00000000 != 0 {
		return -1
	}
	return int(val)
}

func (tk *numTokenizer) nextHexValue() bool {
	n := parseHexLiteral(tk.p)
	if n == 0 {
		return false
	}
	if n < 0 {
		tk.setError(ErrInvalidHexadecimalNumber)
		return true
	}
	val := decodeHexLiteral(tk.p[:n])
	if val < 0 {
		tk.setError(ErrNumberOverflow)
		return true
	}
	tk.setToken(tagIntegerVal, val)
	tk.popBytes(n)
	return true
}

func isOctDigit(v byte) bool {
	return inRange(v, '0', '7')
}

// return the number of bytes parsed or 0 if not valid digits
func parseOctDigits(v []byte) int {
	l := len(v)
	if l == 0 || !isOctDigit(v[0]) {
		return 0
	}
	for p := 1; p < l; p++ {
		if v[p] == '_' {
			p++
			if p == l {
				return -1
			}
		}
		if !isOctDigit(v[p]) {
			if v[p-1] == '_' {
				return -1
			}
			return p
		}
	}
	return l
}

// return 0 if not literal, -1 if invalid literal, n when valid and length is n bytes.
func parseOctLiteral(v []byte) int {
	if len(v) < 1 || v[0] != '0' {
		return 0
	}
	var n int
	if len(v) >= 2 && v[1]&0b11011111 == 'O' {
		if n, v = skipHeaderAndOptionalUnderscore(2, v); n >= 0 {
			if p := parseOctDigits(v); p > 0 {
				return n + p
			}
		}
		return -1
	}
	// a 0 at end of input or followed by anything different from _ and an oct digit
	// is not an octal number. It’s thus not invalid.
	if len(v) < 2 || (v[1] != '_' && !isOctDigit(v[1])) {
		return 0
	}
	if n, v = skipHeaderAndOptionalUnderscore(1, v); n >= 0 {
		if p := parseOctDigits(v); p > 0 {
			return n + p
		}
	}
	return -1
}

// return the value which is in the range 0 to MAX_INT, or -1 if overflows.
func decodeOctLiteral(v []byte) int {
	var val uint64
	if v[1]&0b11011111 == 'O' {
		v = v[2:]
	} else {
		v = v[1:]
	}
	for p := 0; p < len(v); p++ {
		if v[p] == '_' {
			continue
		}
		if val&0xF0000000_00000000 != 0 {
			return -1
		}
		val = val<<3 | uint64(v[p]-'0')
	}
	return int(val)
}

func (tk *numTokenizer) nextOctValue() bool {
	n := parseOctLiteral(tk.p)
	if n == 0 {
		return false
	}
	if n < 0 {
		tk.setError(ErrInvalidOctalNumber)
		return true
	}
	val := decodeOctLiteral(tk.p[:n])
	if val < 0 {
		tk.setError(ErrNumberOverflow)
		return true
	}
	tk.setToken(tagIntegerVal, val)
	tk.popBytes(n)
	return true
}

func isIntDigit(v byte) bool {
	return inRange(v, '0', '9')
}

// return the number of bytes parsed or 0 if not valid digits
func parseIntDigits(v []byte) int {
	l := len(v)
	if l == 0 || !isIntDigit(v[0]) {
		return 0
	}
	for p := 1; p < l; p++ {
		if v[p] == '_' {
			p++
			if p == l {
				return -1
			}
		}
		if !isIntDigit(v[p]) {
			if v[p-1] == '_' {
				return -1
			}
			return p
		}
	}
	return l
}

// return 0 if not literal or n when valid and length is n bytes.
func parseIntLiteral(v []byte) int {
	if inRange(v[0], '1', '9') {
		return parseIntDigits(v)
	}
	if v[0] != '0' {
		return 0
	}
	if len(v) > 1 && (v[1] == '_' || isIntDigit(v[1])) {
		return -1
	}
	return 1
}

// return the value which is in the range 0 to MAX_INT, or -1 if overflows.
func decodeIntLiteral(v []byte) int {
	var val uint64
	for p := 0; p < len(v); p++ {
		if v[p] == '_' {
			continue
		}
		if val > 0x1999999999999999 {
			return -1
		}
		val = val*10 + uint64(v[p]-'0')
	}
	if val&0x80000000_00000000 != 0 {
		return -1
	}
	return int(val)
}

func (tk *numTokenizer) nextIntValue() bool {
	n := parseIntLiteral(tk.p) // n is allways >= 0
	if n == 0 {
		return false
	}
	if n < 0 {
		tk.setError(ErrInvalidIntegerNumber)
		return true
	}
	val := decodeIntLiteral(tk.p[:n])
	if val < 0 {
		tk.setError(ErrNumberOverflow)
		return true
	}
	tk.setToken(tagIntegerVal, val)
	tk.popBytes(n)
	return true
}

// return 0 if not exponent, -1 if invalid exponent, n when valid and length is n bytes.
func parseExponent(v []byte) int {
	if len(v) == 0 || v[0]&0b11011111 != 'E' {
		return 0
	}
	n := 1
	v = v[1:]
	if len(v) == 0 {
		return -1
	}
	if v[0] == '+' || v[0] == '-' {
		n++
		v = v[1:]
		if len(v) == 0 {
			return -1
		}
	}
	if p := parseIntDigits(v); p > 0 {
		return n + p
	}
	return -1
}

// return 0 if not literal, -1 if invalid literal, n when valid and length is n bytes.
func parseDecLiteral(v []byte) int {
	p := parseIntDigits(v)
	if p < 0 {
		return 0
	}
	if p == 0 {
		// numbers must be of the form .123[e[+/-]145]
		if v[0] != '.' || len(v) < 2 {
			return 0
		}
		p := parseIntDigits(v[1:])
		if p < 0 {
			return -1
		}
		if p == 0 {
			if len(v) > 1 && (v[1] == '_' || v[1]&0b11011111 == 'E') {
				return -1
			}
			return 0
		}
		q := parseExponent(v[p+1:])
		if q < 0 {
			return -1
		}
		return 1 + p + q
	}
	//  numbers must be of the form 123e[+/-]145 or 123.456[e[+/-]789]
	n := p
	v = v[p:]
	q := parseExponent(v)
	if q < 0 {
		return -1
	}
	if q > 0 {
		return p + q
	}
	// numbers must be of the form 123.456[e[+/-]789]
	if len(v) == 0 {
		return 0
	}
	if v[0] != '.' {
		return 0 // not invalid, but not a decimal number
	}
	n++
	v = v[1:]
	q = parseIntDigits(v)
	if q > 0 {
		n += q
		v = v[q:]
	}
	if q < 0 {
		return -1
	}
	p = parseExponent(v)
	if p < 0 {
		return -1
	}
	n += p
	if len(v) > p && v[p] == '_' {
		return -1
	}
	return n
}

// return the value which is in the range 0 to MAX_INT, or -1 if overflows.
func decodeDecLiteral(v []byte) float64 {
	x, err := strconv.ParseFloat(string(v), 64)
	if err != nil {
		return -1
	}
	return x
}

func (tk *numTokenizer) nextDecValue() bool {
	n := parseDecLiteral(tk.p)
	if n == 0 {
		return false
	}
	if n < 0 {
		tk.setError(ErrInvalidDecimalNumber)
		return true
	}
	val := decodeDecLiteral(tk.p[:n])
	if val < 0 {
		tk.setError(ErrInvalidDecimalNumber)
		return true
	}
	tk.setToken(tagDecimalVal, val)
	tk.popBytes(n)
	return true
}

// see https://fr.wikipedia.org/wiki/ISO_8601 (ex: 1997−07−16T19:20+01:00) RFC3339
func parseISODateTimeLiteral(v []byte) int {
	// must start with date
	if len(v) < 11 || v[10] != 'T' || v[4] != '-' || v[7] != '-' ||
		!isIntDigit(v[0]) || !isIntDigit(v[1]) || !isIntDigit(v[2]) || !isIntDigit(v[3]) ||
		!isIntDigit(v[5]) || !isIntDigit(v[6]) || !isIntDigit(v[8]) || !isIntDigit(v[9]) {
		return 0
	}
	n := 11
	v = v[11:]
	if len(v) == 0 {
		return n
	}
	if len(v) < 5 || !isIntDigit(v[0]) || !isIntDigit(v[1]) || !isIntDigit(v[3]) || !isIntDigit(v[4]) || v[2] != ':' {
		return n
	}
	n += 5
	v = v[5:]
	if len(v) == 0 {
		return n
	}
	if v[0] == 'Z' {
		return n + 1
	}
	if v[0] != ':' {
		return n
	}
	if len(v) < 3 || !isIntDigit(v[1]) || !isIntDigit(v[2]) {
		return -1
	}
	n += 3
	v = v[3:]
	if len(v) == 0 {
		return n
	}
	if v[0] == 'Z' {
		return n + 1
	}
	if v[0] != '.' && v[0] != '+' && v[0] != '-' {
		return n
	}
	// milli or micro seconds
	if v[0] == '.' {
		n++
		v = v[1:]
		var p int
		for len(v) > p && isIntDigit(v[p]) {
			p++
		}
		if p != 6 && p != 3 {
			return -1
		}
		n += p
		v = v[p:]
	}
	if len(v) == 0 {
		return n
	}
	if v[0] == 'Z' {
		return n + 1
	}
	// optional time offset
	if v[0] == '+' || v[0] == '-' {
		n++
		v = v[1:]
		if len(v) < 5 || v[2] != ':' || !isIntDigit(v[0]) || !isIntDigit(v[1]) ||
			!isIntDigit(v[3]) || !isIntDigit(v[4]) {
			return -1
		}
		n += 5
	}
	return n
}

// return -1 if decoding failed
func decodeISODateTimeLiteral(v []byte) float64 {
	s := string(v)
	var t time.Time
	var err error
	layouts := []string{
		"2006-01-02T15:04:05.999999Z",
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04Z",
		"2006-01-02T15:04:05.999999-07:00",
		"2006-01-02T15:04:05.999-07:00",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000000",
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T",
	}
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			break
		}
	}
	if err != nil {
		return -1
	}
	x := float64(t.Unix()) + float64(t.Nanosecond())/1e9
	if x < 0 {
		return -1
	}
	return x
}

func (tk *numTokenizer) nextISODateTimeValue() bool {
	n := parseISODateTimeLiteral(tk.p)
	if n == 0 {
		return false
	}
	if n < 0 {
		tk.setError(ErrInvalidISODateTime)
		return true
	}
	val := decodeISODateTimeLiteral(tk.p[:n])
	if val < 0 {
		tk.setError(ErrInvalidISODateTime)
		return true
	}
	tk.setToken(tagDecimalVal, val)
	tk.popBytes(n)
	return true
}
