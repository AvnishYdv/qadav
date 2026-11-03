package checker

import (
	"fmt"
	"strings"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

// Error represents a type-checking diagnostic error.
type Error struct {
	Pos ast.Position
	Msg string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Msg)
}

// Checker traverses the AST and verifies semantic and type correctness.
type Checker struct {
	globalScope          *Scope
	curScope             *Scope
	curRetType           *types.Type
	exprTypes            map[ast.Expr]*types.Type
	networks             map[string]*NetworkInfo
	errors               []Error
	IsTest               bool
	activeReasoningChain *ast.Ident
	verifiedSyndromes    map[string]bool
	recallTokens         map[string]*RecallTokenInfo
	tools                map[string]*ToolInfo
	sandboxes            map[string]*SandboxInfo
}

// New creates an initialized Checker.
func New() *Checker {
	global := NewScope(nil)
	return &Checker{
		globalScope:       global,
		curScope:          global,
		exprTypes:         make(map[ast.Expr]*types.Type),
		networks:          make(map[string]*NetworkInfo),
		verifiedSyndromes: make(map[string]bool),
		recallTokens:      make(map[string]*RecallTokenInfo),
		tools:             make(map[string]*ToolInfo),
		sandboxes:         make(map[string]*SandboxInfo),
	}
}

// Networks returns all checked network metadata.
func (c *Checker) Networks() map[string]*NetworkInfo {
	return c.networks
}

// ExprType returns the deduced type for an expression node.
func (c *Checker) ExprType(e ast.Expr) *types.Type {
	return c.exprTypes[e]
}

// ExprTypes returns the full map of expression types.
func (c *Checker) ExprTypes() map[ast.Expr]*types.Type {
	return c.exprTypes
}

// Errors returns all accumulated type check errors.
func (c *Checker) Errors() []Error {
	return c.errors
}

func (c *Checker) errorf(pos ast.Position, format string, args ...any) {
	c.errors = append(c.errors, Error{
		Pos: pos,
		Msg: fmt.Sprintf(format, args...),
	})
}

func (c *Checker) enterScope() {
	c.curScope = NewScope(c.curScope)
}

func (c *Checker) leaveScope() {
	if c.curScope.parent != nil {
		c.curScope = c.curScope.parent
	}
}

// Check type-checks an entire Program.
func (c *Checker) Check(prog *ast.Program) []Error {
	if prog == nil {
		return nil
	}

	// First pass: register top-level process signatures and networks
	for _, decl := range prog.Decls {
		if proc, ok := decl.(*ast.ProcessDecl); ok {
			c.registerProcessSignature(proc)
		} else if net, ok := decl.(*ast.NetworkDecl); ok {
			c.registerNetworkSignature(net)
		}
	}

	// Second pass: check declarations and function bodies
	for _, decl := range prog.Decls {
		c.checkDecl(decl)
	}

	return c.errors
}

func (c *Checker) registerProcessSignature(proc *ast.ProcessDecl) {
	var paramTypes []*types.Type
	for _, param := range proc.Params {
		pType := c.resolveType(param.Type)
		if pType == nil {
			c.errorf(param.Pos(), "unknown type for parameter %q", param.Name)
			pType = types.Void
		}
		paramTypes = append(paramTypes, pType)
	}

	retType := types.Void
	if proc.Returns != nil {
		retType = c.resolveType(proc.Returns)
		if retType == nil {
			c.errorf(proc.Returns.Pos(), "unknown return type for process %q", proc.Name)
			retType = types.Void
		}
	}

	procType := &types.Type{
		Kind:    types.KindProcess,
		Name:    proc.Name,
		Params:  paramTypes,
		Results: []*types.Type{retType},
	}

	sym := &Symbol{
		Name: proc.Name,
		Type: procType,
		Pos:  proc.Pos(),
	}

	if err := c.globalScope.Define(sym); err != nil {
		c.errorf(proc.Pos(), "%s", err.Error())
	}
}

func (c *Checker) resolveType(t ast.Type) *types.Type {
	if t == nil {
		return types.Void
	}
	switch node := t.(type) {
	case *ast.PrimitiveType:
		pt := types.FromPrimitiveName(node.Name)
		if pt != nil {
			return pt
		}
		return nil
	case *ast.TensorType:
		if node.Elem == nil {
			return types.Tensor
		}
		elemType := c.resolveType(node.Elem)
		if elemType == nil {
			elemType = types.Float64
		}
		if len(node.Dims) > 0 {
			return types.TensorOfShape(elemType, node.Dims)
		}
		return types.TensorOf(elemType)
	case *ast.NamedType:
		if pt := types.FromPrimitiveName(node.Name); pt != nil {
			return pt
		}
		if strings.HasPrefix(node.Name, "confident⟨") {
			return types.ConfidentOf(types.Float64, 0.95)
		}
		if strings.HasPrefix(node.Name, "explore⟨") {
			return types.ExploreOf(types.Float64)
		}
		return nil
	default:
		return nil
	}
}

func (c *Checker) checkDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.ProcessDecl:
		c.checkProcessDecl(d)
	case *ast.ProcedureDecl:
		c.checkProcedureDecl(d)
	case *ast.NetworkDecl:
		c.checkNetworkDecl(d)
	case *ast.LearnDecl:
		c.checkLearnDecl(d)
	case *ast.EthicalBoundaryDecl:
		c.checkEthicalBoundaryDecl(d)
	case *ast.VaultDecl:
		c.checkVaultDecl(d)
	case *ast.ProvenanceDecl:
		c.checkProvenanceDecl(d)
	case *ast.MemoryProvenanceDecl:
		c.checkMemoryProvenanceDecl(d)
	case *ast.LicenseDecl:
		c.checkLicenseDecl(d)
	case *ast.ReasoningChainDecl:
		c.checkReasoningChainDecl(d)
	case *ast.RecallTokenDecl:
		c.checkRecallTokenDecl(d)
	case *ast.ToolRegisterDecl:
		c.checkToolRegisterDecl(d)
	case *ast.AbstainThresholdDecl:
		c.checkAbstainThresholdDecl(d)
	case *ast.AccessDecl, *ast.ExposeDecl:
		// Module operations - acknowledged
	}
}

func (c *Checker) checkProcessDecl(proc *ast.ProcessDecl) {
	sym, ok := c.globalScope.Lookup(proc.Name)
	if !ok {
		return
	}
	procType := sym.Type

	c.enterScope()
	defer c.leaveScope()

	oldRet := c.curRetType
	if len(procType.Results) > 0 {
		c.curRetType = procType.Results[0]
	} else {
		c.curRetType = types.Void
	}
	defer func() { c.curRetType = oldRet }()

	// Define parameters in process scope
	for i, param := range proc.Params {
		pType := procType.Params[i]
		pSym := &Symbol{
			Name:    param.Name,
			Type:    pType,
			Pos:     param.Pos(),
			IsParam: true,
		}
		if err := c.curScope.Define(pSym); err != nil {
			c.errorf(param.Pos(), "duplicate parameter %q: %s", param.Name, err.Error())
		}
	}

	if proc.Body != nil {
		for _, stmt := range proc.Body.Statements {
			c.checkStmt(stmt)
		}
	}
}

func (c *Checker) checkStmt(stmt ast.Stmt) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.Block:
		c.enterScope()
		for _, inner := range s.Statements {
			c.checkStmt(inner)
		}
		c.leaveScope()

	case *ast.BindStmt:
		valType := c.checkExpr(s.Value)
		var targetType *types.Type

		if s.Type != nil {
			targetType = c.resolveType(s.Type)
			if targetType == nil {
				c.errorf(s.Type.Pos(), "unknown type annotation in bind")
			} else if valType != nil && !c.isAssignable(targetType, valType) {
				c.errorf(s.Pos(), "type mismatch in bind: cannot bind value of type %s to declared type %s", valType, targetType)
			}
		} else {
			targetType = valType
		}

		if targetType == nil {
			targetType = types.Void
		}

		sym := &Symbol{
			Name:    s.Name,
			Type:    targetType,
			Pos:     s.Pos(),
			IsConst: false,
		}

		if err := c.curScope.Define(sym); err != nil {
			c.errorf(s.Pos(), "%s", err.Error())
		}

	case *ast.AssignStmt:
		rhsType := c.checkExpr(s.RHS)
		var lhsType *types.Type

		if ident, ok := s.LHS.(*ast.Ident); ok {
			sym, found := c.curScope.Lookup(ident.Name)
			if !found {
				c.errorf(s.LHS.Pos(), "undefined identifier %q in assignment", ident.Name)
			} else {
				if sym.IsConst {
					c.errorf(s.LHS.Pos(), "cannot assign to immutable binding %q", ident.Name)
				}
				lhsType = sym.Type
				c.exprTypes[ident] = lhsType
			}
		} else {
			lhsType = c.checkExpr(s.LHS)
		}

		if lhsType != nil && rhsType != nil {
			if s.Op == token.ASSIGN {
				if !c.isAssignable(lhsType, rhsType) {
					c.errorf(s.Pos(), "type mismatch in assignment: cannot assign %s to %s", rhsType, lhsType)
				}
			} else {
				// Compound assignment: e.g. ↦⊕, ↦⊖, ↦⊗, ↦⊘
				if !lhsType.IsNumeric() || !rhsType.IsNumeric() {
					c.errorf(s.Pos(), "compound assignment requires numeric types, got %s and %s", lhsType, rhsType)
				}
			}
		}

	case *ast.EmitStmt:
		valType := c.checkExpr(s.Value)
		if valType != nil && valType.Kind == types.KindVoid {
			c.errorf(s.Pos(), "cannot emit void value")
		}

	case *ast.ObserveStmt:
		c.checkExpr(s.Value)

	case *ast.YieldStmt:
		if s.Value != nil {
			valType := c.checkExpr(s.Value)
			if c.curRetType != nil && c.curRetType.Kind == types.KindVoid {
				c.errorf(s.Pos(), "cannot yield value from void process")
			} else if c.curRetType != nil && valType != nil && !c.isAssignable(c.curRetType, valType) {
				c.errorf(s.Pos(), "yield type mismatch: process expects %s, got %s", c.curRetType, valType)
			}
		} else {
			if c.curRetType != nil && c.curRetType.Kind != types.KindVoid {
				c.errorf(s.Pos(), "yield missing value; expected %s", c.curRetType)
			}
		}

	case *ast.WhenStmt:
		condType := c.checkExpr(s.Condition)
		if condType != nil && condType.Kind != types.KindLogic {
			c.errorf(s.Condition.Pos(), "when condition must be logic, got %s", condType)
		}
		if s.Then != nil {
			c.checkStmt(s.Then)
		}
		if s.Else != nil {
			c.checkStmt(s.Else)
		}

	case *ast.CycleStmt:
		if s.Condition != nil {
			condType := c.checkExpr(s.Condition)
			if condType != nil && condType.Kind != types.KindLogic {
				c.errorf(s.Condition.Pos(), "cycle condition must be logic, got %s", condType)
			}
		}
		if s.Body != nil {
			c.checkStmt(s.Body)
		}

	case *ast.IterateStmt:
		c.checkExpr(s.Collection)
		c.enterScope()
		// Bind iteration variable
		_ = c.curScope.Define(&Symbol{Name: s.Var, Type: types.Int64, Pos: s.Pos()})
		if s.Body != nil {
			for _, inner := range s.Body.Statements {
				c.checkStmt(inner)
			}
		}
		c.leaveScope()

	case *ast.SuperposeStmt:
		qType := c.checkExpr(s.Qubit)
		if qType != nil && !qType.IsQuantum() {
			c.errorf(s.Pos(), "superpose requires qubit, got %s", qType)
		}

	case *ast.EntangleStmt:
		q1 := c.checkExpr(s.Q1)
		q2 := c.checkExpr(s.Q2)
		if q1 != nil && !q1.IsQuantum() {
			c.errorf(s.Q1.Pos(), "entangle target must be qubit, got %s", q1)
		}
		if q2 != nil && !q2.IsQuantum() {
			c.errorf(s.Q2.Pos(), "entangle target must be qubit, got %s", q2)
		}

	case *ast.CollapseStmt:
		q := c.checkExpr(s.Qubit)
		if q != nil && !q.IsQuantum() {
			c.errorf(s.Pos(), "collapse target must be qubit, got %s", q)
		}
		if s.Into != nil {
			c.checkExpr(s.Into)
		}

	case *ast.QubitDecl:
		for _, name := range s.Names {
			_ = c.curScope.Define(&Symbol{
				Name: name,
				Type: types.Qubit,
				Pos:  s.Pos(),
			})
		}

	case *ast.LearnDecl:
		c.checkLearnDecl(s)

	case *ast.DatasetDecl:
		c.checkDatasetDecl(s)

	case *ast.NormalizeStmt:
		c.checkNormalizeStmt(s)

	case *ast.ShuffleStmt:
		c.checkShuffleStmt(s)

	case *ast.BatchStmt:
		c.checkBatchStmt(s)

	case *ast.EnergyDecl:
		c.checkEnergyDecl(s)

	case *ast.MinimizeStmt:
		c.checkMinimizeStmt(s)

	case *ast.TileStmt:
		c.checkTileStmt(s)

	case *ast.AssertStmt:
		c.checkAssertStmt(s)

	case *ast.InferStmt:
		c.checkInferStmt(s)

	case *ast.PropagateStmt:
		c.checkPropagateStmt(s)

	case *ast.MemoryDecl:
		c.checkMemoryDecl(s)

	case *ast.RememberStmt:
		c.checkRememberStmt(s)

	case *ast.GraphDecl:
		c.checkGraphDecl(s)

	case *ast.NodeStmt:
		c.checkNodeStmt(s)

	case *ast.LinkStmt:
		c.checkLinkStmt(s)

	case *ast.ExecuteStmt:
		c.checkExecuteStmt(s)

	case *ast.EthicalBoundaryDecl:
		c.checkEthicalBoundaryDecl(s)

	case *ast.EnforceStmt:
		c.checkEnforceStmt(s)

	case *ast.DecoherenceStmt:
		c.checkDecoherenceStmt(s)

	case *ast.VaultDecl:
		c.checkVaultDecl(s)

	case *ast.SealStmt:
		c.checkSealStmt(s)

	case *ast.UnsealStmt:
		c.checkUnsealStmt(s)

	case *ast.WipeStmt:
		c.checkWipeStmt(s)

	case *ast.ProvenanceDecl:
		c.checkProvenanceDecl(s)

	case *ast.CertifyStmt:
		c.checkCertifyStmt(s)

	case *ast.MemoryProvenanceDecl:
		c.checkMemoryProvenanceDecl(s)

	case *ast.VerifyMemoryStmt:
		c.checkVerifyMemoryStmt(s)

	case *ast.AuditMemoryStmt:
		c.checkAuditMemoryStmt(s)

	case *ast.CorruptShareStmt:
		c.checkCorruptShareStmt(s)

	case *ast.LicenseDecl:
		c.checkLicenseDecl(s)

	case *ast.TransferStmt:
		c.checkTransferStmt(s)

	case *ast.DistillStmt:
		c.checkDistillStmt(s)

	case *ast.ReasoningChainDecl:
		c.checkReasoningChainDecl(s)

	case *ast.ReasonStepStmt:
		c.checkReasonStepStmt(s)

	case *ast.VerifyChainStmt:
		c.checkVerifyChainStmt(s)

	case *ast.AuditReasoningStmt:
		c.checkAuditReasoningStmt(s)

	case *ast.CorruptReasoningStepStmt:
		c.checkCorruptReasoningStepStmt(s)

	case *ast.ModalityDecl:
		c.checkModalityDecl(s)

	case *ast.CrossModalReasonStmt:
		c.checkCrossModalReasonStmt(s)

	case *ast.QECMemoryDecl:
		c.checkQECMemoryDecl(s)

	case *ast.SyndromeStmt:
		c.checkSyndromeStmt(s)

	case *ast.VerifySyndromeStmt:
		c.checkVerifySyndromeStmt(s)

	case *ast.CorrectStmt:
		c.checkCorrectStmt(s)

	case *ast.CorruptSyndromeStmt:
		c.checkCorruptSyndromeStmt(s)

	case *ast.ConsolidateStmt:
		c.checkConsolidateStmt(s)

	case *ast.RecallTokenDecl:
		c.checkRecallTokenDecl(s)

	case *ast.GrantStmt:
		c.checkGrantStmt(s)

	case *ast.RevokeStmt:
		c.checkRevokeStmt(s)

	case *ast.ToolRegisterDecl:
		c.checkToolRegisterDecl(s)

	case *ast.ToolCallStmt:
		c.checkToolCallStmt(s)

	case *ast.CertifiedReasonStmt:
		c.checkCertifiedReasonStmt(s)

	case *ast.AbstainThresholdDecl:
		c.checkAbstainThresholdDecl(s)

	case *ast.AbstainCheckStmt:
		c.checkAbstainCheckStmt(s)

	case *ast.SandboxCreateStmt:
		c.checkSandboxCreateStmt(s)

	case *ast.SandboxRunStmt:
		c.checkSandboxRunStmt(s)

	case *ast.SandboxVerifyStmt:
		c.checkSandboxVerifyStmt(s)

	case *ast.GPUMatmulStmt:
		c.checkGPUMatmulStmt(s)

	case *ast.GPUSoftmaxStmt:
		c.checkGPUSoftmaxStmt(s)

	case *ast.GPULayernormStmt:
		c.checkGPULayernormStmt(s)

	case *ast.AGKEAttestStmt:
		c.checkAGKEAttestStmt(s)

	case *ast.AGKEVerifyStmt:
		c.checkAGKEVerifyStmt(s)

	case *ast.CorruptGPUDigestStmt:
		c.checkCorruptGPUDigestStmt(s)

	case *ast.ExprStmt:
		c.checkExpr(s.Expr)

	case *ast.HaltStmt, *ast.SkipStmt:
		// Valid in loop contexts
	}
}

func (c *Checker) checkExpr(expr ast.Expr) *types.Type {
	if expr == nil {
		return nil
	}

	var result *types.Type

	switch e := expr.(type) {
	case *ast.IntLit:
		result = types.Int64

	case *ast.FloatLit:
		if e.Imaginary {
			result = types.Complex128
		} else {
			result = types.Float64
		}

	case *ast.StringLit:
		result = types.Text

	case *ast.TensorLit:
		result = c.checkTensorLit(e)

	case *ast.CollapseExpr:
		qType := c.checkExpr(e.Qubit)
		if qType != nil && !qType.IsQuantum() {
			c.errorf(e.Pos(), "collapse target must be qubit, got %s", qType)
		}
		result = types.Int64

	case *ast.ConceptLit:
		result = types.Concept

	case *ast.QuasiLit:
		result = types.Quasi

	case *ast.DecomposeExpr:
		inType := c.checkExpr(e.Input)
		if inType != nil && inType.Kind != types.KindConcept && !inType.IsInteger() {
			c.errorf(e.Pos(), "decompose argument must be concept or integer, got %s", inType)
		}
		result = types.TensorOf(types.Int64)

	case *ast.Ident:
		switch e.Name {
		case "affirm", "⊤":
			result = types.Logic
		case "deny", "⊥":
			result = types.Logic
		default:
			sym, ok := c.curScope.Lookup(e.Name)
			if !ok {
				c.errorf(e.Pos(), "undefined identifier %q", e.Name)
				result = types.Void
			} else {
				result = sym.Type
			}
		}

	case *ast.BinaryExpr:
		leftType := c.checkExpr(e.Left)
		rightType := c.checkExpr(e.Right)

		switch e.Op {
		case token.OSUM, token.ODIFF, token.OTIM, token.ODIV, token.OMOD, token.OPOW:
			// Arithmetic: tensor vs scalar vs numeric
			if leftType != nil && rightType != nil && (leftType.IsTensor() || rightType.IsTensor()) {
				res, err := c.checkTensorBinaryOp(e.Pos(), e.Op, leftType, rightType)
				if err != nil {
					c.errorf(e.Pos(), "%s", err.Error())
					result = types.Void
				} else {
					result = res
				}
			} else if e.Op == token.OSUM && leftType != nil && rightType != nil &&
				leftType.Kind == types.KindText && rightType.Kind == types.KindText {
				result = types.Text
			} else if (e.Op == token.OTIM || e.Op == token.ODIV) && leftType != nil && rightType != nil &&
				(leftType.Kind == types.KindConcept || rightType.Kind == types.KindConcept) {
				result = types.Concept
			} else if leftType != nil && rightType != nil && (leftType.Kind == types.KindConfident || rightType.Kind == types.KindConfident) {
				result = CheckConfidentBinary(leftType, rightType, e.Op == token.OSUM)
			} else if leftType != nil && rightType != nil && leftType.Kind == types.KindQuasi && rightType.Kind == types.KindQuasi {
				result = types.Quasi
			} else if leftType != nil && rightType != nil {
				prom := types.Promote(leftType, rightType)
				if prom == nil {
					c.errorf(e.Pos(), "incompatible types for arithmetic %s: %s and %s", e.Op, leftType, rightType)
					result = types.Void
				} else {
					result = prom
				}
			}

		case token.MATMUL:
			if leftType == nil || rightType == nil || !leftType.IsTensor() || !rightType.IsTensor() {
				c.errorf(e.Pos(), "scalar matmul invalid; use ⊗ for scalar mul")
				result = types.Void
			} else {
				res, err := c.checkMatmul(e.Pos(), leftType, rightType)
				if err != nil {
					c.errorf(e.Pos(), "%s", err.Error())
					result = types.Void
				} else {
					result = res
				}
			}

		case token.EQUIV, token.NEQ, token.PREC, token.SUCC, token.PREQ, token.SUEQ, token.APPROX:
			// Comparison
			if leftType != nil && rightType != nil {
				if leftType.IsTensor() || rightType.IsTensor() {
					c.errorf(e.Pos(), "tensor comparison not supported in MVP")
				} else if !c.areComparable(leftType, rightType) {
					c.errorf(e.Pos(), "incompatible types for comparison %s: %s and %s", e.Op, leftType, rightType)
				}
			}
			result = types.Logic

		case token.WEDGE, token.VEE:
			// Logical
			if leftType != nil && leftType.Kind != types.KindLogic {
				c.errorf(e.Left.Pos(), "logical operator requires logic operands, got %s", leftType)
			}
			if rightType != nil && rightType.Kind != types.KindLogic {
				c.errorf(e.Right.Pos(), "logical operator requires logic operands, got %s", rightType)
			}
			result = types.Logic

		case token.XVEE:
			// Logical or Bitwise XOR
			if leftType != nil && leftType.Kind == types.KindLogic && rightType != nil && rightType.Kind == types.KindLogic {
				result = types.Logic
			} else if leftType != nil && rightType != nil && leftType.IsInteger() && rightType.IsInteger() {
				prom := types.Promote(leftType, rightType)
				if prom == nil {
					prom = types.Int64
				}
				result = prom
			} else {
				c.errorf(e.Pos(), "operator ⊻ requires logic or integer operands, got %s and %s", leftType, rightType)
				result = types.Void
			}

		case token.MEMBER, token.NOTMEMBER:
			c.errorf(e.Pos(), "Phase 2 feature: membership operator ∈")
			result = types.Logic

		default:
			c.errorf(e.Pos(), "unsupported binary operator %s", e.Op)
			result = types.Void
		}

	case *ast.UnaryExpr:
		operandType := c.checkExpr(e.Right)
		switch e.Op {
		case token.NEG:
			if operandType != nil && operandType.Kind != types.KindLogic {
				c.errorf(e.Pos(), "logical negation requires logic operand, got %s", operandType)
			}
			result = types.Logic
		case token.ODIFF:
			if operandType != nil && !operandType.IsNumeric() && !operandType.IsTensor() {
				c.errorf(e.Pos(), "arithmetic negation requires numeric operand, got %s", operandType)
			}
			result = operandType
		case token.SUM:
			if operandType == nil || !operandType.IsTensor() {
				c.errorf(e.Pos(), "∑ requires tensor, got %s", operandType)
				result = types.Void
			} else {
				result = operandType.TensorElem()
				if result == nil {
					result = types.Float64
				}
			}
		case token.PRODUCT:
			if operandType == nil || !operandType.IsTensor() {
				c.errorf(e.Pos(), "∏ requires tensor, got %s", operandType)
				result = types.Void
			} else {
				result = operandType.TensorElem()
				if result == nil {
					result = types.Float64
				}
			}
		case token.ADJOINT:
			if operandType == nil || !operandType.IsTensor() {
				c.errorf(e.Pos(), "⋆ requires tensor, got %s", operandType)
				result = types.Void
			} else if operandType.Rank() != 2 {
				c.errorf(e.Pos(), "transpose requires 2D tensor, got rank %d", operandType.Rank())
				result = types.Void
			} else {
				resDims := []int{operandType.Shape[1], operandType.Shape[0]}
				result = types.TensorOfShape(operandType.TensorElem(), resDims)
			}
		default:
			result = operandType
		}

	case *ast.GroupExpr:
		result = c.checkExpr(e.Inner)

	case *ast.CallExpr:
		procSym, ok := c.lookupProcess(e.Func)
		if !ok {
			c.errorf(e.Pos(), "cannot call non-process expression")
			result = types.Void
		} else {
			procType := procSym.Type
			if len(e.Args) != len(procType.Params) {
				c.errorf(e.Pos(), "process %q expects %d arguments, got %d", procSym.Name, len(procType.Params), len(e.Args))
			} else {
				for i, arg := range e.Args {
					argType := c.checkExpr(arg)
					paramType := procType.Params[i]
					if argType != nil && !c.isAssignable(paramType, argType) {
						c.errorf(arg.Pos(), "argument %d type mismatch: expected %s, got %s", i+1, paramType, argType)
					}
				}
			}
			if len(procType.Results) > 0 {
				result = procType.Results[0]
			} else {
				result = types.Void
			}
		}

	case *ast.IndexExpr:
		objType := c.checkExpr(e.Object)
		c.checkExpr(e.Index)
		if objType != nil && objType.Elem != nil {
			result = objType.Elem
		} else {
			result = types.Int64
		}

	case *ast.MemberExpr:
		c.checkExpr(e.Object)
		result = types.Int64

	case *ast.RangeExpr:
		c.checkExpr(e.Start)
		c.checkExpr(e.End)
		result = types.Tensor

	case *ast.ConfidentLit:
		result = c.checkConfidentLit(e)

	case *ast.RecallExpr:
		result = c.checkRecallExpr(e)

	case *ast.TraceExpr:
		result = c.checkTraceExpr(e)

	case *ast.ExecuteExpr:
		result = c.checkExecuteExpr(e)

	case *ast.VerifyTrainingExpr:
		result = c.checkVerifyTrainingExpr(e)

	case *ast.ScopedRecallExpr:
		result = c.checkScopedRecallExpr(e)

	case *ast.ToolCallExpr:
		result = c.checkToolCallExpr(e)

	case *ast.AbstainCheckExpr:
		result = c.checkAbstainCheckExpr(e)

	case *ast.SandboxVerifyExpr:
		result = c.checkSandboxVerifyExpr(e)

	case *ast.GPUMatmulExpr:
		result = c.checkGPUMatmulExpr(e)

	case *ast.GPUSoftmaxExpr:
		result = c.checkGPUSoftmaxExpr(e)

	case *ast.GPULayernormExpr:
		result = c.checkGPULayernormExpr(e)

	case *ast.AGKEVerifyExpr:
		result = c.checkAGKEVerifyExpr(e)

	default:
		result = types.Void
	}

	if result == nil {
		result = types.Void
	}

	c.exprTypes[expr] = result
	return result
}

func (c *Checker) lookupProcess(fnExpr ast.Expr) (*Symbol, bool) {
	if ident, ok := fnExpr.(*ast.Ident); ok {
		sym, found := c.globalScope.Lookup(ident.Name)
		if found && sym.Type.Kind == types.KindProcess {
			return sym, true
		}
	}
	return nil, false
}

func (c *Checker) isAssignable(target, source *types.Type) bool {
	if target == nil || source == nil {
		return false
	}
	if target.Equal(source) {
		return true
	}
	if target.Kind == types.KindConcept && (source.IsInteger() || source.Kind == types.KindConcept) {
		return true
	}
	if target.IsInteger() && source.Kind == types.KindConcept {
		return true
	}
	if target.IsProbabilistic() {
		if target.Elem != nil && c.isAssignable(target.Elem, source) {
			return true
		}
	}
	if target.IsTensor() && source.IsTensor() {
		tElem := target.TensorElem()
		sElem := source.TensorElem()
		if tElem != nil && sElem != nil && !c.isAssignable(tElem, sElem) {
			return false
		}
		if len(target.Shape) > 0 && len(source.Shape) > 0 {
			if len(target.Shape) != len(source.Shape) {
				return false
			}
			for i := range target.Shape {
				if target.Shape[i] != source.Shape[i] {
					return false
				}
			}
		}
		return true
	}
	// Integer compatibility (literals default to int64, allow assigning to int8..int64)
	if target.IsInteger() && source.IsInteger() {
		return true
	}
	// Float compatibility (literals default to float64, allow assigning to float32..float64)
	if target.IsFloat() && source.IsFloat() {
		return true
	}
	// Widening numeric promotion: source can widen to target
	if target.IsNumeric() && source.IsNumeric() {
		prom := types.Promote(target, source)
		if prom != nil && prom.Equal(target) {
			return true
		}
	}
	return false
}

func (c *Checker) areComparable(a, b *types.Type) bool {
	if a.Equal(b) {
		return true
	}
	if a.IsNumeric() && b.IsNumeric() {
		return types.Promote(a, b) != nil
	}
	return false
}

func (c *Checker) checkTensorLit(lit *ast.TensorLit) *types.Type {
	if len(lit.Items) == 0 {
		return types.TensorOfShape(types.Float64, lit.Dims)
	}
	var elemType *types.Type
	for _, item := range lit.Items {
		itemType := c.checkExpr(item)
		if itemType == nil {
			continue
		}
		if itemType.IsTensor() {
			itemType = itemType.TensorElem()
		}
		if elemType == nil {
			elemType = itemType
		} else {
			prom := types.Promote(elemType, itemType)
			if prom != nil {
				elemType = prom
			}
		}
	}
	if elemType == nil {
		elemType = types.Float64
	}
	return types.TensorOfShape(elemType, lit.Dims)
}

func (c *Checker) checkTensorBinaryOp(_ ast.Position, op token.Type, a, b *types.Type) (*types.Type, error) {
	if a.IsTensor() && b.IsTensor() {
		outShape, err := checkShapeOp(a.Shape, b.Shape, string(op))
		if err != nil {
			return nil, err
		}
		prom, err := promoteTensors(a, b)
		if err != nil {
			return nil, err
		}
		return types.TensorOfShape(prom.TensorElem(), outShape), nil
	} else if a.IsTensor() {
		prom := types.Promote(a.TensorElem(), b)
		if prom == nil {
			return nil, fmt.Errorf("incompatible element type for %s: %s and %s", op, a.TensorElem(), b)
		}
		return types.TensorOfShape(prom, a.Shape), nil
	} else {
		prom := types.Promote(a, b.TensorElem())
		if prom == nil {
			return nil, fmt.Errorf("incompatible element type for %s: %s and %s", op, a, b.TensorElem())
		}
		return types.TensorOfShape(prom, b.Shape), nil
	}
}

func checkShapeOp(a, b []int, op string) ([]int, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("shape mismatch for %s: %v vs %v", op, a, b)
	}
	for i := range a {
		if a[i] != b[i] {
			return nil, fmt.Errorf("shape mismatch for %s: %v vs %v", op, a, b)
		}
	}
	return a, nil
}

func promoteTensors(a, b *types.Type) (*types.Type, error) {
	elemA := a.TensorElem()
	if elemA == nil {
		elemA = types.Float64
	}
	elemB := b.TensorElem()
	if elemB == nil {
		elemB = types.Float64
	}
	prom := types.Promote(elemA, elemB)
	if prom == nil {
		return nil, fmt.Errorf("incompatible tensor element types: %s and %s", elemA, elemB)
	}
	return types.TensorOf(prom), nil
}

func (c *Checker) checkMatmul(_ ast.Position, a, b *types.Type) (*types.Type, error) {
	prom, err := promoteTensors(a, b)
	if err != nil {
		return nil, err
	}
	elem := prom.TensorElem()
	outDims, err := checkMatmulShape(a.Shape, b.Shape)
	if err != nil {
		return nil, err
	}
	if len(outDims) == 0 {
		return elem, nil
	}
	return types.TensorOfShape(elem, outDims), nil
}

func checkMatmulShape(a, b []int) ([]int, error) {
	rankA := len(a)
	rankB := len(b)

	if rankA == 2 && rankB == 2 {
		if a[1] != b[0] {
			return nil, fmt.Errorf("matmul shape mismatch: inner dims %d vs %d don't match", a[1], b[0])
		}
		return []int{a[0], b[1]}, nil
	}
	if rankA == 1 && rankB == 2 {
		if a[0] != b[0] {
			return nil, fmt.Errorf("matmul shape mismatch: inner dims %d vs %d don't match", a[0], b[0])
		}
		return []int{b[1]}, nil
	}
	if rankA == 2 && rankB == 1 {
		if a[1] != b[0] {
			return nil, fmt.Errorf("matmul shape mismatch: inner dims %d vs %d don't match", a[1], b[0])
		}
		return []int{a[0]}, nil
	}
	if rankA == 1 && rankB == 1 {
		if a[0] != b[0] {
			return nil, fmt.Errorf("matmul shape mismatch: dims %d vs %d don't match", a[0], b[0])
		}
		return []int{}, nil
	}

	return nil, fmt.Errorf("unsupported tensor ranks for matmul: %d and %d", rankA, rankB)
}
