package compiler

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/123654lkj/zero/go/vm"
)

// compileAndRun compiles the source, runs it via the VM, and captures stdout.
func compileAndRun(t *testing.T, source string) string {
	t.Helper()

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	var runErr interface{}
	func() {
		defer func() {
			runErr = recover()
		}()
		comp := NewCompiler()
		c := comp.Compile(source)
		v := vm.NewVM()
		v.RunChunk(c)
	}()

	w.Close()
	os.Stdout = old

	if runErr != nil {
		t.Fatalf("runtime error: %v", runErr)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrint42(t *testing.T) {
	got := compileAndRun(t, "print(42)")
	want := "42\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrint2Plus3(t *testing.T) {
	got := compileAndRun(t, "print(2 + 3)")
	want := "5\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrint2Mul3Plus1(t *testing.T) {
	got := compileAndRun(t, "print(2 * 3 + 1)")
	want := "7\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLetXPrint(t *testing.T) {
	got := compileAndRun(t, "let x = 10\nprint(x)")
	want := "10\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStringConcat(t *testing.T) {
	got := compileAndRun(t, `print("hello" + " world")`)
	want := "hello world\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLetAB(t *testing.T) {
	got := compileAndRun(t, "let a = 5\nlet b = 3\nprint(a + b)")
	want := "8\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIfElse(t *testing.T) {
	got := compileAndRun(t, "if 5 > 3 {\n  print(\"yes\")\n} else {\n  print(\"no\")\n}")
	want := "yes\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIfElseFalse(t *testing.T) {
	got := compileAndRun(t, "if 2 > 5 {\n  print(\"yes\")\n} else {\n  print(\"no\")\n}")
	want := "no\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWhileLoop(t *testing.T) {
	got := compileAndRun(t, "let i = 3\nwhile i > 0 {\n  print(i)\n  i = i - 1\n}")
	want := "3\n2\n1\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTypeFn(t *testing.T) {
	got := compileAndRun(t, "print(type(42))")
	want := "Int\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLenFn(t *testing.T) {
	got := compileAndRun(t, `print(len("hello"))`)
	want := "5\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestArrayPrint(t *testing.T) {
	got := compileAndRun(t, "print([1, 2, 3])")
	want := "[1 2 3]\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmptyIf(t *testing.T) {
	got := compileAndRun(t, "if true {\n  print(1)\n}")
	want := "1\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNestedExpressions(t *testing.T) {
	got := compileAndRun(t, "print((2 + 3) * 4)")
	want := "20\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAssign(t *testing.T) {
	got := compileAndRun(t, "let x = 1\nx = 2\nprint(x)")
	want := "2\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBoolAndNil(t *testing.T) {
	got := compileAndRun(t, "print(true)\nprint(false)\nprint(nil)")
	want := "true\nfalse\nnil\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestComparison(t *testing.T) {
	got := compileAndRun(t, "print(3 < 5)\nprint(3 > 5)\nprint(3 == 3)\nprint(3 != 5)")
	want := "true\nfalse\ntrue\ntrue\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
