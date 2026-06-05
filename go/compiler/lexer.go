package compiler

// Lexer tokenizes Zero language source code.
type Lexer struct {
	source string
	pos    int
	line   int
	col    int
	ch     byte
}

// NewLexer creates a new Lexer initialized with the first character.
func NewLexer(source string) *Lexer {
	l := &Lexer{source: source, line: 1, col: 1}
	if len(source) > 0 {
		l.ch = source[0]
	}
	return l
}

// peekChar returns the current character without advancing, or 0 at end.
func (l *Lexer) peekChar() byte {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

// advanceChar advances the position and returns the character before advancing.
func (l *Lexer) advanceChar() byte {
	if l.pos >= len(l.source) {
		return 0
	}
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	// Update l.ch to the new current character
	if l.pos < len(l.source) {
		l.ch = l.source[l.pos]
	} else {
		l.ch = 0
	}
	return ch
}

// skipWhitespace skips spaces, tabs, carriage returns (not newlines) and line comments.
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.source) {
		ch := l.ch
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advanceChar()
		} else if ch == '/' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '/' {
			// line comment: skip until \n
			for l.pos < len(l.source) && l.ch != '\n' {
				l.advanceChar()
			}
		} else {
			break
		}
	}
}

// readNumber reads an integer or float literal.
func (l *Lexer) readNumber() Token {
	startLine := l.line
	startCol := l.col
	start := l.pos
	isFloat := false

	for l.pos < len(l.source) && l.ch >= '0' && l.ch <= '9' {
		l.advanceChar()
	}

	if l.pos < len(l.source) && l.ch == '.' {
		// Check that next char is a digit (to avoid consuming dots after identifiers)
		if l.pos+1 < len(l.source) && l.source[l.pos+1] >= '0' && l.source[l.pos+1] <= '9' {
			isFloat = true
			l.advanceChar() // consume '.'
			for l.pos < len(l.source) && l.ch >= '0' && l.ch <= '9' {
				l.advanceChar()
			}
		}
	}

	lit := l.source[start:l.pos]
	if isFloat {
		return Token{Type: TOKEN_FLOAT, Literal: lit, Line: startLine, Col: startCol}
	}
	return Token{Type: TOKEN_INT, Literal: lit, Line: startLine, Col: startCol}
}

// readString reads a double-quoted string with escape sequences.
func (l *Lexer) readString() Token {
	startLine := l.line
	startCol := l.col
	l.advanceChar() // consume opening "

	var buf []byte
	for l.pos < len(l.source) && l.ch != '"' {
		if l.ch == '\\' {
			l.advanceChar() // consume backslash
			switch l.ch {
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case '\\':
				buf = append(buf, '\\')
			case '"':
				buf = append(buf, '"')
			default:
				buf = append(buf, '\\')
				buf = append(buf, l.ch)
			}
			l.advanceChar()
		} else if l.ch == '\n' {
			buf = append(buf, '\n')
			l.advanceChar()
		} else {
			buf = append(buf, l.ch)
			l.advanceChar()
		}
	}

	if l.pos >= len(l.source) {
		return Token{Type: TOKEN_ERROR, Literal: "unterminated string", Line: startLine, Col: startCol}
	}

	l.advanceChar() // consume closing "

	return Token{Type: TOKEN_STRING, Literal: string(buf), Line: startLine, Col: startCol}
}

// readIdent reads an identifier or keyword.
func (l *Lexer) readIdent() Token {
	startLine := l.line
	startCol := l.col
	start := l.pos

	for l.pos < len(l.source) {
		ch := l.ch
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			l.advanceChar()
		} else {
			break
		}
	}

	lit := l.source[start:l.pos]
	if tt, ok := keywords[lit]; ok {
		return Token{Type: tt, Literal: lit, Line: startLine, Col: startCol}
	}
	return Token{Type: TOKEN_IDENT, Literal: lit, Line: startLine, Col: startCol}
}

// NextToken returns the next token from the source.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	startLine := l.line
	startCol := l.col

	if l.pos >= len(l.source) {
		return Token{Type: TOKEN_EOF, Literal: "", Line: startLine, Col: startCol}
	}

	ch := l.ch

	// Newline
	if ch == '\n' {
		l.advanceChar()
		return Token{Type: TOKEN_NEWLINE, Literal: "\n", Line: startLine, Col: startCol}
	}

	// Numbers
	if ch >= '0' && ch <= '9' {
		return l.readNumber()
	}

	// Strings
	if ch == '"' {
		return l.readString()
	}

	// Identifiers and keywords
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
		return l.readIdent()
	}

	// Two-character operators
	if ch == '=' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
		l.advanceChar()
		l.advanceChar()
		return Token{Type: TOKEN_EQ, Literal: "==", Line: startLine, Col: startCol}
	}
	if ch == '!' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
		l.advanceChar()
		l.advanceChar()
		return Token{Type: TOKEN_NEQ, Literal: "!=", Line: startLine, Col: startCol}
	}
	if ch == '<' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
		l.advanceChar()
		l.advanceChar()
		return Token{Type: TOKEN_LTE, Literal: "<=", Line: startLine, Col: startCol}
	}
	if ch == '>' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
		l.advanceChar()
		l.advanceChar()
		return Token{Type: TOKEN_GTE, Literal: ">=", Line: startLine, Col: startCol}
	}
	if ch == '-' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '>' {
		l.advanceChar()
		l.advanceChar()
		return Token{Type: TOKEN_ARROW, Literal: "->", Line: startLine, Col: startCol}
	}
	if ch == '.' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '.' {
		l.advanceChar()
		l.advanceChar()
		return Token{Type: TOKEN_DOTDOT, Literal: "..", Line: startLine, Col: startCol}
	}
	if ch == '&' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '&' {
		l.advanceChar()
		l.advanceChar()
		return Token{Type: TOKEN_AND, Literal: "&&", Line: startLine, Col: startCol}
	}
	if ch == '|' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '|' {
		l.advanceChar()
		l.advanceChar()
		return Token{Type: TOKEN_OR, Literal: "||", Line: startLine, Col: startCol}
	}
	if ch == '|' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '>' {
		l.advanceChar()
		l.advanceChar()
		return Token{Type: TOKEN_PIPE, Literal: "|>", Line: startLine, Col: startCol}
	}

	// Single-character tokens
	l.advanceChar()
	switch ch {
	case '+':
		return Token{Type: TOKEN_PLUS, Literal: "+", Line: startLine, Col: startCol}
	case '-':
		return Token{Type: TOKEN_MINUS, Literal: "-", Line: startLine, Col: startCol}
	case '*':
		return Token{Type: TOKEN_STAR, Literal: "*", Line: startLine, Col: startCol}
	case '/':
		return Token{Type: TOKEN_SLASH, Literal: "/", Line: startLine, Col: startCol}
	case '%':
		return Token{Type: TOKEN_PERCENT, Literal: "%", Line: startLine, Col: startCol}
	case '=':
		return Token{Type: TOKEN_ASSIGN, Literal: "=", Line: startLine, Col: startCol}
	case '<':
		return Token{Type: TOKEN_LT, Literal: "<", Line: startLine, Col: startCol}
	case '>':
		return Token{Type: TOKEN_GT, Literal: ">", Line: startLine, Col: startCol}
	case '!':
		return Token{Type: TOKEN_NOT, Literal: "!", Line: startLine, Col: startCol}
	case '(':
		return Token{Type: TOKEN_LPAREN, Literal: "(", Line: startLine, Col: startCol}
	case ')':
		return Token{Type: TOKEN_RPAREN, Literal: ")", Line: startLine, Col: startCol}
	case '{':
		return Token{Type: TOKEN_LBRACE, Literal: "{", Line: startLine, Col: startCol}
	case '}':
		return Token{Type: TOKEN_RBRACE, Literal: "}", Line: startLine, Col: startCol}
	case '[':
		return Token{Type: TOKEN_LBRACKET, Literal: "[", Line: startLine, Col: startCol}
	case ']':
		return Token{Type: TOKEN_RBRACKET, Literal: "]", Line: startLine, Col: startCol}
	case ',':
		return Token{Type: TOKEN_COMMA, Literal: ",", Line: startLine, Col: startCol}
	case ':':
		return Token{Type: TOKEN_COLON, Literal: ":", Line: startLine, Col: startCol}
	case '.':
		return Token{Type: TOKEN_DOT, Literal: ".", Line: startLine, Col: startCol}
	case ';':
		return Token{Type: TOKEN_SEMICOLON, Literal: ";", Line: startLine, Col: startCol}
	default:
		return Token{Type: TOKEN_ERROR, Literal: string(ch), Line: startLine, Col: startCol}
	}
}
