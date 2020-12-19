package qjson

import (
	"fmt"
	"reflect"
)

// inspired by https://eli.thegreenplace.net/2010/01/02/top-down-operator-precedence-parsing

// operator precedence
// 3             w  d  h  m  s
// 2             *  /  %  <<  >>  &  &^
// 1             +  -  |  ^  ~
// 0
var precedenceTable = [256]byte{
	0, // tagUnknown
	0, // tagError
	0, // tagIntegerVal
	0, // tagDecimalVal
	1, // tagPlus
	1, // tagMinus
	2, // tagMultiplication
	2, // tagDivision
	1, // tagXor
	2, // tagAnd
	1, // tagOr
	1, // tagInverse
	2, // tagModulo
	0, // tagOpenParen
	0, // tagCloseParen
	3, // tagWeeks
	3, // tagDays
	3, // tagHours
	3, // tagMinutes
	3, // tagSeconds
}

const highestPrecedence = 3

// A nudXXX function returns the result of evaluating the expression at the current
// location. It returns nil when the end is reached or an error occured, otherwise
// it returns an int or a float64.
type nudFunc func(tk *numTokenizer, t numToken) interface{}

// A ledXXX function is an n-ary operation function. It is given its left operand
// as an argument and may get the right operand if needed from a call to expression.
// It returns nil when the end is reached or an error occured, otherwise it returns
// an int or a float64.
type ledFunc func(tk *numTokenizer, t numToken, left interface{}) interface{}

var nudTable = [256]nudFunc{}

var ledTable = [256]ledFunc{}

// to get rid of initialization cycle error
func init() {
	nudTable = [256]nudFunc{
		nil,           // tagUnknown
		nil,           // tagError
		nudValue,      // tagIntegerVal
		nudValue,      // tagDecimalVal
		nudPlus,       // tagPlus
		nudMinus,      // tagMinus
		nil,           // tagMultiplication
		nil,           // tagDivision
		nil,           // tagXor
		nil,           // tagAnd
		nil,           // tagOr
		nudInverse,    // tagInverse
		nil,           // tagModulo
		nudOpenParen,  // tagOpenParen
		nudCloseParen, // tagCloseParen
		nil,           // tagWeeks
		nil,           // tagDays
		nil,           // tagHours
		nil,           // tagMinutes
		nil,           // tagSeconds
	}
	ledTable = [256]ledFunc{
		nil,               // tagUnknown
		nil,               // tagError
		nil,               // tagIntegerVal
		nil,               // tagDecimalVal
		ledPlus,           // tagPlus
		ledMinus,          // tagMinus
		ledMultiplication, // tagMultiplication
		ledDivision,       // tagDivision
		ledXor,            // tagXor
		ledAnd,            // tagAnd
		ledOr,             // tagOr
		nil,               // tagInverse
		ledModulo,         // tagModulo
		nil,               // tagOpenParen
		nil,               // tagCloseParen
		ledWeeks,          // tagWeeks
		ledDays,           // tagDays
		ledHours,          // tagHours
		ledMinutes,        // tagMinutes
		ledSeconds,        // tagSeconds
	}
}

func (tk *numTokenizer) nud(t numToken) interface{} {
	if f := nudTable[t.tag]; f != nil {
		return f(tk, t)
	}
	tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
	return nil
}

func (tk *numTokenizer) led(t numToken, left interface{}) interface{} {
	if f := ledTable[t.tag]; f != nil {
		return f(tk, t, left)
	}
	tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
	return nil
}

// expression evaluates the expression at the current token position.
// On return, the current token will be the first token after the evaluated
// expression.
// It returns nil when the end of input has been reached, an exprError when
// an error condition has been detected, or the result of the expression
// evaluation.
func (tk *numTokenizer) expression(rbp byte) interface{} {
	t := tk.token()
	if tk.done() {
		return nil
	}
	tk.nextToken()
	left := tk.nud(t)
	for left != nil && rbp < precedenceTable[tk.token().tag] {
		t = tk.token()
		tk.nextToken()
		left = tk.led(t, left)
	}
	return left
}

// evalNumberExpression evaluates the expression in input and
// return the resulting value, otherwise reture the error and
// its index in the input.
func evalNumberExpression(input []byte) (float64, int, error) {
	var tk numTokenizer
	tk.init(input)
	tk.nextToken()
	res := tk.expression(0)
	if tk.tk.tag == tagError {
		if tk.tk.val.(error) != ErrEndOfInput {
			return 0, tk.tk.pos, tk.tk.val.(error)
		}
	} else {
		if tk.tk.tag == tagCloseParen {
			return 0, tk.tk.pos, ErrUnopenedParenthesis
		}
		return 0, tk.tk.pos, ErrInvalidNumericExpression
	}
	switch res.(type) {
	case int:
		return float64(res.(int)), 0, nil
	case float64:
		return res.(float64), 0, nil
	}
	return 0, tk.tk.pos, tk.tk.val.(error)
}

// normalizeTypes ensures that v1 anv v2 are both int, otherwise cast both to float64.
// Requires that v1 anv v2 are int or float64.
func normalizeTypes(v1 interface{}, v2 interface{}) (interface{}, interface{}) {
	if v1Int, ok := v1.(int); ok {
		if v2Float, ok := v2.(float64); ok {
			return float64(v1Int), v2Float
		}
	} else if v1Float, ok := v1.(float64); ok {
		if v2Int, ok := v2.(int); ok {
			return v1Float, float64(v2Int)
		}
	}
	return v1, v2
}

func nudValue(tk *numTokenizer, t numToken) interface{} {
	return t.val
}

func nudPlus(tk *numTokenizer, t numToken) interface{} {
	right := tk.expression(highestPrecedence + 1)
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	return right
}

func ledPlus(tk *numTokenizer, t numToken, left interface{}) interface{} {
	right := tk.expression(precedenceTable[tagPlus])
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	left, right = normalizeTypes(left, right)
	if x, ok := left.(int); ok {
		return x + right.(int)
	}
	return left.(float64) + right.(float64)
}

func nudMinus(tk *numTokenizer, t numToken) interface{} {
	right := tk.expression(highestPrecedence + 1)
	switch right.(type) {
	case nil:
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
	case int:
		return -right.(int)
	case float64:
		return -right.(float64)
	}
	return right
}

func ledMinus(tk *numTokenizer, t numToken, left interface{}) interface{} {
	right := tk.expression(precedenceTable[tagMinus])
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	left, right = normalizeTypes(left, right)
	if x, ok := left.(int); ok {
		return x - right.(int)
	}
	return left.(float64) - right.(float64)
}

func ledMultiplication(tk *numTokenizer, t numToken, left interface{}) interface{} {
	right := tk.expression(precedenceTable[tagMultiplication])
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	left, right = normalizeTypes(left, right)
	if x, ok := left.(int); ok {
		return x * right.(int)
	}
	return left.(float64) * right.(float64)
}

func ledDivision(tk *numTokenizer, t numToken, left interface{}) interface{} {
	right := tk.expression(precedenceTable[tagDivision])
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	left, right = normalizeTypes(left, right)
	if x1, ok := left.(int); ok {
		x2 := right.(int)
		if x2 == 0 {
			tk.setErrorAndPos(ErrDivisionByZero, t.pos)
			return nil
		}
		return x1 / right.(int)
	}
	x2 := right.(float64)
	if x2 == 0 {
		tk.setErrorAndPos(ErrDivisionByZero, t.pos)
		return nil
	}
	return left.(float64) / right.(float64)
}

func nudOpenParen(tk *numTokenizer, t numToken) interface{} {
	right := tk.expression(precedenceTable[tagOpenParen])
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	if tk.token().tag != tagCloseParen {
		tk.setErrorAndPos(ErrUnclosedParenthesis, t.pos)
		return nil
	}
	tk.nextToken()
	return right
}

func nudCloseParen(tk *numTokenizer, t numToken) interface{} {
	tk.setErrorAndPos(ErrUnopenedParenthesis, t.pos)
	return nil
}

func nudInverse(tk *numTokenizer, t numToken) interface{} {
	right := tk.expression(highestPrecedence + 1)
	switch right.(type) {
	case nil:
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
	case int:
		return ^right.(int)
	case float64:
		tk.setErrorAndPos(ErrOperandMustBeInteger, t.pos)
		return nil
	}
	return right
}

func ledModulo(tk *numTokenizer, t numToken, left interface{}) interface{} {
	right := tk.expression(precedenceTable[tagModulo])
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	left, right = normalizeTypes(left, right)
	if x1, ok := left.(int); ok {
		x2 := right.(int)
		if x2 == 0 {
			tk.setErrorAndPos(ErrDivisionByZero, t.pos)
			return nil
		}
		return x1 % x2
	}
	tk.setErrorAndPos(ErrOperandsMustBeInteger, t.pos)
	return nil
}

func ledAnd(tk *numTokenizer, t numToken, left interface{}) interface{} {
	right := tk.expression(precedenceTable[tagAnd])
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	left, right = normalizeTypes(left, right)
	if x, ok := left.(int); ok {
		return x & right.(int)
	}
	tk.setErrorAndPos(ErrOperandsMustBeInteger, t.pos)
	return nil
}

func ledOr(tk *numTokenizer, t numToken, left interface{}) interface{} {
	right := tk.expression(precedenceTable[tagOr])
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	left, right = normalizeTypes(left, right)
	if x, ok := left.(int); ok {
		return x | right.(int)
	}
	tk.setErrorAndPos(ErrOperandsMustBeInteger, t.pos)
	return nil
}

func ledXor(tk *numTokenizer, t numToken, left interface{}) interface{} {
	right := tk.expression(precedenceTable[tagXor])
	if right == nil {
		if tk.tk.val.(error) == ErrEndOfInput {
			tk.setErrorAndPos(ErrInvalidNumericExpression, t.pos)
		}
		return nil
	}
	left, right = normalizeTypes(left, right)
	if x, ok := left.(int); ok {
		return x ^ right.(int)
	}
	tk.setErrorAndPos(ErrOperandsMustBeInteger, t.pos)
	return nil
}

func toFloat64(v interface{}) float64 {
	switch v.(type) {
	case int:
		return float64(v.(int))
	case float64:
		return v.(float64)
	case nil:
		panic("invalid nil value")
	default:
		panic(fmt.Sprintf("invalid type %v", reflect.TypeOf(v)))
	}
}

func ledWeeks(tk *numTokenizer, t numToken, left interface{}) interface{} {
	const duration float64 = 3600 * 24 * 7
	right := tk.expression(precedenceTable[tagWeeks])
	if right == nil { // right hand operand is optional
		if tk.tk.val.(error) == ErrEndOfInput {
			return toFloat64(left) * duration
		}
		return nil
	}
	return toFloat64(left)*duration + toFloat64(right)
}

func ledDays(tk *numTokenizer, t numToken, left interface{}) interface{} {
	const duration float64 = 3600 * 24
	right := tk.expression(precedenceTable[tagDays])
	if right == nil { // right hand operand is optional
		if tk.tk.val.(error) == ErrEndOfInput {
			return toFloat64(left) * duration
		}
		return nil
	}
	return toFloat64(left)*duration + toFloat64(right)
}

func ledHours(tk *numTokenizer, t numToken, left interface{}) interface{} {
	const duration float64 = 3600
	right := tk.expression(precedenceTable[tagDays])
	if right == nil { // right hand operand is optional
		if tk.tk.val.(error) == ErrEndOfInput {
			return toFloat64(left) * duration
		}
		return nil
	}
	return toFloat64(left)*duration + toFloat64(right)
}

func ledMinutes(tk *numTokenizer, t numToken, left interface{}) interface{} {
	const duration float64 = 60
	right := tk.expression(precedenceTable[tagDays])
	if right == nil { // right hand operand is optional
		if tk.tk.val.(error) == ErrEndOfInput {
			return toFloat64(left) * duration
		}
		return nil
	}
	return toFloat64(left)*duration + toFloat64(right)
}

func ledSeconds(tk *numTokenizer, t numToken, left interface{}) interface{} {
	right := tk.expression(precedenceTable[tagDays])
	if right == nil { // right hand operand is optional
		if tk.tk.val.(error) == ErrEndOfInput {
			return toFloat64(left)
		}
		return nil
	}
	return toFloat64(left) + toFloat64(right)
}
