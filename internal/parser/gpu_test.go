package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseGPUMatmulStmt(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    @gpu gpu_matmul a ⸴ b into c ·
    yield 0 ·
⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	stmt, ok := proc.Body.Statements[0].(*ast.GPUMatmulStmt)
	if !ok {
		t.Fatalf("expected GPUMatmulStmt, got %T", proc.Body.Statements[0])
	}
	if !stmt.IsAnnotated {
		t.Errorf("expected IsAnnotated == true")
	}
}

func TestParseGPUSoftmaxAndLayernorm(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    gpu_softmax x into y ·
    @gpu gpu_layernorm y into z audit audit_chain ·
    yield 0 ·
⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	s1, ok := proc.Body.Statements[0].(*ast.GPUSoftmaxStmt)
	if !ok || s1.IsAnnotated {
		t.Fatalf("unexpected stmt 1: %+v", proc.Body.Statements[0])
	}
	s2, ok := proc.Body.Statements[1].(*ast.GPULayernormStmt)
	if !ok || !s2.IsAnnotated || s2.Audit == nil {
		t.Fatalf("unexpected stmt 2: %+v", proc.Body.Statements[1])
	}
}

func TestParseAGKEAttestAndVerify(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    agke_attest my_chain ⦅ "matmul_f64" ⸴ in_t ⸴ out_t ⦆ ·
    agke_attest my_chain kernel ↦ "conv2d" input ↦ in_t output ↦ out_t ·
    agke_verify my_chain into verify_res ·
    corrupt_gpu_digest my_chain at ↦ 0 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if _, ok := proc.Body.Statements[0].(*ast.AGKEAttestStmt); !ok {
		t.Errorf("stmt 0 not AGKEAttestStmt: %T", proc.Body.Statements[0])
	}
	if _, ok := proc.Body.Statements[1].(*ast.AGKEAttestStmt); !ok {
		t.Errorf("stmt 1 not AGKEAttestStmt: %T", proc.Body.Statements[1])
	}
	if _, ok := proc.Body.Statements[2].(*ast.AGKEVerifyStmt); !ok {
		t.Errorf("stmt 2 not AGKEVerifyStmt: %T", proc.Body.Statements[2])
	}
	if _, ok := proc.Body.Statements[3].(*ast.CorruptGPUDigestStmt); !ok {
		t.Errorf("stmt 3 not CorruptGPUDigestStmt: %T", proc.Body.Statements[3])
	}
}

func TestParseGPUFallbackAndDefault(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    @gpu \gpu_matmul⦅a ⸴ b⦆ into c audit my_chain ·
    @gpu \gpu_softmax⦅c⦆ into s audit my_chain ·
    @gpu \gpu_layernorm⦅s⦆ into ln audit my_chain ·
    @gpu emit 42 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if len(proc.Body.Statements) != 5 {
		t.Fatalf("expected 5 statements, got %d", len(proc.Body.Statements))
	}
}

func TestParseGPUExpressions(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬2⸴ 2⟭ c ↦ gpu_matmul⦅a ⸴ b⦆ ·
    tensor⟨float64⟩⟬2⸴ 2⟭ s ↦ gpu_softmax⦅c⦆ ·
    tensor⟨float64⟩⟬2⸴ 2⟭ ln ↦ gpu_layernorm⦅s⦆ ·
    int64 v ↦ agke_verify my_chain ·
    yield 0 ·
⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if len(proc.Body.Statements) != 5 {
		t.Fatalf("expected 5 body statements, got %d", len(proc.Body.Statements))
	}
}
