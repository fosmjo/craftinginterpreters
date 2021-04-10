package scanner

import "strconv"

type Scanner struct {
	source string
	tokens []Token
	lox    loxer

	start, current, line int
}

type loxer interface {
	Error(line int, msg string)
}

func New(source string, lox loxer) *Scanner {
	return &Scanner{
		source:  source,
		lox:     lox,
		tokens:  make([]Token, 0),
		start:   0,
		current: 0,
		line:    0,
	}
}

func (s *Scanner) ScanTokens() []Token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	eof := NewToken(EOF, "", nil, s.line)
	s.tokens = append(s.tokens, eof)
	return s.tokens
}

func (s *Scanner) scanToken() {
	ch := s.advance()
	switch ch {
	case '(':
		s.addToken(LEFT_PAREN)
	case ')':
		s.addToken(RIGHT_PAREN)
	case '{':
		s.addToken(LEFT_BRACE)
	case '}':
		s.addToken(RIGHT_BRACE)
	case ',':
		s.addToken(COMMA)
	case '.':
		s.addToken(DOT)
	case '-':
		s.addToken(MINUS)
	case '+':
		s.addToken(PLUS)
	case ';':
		s.addToken(SEMICOLON)
	case '*':
		s.addToken(STAR)
	case '!':
		tokenType := BANG
		if s.match('=') {
			tokenType = BANG_EQUAL
		}
		s.addToken(tokenType)
	case '=':
		tokenType := EQUAL
		if s.match('=') {
			tokenType = EQUAL_EQUAL
		}
		s.addToken(tokenType)
	case '<':
		tokenType := LESS
		if s.match('=') {
			tokenType = BANG_EQUAL
		}
		s.addToken(tokenType)
	case '>':
		tokenType := GREATER
		if s.match('=') {
			tokenType = GREATER_EQUAL
		}
		s.addToken(tokenType)
	case '/':
		if s.match('/') {
			// A comment goes until the end of the line.
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addToken(SLASH)
		}
	case ' ', '\r', '\t':
		// ignore whitespace
	case '\n':
		s.line++
	case '"':
		s.scanString()
	default:
		if s.isDigit(ch) {
			s.scanNumber()
		} else if s.isAlpha(ch) {
			s.scanIdentifier()
		} else {
			s.lox.Error(s.line, "Unexpected character")
		}
	}
}

func (s *Scanner) scanString() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		s.lox.Error(s.line, "Unterminated string.")
		return
	}
	// Consume closing "
	s.advance()
	// Trim the surrounding quotes
	str := s.source[s.start+1 : s.current-1]
	s.addToken(STRING, str)
}

func (s *Scanner) scanNumber() {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		s.advance()
		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	num, err := strconv.ParseFloat(s.currentLexeme(), 64)
	if err != nil {
		s.lox.Error(s.line, "Invalid number")
	}

	s.addToken(NUMBER, num)
}

func (s *Scanner) scanIdentifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	id := s.currentLexeme()
	tokenType, isKeyword := keywords[id]
	if !isKeyword {
		tokenType = IDENTIFIER
	}

	s.addToken(tokenType)
}

func (s *Scanner) addToken(tokenType TokenType, literal ...interface{}) {
	var lit interface{}
	if len(literal) > 0 {
		lit = literal[0]
	}

	text := s.currentLexeme()
	token := NewToken(tokenType, text, lit, s.line)
	s.tokens = append(s.tokens, token)
}

func (s *Scanner) currentLexeme() string {
	return string(s.source[s.start:s.current])
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (s *Scanner) isAlpha(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func (s *Scanner) isAlphaNumeric(ch byte) bool {
	return s.isAlpha(ch) || s.isDigit(ch)
}

func (s *Scanner) advance() byte {
	ch := s.source[s.current]
	s.current++
	return ch
}

func (s *Scanner) match(ch byte) bool {
	if s.isAtEnd() {
		return false
	}

	if s.source[s.current] != ch {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return 0
	}
	return s.source[s.current]
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return 0
	}
	return s.source[s.current+1]
}
