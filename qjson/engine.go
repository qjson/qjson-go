package qjson

import (
	"bytes"
	"fmt"
	"strconv"
)

// Decode accept QJSON text as input and return a JSON text or return an error.
func Decode(input []byte) ([]byte, error) {
	if input == nil {
		return []byte("{}"), nil
	}
	var e engine
	e.init(input)
	e.members()
	if e.token().tag == tagCloseBrace {
		e.setError(ErrUnexpectedCloseBrace)
	}
	t := e.token()
	if t.tag == tagError && t.val.(error) != ErrEndOfInput {
		return nil, fmt.Errorf("%w at line %d col %d", t.val.(error), t.pos.l+1, column(e.in[t.pos.s:t.pos.b])+1)
	}
	return e.out.Bytes(), nil
}

var maxDepth = 200

// engine to convert QJSON to JSON.
type engine struct {
	tokenizer
	depth int
	out   bytes.Buffer
}

func (e *engine) init(input []byte) {
	e.tokenizer.init(input)
	e.out.Reset()
	e.depth = 0
	e.nextToken()
}

func (e *engine) done() bool {
	return e.tk.tag == tagError
}
func (e *engine) setError(err error) {
	e.setErrorAndPos(err, e.tk.pos)
}

func (e *engine) setErrorAndPos(err error, p pos) {
	e.tk = token{tag: tagError, pos: p, val: err}
}

// value process a value. If an error occurred it returns with the error set,
// otherwise calls nextToken() and return its result.
func (e *engine) value() bool {
	switch e.tk.tag {
	case tagCloseSquare:
		e.setError(ErrUnexpectedCloseSquare)
		return false
	case tagCloseBrace:
		e.setError(ErrUnexpectedCloseBrace)
		return false
	case tagDoubleQuotedString:
		e.outputDoubleQuotedString()
	case tagSingleQuotedString:
		e.outputSingleQuotedString()
	case tagMultilineString:
		e.outputMultilineString()
	case tagQuotelessString:
		val := e.tk.val.([]byte)
		if str := isLiteralValue(val); str != "" {
			e.out.WriteString(str)
		} else if isNumberExpr(val) {
			res, pos, err := evalNumberExpression(val)
			if err != nil {
				p := e.tk.pos
				p.b += pos
				e.setErrorAndPos(err, p)
				return true
			}
			e.out.WriteString(strconv.FormatFloat(res, 'g', 16, 64))
		} else {
			e.outputQuotelessString()
		}
	case tagOpenBrace:
		startPos := e.tk.pos
		e.nextToken()
		if e.done() {
			if e.tk.val.(error) == ErrEndOfInput {
				e.setErrorAndPos(ErrUnclosedObject, startPos)
			}
			return true
		}
		if e.depth == maxDepth {
			e.setError(ErrMaxObjectArrayDepth)
			return true
		}
		e.depth++
		if e.members() {
			if e.tk.val.(error) == ErrEndOfInput {
				e.setErrorAndPos(ErrUnclosedObject, startPos)
			}
			return true
		}
		e.depth--
	case tagOpenSquare:
		e.nextToken()
		if e.done() {
			if e.tk.val.(error) == ErrEndOfInput {
				e.setError(ErrUnclosedArray)
			}
			return true
		}
		startPos := e.tk.pos
		if e.depth == maxDepth {
			e.setError(ErrMaxObjectArrayDepth)
			return true
		}
		e.depth++
		if e.values() {
			if e.tk.val.(error) == ErrEndOfInput {
				e.setErrorAndPos(ErrUnclosedArray, startPos)
			}
			return true
		}
		e.depth--
	default:
		e.setError(ErrSyntaxError)
		//		e.setError(fmt.Errorf("expected value, got %v", e.tk))
		return false
	}
	e.nextToken()
	return e.done()
}

// values process 0 or more values and pops the ending ]. Return done().
func (e *engine) values() bool {
	var notFirst bool
	e.out.WriteByte('[')
	for !e.done() && e.tk.tag != tagCloseSquare {
		if notFirst {
			e.out.WriteByte(',')
			if e.tk.tag == tagComma {
				e.nextToken()
				if e.done() {
					if e.tk.val.(error) == ErrEndOfInput {
						e.setError(ErrExpectValueAfterComma)
					}
					break
				}
				if e.tk.tag == tagCloseBrace || e.tk.tag == tagCloseSquare {
					e.setError(ErrExpectValueAfterComma)
					break
				}
			}
		} else {
			notFirst = true
		}
		if e.value() {
			break
		}
	}
	e.out.WriteByte(']')
	return e.done()
}

func (e *engine) member() bool {
	switch e.tk.tag {
	case tagCloseSquare:
		e.setError(ErrUnexpectedCloseSquare)
		return false
	case tagDoubleQuotedString:
		e.outputDoubleQuotedString()
	case tagSingleQuotedString:
		e.outputSingleQuotedString()
	case tagQuotelessString:
		e.outputQuotelessString()
	default:
		e.setError(ErrExpectStringIdentifier)
	}
	e.nextToken()
	if e.done() {
		if e.tk.val.(error) == ErrEndOfInput {
			e.setError(ErrUnexpectedEndOfInput)
		}
		return true
	}
	if e.tk.tag != tagColon {
		e.setError(ErrExpectColon)
		return true
	}
	e.out.WriteByte(':')
	e.nextToken()
	if e.done() {
		if e.tk.val.(error) == ErrEndOfInput {
			e.setError(ErrUnexpectedEndOfInput)
		}
		return true
	}
	return e.value()
}

// values process 0 or more members (identifiers : value) and pops the ending }. Return done().
func (e *engine) members() bool {
	var notFirst bool
	e.out.WriteByte('{')
	for !e.done() && e.tk.tag != tagCloseBrace {
		if notFirst {
			e.out.WriteByte(',')
			if e.tk.tag == tagComma {
				e.nextToken()
				if e.done() {
					if e.tk.val.(error) == ErrEndOfInput {
						e.setError(ErrExpectIdentifierAfterComma)
					}
					break
				}
				if e.tk.tag == tagCloseBrace || e.tk.tag == tagCloseSquare {
					e.setError(ErrExpectIdentifierAfterComma)
					break
				}
			}
		} else {
			notFirst = true
		}
		if e.member() {
			break
		}
	}
	e.out.WriteByte('}')
	return e.done()
}

// isNullValue return true when p is equal to "null", "Null", "NULL".
func isLiteralValue(p []byte) string {
	switch len(p) {
	case 5:
		if (p[0] == 'f' || p[0] == 'F') &&
			((p[1] == 'a' && p[2] == 'l' && p[3] == 's' && p[4] == 'e') ||
				(p[1] == 'A' && p[2] == 'L' && p[3] == 'S' && p[4] == 'E')) {
			return "false"
		}
	case 4:
		if (p[0] == 'n' || p[0] == 'N') &&
			((p[1] == 'u' && p[2] == 'l' && p[3] == 'l') || (p[1] == 'U' && p[2] == 'L' && p[3] == 'L')) {
			return "null"
		}
		if (p[0] == 't' || p[0] == 'T') &&
			((p[1] == 'r' && p[2] == 'u' && p[3] == 'e') || (p[1] == 'R' && p[2] == 'U' && p[3] == 'E')) {
			return "true"
		}
	case 3:
		if (p[0] == 'y' || p[0] == 'Y') &&
			((p[1] == 'e' && p[2] == 's') || (p[1] == 'E' && p[2] == 'S')) {
			return "true"
		}
		if (p[0] == 'o' || p[0] == 'O') &&
			((p[1] == 'f' && p[2] == 'f') || (p[1] == 'F' && p[2] == 'F')) {
			return "false"
		}
	case 2:
		if (p[0] == 'o' || p[0] == 'O') && (p[1] == 'n' || p[1] == 'N') {
			return "true"
		}
		if (p[0] == 'n' || p[0] == 'N') && (p[1] == 'o' || p[1] == 'O') {
			return "false"
		}
	}
	return ""
}

func (e *engine) outputDoubleQuotedString() {
	str := e.tk.val.([]byte)
	e.out.WriteByte('"')
	for i := 1; i < len(str)-1; i++ {
		switch str[i] {
		case '/':
			if str[i-1] == '<' {
				e.out.WriteByte('\\')
			}
		case '\t':
			e.out.WriteByte('\\')
			e.out.WriteByte('t')
			continue
		case '\\':
			c := str[i+1]
			if c != 't' && c != 'n' && c != 'r' && c != 'f' && c != 'b' && c != '/' && c != '\\' && c != '"' &&
				!(c == 'u' && len(str) >= i+6 && isHexDigit(str[i+2]) && isHexDigit(str[i+3]) && isHexDigit(str[i+5]) && isHexDigit(str[i+5])) {
				p := e.tk.pos
				p.b += i
				e.setErrorAndPos(ErrInvalidEscapeSequence, p)
				return
			}
		}
		e.out.WriteByte(str[i])
	}
	e.out.WriteByte('"')
}

func (e *engine) outputSingleQuotedString() {
	str := e.tk.val.([]byte)
	e.out.WriteByte('"')
	for i := 1; i < len(str)-1; i++ {
		switch str[i] {
		case '/':
			if str[i-1] == '<' {
				e.out.WriteByte('\\')
			}
		case '\t':
			e.out.WriteByte('\\')
			e.out.WriteByte('t')
			continue
		case '\\':
			c := str[i+1]
			if c != 't' && c != 'n' && c != 'r' && c != 'f' && c != 'b' && c != '/' && c != '\\' && c != '\'' &&
				!(c == 'u' && len(str) >= i+6 && isHexDigit(str[i+2]) && isHexDigit(str[i+3]) && isHexDigit(str[i+5]) && isHexDigit(str[i+5])) {
				p := e.tk.pos
				p.b += i
				e.setErrorAndPos(ErrInvalidEscapeSequence, p)
				return
			}
			if c == '\'' {
				continue
			}
		case '"':
			e.out.WriteByte('\\')
		}
		e.out.WriteByte(str[i])
	}
	e.out.WriteByte('"')
}

func (e *engine) outputQuotelessString() {
	str := e.tk.val.([]byte)
	e.out.WriteByte('"')
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case '"':
			e.out.WriteByte('\\')
		case '\t':
			e.out.WriteByte('\\')
			e.out.WriteByte('t')
			continue
		case '/':
			if i > 0 && str[i-1] == '<' {
				e.out.WriteByte('\\')
			}
		case '\\':
			e.out.WriteByte('\\')
		}
		e.out.WriteByte(str[i])
	}
	e.out.WriteByte('"')
}

func (e *engine) outputMultilineString() {
	str := e.tk.val.([]byte)
	var p int
	for str[p] != '`' {
		p++
	}
	margin := str[:p]
	str = str[p+1:]
	for n := whitespace(str); n > 0; n = whitespace(str) {
		str = str[n:]
	}
	str = str[1:] // skip \
	var nl []byte
	if str[0] == 'n' {
		nl = []byte("\\n")
		str = str[1:]
	} else {
		nl = []byte("\\r\\n")
		str = str[3:]
	}
	for str[0] != '\n' {
		str = str[1:]
	}
	// skip \n with margin of first line, and drop closing `
	str = str[1+len(margin) : len(str)-1]
	e.out.WriteByte('"')

	for len(str) > 0 {
		if n := newline(str); n != 0 {
			e.out.Write(nl)
			str = str[n+len(margin):]
		} else if str[0] < 0x20 {
			switch str[0] {
			case '\b':
				e.out.WriteString("\\b")
			case '\t':
				e.out.WriteString("\\t")
			case '\r':
				e.out.WriteString("\\r")
			case '\f':
				e.out.WriteString("\\f")
			default:
				e.out.WriteString(fmt.Sprintf("\\u00%0X", str[0]))
			}
			str = str[1:]
		} else if str[0] == '<' {
			e.out.WriteByte('<')
			if len(str) > 1 && str[1] == '/' {
				e.out.WriteByte('\\')
			}
			str = str[1:]
		} else if str[0] == '"' {
			e.out.WriteByte('\\')
			e.out.WriteByte('"')
			str = str[1:]
		} else if str[0] == '`' && len(str) > 1 && str[1] == '\\' {
			e.out.WriteByte('`')
			str = str[2:]
		} else if str[0] == '\\' {
			e.out.WriteByte('\\')
			e.out.WriteByte('\\')
			str = str[1:]
		} else {
			e.out.WriteByte(str[0])
			str = str[1:]
		}
	}
	e.out.WriteByte('"')
}
