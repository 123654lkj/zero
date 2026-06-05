package compiler

import (
	"fmt"
	"strconv"
)

// Parser is a recursive descent parser for the Zero language.
type Parser struct {
	lexer   *Lexer
	current Token // last consumed token
	peek    Token // lookahead token
}

// NewParser creates a new Parser that parses the given source string.
func NewParser(source string) *Parser {
	p := &Parser{lexer: NewLexer(source)}
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances peek into current and reads a new token into peek.
func (p *Parser) nextToken() {
	p.current = p.peek
	p.peek = p.lexer.NextToken()
}

// advance consumes the current token and returns it.
func (p *Parser) advance() Token {
	tok := p.current
	p.nextToken()
	return tok
}

// expect consumes the current token if it matches typ, else panics.
func (p *Parser) expect(typ TokenType) Token {
	if p.current.Type != typ {
		panic(fmt.Sprintf("parser: expected %s, got %s at line %d col %d",
			typ, p.current.Type, p.current.Line, p.current.Col))
	}
	return p.advance()
}

// match consumes the current token if it matches typ and returns true.
func (p *Parser) match(typ TokenType) bool {
	if p.current.Type == typ {
		p.advance()
		return true
	}
	return false
}

// Parse parses a complete Zero program and returns the list of statements.
func (p *Parser) Parse() []Stmt {
	var stmts []Stmt
	for p.current.Type != TOKEN_EOF {
		// skip newlines between statements
		if p.match(TOKEN_NEWLINE) {
			continue
		}
		stmts = append(stmts, p.parseStatement())
	}
	return stmts
}

// parseStatement dispatches based on the current token type.
func (p *Parser) parseStatement() Stmt {
	switch p.current.Type {
	case TOKEN_LET:
		return p.parseVarDecl()
	case TOKEN_FN:
		return p.parseFuncDef()
	case TOKEN_IF:
		return p.parseIf()
	case TOKEN_WHILE:
		return p.parseWhile()
	case TOKEN_RETURN:
		return p.parseReturn()
	case TOKEN_IDENT:
		// Check if this is an assignment: IDENT ASSIGN
		if p.peek.Type == TOKEN_ASSIGN {
			return p.parseAssign()
		}
		// Check if this is index assignment: IDENT[expr] = expr
		if p.peek.Type == TOKEN_LBRACKET {
			return p.parseIndexAssign()
		}
		return p.parseExprStmt()
	default:
		return p.parseExprStmt()
	}
}

// parseBlock parses a block: { stmt* }
func (p *Parser) parseBlock() Block {
	tok := p.expect(TOKEN_LBRACE)
	var stmts []Stmt
	for p.current.Type != TOKEN_RBRACE && p.current.Type != TOKEN_EOF {
		if p.match(TOKEN_NEWLINE) {
			continue
		}
		stmts = append(stmts, p.parseStatement())
	}
	p.expect(TOKEN_RBRACE)
	return Block{Token: tok, Stmts: stmts}
}

// parseVarDecl parses: let ident = expression
func (p *Parser) parseVarDecl() *VarDecl {
	tok := p.expect(TOKEN_LET)
	name := p.expect(TOKEN_IDENT).Literal
	p.expect(TOKEN_ASSIGN)
	value := p.parseExpression()
	return &VarDecl{Token: tok, Name: name, Value: value}
}

// parseAssign parses: ident = expression
func (p *Parser) parseAssign() *Assign {
	nameTok := p.expect(TOKEN_IDENT)
	p.expect(TOKEN_ASSIGN)
	value := p.parseExpression()
	return &Assign{Token: nameTok, Name: nameTok.Literal, Value: value}
}

// parseFuncDef parses: fn name(params) { body }
func (p *Parser) parseFuncDef() *FuncDef {
	tok := p.expect(TOKEN_FN)
	name := p.expect(TOKEN_IDENT).Literal

	p.expect(TOKEN_LPAREN)
	var params []Param
	if p.current.Type != TOKEN_RPAREN {
		for {
			paramTok := p.expect(TOKEN_IDENT)
			params = append(params, Param{Token: paramTok, Name: paramTok.Literal})
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
	}
	p.expect(TOKEN_RPAREN)
	body := p.parseBlock()
	return &FuncDef{Token: tok, Name: name, Params: params, Body: body}
}

// skipNewlines advances past any consecutive TOKEN_NEWLINE tokens.
func (p *Parser) skipNewlines() {
	for p.match(TOKEN_NEWLINE) {
	}
}

// parseIndexAssign parses: ident[expr] = expr
func (p *Parser) parseIndexAssign() Stmt {
	nameTok := p.expect(TOKEN_IDENT)
	p.expect(TOKEN_LBRACKET)
	index := p.parseExpression()
	p.expect(TOKEN_RBRACKET)
	if p.current.Type == TOKEN_ASSIGN {
		p.advance()
		value := p.parseExpression()
		return &IndexAssign{
			Token:  nameTok,
			Object: &Identifier{Token: nameTok, Name: nameTok.Literal},
			Index:  index,
			Value:  value,
		}
	}
	// Not assignment, build IndexExpr and wrap in ExprStmt
	// (this shouldn't normally happen since parseStatement checks for LBRACKET first)
	return &ExprStmt{Expr: &IndexExpr{Token: nameTok, Object: &Identifier{Token: nameTok, Name: nameTok.Literal}, Index: index}}
}

// parseIf parses: if expr block (else block)?
func (p *Parser) parseIf() *IfStmt {
	tok := p.expect(TOKEN_IF)
	cond := p.parseExpression()
	then := p.parseBlock()
	var elseBlock Block
	p.skipNewlines()
	if p.match(TOKEN_ELSE) {
		if p.current.Type == TOKEN_IF {
			// else if: wrap inner if in a block
			innerIf := p.parseIf()
			elseBlock = Block{Token: innerIf.Token, Stmts: []Stmt{innerIf}}
		} else {
			elseBlock = p.parseBlock()
		}
	}
	return &IfStmt{Token: tok, Cond: cond, Then: then, Else: elseBlock}
}

// parseWhile parses: while expr block
func (p *Parser) parseWhile() *WhileStmt {
	tok := p.expect(TOKEN_WHILE)
	cond := p.parseExpression()
	body := p.parseBlock()
	return &WhileStmt{Token: tok, Cond: cond, Body: body}
}

// parseReturn parses: return expr?
func (p *Parser) parseReturn() *ReturnStmt {
	tok := p.expect(TOKEN_RETURN)
	var value Expr
	if p.current.Type != TOKEN_NEWLINE && p.current.Type != TOKEN_EOF && p.current.Type != TOKEN_RBRACE {
		value = p.parseExpression()
	}
	return &ReturnStmt{Token: tok, Value: value}
}

// parseExprStmt parses an expression followed by a statement terminator.
func (p *Parser) parseExprStmt() *ExprStmt {
	expr := p.parseExpression()
	return &ExprStmt{Expr: expr}
}

// parseExpression entry point for Pratt parser.
func (p *Parser) parseExpression() Expr {
	return p.parsePrecedence(0)
}

// precedence returns the binding power of a binary operator token type.
// Higher values bind more tightly.
func precedence(typ TokenType) int {
	switch typ {
	case TOKEN_OR:
		return 1
	case TOKEN_AND:
		return 2
	case TOKEN_EQ, TOKEN_NEQ:
		return 3
	case TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE:
		return 4
	case TOKEN_PLUS, TOKEN_MINUS:
		return 5
	case TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT:
		return 6
	default:
		return 0
	}
}

// isBinaryOp returns true if the token is a binary operator.
func isBinaryOp(typ TokenType) bool {
	switch typ {
	case TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT,
		TOKEN_EQ, TOKEN_NEQ, TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE,
		TOKEN_AND, TOKEN_OR:
		return true
	}
	return false
}

// parsePrecedence implements Pratt parsing / precedence climbing.
func (p *Parser) parsePrecedence(minPrec int) Expr {
	left := p.parsePrimary()

	// Handle postfix operators: call ( ) and index [ ]
	for {
		if p.current.Type == TOKEN_LPAREN {
			left = p.parseCall(left)
		} else if p.current.Type == TOKEN_LBRACKET {
			left = p.parseIndex(left)
		} else {
			break
		}
	}

	// Handle binary operators
	for isBinaryOp(p.current.Type) && precedence(p.current.Type) >= minPrec {
		opTok := p.advance()
		prec := precedence(opTok.Type)
		right := p.parsePrecedence(prec + 1)
		left = &BinaryExpr{Token: opTok, Op: opTok.Type, Left: left, Right: right}

		// After binary op, still handle postfix on the result
		for {
			if p.current.Type == TOKEN_LPAREN {
				left = p.parseCall(left)
			} else if p.current.Type == TOKEN_LBRACKET {
				left = p.parseIndex(left)
			} else {
				break
			}
		}
	}

	return left
}

// parseCall parses: callee(args...)
func (p *Parser) parseCall(callee Expr) Expr {
	tok := p.expect(TOKEN_LPAREN)
	var args []Expr
	if p.current.Type != TOKEN_RPAREN {
		args = append(args, p.parseExpression())
		for p.match(TOKEN_COMMA) {
			args = append(args, p.parseExpression())
		}
	}
	p.expect(TOKEN_RPAREN)
	return &CallExpr{Token: tok, Callee: callee, Args: args}
}

// parseIndex parses: object[index]
func (p *Parser) parseIndex(object Expr) Expr {
	tok := p.expect(TOKEN_LBRACKET)
	index := p.parseExpression()
	p.expect(TOKEN_RBRACKET)
	return &IndexExpr{Token: tok, Object: object, Index: index}
}

// parsePrimary parses literals, identifiers, grouped expressions,
// unary operators, array literals, and map literals.
func (p *Parser) parsePrimary() Expr {
	switch p.current.Type {
	case TOKEN_INT:
		tok := p.advance()
		val, _ := strconv.ParseInt(tok.Literal, 10, 64)
		return &IntLiteral{Token: tok, Value: val}

	case TOKEN_FLOAT:
		tok := p.advance()
		val, _ := strconv.ParseFloat(tok.Literal, 64)
		return &FloatLiteral{Token: tok, Value: val}

	case TOKEN_STRING:
		tok := p.advance()
		return &StringLiteral{Token: tok, Value: tok.Literal}

	case TOKEN_TRUE:
		tok := p.advance()
		return &BoolLiteral{Token: tok, Value: true}

	case TOKEN_FALSE:
		tok := p.advance()
		return &BoolLiteral{Token: tok, Value: false}

	case TOKEN_NIL:
		tok := p.advance()
		return &NilLiteral{Token: tok}

	case TOKEN_IDENT:
		tok := p.advance()
		return &Identifier{Token: tok, Name: tok.Literal}

	case TOKEN_LPAREN:
		p.advance()
		expr := p.parseExpression()
		p.expect(TOKEN_RPAREN)
		return expr

	case TOKEN_LBRACKET:
		return p.parseArrayLiteral()

	case TOKEN_LBRACE:
		return p.parseMapLiteral()

	case TOKEN_MINUS, TOKEN_NOT:
		tok := p.advance()
		operand := p.parsePrecedence(7) // unary precedence
		return &UnaryExpr{Token: tok, Op: tok.Type, Operand: operand}

	default:
		panic(fmt.Sprintf("parser: unexpected token %s at line %d col %d",
			p.current.Type, p.current.Line, p.current.Col))
	}
}

// parseArrayLiteral parses: [expr, expr, ...]
func (p *Parser) parseArrayLiteral() *ArrayLiteral {
	tok := p.expect(TOKEN_LBRACKET)
	var elements []Expr
	if p.current.Type != TOKEN_RBRACKET {
		elements = append(elements, p.parseExpression())
		for p.match(TOKEN_COMMA) {
			if p.current.Type == TOKEN_RBRACKET {
				break // trailing comma
			}
			elements = append(elements, p.parseExpression())
		}
	}
	p.expect(TOKEN_RBRACKET)
	return &ArrayLiteral{Token: tok, Elements: elements}
}

// parseMapLiteral parses: { key: value, ... }
func (p *Parser) parseMapLiteral() *MapLiteral {
	tok := p.expect(TOKEN_LBRACE)
	var keys []Expr
	var values []Expr
	if p.current.Type != TOKEN_RBRACE {
		key := p.parseExpression()
		p.expect(TOKEN_COLON)
		val := p.parseExpression()
		keys = append(keys, key)
		values = append(values, val)
		for p.match(TOKEN_COMMA) {
			if p.current.Type == TOKEN_RBRACE {
				break // trailing comma
			}
			key = p.parseExpression()
			p.expect(TOKEN_COLON)
			val = p.parseExpression()
			keys = append(keys, key)
			values = append(values, val)
		}
	}
	p.expect(TOKEN_RBRACE)
	return &MapLiteral{Token: tok, Keys: keys, Values: values}
}
