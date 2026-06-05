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
	TOKEN_MATCH
	TOKEN_CASE
	TOKEN_ENUM
	TOKEN_SPAWN
	TOKEN_SEND
	TOKEN_RECEIVE
	TOKEN_CHANNEL
	// Effect system keywords (Phase 2)
	TOKEN_PURE    // pure keyword
	TOKEN_EFFECTS // effects keyword

	// Snapshot keywords (Phase 2)
	TOKEN_SNAPSHOT  // snapshot keyword
	TOKEN_RESTORE   // restore keyword
	TOKEN_REPLAY    // replay keyword

	// AI keywords (Phase 2)
	TOKEN_AI      // ai keyword
	TOKEN_CACHE   // cache keyword
	TOKEN_PROMPT  // prompt keyword

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
	TOKEN_PIPE      // |>

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
	TOKEN_DOTDOT    // ..

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
	TOKEN_SPAWN: "SPAWN", TOKEN_SEND: "SEND", TOKEN_RECEIVE: "RECEIVE",
 TOKEN_CHANNEL: "CHANNEL",
 TOKEN_PURE: "PURE", TOKEN_EFFECTS: "EFFECTS",
 TOKEN_SNAPSHOT: "SNAPSHOT", TOKEN_RESTORE: "RESTORE", TOKEN_REPLAY: "REPLAY",
 TOKEN_AI: "AI", TOKEN_CACHE: "CACHE", TOKEN_PROMPT: "PROMPT",
	TOKEN_PLUS: "PLUS", TOKEN_MINUS: "MINUS", TOKEN_STAR: "STAR",
	TOKEN_SLASH: "SLASH", TOKEN_PERCENT: "PERCENT",
	TOKEN_EQ: "EQ", TOKEN_NEQ: "NEQ", TOKEN_LT: "LT", TOKEN_GT: "GT",
	TOKEN_LTE: "LTE", TOKEN_GTE: "GTE", TOKEN_AND: "AND", TOKEN_OR: "OR",
	TOKEN_NOT: "NOT", TOKEN_ASSIGN: "ASSIGN", TOKEN_ARROW: "ARROW",
	TOKEN_PIPE: "PIPE",
	TOKEN_LPAREN: "LPAREN", TOKEN_RPAREN: "RPAREN",
	TOKEN_LBRACE: "LBRACE", TOKEN_RBRACE: "RBRACE",
	TOKEN_LBRACKET: "LBRACKET", TOKEN_RBRACKET: "RBRACKET",
	TOKEN_COMMA: "COMMA", TOKEN_COLON: "COLON", TOKEN_DOT: "DOT",
 TOKEN_SEMICOLON: "SEMICOLON", TOKEN_NEWLINE: "NEWLINE",
 TOKEN_DOTDOT: "DOTDOT",
 TOKEN_MATCH: "MATCH", TOKEN_CASE: "CASE", TOKEN_ENUM: "ENUM",
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
	"match":  TOKEN_MATCH,
	"case":   TOKEN_CASE,
	"enum":   TOKEN_ENUM,
	"spawn":   TOKEN_SPAWN,
	"send":    TOKEN_SEND,
	"receive": TOKEN_RECEIVE,
	"ai":      TOKEN_AI,
	"cache":   TOKEN_CACHE,
	"prompt":  TOKEN_PROMPT,
	"channel": TOKEN_CHANNEL,
	"pure":    TOKEN_PURE,
	"effects": TOKEN_EFFECTS,
	"snapshot": TOKEN_SNAPSHOT,
	"restore":  TOKEN_RESTORE,
	"replay":   TOKEN_REPLAY,
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

// MatchExpression represents: match expr { case pattern -> expr, ... }
type MatchExpression struct {
	Token  Token
	Value  Expr       // expression to match against
	Cases  []MatchCase
}

type MatchCase struct {
	Token   Token
	Pattern Expr       // pattern to match (IntLiteral, StringLiteral, Identifier, etc.)
	Body    Expr       // expression to evaluate if matched
}

// EnumDef represents: enum Name { Variant1, Variant2(args...) }
type EnumDef struct {
	Token    Token
	Name     string
	Variants []EnumVariant
}

type EnumVariant struct {
	Token  Token
	Name   string
	Arity  int  // number of payload fields (0 = simple variant)
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

// DotExpr represents: expr.field (property/variant access)
type DotExpr struct {
	Token  Token
	Object Expr
	Field  string
}

func (DotExpr) exprNode()      {}
func (ArrayLiteral) exprNode()  {}
func (MapLiteral) exprNode()    {}
func (MatchExpression) exprNode() {}

// --- Actor Model AST nodes ---

// SpawnExpr represents: spawn fn_name(args)
type SpawnExpr struct {
	Token  Token
	Callee Expr
	Args   []Expr
}

// SendExpr represents: send(channel, value)
type SendExpr struct {
	Token   Token
	Channel Expr
	Value   Expr
}

// ReceiveExpr represents: receive(channel)
type ReceiveExpr struct {
	Token   Token
	Channel Expr
}

// ChannelExpr represents: channel() — creates a new channel
type ChannelExpr struct {
	Token Token
}

// Implement exprNode for actor expression types
func (SpawnExpr) exprNode()    {}
func (SendExpr) exprNode()     {}
func (ReceiveExpr) exprNode()  {}
func (ChannelExpr) exprNode()  {}

// EffectAnnot represents a single effect annotation (e.g., "io", "netio", "pure").
type EffectAnnot struct {
	Token Token
	Name  string // "pure", "io", "netio"
}

// SnapshotExpr represents: snapshot() — captures the entire VM state
type SnapshotExpr struct {
	Token Token
}

// RestoreExpr represents: restore(index) — restores VM state without replay
type RestoreExpr struct {
	Token Token
	Index Expr
}

// ReplayExpr represents: replay(index) — restores VM state and re-executes
type ReplayExpr struct {
	Token Token
	Index Expr
}

func (SnapshotExpr) exprNode() {}
func (RestoreExpr) exprNode()  {}
func (ReplayExpr) exprNode()   {}

// PipelineExpr represents: expr |> fn1(args...) |> fn2(args...) |> ...
// The left-hand expression is the first step; subsequent steps are piped.
type PipelineExpr struct {
	Token Token
	Expr  Expr           // the expression being piped (left-hand side)
	Steps []PipelineStep // each |> step
}

// PipelineStep represents one step in a pipeline: |> fn_name args...
// The piped value becomes the first argument.
type PipelineStep struct {
	Token Token
	Name  string // function name
	Args  []Expr // additional arguments (piped value is prepended at compile time)
}

func (PipelineExpr) exprNode() {}

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

// IndexAssign represents index assignment: ident[expr] = expr
type IndexAssign struct {
	Token  Token
	Object Expr
	Index  Expr
	Value  Expr
}
func (IndexAssign) stmtNode() {}

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
	Token      Token
	Name       string
	Params     []Param
	Body       Block
	Effects    []EffectAnnot // effect annotations (empty = no effects declared)
	ReturnType string        // optional return type annotation
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
func (EnumDef) stmtNode()     {}
func (AIFuncDef) stmtNode()   {}

// AIFuncDef represents: ai fn name(args) -> type { prompt: "..." [model: "..."] [cache: true] }
type AIFuncDef struct {
	Token      Token
	Name       string
	Params     []Param
	ReturnType string // optional return type annotation
	Prompt     string // the prompt template
	Model      string // optional model specification
	Cache      bool   // whether to cache responses
}

// AICallExpr represents calling an AI function (handled at runtime)
type AICallExpr struct {
	Token Token
	Name  string
	Args  []Expr // arguments for prompt template
}

func (AICallExpr) exprNode() {}
