package qjson

import (
	"fmt"
	"reflect"
	"strings"
)

type pos struct {
	b int // index of first byte of p
	s int // start of line index
	l int // line index number
}

func (p pos) String() string {
	if p.s == 0 && p.l == 0 {
		return fmt.Sprintf("pos{b: %d}", p.b)
	}
	return fmt.Sprintf("pos{b: %d, s: %d, l:%d}", p.b, p.s, p.l)
}

type tokenTag byte

const (
	tagUnknown tokenTag = iota
	tagError
	tagIntegerVal
	tagDecimalVal
	tagPlus
	tagMinus
	tagMultiplication
	tagDivision
	tagXor
	tagAnd
	tagOr
	tagInverse
	tagModulo
	tagOpenParen
	tagCloseParen
	tagOpenBrace
	tagCloseBrace
	tagOpenSquare
	tagCloseSquare
	tagColon
	tagQuotelessString
	tagDoubleQuotedString
	tagSingleQuotedString
	tagMultilineString
	tagComma
)

var tagStr = map[tokenTag]string{
	tagUnknown:            "tagUnknown",
	tagError:              "tagError",
	tagIntegerVal:         "tagIntegerVal",
	tagDecimalVal:         "tagDecimalVal",
	tagPlus:               "tagPlus",
	tagMinus:              "tagMinus",
	tagMultiplication:     "tagMultiplication",
	tagDivision:           "tagDivision",
	tagXor:                "tagXor",
	tagAnd:                "tagAnd",
	tagOr:                 "tagOr",
	tagInverse:            "tagInverse",
	tagModulo:             "tagModulo",
	tagOpenParen:          "tagOpenParen",
	tagCloseParen:         "tagCloseParen",
	tagOpenBrace:          "tagOpenBrace",
	tagCloseBrace:         "tagCloseBrace",
	tagOpenSquare:         "tagOpenSquare",
	tagCloseSquare:        "tagCloseSquare",
	tagColon:              "tagColon",
	tagQuotelessString:    "tagQuotelessString",
	tagDoubleQuotedString: "tagDoubleQuotedString",
	tagSingleQuotedString: "tagSingleQuotedString",
	tagMultilineString:    "tagMultilineString",
	tagComma:              "tagComma",
}

func b2q(b []byte) string {
	if b == nil {
		return "nil"
	}
	return fmt.Sprintf("%q", b)
}

type token struct {
	pos
	val interface{}
	tag tokenTag
}

func (t token) String() string {
	var buf strings.Builder
	if str, ok := tagStr[t.tag]; ok {
		buf.WriteString(fmt.Sprintf("{tag: %s", str))
	} else {
		buf.WriteString(fmt.Sprintf("{tag: %d", t.tag))
	}
	if t.b != 0 || t.s != 0 || t.l != 0 {
		buf.WriteString(fmt.Sprintf(", pos: %v", t.pos))
	}
	switch t.val.(type) {
	case nil:
	case []byte:
		buf.WriteString(fmt.Sprintf(", val: %s", b2q(t.val.([]byte))))
	case Error:
		buf.WriteString(fmt.Sprintf(", val: %s", errStr(t.val.(error))))
	case int, float64:
		buf.WriteString(fmt.Sprintf(", val: %v", t.val))
	default:
		buf.WriteString(fmt.Sprintf(", val: %v type: %v", t.val, reflect.TypeOf(t.val)))
	}
	buf.WriteString("}")
	return buf.String()
}

// A tokenizer produces QJSON tokens from input.
type tokenizer struct {
	pos
	in     []byte // input text
	p      []byte // text left to parse
	tk     token
	margin []byte // small buffer holding the last margin
}

// init resets the tokenizer. Requires that nexToken() is called afterward.
func (tk *tokenizer) init(in []byte) {
	tk.pos = pos{}
	tk.in = in
	tk.p = tk.in
	tk.tk = token{}
	tk.margin = tk.margin[:0]
}

// whitespace returns the byte size of the first whitechar of p.
// It return 0 if p is empty or the first char of p is not a whitespace.
func whitespace(p []byte) int {
	if len(p) == 0 {
		return 0
	}
	if p[0] == ' ' || p[0] == '\t' {
		return 1
	}
	if p[0] == 0xC2 && len(p) >= 2 && p[1] == 0xA0 {
		return 2
	}
	return 0
}

// newline returns the byte lenght of the newline at start of p.
// It return 0 if p is empty or there is no newline, otherwise, it returns 1 or 2.
func newline(p []byte) int {
	if len(p) == 0 {
		return 0
	}
	if p[0] == '\n' {
		return 1
	}
	if p[0] == '\r' && len(p) >= 2 && p[1] == '\n' {
		return 2
	}
	return 0
}

// popBytes pops n bytes from the bytes left to parse. Return an error
// if there are not enough bytes to pop.
// The bytes must not contain a newline. Use popNewline() for newlines.
func (tk *tokenizer) popBytes(n int) {
	tk.b += n
	tk.p = tk.p[n:]
}

// popNewline returns true if a newline was popped from the front of tk.p.
// It return false if tk.p is empty or there is no newline in front of tk.p.
func (tk *tokenizer) popNewline() bool {
	if n := newline(tk.p); n != 0 {
		tk.s = tk.b + n
		tk.l++
		tk.popBytes(n)
		return true
	}
	return false
}

func (tk *tokenizer) token() token {
	return tk.tk
}

const (
	s0 byte = 0x00 // invalid character (e.g. control characters)
	s1      = 0x01 // valid characters (printable ascii characters)
	s2      = 0x12 // rule 1, 2 byte long
	s3      = 0x23 // rule 2, 3 byte long
	s4      = 0x13 // rule 1, 3 byte long
	s5      = 0x33 // rule 3, 3 byqte long
	s6      = 0x44 // rule 4, 4 byte long
	s7      = 0x14 // rule 1, 4 byte long
	s8      = 0x54 // rule 5, 4 byte long
)

// All control characters except \t are invalid.
// skipChar and readChar return false when the char is \n or \n.
// All valid utf8 character is valid.
var utf8Table = [256]byte{
	s0, s0, s0, s0, s0, s0, s0, s0, s0, s1, s0, s0, s0, s0, s0, s0, // 00
	s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, // 10
	s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 20
	s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 30
	s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 40
	s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 50
	s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 60
	s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, s1, // 70
	s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, // 80
	s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, // 90
	s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, // A0
	s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, // B0
	s0, s0, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, // C0
	s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, s2, // D0
	s3, s4, s4, s4, s4, s4, s4, s4, s4, s4, s4, s4, s4, s5, s4, s4, // E0
	s6, s7, s7, s7, s8, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, s0, // F0
}

const (
	utf8lo = 0b10000000
	utf8hi = 0b10111111
)

var utf8Range = [16]byte{
	0, 0,
	utf8lo, utf8hi,
	0xA0, utf8hi,
	utf8lo, 0x9F,
	0x90, utf8hi,
	utf8lo, 0x8F,
}

// char returns the byte length of the char in front of tk.p and nil,
// otherwise it returns 0 and the error which can be ErrEndOfInput,
// ErrInvalidChar or ErrTruncatedChar. The char is not popped.
func (tk *tokenizer) char() (int, *atError) {
	if len(tk.p) == 0 {
		return 0, nil
	}
	x := utf8Table[tk.p[0]]
	if x == s1 {
		return 1, nil
	}
	// Called by char for mid-stack inlining.
	return tk.charX(x)
}

// charX requires that x == s0 || x >= s3.
func (tk *tokenizer) charX(x byte) (int, *atError) {
	if x == s0 {
		return 0, &atError{pos: tk.pos, err: ErrInvalidChar}
	}
	n := int(x & 0xF)
	if n > len(tk.p) {
		return 0, &atError{pos: tk.pos, err: ErrTruncatedChar}
	}
	b2 := tk.p[1]
	r := (x >> 4) << 1
	if b2 < utf8Range[r] || b2 > utf8Range[r+1] {
		return 0, &atError{pos: tk.pos, err: ErrInvalidChar}
	}
	if n >= 3 {
		if tk.p[2] < utf8lo || tk.p[2] > utf8hi {
			return 0, &atError{pos: tk.pos, err: ErrInvalidChar}
		}
		if n == 4 {
			if tk.p[3] < utf8lo || tk.p[3] > utf8hi {
				return 0, &atError{pos: tk.pos, err: ErrInvalidChar}
			}
		}
	}
	return n, nil
}

// column return the number of runes in p. It requires that p contains
// a sequence of valid utf8 encoded runes.
func column(p []byte) int {
	var cnt int
	for len(p) != 0 {
		n := int(utf8Table[p[0]] & 0xF)
		if n == 0 || n > len(p) {
			break
		}
		p = p[n:]
		cnt++
	}
	return cnt
}

// skipRestOfLine pops all characters until an error occurs, a newline is met, or the
// end of input is met. In the later case no error is returned.
func (tk *tokenizer) skipRestOfLine() *atError {
	for {
		if tk.popNewline() || len(tk.p) == 0 {
			return nil
		}
		n, err := tk.char()
		if err != nil {
			return err
		}
		tk.popBytes(n)
	}
}

// skipLineComment return true and nil error if it successfully skipped #... or //... comments
// including the newline or the end of input is reached. Otherwise return false with the error.
func (tk *tokenizer) skipLineComment() (bool, *atError) {
	if len(tk.p) == 0 {
		return false, nil
	}
	if tk.p[0] == '#' || (tk.p[0] == '/' && len(tk.p) >= 2 && tk.p[1] == '/') {
		if err := tk.skipRestOfLine(); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// skipMultilineComment return false and nil when tk.p is not the start of a
// multiline comment. Return true and nil if successfully skipped a /*...*/ comment.
// Otherwise it returns false and an error.
func (tk *tokenizer) skipMultilineComment() (bool, *atError) {
	if len(tk.p) == 0 {
		return false, nil
	}
	if tk.p[0] != '/' || len(tk.p) < 2 || tk.p[1] != '*' {
		return false, nil
	}
	startPos := tk.pos
	tk.popBytes(2)
	for {
		if len(tk.p) == 0 {
			return false, &atError{pos: startPos, err: ErrUnclosedSlashStarComment}
		}
		if tk.p[0] == '*' && len(tk.p) >= 2 && tk.p[1] == '/' {
			tk.popBytes(2)
			return true, nil
		}
		if tk.popNewline() {
			continue
		}
		if tk.p[0] < 0x20 { // control characters are valid
			tk.popBytes(1)
			continue
		}
		n, err := tk.char()
		if err != nil {
			return false, err
		}
		tk.popBytes(n)
	}
}

// skips all whitespace characters.
func (tk *tokenizer) skipWhitespaces() {
	for n := whitespace(tk.p); n != 0; n = whitespace(tk.p) {
		tk.popBytes(n)
	}
}

// doubleQuotedString tries to parse a double quoted string. It returns nil
// if there is no double quoted string in front of tk.p. Requires tk is not
// done an dk.p is not empty.
func (tk *tokenizer) doubleQuotedString() ([]byte, *atError) {
	startPos := tk.pos
	if len(tk.p) == 0 || tk.p[0] != '"' {
		return nil, nil
	}
	tk.popBytes(1)
	for {
		if len(tk.p) == 0 {
			return nil, &atError{pos: startPos, err: ErrUnclosedDoubleQuoteString}
		}
		if tk.p[0] == '\\' && len(tk.p) > 1 && tk.p[1] == '"' {
			tk.popBytes(2)
			continue
		}
		if tk.p[0] == '"' {
			tk.popBytes(1)
			return tk.in[startPos.b:tk.b], nil
		}
		if newline(tk.p) != 0 {
			return nil, &atError{pos: startPos, err: ErrNewlineInDoubleQuoteString}
		}
		n, err := tk.char()
		if err != nil {
			return nil, err
		}
		tk.popBytes(n)
	}
}

// singleQuotedString tries to parse a singgle quoted string. It returns nil
// if there is no single quoted string in front of tk.p. Requires tk is not
// done an dk.p is not empty.
func (tk *tokenizer) singleQuotedString() ([]byte, *atError) {
	startPos := tk.pos
	if len(tk.p) == 0 || tk.p[0] != '\'' {
		return nil, nil
	}
	tk.popBytes(1)
	for {
		if len(tk.p) == 0 {
			return nil, &atError{pos: startPos, err: ErrUnclosedSingleQuoteString}
		}
		if tk.p[0] == '\\' && len(tk.p) >= 2 && tk.p[1] == '\'' {
			tk.popBytes(2)
			continue
		}
		if tk.p[0] == '\'' {
			tk.popBytes(1)
			return tk.in[startPos.b:tk.b], nil
		}
		if newline(tk.p) != 0 {
			return nil, &atError{pos: startPos, err: ErrNewlineInSingleQuoteString}
		}
		n, err := tk.char()
		if err != nil {
			return nil, err
		}
		tk.popBytes(n)
	}
}

// quotelessString include any valid characters until any of
// , { } [ ] : \n \r\n // /*, the end of input or an error is met.
// The quoteless string is right trimmed of whitespace characters.
// It return nil, nil, when the quoteles string is empty.
func (tk *tokenizer) quotelessString() ([]byte, *atError) {
	var stopByte = [256]byte{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 0, // 00  \n \r
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 10
		0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, // 20  # , /
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, // 30  :
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 40
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, // 50 [ ]
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 60
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, // 70 { }
	}
	startPos := tk.pos
	endIdx := startPos.b
	for {
		if len(tk.p) == 0 {
			break
		}
		if whitespace(tk.p) != 0 {
			tk.skipWhitespaces()
			continue
		}
		if stopByte[tk.p[0]] != 0 {
			if (tk.p[0] == '/' && len(tk.p) > 1 && (tk.p[1] == '/' || tk.p[1] == '*')) ||
				newline(tk.p) != 0 || (tk.p[0] != '\r' && tk.p[0] != '/') {
				// we met any of , { } [ ] # \n \r\n // /*
				break
			}
		}
		n, err := tk.char()
		if err != nil {
			return nil, err
		}
		tk.popBytes(n)
		endIdx = tk.b
	}
	if startPos.b == endIdx {
		return nil, nil
	}
	return tk.in[startPos.b:endIdx], nil
}

var tkTagTable = [256]tokenTag{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 00
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 10
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, tagComma, 0, 0, 0, // 20
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, tagColon, 0, 0, 0, 0, 0, // 30
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 40
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, tagOpenSquare, 0, tagCloseSquare, 0, 0, // 50
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 60
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, tagOpenBrace, 0, tagCloseBrace, 0, 0, // 70
}

// delimiter return the delimiter tag or tagUnknown.
func (tk *tokenizer) delimiter() tokenTag {
	tag := tkTagTable[tk.p[0]]
	if tag != tagUnknown {
		tk.popBytes(1)
	}
	return tag
}

// skipSpaces skips whitespaces, comments, newlines, and return
// error if any. Return nil if the end of input is reached.
func (tk *tokenizer) skipSpaces() *atError {
	var err *atError
	var ok bool
	for err == nil && len(tk.p) > 0 {
		tk.skipWhitespaces()
		if ok, err = tk.skipLineComment(); ok || err != nil {
			continue
		}
		if ok, err = tk.skipMultilineComment(); ok || err != nil {
			continue
		}
		if !tk.popNewline() {
			break
		}
	}
	return err
}

func matchingMarginLength(margin, line []byte) int {
	n := len(margin)
	if len(line) < len(margin) {
		n = len(line)
	}
	for i := 0; i < n; i++ {
		if line[i] != margin[i] {
			return i
		}
	}
	return n
}

func newlineSpecifier(p []byte) int {
	if p[0] == '\\' {
		if len(p) > 1 && p[1] == 'n' {
			return 2
		}
		if len(p) > 3 && p[1] == 'r' && p[2] == '\\' && p[3] == 'n' {
			return 4
		}
	}
	return 0
}

// return the index in p of the end of the margin.
func getMargin(p []byte) int {
	var b int
	for len(p) > 0 {
		if n := whitespace(p); n != 0 {
			p = p[n:]
			b += n
		} else {
			break
		}
	}
	return b
}

// multilineString test if e.p starts with a multiline string. If it is,
// it returns the multiline string including the margin and trailing `, and
// pops it from e.p. Otherwise it return nil. It returns a non nil slice of
// lenght 0 if an error occured.
func (tk *tokenizer) multilineString() ([]byte, *atError) {
	if len(tk.p) == 0 || tk.p[0] != '`' {
		return nil, nil
	}
	b := getMargin(tk.in[tk.s:tk.b]) + tk.s
	if b != tk.b {
		return nil, &atError{pos: pos{b: b, s: tk.s, l: tk.l}, err: ErrMarginMustBeWhitespaceOnly}
	}
	margin := tk.in[tk.s:b]
	startPos := tk.pos // for error reporting
	tk.popBytes(1)     // pops starting `
	tk.skipWhitespaces()
	if len(tk.p) == 0 {
		return nil, &atError{pos: startPos, err: ErrMissingNewlineSpecifier}
	}
	if n := newlineSpecifier(tk.p); n != 0 {
		tk.popBytes(n)
	} else {
		return nil, &atError{pos: startPos, err: ErrInvalidNewlineSpecifier}
	}
	tk.skipWhitespaces()
	if !tk.popNewline() {
		ok, err := tk.skipLineComment()
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, &atError{pos: startPos, err: ErrInvalidMultilineStart}
		}
	}
	if len(tk.p) == 0 {
		return nil, &atError{pos: startPos, err: ErrUnclosedMultiline}
	}
	n := matchingMarginLength(margin, tk.p)
	if n != len(margin) {
		return nil, &atError{pos: pos{b: tk.b + n, s: tk.s, l: tk.l}, err: ErrInvalidMarginChar}
	}
	tk.popBytes(n)
	for len(tk.p) > 0 {
		if tk.popNewline() {
			n := matchingMarginLength(margin, tk.p)
			if n != len(margin) {
				return nil, &atError{pos: pos{b: tk.b + n, s: tk.s, l: tk.l}, err: ErrInvalidMarginChar}
			}
			if n > 0 {
				tk.popBytes(n)
			}
			continue
		}
		if tk.p[0] < 0x20 {
			tk.popBytes(1)
			continue
		}
		if tk.p[0] == '`' {
			tk.popBytes(1)
			if len(tk.p) == 0 || tk.p[0] != '\\' {
				return tk.in[startPos.s:tk.b], nil // we reached the end of the multiline
			}
			continue
		}
		n, err := tk.char()
		if err != nil {
			return nil, err
		}
		tk.popBytes(n)
	}
	return nil, &atError{pos: startPos, err: ErrUnclosedMultiline}
}

// nextToken reads the next token. The token is accessed with the token()
// method. The token is set to an error if an error is detected or the
// end of input is reached.
func (tk *tokenizer) nextToken() {
	if tk.tk.tag == tagError {
		return
	}
	var tokenPos pos
	if err := tk.skipSpaces(); err != nil {
		tk.tk = token{tag: tagError, pos: err.pos, val: err.err}
	} else if tokenPos = tk.pos; len(tk.p) == 0 {
		tk.tk = token{tag: tagError, pos: tk.pos, val: ErrEndOfInput}
	} else if tag := tk.delimiter(); tag != tagUnknown {
		tk.tk = token{tag: tag, pos: tokenPos}
	} else if s, err := tk.doubleQuotedString(); err != nil || s != nil {
		if err != nil {
			tk.tk = token{tag: tagError, pos: err.pos, val: err.err}
		} else {
			tk.tk = token{tag: tagDoubleQuotedString, pos: tokenPos, val: s}
		}
	} else if s, err := tk.singleQuotedString(); err != nil || s != nil {
		if err != nil {
			tk.tk = token{tag: tagError, pos: err.pos, val: err.err}
		} else {
			tk.tk = token{tag: tagSingleQuotedString, pos: tokenPos, val: s}
		}
	} else if s, err := tk.multilineString(); err != nil || s != nil {
		if err != nil {
			tk.tk = token{tag: tagError, pos: err.pos, val: err.err}
		} else {
			tk.tk = token{tag: tagMultilineString, pos: tokenPos, val: s}
		}
	} else {
		s, err := tk.quotelessString()
		if err != nil {
			tk.tk = token{tag: tagError, pos: err.pos, val: err.err}
		} else {
			tk.tk = token{tag: tagQuotelessString, pos: tokenPos, val: s}
		}
	}
}
