package lexer

import (
	"unicode"

	"github.com/qadav-lang/qadav/internal/token"
)

// Lexer scans Qadav source code into a stream of tokens.
type Lexer struct {
	input    []rune
	position int
	line     int
	col      int
	file     string
}

// New creates and initializes a new Lexer for the provided input.
func New(input string, filename string) *Lexer {
	return &Lexer{
		input: []rune(input),
		line:  1,
		col:   1,
		file:  filename,
	}
}

// NextToken scans and returns the next token from the input stream.
func (l *Lexer) NextToken() token.Token {
	for {
		l.skipWhitespace()

		if l.position >= len(l.input) {
			return token.Token{
				Type:    token.EOF,
				Literal: "",
				Line:    l.line,
				Col:     l.col,
				File:    l.file,
			}
		}

		// Comments: // or /* */
		if l.peek() == '/' && l.peekNext() == '/' {
			l.skipLineComment()
			continue
		}
		if l.peek() == '/' && l.peekNext() == '*' {
			l.skipBlockComment()
			continue
		}

		break
	}

	if l.position >= len(l.input) {
		return token.Token{
			Type:    token.EOF,
			Literal: "",
			Line:    l.line,
			Col:     l.col,
			File:    l.file,
		}
	}

	startLine := l.line
	startCol := l.col
	ch := l.peek()

	// 1. Check for LaTeX fallback: \ followed by letter (e.g. \oplus, \bind)
	if ch == '\\' && isLetter(l.peekNext()) {
		return l.readFallback()
	}

	// 1b. Check for legacy double pipe: ||
	if l.peek() == '|' && l.peekNext() == '|' {
		l.advance()
		l.advance()
		return token.Token{
			Type:    token.ILLEGAL,
			Literal: "||",
			Line:    startLine,
			Col:     startCol,
			File:    l.file,
		}
	}

	// 2. Check 2-rune compound operators (longest match first)
	if l.position+1 < len(l.input) {
		twoRune := string(l.input[l.position : l.position+2])
		if tokType, ok := token.LookupUnicode(twoRune); ok {
			l.advance()
			l.advance()
			return token.Token{
				Type:    tokType,
				Literal: twoRune,
				Line:    startLine,
				Col:     startCol,
				File:    l.file,
			}
		}
	}

	// 3. Check 1-rune Unicode operators and quantum/math symbols
	oneRune := string(ch)
	if tokType, ok := token.LookupUnicode(oneRune); ok {
		l.advance()
		return token.Token{
			Type:    tokType,
			Literal: oneRune,
			Line:    startLine,
			Col:     startCol,
			File:    l.file,
		}
	}

	// 4. Strings: double quotes " or triple quotes """
	if ch == '"' {
		return l.readString()
	}

	// 5. Numbers: integers, floats, hex, binary, octal, complex
	if isDigit(ch) {
		return l.readNumber()
	}

	// 6. Identifiers and Keywords
	if isIdentStart(ch) {
		return l.readIdent()
	}

	// 6b. Annotation: @gpu
	if ch == '@' {
		l.advance()
		if l.position < len(l.input) && isIdentStart(l.peek()) {
			start := l.position
			for l.position < len(l.input) && isIdentPart(l.peek()) {
				l.advance()
			}
			id := string(l.input[start:l.position])
			if id == "gpu" {
				return token.Token{
					Type:    token.AT_GPU,
					Literal: "@gpu",
					Line:    startLine,
					Col:     startCol,
					File:    l.file,
				}
			}
			if id == "multi_device" {
				return token.Token{
					Type:    token.MULTI_DEVICE,
					Literal: "@multi_device",
					Line:    startLine,
					Col:     startCol,
					File:    l.file,
				}
			}
			if id == "distributed" {
				return token.Token{
					Type:    token.DIST_INIT,
					Literal: "@distributed",
					Line:    startLine,
					Col:     startCol,
					File:    l.file,
				}
			}
			if id == "qhal_backend" {
				return token.Token{
					Type:    token.QHAL_BACKEND,
					Literal: "@qhal_backend",
					Line:    startLine,
					Col:     startCol,
					File:    l.file,
				}
			}
			return token.Token{
				Type:    token.ILLEGAL,
				Literal: "@" + id,
				Line:    startLine,
				Col:     startCol,
				File:    l.file,
			}
		}
		return token.Token{
			Type:    token.ILLEGAL,
			Literal: "@",
			Line:    startLine,
			Col:     startCol,
			File:    l.file,
		}
	}

	// 6c. Dot for record member access: t.kind
	if ch == '.' {
		l.advance()
		return token.Token{
			Type:    token.DOT,
			Literal: ".",
			Line:    startLine,
			Col:     startCol,
			File:    l.file,
		}
	}

	// 6d. Group open/close aliases: ( and )
	if ch == '(' {
		l.advance()
		return token.Token{
			Type:    token.GROUP_OPEN,
			Literal: "(",
			Line:    startLine,
			Col:     startCol,
			File:    l.file,
		}
	}
	if ch == ')' {
		l.advance()
		return token.Token{
			Type:    token.GROUP_CLOSE,
			Literal: ")",
			Line:    startLine,
			Col:     startCol,
			File:    l.file,
		}
	}

	// 7. Unknown / Legacy character -> ILLEGAL token
	l.advance()
	return token.Token{
		Type:    token.ILLEGAL,
		Literal: oneRune,
		Line:    startLine,
		Col:     startCol,
		File:    l.file,
	}
}

func (l *Lexer) peek() rune {
	if l.position >= len(l.input) {
		return 0
	}
	return l.input[l.position]
}

func (l *Lexer) peekNext() rune {
	if l.position+1 >= len(l.input) {
		return 0
	}
	return l.input[l.position+1]
}

func (l *Lexer) advance() rune {
	if l.position >= len(l.input) {
		return 0
	}
	ch := l.input[l.position]
	l.position++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.position < len(l.input) {
		ch := l.input[l.position]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) skipLineComment() {
	l.advance() // consume '/'
	l.advance() // consume '/'
	for l.position < len(l.input) {
		ch := l.advance()
		if ch == '\n' {
			break
		}
	}
}

func (l *Lexer) skipBlockComment() {
	l.advance() // consume '/'
	l.advance() // consume '*'
	depth := 1
	for l.position < len(l.input) && depth > 0 {
		if l.peek() == '/' && l.peekNext() == '*' {
			l.advance()
			l.advance()
			depth++
			continue
		}
		if l.peek() == '*' && l.peekNext() == '/' {
			l.advance()
			l.advance()
			depth--
			continue
		}
		l.advance()
	}
}

func (l *Lexer) readFallback() token.Token {
	startLine := l.line
	startCol := l.col
	l.advance() // consume '\'
	start := l.position
	for l.position < len(l.input) && (isLetter(l.peek()) || isDigit(l.peek())) {
		l.advance()
	}
	name := `\` + string(l.input[start:l.position])
	if tokType, ok := token.LookupFallback(name); ok {
		return token.Token{
			Type:    tokType,
			Literal: name,
			Line:    startLine,
			Col:     startCol,
			File:    l.file,
		}
	}
	return token.Token{
		Type:    token.ILLEGAL,
		Literal: name,
		Line:    startLine,
		Col:     startCol,
		File:    l.file,
	}
}

func (l *Lexer) readIdent() token.Token {
	startLine := l.line
	startCol := l.col
	start := l.position
	for l.position < len(l.input) && isIdentPart(l.peek()) {
		l.advance()
	}
	literal := string(l.input[start:l.position])
	tokType := token.LookupIdent(literal)
	return token.Token{
		Type:    tokType,
		Literal: literal,
		Line:    startLine,
		Col:     startCol,
		File:    l.file,
	}
}

func (l *Lexer) readNumber() token.Token {
	startLine := l.line
	startCol := l.col
	start := l.position
	isFloating := false

	// Check hex (0x), binary (0b), octal (0o)
	if l.peek() == '0' && l.position+1 < len(l.input) {
		next := l.peekNext()
		switch next {
		case 'x', 'X':
			l.advance()
			l.advance()
			for l.position < len(l.input) && (isHexDigit(l.peek()) || l.peek() == '_') {
				l.advance()
			}
			return token.Token{
				Type:    token.INT,
				Literal: string(l.input[start:l.position]),
				Line:    startLine,
				Col:     startCol,
				File:    l.file,
			}
		case 'b', 'B':
			l.advance()
			l.advance()
			for l.position < len(l.input) && (l.peek() == '0' || l.peek() == '1' || l.peek() == '_') {
				l.advance()
			}
			return token.Token{
				Type:    token.INT,
				Literal: string(l.input[start:l.position]),
				Line:    startLine,
				Col:     startCol,
				File:    l.file,
			}
		case 'o', 'O':
			l.advance()
			l.advance()
			for l.position < len(l.input) && ((l.peek() >= '0' && l.peek() <= '7') || l.peek() == '_') {
				l.advance()
			}
			return token.Token{
				Type:    token.INT,
				Literal: string(l.input[start:l.position]),
				Line:    startLine,
				Col:     startCol,
				File:    l.file,
			}
		}
	}

	for l.position < len(l.input) {
		ch := l.peek()
		if isDigit(ch) || ch == '_' {
			l.advance()
		} else if ch == '.' && isDigit(l.peekNext()) {
			isFloating = true
			l.advance() // consume '.'
			for l.position < len(l.input) && (isDigit(l.peek()) || l.peek() == '_') {
				l.advance()
			}
		} else if ch == 'e' || ch == 'E' {
			isFloating = true
			l.advance()
			if l.peek() == '+' || l.peek() == '-' {
				l.advance()
			}
			for l.position < len(l.input) && (isDigit(l.peek()) || l.peek() == '_') {
				l.advance()
			}
		} else {
			break
		}
	}

	// Complex number suffix: e.g. 3.14i or 2j
	if l.position < len(l.input) && (l.peek() == 'i' || l.peek() == 'j') {
		isFloating = true
		l.advance()
	}

	tokType := token.INT
	if isFloating {
		tokType = token.FLOAT
	}

	return token.Token{
		Type:    tokType,
		Literal: string(l.input[start:l.position]),
		Line:    startLine,
		Col:     startCol,
		File:    l.file,
	}
}

func (l *Lexer) readString() token.Token {
	startLine := l.line
	startCol := l.col

	// Check for triple-quote """
	if l.position+2 < len(l.input) && l.input[l.position] == '"' && l.input[l.position+1] == '"' && l.input[l.position+2] == '"' {
		l.advance() // "
		l.advance() // "
		l.advance() // "
		start := l.position
		for l.position < len(l.input) {
			if l.position+2 < len(l.input) && l.input[l.position] == '"' && l.input[l.position+1] == '"' && l.input[l.position+2] == '"' {
				literal := string(l.input[start:l.position])
				l.advance()
				l.advance()
				l.advance()
				return token.Token{
					Type:    token.STRING,
					Literal: literal,
					Line:    startLine,
					Col:     startCol,
					File:    l.file,
				}
			}
			l.advance()
		}
		// Unterminated multiline string
		return token.Token{
			Type:    token.ILLEGAL,
			Literal: string(l.input[start:l.position]),
			Line:    startLine,
			Col:     startCol,
			File:    l.file,
		}
	}

	// Single double-quoted string
	l.advance() // consume opening "
	start := l.position
	for l.position < len(l.input) {
		ch := l.peek()
		if ch == '"' {
			literal := string(l.input[start:l.position])
			l.advance() // consume closing "
			return token.Token{
				Type:    token.STRING,
				Literal: literal,
				Line:    startLine,
				Col:     startCol,
				File:    l.file,
			}
		}
		if ch == '\\' {
			l.advance() // consume backslash in escape sequence
			if l.position < len(l.input) {
				l.advance() // consume escaped character
			}
			continue
		}
		if ch == '\n' {
			// Unterminated single-line string
			return token.Token{
				Type:    token.ILLEGAL,
				Literal: string(l.input[start:l.position]),
				Line:    startLine,
				Col:     startCol,
				File:    l.file,
			}
		}
		l.advance()
	}

	return token.Token{
		Type:    token.ILLEGAL,
		Literal: string(l.input[start:l.position]),
		Line:    startLine,
		Col:     startCol,
		File:    l.file,
	}
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch rune) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isIdentStart(ch rune) bool {
	return isLetter(ch)
}

func isIdentPart(ch rune) bool {
	return isLetter(ch) || isDigit(ch)
}
