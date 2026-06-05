package compiler

import (
	"testing"
)

// collectTokens lexes source and returns all tokens (excluding EOF, but including it at the end).
func collectTokens(source string) []Token {
	l := NewLexer(source)
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}

// expectTypes checks that the token types match the expected list.
func expectTypes(t *testing.T, tokens []Token, expected []TokenType) {
	t.Helper()
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
		for i, tok := range tokens {
			t.Logf("  [%d] %s %q", i, tok.Type, tok.Literal)
		}
	}
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token %d: expected %s, got %s (literal=%q)", i, expected[i], tok.Type, tok.Literal)
		}
	}
}

func TestLexerLetAssignment(t *testing.T) {
	tokens := collectTokens("let x = 5")
	expectTypes(t, tokens, []TokenType{
		TOKEN_LET, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_INT, TOKEN_EOF,
	})
	if tokens[1].Literal != "x" {
		t.Errorf("expected ident literal 'x', got %q", tokens[1].Literal)
	}
	if tokens[3].Literal != "5" {
		t.Errorf("expected int literal '5', got %q", tokens[3].Literal)
	}
}

func TestLexerArithmetic(t *testing.T) {
	tokens := collectTokens("2 + 3 * 4")
	expectTypes(t, tokens, []TokenType{
		TOKEN_INT, TOKEN_PLUS, TOKEN_INT, TOKEN_STAR, TOKEN_INT, TOKEN_EOF,
	})
}

func TestLexerString(t *testing.T) {
	tokens := collectTokens(`"hello world"`)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TOKEN_STRING {
		t.Errorf("expected STRING, got %s", tokens[0].Type)
	}
	if tokens[0].Literal != "hello world" {
		t.Errorf("expected 'hello world', got %q", tokens[0].Literal)
	}
}

func TestLexerIfGt(t *testing.T) {
	tokens := collectTokens("if x > 5")
	expectTypes(t, tokens, []TokenType{
		TOKEN_IF, TOKEN_IDENT, TOKEN_GT, TOKEN_INT, TOKEN_EOF,
	})
}

func TestLexerFuncDef(t *testing.T) {
	tokens := collectTokens("fn add(a, b)")
	expectTypes(t, tokens, []TokenType{
		TOKEN_FN, TOKEN_IDENT, TOKEN_LPAREN, TOKEN_IDENT, TOKEN_COMMA,
		TOKEN_IDENT, TOKEN_RPAREN, TOKEN_EOF,
	})
}

func TestLexerCommentsSkipped(t *testing.T) {
	tokens := collectTokens("let x = 1 // this is a comment\nlet y = 2")
	expectTypes(t, tokens, []TokenType{
		TOKEN_LET, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_INT, TOKEN_NEWLINE,
		TOKEN_LET, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_INT, TOKEN_EOF,
	})
}

func TestLexerMultiline(t *testing.T) {
	src := "let a = 1\nlet b = 2\n"
	tokens := collectTokens(src)
	expectTypes(t, tokens, []TokenType{
		TOKEN_LET, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_INT, TOKEN_NEWLINE,
		TOKEN_LET, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_INT, TOKEN_NEWLINE,
		TOKEN_EOF,
	})
	if tokens[1].Line != 1 {
		t.Errorf("first ident should be on line 1, got line %d", tokens[1].Line)
	}
	if tokens[6].Line != 2 {
		t.Errorf("second ident should be on line 2, got line %d", tokens[6].Line)
	}
}

func TestLexerFloat(t *testing.T) {
	tokens := collectTokens("3.14")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TOKEN_FLOAT {
		t.Errorf("expected FLOAT, got %s", tokens[0].Type)
	}
	if tokens[0].Literal != "3.14" {
		t.Errorf("expected '3.14', got %q", tokens[0].Literal)
	}
}

func TestLexerStringEscapes(t *testing.T) {
	tokens := collectTokens(`"line1\nline2\tescaped\"done"`)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TOKEN_STRING {
		t.Errorf("expected STRING, got %s", tokens[0].Type)
	}
	expected := "line1\nline2\tescaped\"done"
	if tokens[0].Literal != expected {
		t.Errorf("expected %q, got %q", expected, tokens[0].Literal)
	}
}

func TestLexerComparisonOperators(t *testing.T) {
	tokens := collectTokens("a == b != c <= d >= e")
	expectTypes(t, tokens, []TokenType{
		TOKEN_IDENT, TOKEN_EQ, TOKEN_IDENT, TOKEN_NEQ, TOKEN_IDENT,
		TOKEN_LTE, TOKEN_IDENT, TOKEN_GTE, TOKEN_IDENT, TOKEN_EOF,
	})
}

func TestLexerArrow(t *testing.T) {
	tokens := collectTokens("x -> y")
	expectTypes(t, tokens, []TokenType{
		TOKEN_IDENT, TOKEN_ARROW, TOKEN_IDENT, TOKEN_EOF,
	})
}

func TestLexerLogicalOperators(t *testing.T) {
	tokens := collectTokens("a && b || !c")
	expectTypes(t, tokens, []TokenType{
		TOKEN_IDENT, TOKEN_AND, TOKEN_IDENT, TOKEN_OR, TOKEN_NOT,
		TOKEN_IDENT, TOKEN_EOF,
	})
}

func TestLexerDelimiters(t *testing.T) {
	tokens := collectTokens("[1, 2]; {a: b}")
	expectTypes(t, tokens, []TokenType{
		TOKEN_LBRACKET, TOKEN_INT, TOKEN_COMMA, TOKEN_INT, TOKEN_RBRACKET,
		TOKEN_SEMICOLON,
		TOKEN_LBRACE, TOKEN_IDENT, TOKEN_COLON, TOKEN_IDENT, TOKEN_RBRACE,
		TOKEN_EOF,
	})
}

func TestLexerEmptySource(t *testing.T) {
	tokens := collectTokens("")
	if len(tokens) != 1 || tokens[0].Type != TOKEN_EOF {
		t.Errorf("expected single EOF, got %d tokens", len(tokens))
	}
}

func TestLexerKeywords(t *testing.T) {
	tokens := collectTokens("true false nil return while else")
	expectTypes(t, tokens, []TokenType{
		TOKEN_TRUE, TOKEN_FALSE, TOKEN_NIL, TOKEN_RETURN, TOKEN_WHILE,
		TOKEN_ELSE, TOKEN_EOF,
	})
}
