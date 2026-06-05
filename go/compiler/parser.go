package compiler

import (
	"fmt"
	"strconv"
	"strings"
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
	case TOKEN_PURE:
		return p.parsePureFuncDef()
	case TOKEN_IF:
		return p.parseIf()
	case TOKEN_WHILE:
		return p.parseWhile()
	case TOKEN_RETURN:
		return p.parseReturn()
	case TOKEN_MATCH:
		return &ExprStmt{Expr: p.parseMatch()}
	case TOKEN_ENUM:
		return p.parseEnum()
	case TOKEN_AI:
		return p.parseAIFuncDef()
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

// parseFuncDef parses: fn name(params) [effects: [io, netio]] [-> type] { body }
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

	// Optional: effects: [io, netio, ...]
	var effects []EffectAnnot
	if p.match(TOKEN_EFFECTS) {
		p.expect(TOKEN_COLON)
		p.expect(TOKEN_LBRACKET)
		if p.current.Type != TOKEN_RBRACKET {
			for {
				effectTok := p.current
				p.advance()
				effects = append(effects, EffectAnnot{
					Token: effectTok,
					Name:  effectTok.Literal,
				})
				if !p.match(TOKEN_COMMA) {
					break
				}
			}
		}
		p.expect(TOKEN_RBRACKET)
	}

	// Optional: -> ReturnType
	returnType := ""
	if p.match(TOKEN_ARROW) {
		returnType = p.expect(TOKEN_IDENT).Literal
	}

	body := p.parseBlock()
	return &FuncDef{
		Token:      tok,
		Name:       name,
		Params:     params,
		Body:       body,
		Effects:    effects,
		ReturnType: returnType,
	}
}

// parsePureFuncDef parses: pure fn name(params) [-> type] { body }
// A pure function has no side effects — the compiler tracks this.
func (p *Parser) parsePureFuncDef() *FuncDef {
	tok := p.expect(TOKEN_PURE)
	p.expect(TOKEN_FN)
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

	// Optional: -> ReturnType
	returnType := ""
	if p.match(TOKEN_ARROW) {
		returnType = p.expect(TOKEN_IDENT).Literal
	}

	body := p.parseBlock()

	// Mark as pure by adding a "pure" effect annotation
	effects := []EffectAnnot{
		{Token: tok, Name: "pure"},
	}

	return &FuncDef{
		Token:      tok,
		Name:       name,
		Params:     params,
		Body:       body,
		Effects:    effects,
		ReturnType: returnType,
	}
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
	case TOKEN_PIPE:
		return 0
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
		TOKEN_AND, TOKEN_OR, TOKEN_PIPE:
		return true
	}
	return false
}

// parsePrecedence implements Pratt parsing / precedence climbing.
func (p *Parser) parsePrecedence(minPrec int) Expr {
	left := p.parsePrimary()

	// Handle postfix operators: call ( ), index [ ], and dot .
	for {
		if p.current.Type == TOKEN_LPAREN {
			left = p.parseCall(left)
		} else if p.current.Type == TOKEN_LBRACKET {
			left = p.parseIndex(left)
		} else if p.current.Type == TOKEN_DOT {
			p.advance()
			field := p.expect(TOKEN_IDENT).Literal
			left = &DotExpr{Token: p.current, Object: left, Field: field}
		} else {
			break
		}
	}

	// Handle binary operators
	for isBinaryOp(p.current.Type) && precedence(p.current.Type) >= minPrec {
		// Pipe operator has special parsing: args follow without parentheses
		if p.current.Type == TOKEN_PIPE {
			p.advance() // consume |>
			tok := p.current
			name := p.expect(TOKEN_IDENT).Literal
			var args []Expr
			// Collect arguments until we hit another |>, newline, or EOF
			for p.current.Type != TOKEN_PIPE &&
				p.current.Type != TOKEN_NEWLINE &&
				p.current.Type != TOKEN_EOF &&
				p.current.Type != TOKEN_RBRACE &&
				p.current.Type != TOKEN_RPAREN {
				if p.current.Type == TOKEN_COMMA {
					break
				}
				args = append(args, p.parsePrimary())
			}
			step := PipelineStep{Token: tok, Name: name, Args: args}
			if pipeExpr, ok := left.(*PipelineExpr); ok {
				pipeExpr.Steps = append(pipeExpr.Steps, step)
				left = pipeExpr
			} else {
				left = &PipelineExpr{
					Token: tok,
					Expr:  left,
					Steps: []PipelineStep{step},
				}
			}
			continue
		}

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
			} else if p.current.Type == TOKEN_DOT {
				p.advance()
				field := p.expect(TOKEN_IDENT).Literal
				left = &DotExpr{Token: p.current, Object: left, Field: field}
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

	case TOKEN_SPAWN:
		return p.parseSpawn()

	case TOKEN_SEND:
		return p.parseSend()

	case TOKEN_RECEIVE:
		return p.parseReceive()

	case TOKEN_CHANNEL:
		return p.parseChannel()
	case TOKEN_MATCH:
		return p.parseMatch()
	case TOKEN_SNAPSHOT:
		return p.parseSnapshot()
	case TOKEN_RESTORE:
		return p.parseRestore()
	case TOKEN_REPLAY:
		return p.parseReplay()
	case TOKEN_PURE:
		// pure fn as expression: not supported in expression context
		panic(fmt.Sprintf("parser: pure fn definition is only allowed as a statement at line %d", p.current.Line))

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

// parseMatch parses: match expr { case pattern -> expr, ... }
func (p *Parser) parseMatch() *MatchExpression {
	tok := p.expect(TOKEN_MATCH)
	value := p.parseExpression()
	p.skipNewlines()
	p.expect(TOKEN_LBRACE)
	p.skipNewlines()
	var cases []MatchCase
	for p.current.Type != TOKEN_RBRACE && p.current.Type != TOKEN_EOF {
		if p.match(TOKEN_NEWLINE) {
			continue
		}
		p.expect(TOKEN_CASE)
		pattern := p.parseExpression()
		p.expect(TOKEN_ARROW)
		body := p.parseExpression()

		// Extract token from pattern
		var patToken Token
		switch pat := pattern.(type) {
		case *IntLiteral:
			patToken = pat.Token
		case *StringLiteral:
			patToken = pat.Token
		case *FloatLiteral:
			patToken = pat.Token
		case *BoolLiteral:
			patToken = pat.Token
		case *Identifier:
			patToken = pat.Token
		default:
			patToken = tok // fallback
		}

		cases = append(cases, MatchCase{Token: patToken, Pattern: pattern, Body: body})
		p.match(TOKEN_COMMA) // optional trailing comma
		p.skipNewlines()
	}
	p.expect(TOKEN_RBRACE)
	return &MatchExpression{Token: tok, Value: value, Cases: cases}
}

// parseEnum parses: enum Name { Variant1, Variant2(args...) }
func (p *Parser) parseEnum() *EnumDef {
	tok := p.expect(TOKEN_ENUM)
	name := p.expect(TOKEN_IDENT).Literal
	p.skipNewlines()
	p.expect(TOKEN_LBRACE)
	p.skipNewlines()
	var variants []EnumVariant
	for p.current.Type != TOKEN_RBRACE && p.current.Type != TOKEN_EOF {
		if p.match(TOKEN_NEWLINE) {
			continue
		}
		vtok := p.current
		vname := p.advance().Literal
		arity := 0
		if p.current.Type == TOKEN_LPAREN {
			p.advance()
			// count params
			if p.current.Type != TOKEN_RPAREN {
				arity = 1
				p.advance() // consume first param name
				for p.match(TOKEN_COMMA) {
					p.advance() // consume subsequent param names
					arity++
				}
			}
			p.expect(TOKEN_RPAREN)
		}
		variants = append(variants, EnumVariant{Token: vtok, Name: vname, Arity: arity})
		p.match(TOKEN_COMMA) // optional trailing comma
		p.skipNewlines()
	}
	p.expect(TOKEN_RBRACE)
	return &EnumDef{Token: tok, Name: name, Variants: variants}
}

// parseSpawn parses: spawn fn_name(args)
func (p *Parser) parseSpawn() *SpawnExpr {
	tok := p.expect(TOKEN_SPAWN)
	callee := p.parsePrimary()
	var args []Expr
	if p.current.Type == TOKEN_LPAREN {
		p.advance()
		if p.current.Type != TOKEN_RPAREN {
			args = append(args, p.parseExpression())
			for p.match(TOKEN_COMMA) {
				args = append(args, p.parseExpression())
			}
		}
		p.expect(TOKEN_RPAREN)
	}
	return &SpawnExpr{Token: tok, Callee: callee, Args: args}
}

// parseSend parses: send(channel, value)
func (p *Parser) parseSend() *SendExpr {
	tok := p.expect(TOKEN_SEND)
	p.expect(TOKEN_LPAREN)
	channel := p.parseExpression()
	p.expect(TOKEN_COMMA)
	value := p.parseExpression()
	p.expect(TOKEN_RPAREN)
	return &SendExpr{Token: tok, Channel: channel, Value: value}
}

// parseReceive parses: receive(channel)
func (p *Parser) parseReceive() *ReceiveExpr {
	tok := p.expect(TOKEN_RECEIVE)
	p.expect(TOKEN_LPAREN)
	channel := p.parseExpression()
	p.expect(TOKEN_RPAREN)
	return &ReceiveExpr{Token: tok, Channel: channel}
}

// parseChannel parses: channel()
func (p *Parser) parseChannel() *ChannelExpr {
	tok := p.expect(TOKEN_CHANNEL)
	p.expect(TOKEN_LPAREN)
	p.expect(TOKEN_RPAREN)
	return &ChannelExpr{Token: tok}
}

// parseSnapshot parses: snapshot()
func (p *Parser) parseSnapshot() *SnapshotExpr {
	tok := p.expect(TOKEN_SNAPSHOT)
	p.expect(TOKEN_LPAREN)
	p.expect(TOKEN_RPAREN)
	return &SnapshotExpr{Token: tok}
}

// parseRestore parses: restore(index)
func (p *Parser) parseRestore() *RestoreExpr {
	tok := p.expect(TOKEN_RESTORE)
	p.expect(TOKEN_LPAREN)
	index := p.parseExpression()
	p.expect(TOKEN_RPAREN)
	return &RestoreExpr{Token: tok, Index: index}
}

// parseReplay parses: replay(index)
func (p *Parser) parseReplay() *ReplayExpr {
	tok := p.expect(TOKEN_REPLAY)
	p.expect(TOKEN_LPAREN)
	index := p.parseExpression()
	p.expect(TOKEN_RPAREN)
	return &ReplayExpr{Token: tok, Index: index}
}

// parseAIFuncDef parses: ai fn name(args) -> type { prompt: "..." [model: "..."] [cache: true/false] }
func (p *Parser) parseAIFuncDef() *AIFuncDef {
	tok := p.expect(TOKEN_AI)
	p.expect(TOKEN_FN)
	name := p.expect(TOKEN_IDENT).Literal

	// Parse parameters
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

	// Optional return type
	returnType := ""
	if p.match(TOKEN_ARROW) {
		returnType = p.expect(TOKEN_IDENT).Literal
	}

	// Parse AI function body: { prompt: "..." [model: "..."] [cache: true/false] }
	p.expect(TOKEN_LBRACE)
	prompt := ""
	model := ""
	cache := false

	// Parse block contents — support key: value pairs
	for p.current.Type != TOKEN_RBRACE && p.current.Type != TOKEN_EOF {
		if p.match(TOKEN_NEWLINE) {
			continue
		}
		// Accept keywords as keys too (prompt, model, cache are keywords)
		var key string
		if p.current.Type == TOKEN_IDENT {
			key = p.advance().Literal
		} else {
			key = p.advance().Type.String()
			// Lowercase the token name: "PROMPT" -> "prompt"
			key = strings.ToLower(key)
		}
		p.expect(TOKEN_COLON)
		switch key {
		case "prompt":
			prompt = p.expect(TOKEN_STRING).Literal
		case "model":
			model = p.expect(TOKEN_STRING).Literal
		case "cache":
			cache = p.current.Type == TOKEN_TRUE
			p.advance()
		default:
			panic(fmt.Sprintf("parser: unknown AI function property %q at line %d col %d",
				key, p.current.Line, p.current.Col))
		}
		// optional newline between properties
		p.match(TOKEN_NEWLINE)
	}
	p.expect(TOKEN_RBRACE)

	return &AIFuncDef{
		Token:      tok,
		Name:       name,
		Params:     params,
		ReturnType: returnType,
		Prompt:     prompt,
		Model:      model,
		Cache:      cache,
	}
}
