// Package compiler implements the Zero language three-stage compilation pipeline:
// Lexer (tokens) -> Parser (AST) -> Compiler (bytecode).
package compiler

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Literals
	TOKEN_INT TokenType = iota
	TOKEN_FLOAT
	TOKEN_STRING
	TOKEN_IDENT

	// Keywords
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_NIL
	TOKEN_LET
	TOKEN_FN
	TOKEN_RETURN
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_WHILE

	// Operators
	TOKEN_PLUS      // +
	TOKEN_MINUS     // -
	TOKEN_STAR      // *
	TOKEN_SLASH     // /
	TOKEN_PERCENT   // %
	TOKEN_EQ        // ==
	TOKEN_NEQ       // !=
	TOKEN_LT        // <
	TOKEN_GT        // >
	TOKEN_LTE       // <=
	TOKEN_GTE       // >=
	TOKEN_AND       // &&
	TOKEN_OR        // ||
	TOKEN_NOT       // !
	TOKEN_ASSIGN    // =
	TOKEN_ARROW     // ->

	// Delimiters
	TOKEN_LPAREN    // (
	TOKEN_RPAREN    // )
	TOKEN_LBRACE    // {
	TOKEN_RBRACE    // }
	TOKEN_LBRACKET  // [
	TOKEN_RBRACKET  // ]
	TOKEN_COMMA     // ,
	TOKEN_COLON     // :
	TOKEN_DOT       // .
	TOKEN_SEMICOLON // ;

	// Special
	TOKEN_NEWLINE
	TOKEN_EOF
	TOKEN_ERROR
)

var tokenNames = map[TokenType]string{
	TOKEN_INT: "INT", TOKEN_FLOAT: "FLOAT", TOKEN_STRING: "STRING", TOKEN_IDENT: "IDENT",
	TOKEN_TRUE: "TRUE", TOKEN_FALSE: "FALSE", TOKEN_NIL: "NIL",
	TOKEN_LET: "LET", TOKEN_FN: "FN", TOKEN_RETURN: "RETURN",
	TOKEN_IF: "IF", TOKEN_ELSE: "ELSE", TOKEN_WHILE: "WHILE",
	TOKEN_PLUS: "PLUS", TOKEN_MINUS: "MINUS", TOKEN_STAR: "STAR",
	TOKEN_SLASH: "SLASH", TOKEN_PERCENT: "PERCENT",
	TOKEN_EQ: "EQ", TOKEN_NEQ: "NEQ", TOKEN_LT: "LT", TOKEN_GT: "GT",
	TOKEN_LTE: "LTE", TOKEN_GTE: "GTE", TOKEN_AND: "AND", TOKEN_OR: "OR",
	TOKEN_NOT: "NOT", TOKEN_ASSIGN: "ASSIGN", TOKEN_ARROW: "ARROW",
	TOKEN_LPAREN: "LPAREN", TOKEN_RPAREN: "RPAREN",
	TOKEN_LBRACE: "LBRACE", TOKEN_RBRACE: "RBRACE",
	TOKEN_LBRACKET: "LBRACKET", TOKEN_RBRACKET: "RBRACKET",
	TOKEN_COMMA: "COMMA", TOKEN_COLON: "COLON", TOKEN_DOT: "DOT",
	TOKEN_SEMICOLON: "SEMICOLON", TOKEN_NEWLINE: "NEWLINE",
	TOKEN_EOF: "EOF", TOKEN_ERROR: "ERROR",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// Token represents a single lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

var keywords = map[string]TokenType{
	"true":   TOKEN_TRUE,
	"false":  TOKEN_FALSE,
	"nil":    TOKEN_NIL,
	"let":    TOKEN_LET,
	"fn":     TOKEN_FN,
	"return": TOKEN_RETURN,
	"if":     TOKEN_IF,
	"else":   TOKEN_ELSE,
	"while":  TOKEN_WHILE,
}

// AST Node types

// Expr is the interface for all expression nodes.
type Expr interface {
	exprNode()
}

// Stmt is the interface for all statement nodes.
type Stmt interface {
	stmtNode()
}

// --- Expressions ---

type IntLiteral struct {
	Token Token
	Value int64
}

type FloatLiteral struct {
	Token Token
	Value float64
}

type StringLiteral struct {
	Token Token
	Value string
}

type BoolLiteral struct {
	Token Token
	Value bool
}

type NilLiteral struct {
	Token Token
}

type Identifier struct {
	Token Token
	Name  string
}

type BinaryExpr struct {
	Token Token
	Op    TokenType
	Left  Expr
	Right Expr
}

type UnaryExpr struct {
	Token   Token
	Op      TokenType
	Operand Expr
}

type CallExpr struct {
	Token  Token
	Callee Expr
	Args   []Expr
}

type IndexExpr struct {
	Token  Token
	Object Expr
	Index  Expr
}

type ArrayLiteral struct {
	Token    Token
	Elements []Expr
}

type MapLiteral struct {
	Token  Token
	Keys   []Expr
	Values []Expr
}

// Implement exprNode for all expression types
func (IntLiteral) exprNode()    {}
func (FloatLiteral) exprNode()  {}
func (StringLiteral) exprNode() {}
func (BoolLiteral) exprNode()   {}
func (NilLiteral) exprNode()    {}
func (Identifier) exprNode()    {}
func (BinaryExpr) exprNode()    {}
func (UnaryExpr) exprNode()     {}
func (CallExpr) exprNode()      {}
func (IndexExpr) exprNode()     {}
func (ArrayLiteral) exprNode()  {}
func (MapLiteral) exprNode()    {}

// --- Statements ---

type ExprStmt struct {
	Expr Expr
}

type VarDecl struct {
	Token Token
	Name  string
	Value Expr
}

type Assign struct {
	Token Token
	Name  string
	Value Expr
}

type ReturnStmt struct {
	Token Token
	Value Expr // can be nil for bare return
}

type IfStmt struct {
	Token Token
	Cond  Expr
	Then  Block
	Else  Block // empty if no else
}

type WhileStmt struct {
	Token Token
	Cond  Expr
	Body  Block
}

type FuncDef struct {
	Token  Token
	Name   string
	Params []Param
	Body   Block
}

type Block struct {
	Token Token
	Stmts []Stmt
}

type Param struct {
	Token Token
	Name  string
}

// Implement stmtNode for all statement types
func (ExprStmt) stmtNode()   {}
func (VarDecl) stmtNode()    {}
func (Assign) stmtNode()     {}
func (ReturnStmt) stmtNode() {}
func (IfStmt) stmtNode()     {}
func (WhileStmt) stmtNode()  {}
func (FuncDef) stmtNode()    {}
func (Block) stmtNode()      {}
