package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genAADCertStmt(s *ast.AADCertStmt) {
	modelPtr := g.genStringConst("qec_decoder_model")
	modelLen := 17
	if s.Model != nil {
		if sLit, ok := s.Model.(*ast.StringLit); ok {
			modelPtr = g.genStringConst(sLit.Value)
			modelLen = len(sLit.Value)
		} else if ident, ok := s.Model.(*ast.Ident); ok {
			if slot, ok := g.vars[ident.Name]; ok {
				loaded := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
				modelPtr = loaded
				modelLen = 16
			}
		}
	}

	trainPtr := g.genStringConst("training_samples")
	trainLen := 16
	if s.TrainSet != nil {
		if sLit, ok := s.TrainSet.(*ast.StringLit); ok {
			trainPtr = g.genStringConst(sLit.Value)
			trainLen = len(sLit.Value)
		} else if ident, ok := s.TrainSet.(*ast.Ident); ok {
			if slot, ok := g.vars[ident.Name]; ok {
				loaded := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
				trainPtr = loaded
				trainLen = 16
			}
		}
	}

	maxSyn := int64(64)
	if s.MaxSyndrome != nil {
		if iLit, ok := s.MaxSyndrome.(*ast.IntLit); ok {
			maxSyn = iLit.Value
		}
	}

	maxCorr := int64(8)
	if s.MaxCorrDist != nil {
		if iLit, ok := s.MaxCorrDist.(*ast.IntLit); ok {
			maxCorr = iLit.Value
		}
	}

	certTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @aad_build_training_cert(i8* %s, i64 %d, i8* %s, i64 %d, i64 100, i64 %d, i64 %d)\n",
		certTemp, modelPtr, modelLen, trainPtr, trainLen, maxSyn, maxCorr)

	if s.Into != "" {
		slot, ok := g.vars[s.Into]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i8*",
			}
			g.vars[s.Into] = slot
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", certTemp, slot.ptr)
	}
}

func (g *Generator) genAADAttestStmt(s *ast.AADAttestStmt) {
	certRef := "null"
	if s.Cert != "" {
		if slot, ok := g.vars[s.Cert]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			certRef = loaded
		}
	}

	envTemp := certRef
	if s.Decode != "" {
		decRef := "null"
		if slot, ok := g.vars[s.Decode]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			decRef = loaded
		}

		// Allocate envelope buffer
		envTemp = g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 512)\n", envTemp)

		// Create dummy syndrome buffer
		synBuf := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 8)\n", synBuf)
		fmt.Fprintf(&g.sb, "  store i8 1, i8* %s\n", synBuf)

		wrapRes := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i32 @aad_wrap_decode(i8* %s, i8* %s, i8* %s, i32 1, i8* %s)\n",
			wrapRes, certRef, decRef, synBuf, envTemp)
	}

	// Obtain reasoning chain reference
	chainRef := "null"
	if slot, ok := g.vars["__qera_chain"]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		chainRef = loaded
	} else {
		chainName := g.genStringConst("aad_reason_chain")
		chainTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_qera_create(i8* %s, i32 30, i8* null)\n", chainTemp, chainName)
		alloca := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
		slot := &varSlot{
			ptr:      alloca,
			llvmType: "i8*",
		}
		g.vars["__qera_chain"] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", chainTemp, slot.ptr)
		chainRef = chainTemp
	}

	attestRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @aad_attest_envelope(i8* %s, i8* %s)\n", attestRes, envTemp, chainRef)

	if s.Into != "" {
		slot, ok := g.vars[s.Into]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i8*",
			}
			g.vars[s.Into] = slot
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", envTemp, slot.ptr)
	}
}

func (g *Generator) genMetaFitStmt(s *ast.MetaFitStmt) {
	certRef := "null"
	if s.Cert != "" {
		if slot, ok := g.vars[s.Cert]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			certRef = loaded
		}
	}

	stateTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 1024)\n", stateTemp)

	fitRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @aad_meta_fit(i8* %s, i8* %s)\n", fitRes, stateTemp, certRef)

	if s.Into != "" {
		slot, ok := g.vars[s.Into]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i8*",
			}
			g.vars[s.Into] = slot
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", stateTemp, slot.ptr)
	}
}

func (g *Generator) genMetaAdaptStmt(s *ast.MetaAdaptStmt) {
	stateRef := "null"
	if s.State != "" {
		if slot, ok := g.vars[s.State]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			stateRef = loaded
		}
	}

	envRef := "null"
	if s.Envelope != "" {
		if slot, ok := g.vars[s.Envelope]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			envRef = loaded
		}
	}

	adaptRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @aad_meta_adapt(i8* %s, i8* %s)\n", adaptRes, stateRef, envRef)

	if s.Into != "" {
		slot, ok := g.vars[s.Into]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i8*",
			}
			g.vars[s.Into] = slot
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", stateRef, slot.ptr)
	}
}

func (g *Generator) genMetaVerifyStmt(s *ast.MetaVerifyStmt) {
	stateRef := "null"
	if s.State != "" {
		if slot, ok := g.vars[s.State]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			stateRef = loaded
		}
	}

	vTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @aad_meta_verify(i8* %s)\n", vTemp, stateRef)

	boolTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = icmp ne i32 %s, 0\n", boolTemp, vTemp)

	if s.Into != "" {
		slot, ok := g.vars[s.Into]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i1\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i1",
			}
			g.vars[s.Into] = slot
		}
		fmt.Fprintf(&g.sb, "  store i1 %s, i1* %s\n", boolTemp, slot.ptr)
	}
}

func (g *Generator) genCorruptAADEnvStmt(s *ast.CorruptAADEnvStmt) {
	envRef := "null"
	if s.Env != "" {
		if slot, ok := g.vars[s.Env]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			envRef = loaded
		}
	}

	offset := 16
	if s.Offset != nil {
		if iLit, ok := s.Offset.(*ast.IntLit); ok {
			offset = int(iLit.Value)
		}
	}

	corruptRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @aad_corrupt_envelope(i8* %s, i32 %d)\n", corruptRes, envRef, offset)
}
