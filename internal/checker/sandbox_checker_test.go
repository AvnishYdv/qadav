package checker

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func TestSandboxChecker_Valid(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}
	c.globalScope.Define(&Symbol{Name: "auth_boundary", Type: types.Concept})
	c.tools["safe_tool"] = &ToolInfo{Name: "safe_tool"}

	// 1. Sandbox create with explicit capacity
	create := &ast.SandboxCreateStmt{
		Pos_:     pos,
		Name:     "untrusted_sandbox",
		Boundary: "auth_boundary",
		Capacity: &ast.IntLit{Pos_: pos, Value: 10},
	}
	c.checkSandboxCreateStmt(create)
	if len(c.Errors()) != 0 {
		t.Fatalf("expected 0 errors, got: %v", c.Errors())
	}

	// 1b. Sandbox create with default capacity (nil capacity)
	createNilCap := &ast.SandboxCreateStmt{
		Pos_:     pos,
		Name:     "untrusted_sandbox_default",
		Boundary: "auth_boundary",
		Capacity: nil,
	}
	c.checkSandboxCreateStmt(createNilCap)
	if len(c.Errors()) != 0 {
		t.Fatalf("expected 0 errors, got: %v", c.Errors())
	}

	// 2. Sandbox run
	run := &ast.SandboxRunStmt{
		Pos_:     pos,
		Sandbox:  "untrusted_sandbox",
		ToolName: "safe_tool",
		Args:     []ast.Expr{&ast.StringLit{Pos_: pos, Value: "valid_arg"}},
	}
	c.checkSandboxRunStmt(run)
	if len(c.Errors()) != 0 {
		t.Fatalf("expected 0 errors, got: %v", c.Errors())
	}

	// 3. Sandbox verify expr
	vexpr := &ast.SandboxVerifyExpr{
		Pos_:    pos,
		Sandbox: "untrusted_sandbox",
	}
	resType := c.checkSandboxVerifyExpr(vexpr)
	if resType == nil || resType.Kind != types.KindText {
		t.Fatalf("expected text return type, got: %v", resType)
	}

	// 4. Sandbox verify stmt with into
	vstmt := &ast.SandboxVerifyStmt{
		Pos_:    pos,
		Sandbox: "untrusted_sandbox",
		Into:    "attestation",
	}
	c.checkSandboxVerifyStmt(vstmt)
	if len(c.Errors()) != 0 {
		t.Fatalf("expected 0 errors, got: %v", c.Errors())
	}

	// 4b. Sandbox verify stmt without into
	vstmtNoInto := &ast.SandboxVerifyStmt{
		Pos_:    pos,
		Sandbox: "untrusted_sandbox",
		Into:    "",
	}
	c.checkSandboxVerifyStmt(vstmtNoInto)
	if len(c.Errors()) != 0 {
		t.Fatalf("expected 0 errors, got: %v", c.Errors())
	}
}

func TestSandboxChecker_EmptyNames(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}

	// Empty name in create
	c.checkSandboxCreateStmt(&ast.SandboxCreateStmt{Pos_: pos, Name: ""})
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for empty name in create")
	}

	// Empty name in run
	c.checkSandboxRunStmt(&ast.SandboxRunStmt{Pos_: pos, Sandbox: "", ToolName: "t"})
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for empty sandbox in run")
	}

	c.checkSandboxRunStmt(&ast.SandboxRunStmt{Pos_: pos, Sandbox: "sb", ToolName: ""})
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for empty tool in run")
	}

	// Empty name in verify
	c.checkSandboxVerifyStmt(&ast.SandboxVerifyStmt{Pos_: pos, Sandbox: ""})
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for empty sandbox in verify stmt")
	}

	c.checkSandboxVerifyExpr(&ast.SandboxVerifyExpr{Pos_: pos, Sandbox: ""})
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for empty sandbox in verify expr")
	}
}

func TestSandboxChecker_UndefinedBoundary(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}

	create := &ast.SandboxCreateStmt{
		Pos_:     pos,
		Name:     "untrusted_sandbox",
		Boundary: "missing_boundary",
		Capacity: &ast.IntLit{Pos_: pos, Value: 10},
	}
	c.checkSandboxCreateStmt(create)
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for undefined boundary")
	}
}

func TestSandboxChecker_InvalidCapacity(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}
	c.globalScope.Define(&Symbol{Name: "auth_boundary", Type: types.Concept})

	// Non-positive capacity
	create := &ast.SandboxCreateStmt{
		Pos_:     pos,
		Name:     "untrusted_sandbox",
		Boundary: "auth_boundary",
		Capacity: &ast.IntLit{Pos_: pos, Value: 0},
	}
	c.checkSandboxCreateStmt(create)
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for non-positive capacity")
	}

	// String capacity
	createStr := &ast.SandboxCreateStmt{
		Pos_:     pos,
		Name:     "untrusted_sandbox2",
		Boundary: "auth_boundary",
		Capacity: &ast.StringLit{Pos_: pos, Value: "invalid"},
	}
	c.checkSandboxCreateStmt(createStr)
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for string capacity")
	}
}

func TestSandboxChecker_UndefinedSandboxAndTool(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}

	// Undefined sandbox in run
	run := &ast.SandboxRunStmt{
		Pos_:     pos,
		Sandbox:  "nonexistent_sb",
		ToolName: "safe_tool",
	}
	c.checkSandboxRunStmt(run)
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for undefined sandbox in run")
	}

	// Undefined sandbox in verify expr
	vexpr := &ast.SandboxVerifyExpr{
		Pos_:    pos,
		Sandbox: "nonexistent_sb",
	}
	c.checkSandboxVerifyExpr(vexpr)
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for undefined sandbox in verify expr")
	}

	// Undefined sandbox in verify stmt
	vstmt := &ast.SandboxVerifyStmt{
		Pos_:    pos,
		Sandbox: "nonexistent_sb",
		Into:    "res",
	}
	c.checkSandboxVerifyStmt(vstmt)
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for undefined sandbox in verify stmt")
	}

	// Undefined tool in run
	c.sandboxes["existing_sb"] = &SandboxInfo{Name: "existing_sb"}
	runNoTool := &ast.SandboxRunStmt{
		Pos_:     pos,
		Sandbox:  "existing_sb",
		ToolName: "unregistered_tool",
	}
	c.checkSandboxRunStmt(runNoTool)
	if len(c.Errors()) == 0 {
		t.Fatal("expected error for unregistered tool in run")
	}
}
