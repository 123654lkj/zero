// Package compiler implements the Zero language code generator that transforms
// a parsed AST into bytecode stored in a Chunk.
package compiler

import (
	"github.com/123654lkj/zero/go/chunk"
	"github.com/123654lkj/zero/go/opcode"
	"github.com/123654lkj/zero/go/value"
)

// local represents a local variable (function parameter) tracked during compilation.
type local struct {
	name string
	slot int // 0–3 for function parameters
}

// Compiler transforms a parsed AST into bytecode.
type Compiler struct {
	chunk  *chunk.Chunk
	line   int    // current source line for emitted instructions
	locals []local // local variables (function parameters)
}

// NewCompiler creates a new Compiler with an empty chunk.
func NewCompiler() *Compiler {
	return &Compiler{chunk: chunk.NewChunk()}
}

// Compile compiles Zero source code into a bytecode Chunk.
func (c *Compiler) Compile(source string) *chunk.Chunk {
	p := NewParser(source)
	stmts := p.Parse()
	c.compileStatements(stmts)
	c.emitHalt()
	return c.chunk
}

// ---------------------------------------------------------------------------
// Emit helpers
// ---------------------------------------------------------------------------

func (c *Compiler) emitByte(op byte)                        { c.chunk.WriteByte(op, c.line) }
func (c *Compiler) emitByteWithOperand(op byte, operand byte)  { c.chunk.WriteByteWithOperand(op, operand, c.line) }
func (c *Compiler) emitWordWithOperand(op byte, operand uint16) { c.chunk.WriteWordWithOperand(op, operand, c.line) }
func (c *Compiler) emitOp(op opcode.Opcode)                 { c.emitByte(op.Byte()) }
func (c *Compiler) emitHalt()                               { c.emitOp(opcode.OP_HALT) }
func (c *Compiler) addConstant(v value.Value) int           { return c.chunk.AddConstant(v) }
func (c *Compiler) addName(name string) int                 { return c.chunk.AddName(name) }

// ---------------------------------------------------------------------------
// Line tracking helpers
// ---------------------------------------------------------------------------

func lineOfExpr(e Expr) int {
	switch n := e.(type) {
	case *IntLiteral:
		return n.Token.Line
	case *FloatLiteral:
		return n.Token.Line
	case *StringLiteral:
		return n.Token.Line
	case *BoolLiteral:
		return n.Token.Line
	case *NilLiteral:
		return n.Token.Line
	case *Identifier:
		return n.Token.Line
	case *BinaryExpr:
		return n.Token.Line
	case *UnaryExpr:
		return n.Token.Line
	case *CallExpr:
		return n.Token.Line
	case *IndexExpr:
		return n.Token.Line
	case *ArrayLiteral:
		return n.Token.Line
	case *MapLiteral:
		return n.Token.Line
	}
	return 1
}

func lineOfStmt(s Stmt) int {
	switch n := s.(type) {
	case *ExprStmt:
		return lineOfExpr(n.Expr)
	case *VarDecl:
		return n.Token.Line
	case *Assign:
		return n.Token.Line
	case *ReturnStmt:
		return n.Token.Line
	case *IfStmt:
		return n.Token.Line
	case *WhileStmt:
		return n.Token.Line
	case *FuncDef:
		return n.Token.Line
	case *Block:
		return n.Token.Line
	case *IndexAssign:
		return n.Token.Line
	}
	return 1
}

// ---------------------------------------------------------------------------
// Jump patching
// ---------------------------------------------------------------------------

// patchJump patches a previously emitted JMP or JMP_IFN instruction whose
// 2-byte operand starts at offsetPos. The offset is set to jump to the
// current end of the chunk.
func (c *Compiler) patchJump(offsetPos int) {
	target := c.chunk.Len()
	offset := int16(target - (offsetPos + 2))
	c.chunk.Code[offsetPos] = byte(offset >> 8)
	c.chunk.Code[offsetPos+1] = byte(offset)
}

// ---------------------------------------------------------------------------
// Local variable resolution
// ---------------------------------------------------------------------------

// resolveLocal looks up a name in the current local scope and returns its
// slot index (0–3). Returns false if not found.
func (c *Compiler) resolveLocal(name string) (int, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if c.locals[i].name == name {
			return c.locals[i].slot, true
		}
	}
	return 0, false
}

// ---------------------------------------------------------------------------
// Statement compilation
// ---------------------------------------------------------------------------

func (c *Compiler) compileStatements(stmts []Stmt) {
	for _, stmt := range stmts {
		c.compileStatement(stmt)
	}
}

func (c *Compiler) compileStatement(stmt Stmt) {
	c.line = lineOfStmt(stmt)

	switch s := stmt.(type) {
	case *ExprStmt:
		c.line = lineOfExpr(s.Expr)
		c.compileExpression(s.Expr)
		c.emitOp(opcode.OP_POP)

	case *VarDecl:
		c.compileExpression(s.Value)
		idx := c.addName(s.Name)
		c.emitWordWithOperand(byte(opcode.OP_DEF_GLOBAL), uint16(idx))

	case *Assign:
		c.compileExpression(s.Value)
		if slot, ok := c.resolveLocal(s.Name); ok {
			c.emitStoreLocal(slot)
			c.emitOp(opcode.OP_POP)
		} else {
			idx := c.addName(s.Name)
			c.emitWordWithOperand(byte(opcode.OP_STORE_GLOBAL), uint16(idx))
			c.emitOp(opcode.OP_POP)
		}

	case *IndexAssign:
		c.compileExpression(s.Object)
		c.compileExpression(s.Index)
		c.compileExpression(s.Value)
		c.emitOp(opcode.OP_ARRAY_SET)

	case *ReturnStmt:
		if s.Value != nil {
			c.compileExpression(s.Value)
		} else {
			idx := c.addConstant(value.NilValue())
			c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
		}
		c.emitOp(opcode.OP_RET)

	case *IfStmt:
		c.compileExpression(s.Cond)
		// Jump to else block if condition is false
		c.emitWordWithOperand(byte(opcode.OP_JMP_IFN), 0)
		jmpIfnPos := c.chunk.Len() - 2
		// Pop condition (then path)
		c.emitOp(opcode.OP_POP)
		// Then block
		c.compileStatements(s.Then.Stmts)
		// Jump past else block
		c.emitWordWithOperand(byte(opcode.OP_JMP), 0)
		jmpPastElsePos := c.chunk.Len() - 2
		// Patch JMP_IFN to else block
		c.patchJump(jmpIfnPos)
		// Pop condition (else path)
		c.emitOp(opcode.OP_POP)
		// Else block
		c.compileStatements(s.Else.Stmts)
		// Patch JMP past else
		c.patchJump(jmpPastElsePos)

	case *WhileStmt:
		loopStart := c.chunk.Len()
		c.compileExpression(s.Cond)
		// Jump to exit if condition is false
		c.emitWordWithOperand(byte(opcode.OP_JMP_IFN), 0)
		jmpExitPos := c.chunk.Len() - 2
		// Pop condition (truthy path)
		c.emitOp(opcode.OP_POP)
		// Body
		c.compileStatements(s.Body.Stmts)
		// Jump back to loop start
		offset := int16(loopStart - (c.chunk.Len() + 3))
		c.emitWordWithOperand(byte(opcode.OP_JMP), uint16(offset))
		// Patch exit jump
		c.patchJump(jmpExitPos)
		// Pop condition (falsy path)
		c.emitOp(opcode.OP_POP)

	case *FuncDef:
		c.compileFuncDef(s)

	case *Block:
		c.compileStatements(s.Stmts)
	}
}

// emitLoadLocal emits the appropriate LOAD_n opcode for local slot n.
func (c *Compiler) emitLoadLocal(slot int) {
	switch slot {
	case 0:
		c.emitOp(opcode.OP_LOAD_0)
	case 1:
		c.emitOp(opcode.OP_LOAD_1)
	case 2:
		c.emitOp(opcode.OP_LOAD_2)
	case 3:
		c.emitOp(opcode.OP_LOAD_3)
	default:
		panic("compiler: too many local variables (max 4)")
	}
}

// emitStoreLocal emits the appropriate STORE_n opcode for local slot n.
func (c *Compiler) emitStoreLocal(slot int) {
	switch slot {
	case 0:
		c.emitOp(opcode.OP_STORE_0)
	case 1:
		c.emitOp(opcode.OP_STORE_1)
	case 2:
		c.emitOp(opcode.OP_STORE_2)
	case 3:
		c.emitOp(opcode.OP_STORE_3)
	default:
		panic("compiler: too many local variables (max 4)")
	}
}

// ---------------------------------------------------------------------------
// Function definition compilation
// ---------------------------------------------------------------------------

func (c *Compiler) compileFuncDef(fn *FuncDef) {
	// Emit JMP to skip over function body during normal execution
	c.emitWordWithOperand(byte(opcode.OP_JMP), 0)
	jmpSkipPos := c.chunk.Len() - 2

	// Record body start
	bodyStart := c.chunk.Len()

	// Save current locals and set up function parameter locals
	savedLocals := c.locals
	c.locals = nil
	for i, param := range fn.Params {
		c.locals = append(c.locals, local{name: param.Name, slot: i})
	}

	// Compile function body
	c.compileStatements(fn.Body.Stmts)

	// Ensure the body always ends with a return
	needsReturn := true
	if len(fn.Body.Stmts) > 0 {
		if _, isRet := fn.Body.Stmts[len(fn.Body.Stmts)-1].(*ReturnStmt); isRet {
			needsReturn = false
		}
	}
	if needsReturn {
		idx := c.addConstant(value.NilValue())
		c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
		c.emitOp(opcode.OP_RET)
	}

	bodyEnd := c.chunk.Len()

	// Restore saved locals
	c.locals = savedLocals

	// Patch the skip JMP to jump past the body
	c.patchJump(jmpSkipPos)

	// Add function metadata to the chunk
	c.chunk.AddFunction(chunk.FunctionMeta{
		Name:  fn.Name,
		Arity: len(fn.Params),
		Start: bodyStart,
		End:   bodyEnd,
	})

	// Define the function name as a global holding a string value.
	// The VM matches this string to chunk.Functions at call time.
	idx := c.addConstant(value.StringValue(fn.Name))
	c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
	nameIdx := c.addName(fn.Name)
	c.emitWordWithOperand(byte(opcode.OP_DEF_GLOBAL), uint16(nameIdx))
}

// ---------------------------------------------------------------------------
// Expression compilation
// ---------------------------------------------------------------------------

func (c *Compiler) compileExpression(expr Expr) {
	c.line = lineOfExpr(expr)

	switch e := expr.(type) {
	case *IntLiteral:
		idx := c.addConstant(value.IntValue(e.Value))
		c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))

	case *FloatLiteral:
		idx := c.addConstant(value.FloatValue(e.Value))
		c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))

	case *StringLiteral:
		idx := c.addConstant(value.StringValue(e.Value))
		c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))

	case *BoolLiteral:
		idx := c.addConstant(value.BoolValue(e.Value))
		c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))

	case *NilLiteral:
		idx := c.addConstant(value.NilValue())
		c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))

	case *Identifier:
		if slot, ok := c.resolveLocal(e.Name); ok {
			c.emitLoadLocal(slot)
		} else {
			idx := c.addName(e.Name)
			c.emitWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(idx))
		}

	case *BinaryExpr:
		c.compileExpression(e.Left)
		c.compileExpression(e.Right)
		c.compileBinaryOp(e.Op)

	case *UnaryExpr:
		c.compileExpression(e.Operand)
		switch e.Op {
		case TOKEN_MINUS:
			c.emitOp(opcode.OP_NEG)
		case TOKEN_NOT:
			c.emitOp(opcode.OP_NOT)
		}

	case *CallExpr:
		// Compile callee (must resolve to a function value)
		c.compileExpression(e.Callee)
		// Compile arguments
		for _, arg := range e.Args {
			c.compileExpression(arg)
		}
		// Emit CALL with argument count
		c.emitByteWithOperand(byte(opcode.OP_CALL), byte(len(e.Args)))

	case *IndexExpr:
		c.compileExpression(e.Object)
		c.compileExpression(e.Index)
		// Determine if array or map at compile time we emit both;
		// the VM checks the type at runtime.
		c.emitOp(opcode.OP_ARRAY_GET) // placeholder — VM dispatches by type

	case *ArrayLiteral:
		for _, elem := range e.Elements {
			c.compileExpression(elem)
		}
		// Push element count
		idx := c.addConstant(value.IntValue(int64(len(e.Elements))))
		c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
		c.emitOp(opcode.OP_ARRAY_NEW)

	case *MapLiteral:
		for i := range e.Keys {
			c.compileExpression(e.Keys[i])
			c.compileExpression(e.Values[i])
		}
		// Push pair count
		idx := c.addConstant(value.IntValue(int64(len(e.Keys))))
		c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
		c.emitOp(opcode.OP_MAP_NEW)
	}
}

// compileBinaryOp emits the bytecode for a binary operator.
func (c *Compiler) compileBinaryOp(op TokenType) {
	switch op {
	case TOKEN_PLUS:
		c.emitOp(opcode.OP_ADD)
	case TOKEN_MINUS:
		c.emitOp(opcode.OP_SUB)
	case TOKEN_STAR:
		c.emitOp(opcode.OP_MUL)
	case TOKEN_SLASH:
		c.emitOp(opcode.OP_DIV)
	case TOKEN_PERCENT:
		c.emitOp(opcode.OP_MOD)
	case TOKEN_EQ:
		c.emitOp(opcode.OP_EQ)
	case TOKEN_NEQ:
		c.emitOp(opcode.OP_NEQ)
	case TOKEN_LT:
		c.emitOp(opcode.OP_LT)
	case TOKEN_GT:
		c.emitOp(opcode.OP_GT)
	case TOKEN_LTE:
		c.emitOp(opcode.OP_LTE)
	case TOKEN_GTE:
		c.emitOp(opcode.OP_GTE)
	case TOKEN_AND:
		c.emitOp(opcode.OP_AND)
	case TOKEN_OR:
		c.emitOp(opcode.OP_OR)
	}
}
