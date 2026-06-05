// Package compiler implements the Zero language code generator that transforms
// a parsed AST into bytecode stored in a Chunk.
package compiler

import (
	"fmt"
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
	chunk      *chunk.Chunk
	line       int    // current source line for emitted instructions
	locals     []local // local variables (function parameters)
	aiFunctions map[string]int // AI function name -> index in chunk.AIFunctions
	funcEffects  map[string][]string // function name -> list of effect names
	pureScope    bool               // true when inside a pure function body
	pipeTempCount int              // counter for pipeline temporary global names
}

// NewCompiler creates a new Compiler with an empty chunk.
func NewCompiler() *Compiler {
	return &Compiler{
		chunk:       chunk.NewChunk(),
		aiFunctions: make(map[string]int),
		funcEffects: make(map[string][]string),
	}
}

// PureFuncs returns the map of function names to their effect annotations.
// Useful for external tooling and tests.
func (c *Compiler) PureFuncs() map[string][]string {
	return c.funcEffects
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
	case *DotExpr:
		return n.Token.Line
	case *ArrayLiteral:
		return n.Token.Line
	case *MapLiteral:
		return n.Token.Line
	case *MatchExpression:
		return n.Token.Line
	case *AICallExpr:
		return n.Token.Line
	case *SpawnExpr:
		return n.Token.Line
	case *SendExpr:
		return n.Token.Line
	case *ReceiveExpr:
		return n.Token.Line
	case *ChannelExpr:
		return n.Token.Line
	case *PipelineExpr:
		return n.Token.Line
	case *SnapshotExpr:
		return n.Token.Line
	case *RestoreExpr:
		return n.Token.Line
	case *ReplayExpr:
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
	case *EnumDef:
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

	case *AIFuncDef:
		c.compileAIFuncDef(s)

	case *Block:
		c.compileStatements(s.Stmts)

	case *EnumDef:
		c.compileEnumDef(s)
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
	savedPureScope := c.pureScope
	c.locals = nil
	for i, param := range fn.Params {
		c.locals = append(c.locals, local{name: param.Name, slot: i})
	}

	// Determine if this function is pure
	isPure := false
	for _, eff := range fn.Effects {
		if eff.Name == "pure" {
			isPure = true
			break
		}
	}
	c.pureScope = isPure

	// Compile function body
	c.compileStatements(fn.Body.Stmts)

	// Restore pure scope
	c.pureScope = savedPureScope

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

	// Track effects for this function
	effectNames := make([]string, len(fn.Effects))
	for i, e := range fn.Effects {
		effectNames[i] = e.Name
	}
	c.funcEffects[fn.Name] = effectNames

	// Define the function name as a global holding a string value.
	// The VM matches this string to chunk.Functions at call time.
	idx := c.addConstant(value.StringValue(fn.Name))
	c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
	nameIdx := c.addName(fn.Name)
	c.emitWordWithOperand(byte(opcode.OP_DEF_GLOBAL), uint16(nameIdx))
}

// ---------------------------------------------------------------------------
// AI function definition compilation
// ---------------------------------------------------------------------------

func (c *Compiler) compileAIFuncDef(ai *AIFuncDef) {
	// Store AI function metadata in the chunk
	paramNames := make([]string, len(ai.Params))
	for i, p := range ai.Params {
		paramNames[i] = p.Name
	}
	aiIdx := c.chunk.AddAIFunction(chunk.AIFunctionMeta{
		Name:       ai.Name,
		ParamNames: paramNames,
		Prompt:     ai.Prompt,
		Model:      ai.Model,
		Cache:      ai.Cache,
		ReturnType: ai.ReturnType,
	})

	// Track this AI function for call detection
	c.aiFunctions[ai.Name] = aiIdx

	// Define the function name as a global holding the AI function name string.
	// The VM will use OP_AI_CALL to dispatch to the AI backend.
	idx := c.addConstant(value.StringValue(ai.Name))
	c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
	nameIdx := c.addName(ai.Name)
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
		// Check if this is an AI function call
		if id, ok := e.Callee.(*Identifier); ok {
			if aiIdx, isAI := c.aiFunctions[id.Name]; isAI {
				// AI function call: compile arguments, then emit OP_AI_CALL
				for _, arg := range e.Args {
					c.compileExpression(arg)
				}
				c.emitWordWithOperand(byte(opcode.OP_AI_CALL), uint16(aiIdx))
				return
			}
			// Effect compatibility check
			if c.pureScope {
				if effects, has := c.funcEffects[id.Name]; has {
					for _, eff := range effects {
						if eff == "io" || eff == "netio" || eff == "write" {
							// TODO: Turn this into a compile error in Phase 3
							// For now, emit a compile-time warning
							fmt.Printf("WARNING: pure function calls impure function '%s' (effect: %s) at line %d\n",
								id.Name, eff, e.Token.Line)
						}
					}
				}
			}
		}
		// Check if this is an enum constructor: Color.Red(args...)
		if dot, ok := e.Callee.(*DotExpr); ok {
			if id, ok := dot.Object.(*Identifier); ok {
				fullName := id.Name + "." + dot.Field
				// Compile as array: [variantName, arg1, arg2, ...]
				idx := c.addConstant(value.StringValue(fullName))
				c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
				for _, arg := range e.Args {
					c.compileExpression(arg)
				}
				// Create array with (1 + argCount) elements
				total := 1 + len(e.Args)
				countIdx := c.addConstant(value.IntValue(int64(total)))
				c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(countIdx))
				c.emitOp(opcode.OP_ARRAY_NEW)
				return
			}
		}
		// Normal function call
		c.compileExpression(e.Callee)
		for _, arg := range e.Args {
			c.compileExpression(arg)
		}
		c.emitByteWithOperand(byte(opcode.OP_CALL), byte(len(e.Args)))

	case *IndexExpr:
		c.compileExpression(e.Object)
		c.compileExpression(e.Index)
		// Determine if array or map at compile time we emit both;
		// the VM checks the type at runtime.
		c.emitOp(opcode.OP_ARRAY_GET) // placeholder — VM dispatches by type
	case *DotExpr:
		// Color.Red -> load global "Color.Red"
		if id, ok := e.Object.(*Identifier); ok {
			fullName := id.Name + "." + e.Field
			idx := c.addName(fullName)
			c.emitWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(idx))
		} else {
			// Generic dot access: compile object, then push field name and do map get
			c.compileExpression(e.Object)
			idx := c.addConstant(value.StringValue(e.Field))
			c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
			c.emitOp(opcode.OP_MAP_GET)
 		}

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

	case *MatchExpression:
		c.compileMatch(e)

	case *SpawnExpr:
		// Push the function value (callee) and arguments onto the stack,
		// then emit OP_SPAWN which will launch a goroutine to call the function.
		c.compileExpression(e.Callee)
		for _, arg := range e.Args {
			c.compileExpression(arg)
		}
		// OP_SPAWN: pops args + function, launches goroutine, pushes pid
		c.emitByteWithOperand(byte(opcode.OP_SPAWN), byte(len(e.Args)))

	case *SendExpr:
		// Compile channel and value, then emit OP_SEND
		c.compileExpression(e.Channel)
		c.compileExpression(e.Value)
		c.emitOp(opcode.OP_SEND)

	case *ReceiveExpr:
		// Compile channel, then emit OP_RECEIVE
		c.compileExpression(e.Channel)
		c.emitOp(opcode.OP_RECEIVE)

	case *ChannelExpr:
		// Emit OP_CHAN_NEW to create a new channel
		c.emitOp(opcode.OP_CHAN_NEW)

	case *SnapshotExpr:
		// Emit OP_SNAPSHOT to capture the full VM state.
		// The opcode pushes the snapshot index onto the stack.
		c.emitOp(opcode.OP_SNAPSHOT)

	case *RestoreExpr:
		// Compile the index expression, then emit OP_RESTORE.
		c.compileExpression(e.Index)
		c.emitOp(opcode.OP_RESTORE)

	case *ReplayExpr:
		// Compile the index expression, then emit OP_REPLAY.
		c.compileExpression(e.Index)
		c.emitOp(opcode.OP_REPLAY)

	case *PipelineExpr:
		// Pipeline: expr |> fn1(a, b) |> fn2(c)
		// Compiles as nested calls using temp globals for stack ordering.
		//
		// Strategy per step:
		//   1. Result (piped value) is on the stack
		//   2. DEF_GLOBAL to save to temp (pops value from stack)
		//   3. LOAD_GLOBAL for the function name
		//   4. LOAD_GLOBAL for the temp (piped value goes below fn)
		//   5. Push extra args
		//   6. CALL(1 + extra_arg_count)

		// Compile the base expression
		c.compileExpression(e.Expr)

		for _, step := range e.Steps {
			c.line = step.Token.Line

			// Save piped value to a temporary global (DEF_GLOBAL pops)
			tmpName := fmt.Sprintf("__pipe_tmp_%d", c.pipeTempCount)
			c.pipeTempCount++
			tmpIdx := c.addName(tmpName)
			c.emitWordWithOperand(byte(opcode.OP_DEF_GLOBAL), uint16(tmpIdx))

			// Load the function by name
			fnIdx := c.addName(step.Name)
			c.emitWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(fnIdx))

			// Load the piped value from temp global (becomes first arg)
			c.emitWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(tmpIdx))

			// Push extra arguments
			for _, arg := range step.Args {
				c.compileExpression(arg)
			}

			// Call with 1 (piped) + len(extra) arguments
			c.emitByteWithOperand(byte(opcode.OP_CALL), byte(1+len(step.Args)))

			// Result is now on stack; will be piped to next step or be final result
		}

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

// ---------------------------------------------------------------------------
// Enum compilation
// ---------------------------------------------------------------------------

// compileEnumDef compiles an enum definition.
// Each variant becomes a global variable holding a tagged string like "EnumName.VariantName".
func (c *Compiler) compileEnumDef(e *EnumDef) {
	for _, variant := range e.Variants {
		variantName := e.Name + "." + variant.Name
		idx := c.addConstant(value.StringValue(variantName))
		c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
		nameIdx := c.addName(variantName)
		c.emitWordWithOperand(byte(opcode.OP_DEF_GLOBAL), uint16(nameIdx))
	}
}

// ---------------------------------------------------------------------------
// Match expression compilation
// ---------------------------------------------------------------------------

// compileMatch compiles a match expression using a chain of DUP + LOAD pattern + EQ + JMP_IF.
//
// For each case (except the last which acts as a default/catch-all):
//   DUP the match value, LOAD the pattern, EQ, JMP_IF to the case body
//
// If the pattern is an Identifier and it's the last case, treat it as a
// catch-all (always matches).
func (c *Compiler) compileMatch(m *MatchExpression) {
	// Compile the match value — this stays on the stack throughout
	c.compileExpression(m.Value)

	// Each non-catch-all case:
	//   1. DUP the match value
	//   2. Load pattern, EQ → stack: [..., match_value, bool]
	//   3. JMP_IF to case body (Peek, doesn't pop)
	//   4. Fall-through if not matched: POP the bool
	//
	// On match (JMP_IF taken):
	//   Stack: [..., match_value, true]
	//   POP the true, POP the match_value, compile body → result on stack
	//   JMP past remaining cases
	//
	// Catch-all (identifier in last position):
	//   POP match_value, compile body → result, JMP to end

	// Track "jump to end" positions that need patching
	var endJumps []int
	// Track "jump to next case" positions
	var nextCaseJumps []int

	for i, mc := range m.Cases {
		isLast := (i == len(m.Cases)-1)
		_, isIdentifier := mc.Pattern.(*Identifier)

		// If it's a catch-all (Identifier in last position), skip the comparison
		if isLast && isIdentifier {
			// POP the match value
			c.emitOp(opcode.OP_POP)
			// Compile body expression → result on stack
			c.compileExpression(mc.Body)
			// Jump to end (will be patched)
			c.emitWordWithOperand(byte(opcode.OP_JMP), 0)
			endJumps = append(endJumps, c.chunk.Len()-2)
			continue
		}

		// DUP the match value for comparison
		c.emitOp(opcode.OP_DUP)

		// Load the pattern value
		// Compile pattern — handle enum constructors specially
		switch pat := mc.Pattern.(type) {
		case *DotExpr:
			// Color.Red → load string "Color.Red"
			if id, ok := pat.Object.(*Identifier); ok {
				fullName := id.Name + "." + pat.Field
				idx := c.addConstant(value.StringValue(fullName))
				c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
			} else {
				c.compileExpression(mc.Pattern)
			}
		case *CallExpr:
			// Result.Ok(v) → extract variant name, ignore payload
			if dot, ok := pat.Callee.(*DotExpr); ok {
				if id, ok := dot.Object.(*Identifier); ok {
					fullName := id.Name + "." + dot.Field
					idx := c.addConstant(value.StringValue(fullName))
					c.emitWordWithOperand(byte(opcode.OP_PUSH), uint16(idx))
				} else {
					c.compileExpression(pat.Callee)
				}
			} else {
				c.compileExpression(mc.Pattern)
			}
		default:
			c.compileExpression(mc.Pattern)
		}

		// Compare (consumes DUP'd value and pattern, pushes bool)
		c.emitOp(opcode.OP_EQ)

		// Jump to the case body if matched (JMP_IF peeks, doesn't pop)
		c.emitWordWithOperand(byte(opcode.OP_JMP_IF), 0)
		jmpToBodyPos := c.chunk.Len() - 2

		// Not matched: fall through — pop the bool
		c.emitOp(opcode.OP_POP)
		// Jump to next case (will be patched)
		c.emitWordWithOperand(byte(opcode.OP_JMP), 0)
		nextCaseJumps = append(nextCaseJumps, c.chunk.Len()-2)

		// Patch jump to body
		c.patchJump(jmpToBodyPos)

		// === Matched case body ===
		// Stack: [..., match_value, true]
		// POP the true
		c.emitOp(opcode.OP_POP)
		// POP the original match value (no longer needed)
		c.emitOp(opcode.OP_POP)
		// Compile body expression → result on stack
		c.compileExpression(mc.Body)
		// Jump past remaining cases (to end)
		c.emitWordWithOperand(byte(opcode.OP_JMP), 0)
		endJumps = append(endJumps, c.chunk.Len()-2)
	}

	// === End label ===
	endOffset := c.chunk.Len()

	// Patch all "jump to next case" to point to the end
	for _, jmpPos := range nextCaseJumps {
		offset := int16(endOffset - (jmpPos + 2))
		c.chunk.Code[jmpPos] = byte(offset >> 8)
		c.chunk.Code[jmpPos+1] = byte(offset)
	}

	// Patch all "jump to end" to point to the end
	for _, jmpPos := range endJumps {
		offset := int16(endOffset - (jmpPos + 2))
		c.chunk.Code[jmpPos] = byte(offset >> 8)
		c.chunk.Code[jmpPos+1] = byte(offset)
	}
}
