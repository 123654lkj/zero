package compiler

import (
	"testing"
)

func parse(source string) []Stmt {
	p := NewParser(source)
	return p.Parse()
}

func TestParserVarDecl(t *testing.T) {
	stmts := parse("let x = 5")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	vd, ok := stmts[0].(*VarDecl)
	if !ok {
		t.Fatalf("expected VarDecl, got %T", stmts[0])
	}
	if vd.Name != "x" {
		t.Errorf("expected name 'x', got '%s'", vd.Name)
	}
	intLit, ok := vd.Value.(*IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral, got %T", vd.Value)
	}
	if intLit.Value != 5 {
		t.Errorf("expected value 5, got %d", intLit.Value)
	}
}

func TestParserFuncCall(t *testing.T) {
	stmts := parse("print(42)")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	es, ok := stmts[0].(*ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", stmts[0])
	}
	call, ok := es.Expr.(*CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", es.Expr)
	}
	callee, ok := call.Callee.(*Identifier)
	if !ok {
		t.Fatalf("expected Identifier callee, got %T", call.Callee)
	}
	if callee.Name != "print" {
		t.Errorf("expected callee 'print', got '%s'", callee.Name)
	}
	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}
	arg, ok := call.Args[0].(*IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral arg, got %T", call.Args[0])
	}
	if arg.Value != 42 {
		t.Errorf("expected arg value 42, got %d", arg.Value)
	}
}

func TestParserBinaryExpr(t *testing.T) {
	stmts := parse("2 + 3")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	es, ok := stmts[0].(*ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", stmts[0])
	}
	bin, ok := es.Expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", es.Expr)
	}
	if bin.Op != TOKEN_PLUS {
		t.Errorf("expected PLUS, got %s", bin.Op)
	}
	left, ok := bin.Left.(*IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral left, got %T", bin.Left)
	}
	if left.Value != 2 {
		t.Errorf("expected left 2, got %d", left.Value)
	}
	right, ok := bin.Right.(*IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral right, got %T", bin.Right)
	}
	if right.Value != 3 {
		t.Errorf("expected right 3, got %d", right.Value)
	}
}

func TestParserIfElse(t *testing.T) {
	stmts := parse("if x > 5 {\n}\nelse {\n}")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	ifs, ok := stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", stmts[0])
	}
	bin, ok := ifs.Cond.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr condition, got %T", ifs.Cond)
	}
	if bin.Op != TOKEN_GT {
		t.Errorf("expected GT, got %s", bin.Op)
	}
	if len(ifs.Else.Stmts) != 0 {
		// Just checking else block is non-nil (it's a Block, may be empty)
	}
}

func TestParserWhile(t *testing.T) {
	stmts := parse("while true {\n}")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	ws, ok := stmts[0].(*WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", stmts[0])
	}
	bl, ok := ws.Cond.(*BoolLiteral)
	if !ok {
		t.Fatalf("expected BoolLiteral condition, got %T", ws.Cond)
	}
	if !bl.Value {
		t.Error("expected true")
	}
	if len(ws.Body.Stmts) != 0 {
		t.Errorf("expected empty body, got %d statements", len(ws.Body.Stmts))
	}
}

func TestParserFuncDef(t *testing.T) {
	stmts := parse("fn add(a, b) {\nreturn a + b\n}")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	fd, ok := stmts[0].(*FuncDef)
	if !ok {
		t.Fatalf("expected FuncDef, got %T", stmts[0])
	}
	if fd.Name != "add" {
		t.Errorf("expected name 'add', got '%s'", fd.Name)
	}
	if len(fd.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fd.Params))
	}
	if fd.Params[0].Name != "a" {
		t.Errorf("expected param 'a', got '%s'", fd.Params[0].Name)
	}
	if fd.Params[1].Name != "b" {
		t.Errorf("expected param 'b', got '%s'", fd.Params[1].Name)
	}
	if len(fd.Body.Stmts) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(fd.Body.Stmts))
	}
	ret, ok := fd.Body.Stmts[0].(*ReturnStmt)
	if !ok {
		t.Fatalf("expected ReturnStmt, got %T", fd.Body.Stmts[0])
	}
	bin, ok := ret.Value.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr in return, got %T", ret.Value)
	}
	if bin.Op != TOKEN_PLUS {
		t.Errorf("expected PLUS in return, got %s", bin.Op)
	}
}

func TestParserArrayLiteral(t *testing.T) {
	stmts := parse("[1, 2, 3]")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	es, ok := stmts[0].(*ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", stmts[0])
	}
	arr, ok := es.Expr.(*ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral, got %T", es.Expr)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
	for i, expected := range []int64{1, 2, 3} {
		el, ok := arr.Elements[i].(*IntLiteral)
		if !ok {
			t.Fatalf("expected IntLiteral at index %d, got %T", i, arr.Elements[i])
		}
		if el.Value != expected {
			t.Errorf("expected element %d to be %d, got %d", i, expected, el.Value)
		}
	}
}

func TestParserIndexExpr(t *testing.T) {
	stmts := parse("a[0]")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	es, ok := stmts[0].(*ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", stmts[0])
	}
	idx, ok := es.Expr.(*IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", es.Expr)
	}
	obj, ok := idx.Object.(*Identifier)
	if !ok {
		t.Fatalf("expected Identifier object, got %T", idx.Object)
	}
	if obj.Name != "a" {
		t.Errorf("expected object 'a', got '%s'", obj.Name)
	}
	index, ok := idx.Index.(*IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral index, got %T", idx.Index)
	}
	if index.Value != 0 {
		t.Errorf("expected index 0, got %d", index.Value)
	}
}

func TestParserPrecedence(t *testing.T) {
	stmts := parse("2 + 3 * 4")
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	es, ok := stmts[0].(*ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", stmts[0])
	}
	// Should parse as 2 + (3 * 4), so top-level is PLUS
	bin, ok := es.Expr.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", es.Expr)
	}
	if bin.Op != TOKEN_PLUS {
		t.Errorf("expected top-level PLUS, got %s", bin.Op)
	}
	// Left should be 2
	left, ok := bin.Left.(*IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral left, got %T", bin.Left)
	}
	if left.Value != 2 {
		t.Errorf("expected left 2, got %d", left.Value)
	}
	// Right should be 3 * 4
	right, ok := bin.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr right, got %T", bin.Right)
	}
	if right.Op != TOKEN_STAR {
		t.Errorf("expected right STAR, got %s", right.Op)
	}
	rLeft, ok := right.Left.(*IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral right-left, got %T", right.Left)
	}
	if rLeft.Value != 3 {
		t.Errorf("expected right-left 3, got %d", rLeft.Value)
	}
	rRight, ok := right.Right.(*IntLiteral)
	if !ok {
		t.Fatalf("expected IntLiteral right-right, got %T", right.Right)
	}
	if rRight.Value != 4 {
		t.Errorf("expected right-right 4, got %d", rRight.Value)
	}
}

func TestParserNestedIfElse(t *testing.T) {
	src := `if a > 1 {
if b > 2 {
}
else {
}
}
else {
}`
	stmts := parse(src)
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	outer, ok := stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", stmts[0])
	}
	if len(outer.Then.Stmts) != 1 {
		t.Fatalf("expected 1 then-stmt, got %d", len(outer.Then.Stmts))
	}
	inner, ok := outer.Then.Stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected inner IfStmt, got %T", outer.Then.Stmts[0])
	}
	if inner.Cond == nil {
		t.Error("inner condition should not be nil")
	}
	if len(outer.Else.Stmts) != 0 {
		t.Errorf("expected empty else block, got %d stmts", len(outer.Else.Stmts))
	}
}
