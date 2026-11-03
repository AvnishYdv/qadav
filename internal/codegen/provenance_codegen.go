package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genProvenanceDecl(p *ast.ProvenanceDecl) {
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_mctp_init(i32 1000)\n", resTemp)

	if p.Name != "" {
		ptr := fmt.Sprintf("%%%s.ptr", p.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i8*",
			qtype:    types.Provenance,
		}
		g.vars[p.Name] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
	}
}

func (g *Generator) genCertifyStmt(s *ast.CertifyStmt) {
	pVal := g.genExpr(s.Provenance)
	fmt.Fprintf(&g.sb, "  call void @qadav_mctp_certify(i8* %s)\n", pVal.Ref)
}

func (g *Generator) genVerifyTrainingExpr(e *ast.VerifyTrainingExpr) Value {
	var certRef = "null"
	if e.Certificate != nil {
		cVal := g.genExpr(e.Certificate)
		certRef = cVal.Ref
	}
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_mctp_verify(i8* %s, i8* null, i64 0)\n", resTemp, certRef)
	return Value{
		Type:  "i32",
		Ref:   resTemp,
		QType: types.Logic,
	}
}
