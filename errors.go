package qjson

import "fmt"

type atError struct {
	pos
	err error
}

func (e atError) Error() string {
	return fmt.Sprintf("%s %v", e.err, e.pos)
}

// Error is a constant error type
type Error string

func (e Error) Error() string { return string(e) }

var errMap = map[Error]string{
	ErrEndOfInput:                 "ErrEndOfInput",
	ErrInvalidChar:                "ErrInvalidChar",
	ErrTruncatedChar:              "ErrTruncatedChar",
	ErrSyntaxError:                "ErrSyntaxError",
	ErrUnclosedDoubleQuoteString:  "ErrUnclosedDoubleQuoteString",
	ErrUnclosedSingleQuoteString:  "ErrUnclosedDoubleQuoteString",
	ErrUnclosedSlashStarComment:   "ErrUnclosedSlashStarComment",
	ErrNewlineInDoubleQuoteString: "ErrNewlineInDoubleQuoteString",
	ErrNewlineInSingleQuoteString: "ErrNewlineInSingleQuoteString",
	ErrExpectStringIdentifier:     "ErrExpectStringIdentifier",
	ErrExpectColon:                "ErrExpectColon",
	ErrInvalidValueType:           "ErrInvalidValueType",
	ErrMaxObjectArrayDepth:        "ErrMaxObjectArrayDepth",
	ErrUnclosedObject:             "ErrUnclosedObject",
	ErrUnclosedArray:              "ErrUnclosedArray",
	ErrUnexpectedEndOfInput:       "ErrUnexpectedEndOfInput",
	ErrExpectIdentifierAfterComma: "ErrExpectIdentifierAfterComma",
	ErrExpectValueAfterComma:      "ErrExpectValueAfterComma",
	ErrInvalidEscapeSequence:      "ErrInvalidEscapeSequence",
	ErrInvalidNumericExpression:   "ErrInvalidNumericExpression",
	ErrInvalidBinaryNumber:        "ErrInvalidBinaryNumber",
	ErrInvalidHexadecimalNumber:   "ErrInvalidHexadecimalNumber",
	ErrInvalidOctalNumber:         "ErrInvalidOctalNumber",
	ErrInvalidIntegerNumber:       "ErrInvalidIntegerNumber",
	ErrInvalidDecimalNumber:       "ErrInvalidDecimalNumber",
	ErrNumberOverflow:             "ErrNumberOverflow",
	ErrUnopenedParenthesis:        "ErrUnopenedParenthesis",
	ErrDivisionByZero:             "ErrDivisionByZero",
	ErrUnclosedParenthesis:        "ErrUnclosedParenthesis",
	ErrOperandMustBeInteger:       "ErrOperandMustBeInteger",
	ErrOperandsMustBeInteger:      "ErrOperandsMustBeInteger",
	ErrMarginMustBeWhitespaceOnly: "ErrMarginMustBeWhitespaceOnly",
	ErrUnclosedMultiline:          "ErrUnclosedMultiline",
	ErrInvalidMarginChar:          "ErrInvalidMarginChar",
	ErrMissingNewlineSpecifier:    "ErrMissingNewlineSpecifier",
	ErrInvalidNewlineSpecifier:    "ErrInvalidNewlineSpecifier",
	ErrInvalidMultilineStart:      "ErrInvalidMultilineStart",
	ErrUnexpectedCloseBrace:       "ErrUnexpectedCloseBrace",
	ErrUnexpectedCloseSquare:      "ErrUnexpectedCloseSquare",
}

func errStr(e error) string {
	if e == nil {
		return "nil"
	}
	if err, ok := e.(Error); ok {
		if v, ok := errMap[err]; ok {
			return v
		}
		return fmt.Sprintf("Err??? (%s))", e.Error())
	}
	return fmt.Sprintf("error: %v", e)
}

// ErrEndOfInput is returned when the end of input is reached.
const ErrEndOfInput = Error("end of input")

// ErrInvalidChar is returned when an invalid rune is found in the input stream.
const ErrInvalidChar = Error("invalid character")

// ErrTruncatedChar occurs when the last utf8 char of the input is truncated.
const ErrTruncatedChar = Error("last utf8 char is truncated")

// ErrSyntaxError is returned when a non-expected token is met.
const ErrSyntaxError = Error("syntax error")

// ErrUnclosedDoubleQuoteString is returned when a double quote string is unclosed.
const ErrUnclosedDoubleQuoteString = Error("unclosed double quote string")

// ErrUnclosedSingleQuoteString is returned when a single quote string is unclosed.
const ErrUnclosedSingleQuoteString = Error("unclosed single quote string")

// ErrUnclosedSlashStarComment is returned when the end of input is found in a /*...*/
const ErrUnclosedSlashStarComment = Error("unclosed /*...*/ comment")

// ErrNewlineInDoubleQuoteString is returned when a newline is met in a double quoted string.
const ErrNewlineInDoubleQuoteString = Error("newline in double quoted string")

// ErrNewlineInSingleQuoteString is returned when a newline is met in a single quoted string.
const ErrNewlineInSingleQuoteString = Error("newline in single quoted string")

// ErrExpectStringIdentifier is return when an invalid identifier type is found.
const ErrExpectStringIdentifier = Error("expect string identifier")

// ErrExpectColon is return when an a colon is not found after the identifier.
const ErrExpectColon = Error("expect a colon")

// ErrInvalidValueType is return when an invalid value type is found.
const ErrInvalidValueType = Error("invalid value type")

// ErrMaxObjectArrayDepth is return when the number of encapsuled object has reached a limit.
const ErrMaxObjectArrayDepth = Error("too many object or array encapsulations")

// ErrUnclosedObject is returned when the end of input is met before the object was closed.
const ErrUnclosedObject = Error("unclosed object")

// ErrUnclosedArray is returned when the end of input is met before the array was closed.
const ErrUnclosedArray = Error("unclosed array")

// ErrUnexpectedEndOfInput is returned when the end of input is met in an unexpected location.
const ErrUnexpectedEndOfInput = Error("unexpected end of input")

// ErrExpectIdentifierAfterComma is returned when a comma is at end of input or object.
const ErrExpectIdentifierAfterComma = Error("expect identifier after comma")

// ErrExpectValueAfterComma is returned when a comma is at end of input or array.
const ErrExpectValueAfterComma = Error("expect value after comma")

// ErrInvalidEscapeSequence is returned when an invalid escape sequence is found in a string.
const ErrInvalidEscapeSequence = Error("invalid escape squence")

// ErrInvalidNumericExpression is returned when an unrecognized text is found in a numeric expression.
const ErrInvalidNumericExpression = Error("invalid numeric expression")

// ErrInvalidBinaryNumber is returned when the tokenizer met an invalid binary number.
const ErrInvalidBinaryNumber = Error("invalid binary number")

// ErrInvalidHexadecimalNumber is returned when the tokenizer met an invalid hexadecimal number.
const ErrInvalidHexadecimalNumber = Error("invalid hexadecimal number")

// ErrInvalidOctalNumber is returned when the tokenizer met an invalid octal number.
const ErrInvalidOctalNumber = Error("invalid octal number")

// ErrInvalidIntegerNumber is returned when the tokenizer met an invalid integer number.
const ErrInvalidIntegerNumber = Error("invalid integer number")

// ErrInvalidDecimalNumber is returned when the tokenizer met an invalid decimal number.
const ErrInvalidDecimalNumber = Error("invalid decimal number")

// ErrNumberOverflow is retured when a number would overflow a float64 representation.
const ErrNumberOverflow = Error("number overflow")

// ErrUnopenedParenthesis is returned when a close parenthesis has no matching open parenthesis.
const ErrUnopenedParenthesis = Error("missing open parenthesis")

// ErrDivisionByZero is returned when there is a division by zero in an expression.
const ErrDivisionByZero = Error("division by zero")

// ErrUnclosedParenthesis is returned when an open parenthesis has no matching close parenthesis.
const ErrUnclosedParenthesis = Error("missing close parenthesis")

// ErrOperandMustBeInteger is returned when a binary operation is attempted on a float.
const ErrOperandMustBeInteger = Error("operand must be integer")

// ErrOperandsMustBeInteger is returned when a binary or a modulo operation is attempted on a float.
const ErrOperandsMustBeInteger = Error("operands must be integer")

// ErrMarginMustBeWhitespaceOnly is returned when a non whitespace character is found in front of `.
const ErrMarginMustBeWhitespaceOnly = Error("multiline margin must contain only whitespaces")

// ErrUnclosedMultiline is returned when the end of input is met before the ending `.
const ErrUnclosedMultiline = Error("unclosed multiline")

// ErrInvalidMarginChar is returned when the margin doesn’t match the start of multiline margin.
const ErrInvalidMarginChar = Error("invalid margin character")

// ErrMissingNewlineSpecifier is returned when the starting ` of a multiline is not followed by \n or \r\n.
const ErrMissingNewlineSpecifier = Error("missing \\n or \\r\\n after multiline start")

// ErrInvalidNewlineSpecifier is returned when the starting ` of a multiline is not followed by \n or \r\n.
const ErrInvalidNewlineSpecifier = Error("expect \\n or \\r\\n after `")

// ErrInvalidMultilineStart is returned when non whitespace or line comments follow the opening `.
const ErrInvalidMultilineStart = Error("invalid multiline start line")

// ErrUnexpectedCloseBrace is return when } is met where a value is expected.
const ErrUnexpectedCloseBrace = Error("unexpected }")

// ErrUnexpectedCloseSquare is return when } is met where a value is expected.
const ErrUnexpectedCloseSquare = Error("unexpected ]")
