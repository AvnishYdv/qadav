package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

// Parser parses a Qadav token stream into an Abstract Syntax Tree.
type Parser struct {
	lexer   *lexer.Lexer
	current token.Token
	peeked  token.Token
	errors  ErrorList
	file    string
}

// New creates and initializes a new Parser for the given lexer.
func New(l *lexer.Lexer, file string) *Parser {
	p := &Parser{
		lexer: l,
		file:  file,
	}
	p.nextToken()
	p.nextToken()
	return p
}

// Errors returns the list of errors encountered during parsing.
func (p *Parser) Errors() ErrorList {
	return p.errors
}

func (p *Parser) nextToken() {
	p.current = p.peeked
	p.peeked = p.lexer.NextToken()
}

func (p *Parser) curPos() ast.Position {
	return ast.FromToken(p.current)
}

func (p *Parser) errorf(pos ast.Position, format string, args ...any) {
	p.errors = append(p.errors, ParseError{
		Pos:     pos,
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *Parser) expect(t token.Type) token.Token {
	tok := p.current
	if p.current.Type != t {
		p.errorf(p.curPos(), "expected token %s, got %s (%q)", t, p.current.Type, p.current.Literal)
	} else {
		p.nextToken()
	}
	return tok
}

func (p *Parser) synchronize() {
	for p.current.Type != token.EOF {
		if p.current.Type == token.STMT_END {
			p.nextToken()
			return
		}
		if p.current.Type == token.BLOCK_CLOSE || p.current.Type == token.BLOCK_OPEN {
			return
		}
		p.nextToken()
	}
}

// Parse parses the entire token stream and returns a Program AST.
func (p *Parser) Parse() *ast.Program {
	prog := &ast.Program{
		Pos_:  p.curPos(),
		Decls: []ast.Decl{},
	}

	for p.current.Type != token.EOF {
		switch p.current.Type {
		case token.RECORD:
			decl := p.parseRecordDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.VARIANT:
			decl := p.parseVariantDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.PROCESS:
			decl := p.parseProcessDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.NETWORK:
			decl := p.parseNetworkDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.LEARN, token.TRAIN:
			decl := p.parseLearnDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.ACCESS:
			decl := p.parseAccessDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.EXPOSE:
			decl := p.parseExposeDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.PROCEDURE:
			decl := p.parseProcedureDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.ETHICAL_BOUNDARY:
			decl := p.parseEthicalBoundaryDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.VAULT:
			decl := p.parseVaultDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.PROVENANCE:
			decl := p.parseProvenanceDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.MEMORY_PROV:
			decl := p.parseMemoryProvenanceDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.LICENSE:
			decl := p.parseLicenseDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.REASONING_CHAIN:
			decl := p.parseReasoningChainDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.RECALL_TOKEN:
			decl := p.parseRecallTokenDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.TOOL_REGISTER:
			decl := p.parseToolRegisterDecl()
			if decl != nil {
				prog.Decls = append(prog.Decls, decl)
			}
		case token.DIST_INIT, token.AT_GPU, token.MULTI_DEVICE, token.QHAL_BACKEND:
			p.nextToken()
			if p.current.Type == token.IDENT {
				p.nextToken()
			}
			continue
		default:
			p.errorf(p.curPos(), "unexpected top-level token: %s (%q)", p.current.Type, p.current.Literal)
			p.nextToken()
			p.synchronize()
		}
	}

	return prog
}

// ---------------------------------------------------------------------
// Top-Level Declarations
// ---------------------------------------------------------------------

func (p *Parser) parseProcessDecl() *ast.ProcessDecl {
	pos := p.curPos()
	p.expect(token.PROCESS)

	nameTok := p.expect(token.IDENT)
	decl := &ast.ProcessDecl{
		Pos_:   pos,
		Name:   nameTok.Literal,
		Params: []*ast.Param{},
	}

	p.expect(token.GROUP_OPEN)

	for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
		param := p.parseParam()
		if param != nil {
			decl.Params = append(decl.Params, param)
		}
		if p.current.Type == token.SEP {
			p.nextToken()
		} else {
			break
		}
	}

	p.expect(token.GROUP_CLOSE)

	if p.current.Type == token.YIELDS {
		p.nextToken() // consume →
		decl.Returns = p.parseType()
	} else if isTypeToken(p.current.Type) {
		decl.Returns = p.parseType()
	}

	decl.Body = p.parseBlock()
	return decl
}

func (p *Parser) parseParam() *ast.Param {
	pos := p.curPos()
	if isTypeToken(p.current.Type) || (p.current.Type == token.IDENT && p.peeked.Type == token.IDENT) {
		ptype := p.parseType()
		nameTok := p.expect(token.IDENT)
		return &ast.Param{
			Pos_: pos,
			Name: nameTok.Literal,
			Type: ptype,
		}
	}

	nameTok := p.expect(token.IDENT)
	var ptype ast.Type
	if p.current.Type == token.ANNOT {
		p.nextToken() // consume ∵
		ptype = p.parseType()
	}
	return &ast.Param{
		Pos_: pos,
		Name: nameTok.Literal,
		Type: ptype,
	}
}

func (p *Parser) parseNetworkDecl() *ast.NetworkDecl {
	pos := p.curPos()
	p.expect(token.NETWORK)
	nameTok := p.expect(token.IDENT)

	decl := &ast.NetworkDecl{
		Pos_:   pos,
		Name:   nameTok.Literal,
		Layers: []*ast.Layer{},
	}

	p.expect(token.BLOCK_OPEN)
	for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
		layerPos := p.curPos()
		kind := p.current.Type
		switch kind {
		case token.DENSE, token.RELU, token.SIGMOID, token.SOFTMAX, token.LAYERNORM,
			token.CONV2D, token.ATTENTION, token.DROPOUT:
			p.nextToken()
			var args []ast.Expr
			// Parse optional arguments, e.g. dense ⟬2⸴ 4⟭ or dense ⟬128⟭ or dense 128
			if p.current.Type == token.INDEX_OPEN {
				p.nextToken() // consume ⟬
				for p.current.Type != token.INDEX_CLOSE && p.current.Type != token.EOF {
					arg := p.parseExpression(LOWEST)
					if arg != nil {
						args = append(args, arg)
					}
					if p.current.Type == token.SEP {
						p.nextToken()
					} else {
						break
					}
				}
				p.expect(token.INDEX_CLOSE)
			} else if p.current.Type == token.INT {
				arg := p.parseExpression(LOWEST)
				if arg != nil {
					args = append(args, arg)
				}
			}
			if p.current.Type == token.STMT_END {
				p.nextToken()
			}
			decl.Layers = append(decl.Layers, &ast.Layer{
				Pos_: layerPos,
				Kind: kind,
				Args: args,
			})
		case token.STMT_END:
			p.nextToken()
		default:
			p.errorf(p.curPos(), "unexpected token in network block: %s (%q)", p.current.Type, p.current.Literal)
			p.nextToken()
		}
	}
	p.expect(token.BLOCK_CLOSE)
	return decl
}

func (p *Parser) parseLearnDecl() *ast.LearnDecl {
	pos := p.curPos()
	p.nextToken() // consume LEARN or TRAIN
	nameTok := p.expect(token.IDENT)
	p.expect(token.ON)
	onTok := p.expect(token.IDENT)
	var onY string
	if p.current.Type == token.SEP {
		p.nextToken()
		onYTok := p.expect(token.IDENT)
		onY = onYTok.Literal
	}

	decl := &ast.LearnDecl{
		Pos_:        pos,
		Name:        nameTok.Literal,
		NetworkName: nameTok.Literal,
		OnName:      onTok.Literal,
		OnYName:     onY,
		Loss:        "mse",
		Optimizer:   "sgd",
	}

	p.expect(token.BLOCK_OPEN)
	for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
		fieldTok := p.current
		fieldLiteral := strings.ToLower(fieldTok.Literal)

		switch {
		case fieldTok.Type == token.LOSS || fieldLiteral == "loss":
			p.nextToken()
			lossTok := p.current
			decl.Loss = strings.ToLower(lossTok.Literal)
			p.nextToken()
			if p.current.Type == token.STMT_END {
				p.nextToken()
			}
		case fieldTok.Type == token.OPTIMIZER || fieldLiteral == "optimizer":
			p.nextToken()
			optTok := p.current
			decl.Optimizer = strings.ToLower(optTok.Literal)
			p.nextToken()
			if p.current.Type == token.STMT_END {
				p.nextToken()
			}
		case fieldTok.Type == token.LR || fieldLiteral == "lr":
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			decl.LR = p.parseExpression(LOWEST)
			if p.current.Type == token.STMT_END {
				p.nextToken()
			}
		case fieldTok.Type == token.EPOCHS || fieldLiteral == "epochs":
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			decl.Epochs = p.parseExpression(LOWEST)
			if p.current.Type == token.STMT_END {
				p.nextToken()
			}
		case fieldTok.Type == token.BATCH_SIZE || fieldLiteral == "batch_size":
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			decl.BatchSize = p.parseExpression(LOWEST)
			if p.current.Type == token.STMT_END {
				p.nextToken()
			}
		case fieldTok.Type == token.STMT_END:
			p.nextToken()
		default:
			p.errorf(p.curPos(), "unexpected token in learn block: %s (%q)", p.current.Type, p.current.Literal)
			p.nextToken()
		}
	}
	p.expect(token.BLOCK_CLOSE)
	return decl
}

func (p *Parser) parseAccessDecl() *ast.AccessDecl {
	pos := p.curPos()
	p.expect(token.ACCESS)

	var sb strings.Builder
	for p.current.Type != token.STMT_END && p.current.Type != token.EOF {
		sb.WriteString(p.current.Literal)
		p.nextToken()
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.AccessDecl{
		Pos_: pos,
		Path: sb.String(),
	}
}

func (p *Parser) parseExposeDecl() *ast.ExposeDecl {
	pos := p.curPos()
	p.expect(token.EXPOSE)
	decl := &ast.ExposeDecl{
		Pos_:  pos,
		Names: []string{},
	}

	for p.current.Type != token.STMT_END && p.current.Type != token.EOF {
		if p.current.Type == token.IDENT {
			decl.Names = append(decl.Names, p.current.Literal)
			p.nextToken()
		}
		if p.current.Type == token.SEP {
			p.nextToken()
		} else {
			break
		}
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}
	return decl
}

// ---------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------

func (p *Parser) parseBlock() *ast.Block {
	pos := p.curPos()
	p.expect(token.BLOCK_OPEN)
	block := &ast.Block{
		Pos_:       pos,
		Statements: []ast.Stmt{},
	}

	for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	p.expect(token.BLOCK_CLOSE)
	return block
}

func (p *Parser) parseStatement() ast.Stmt {
	switch {
	case p.current.Type == token.TENSOR && p.peeked.Type == token.INDEX_OPEN:
		return p.parseExprOrAssignStmt()
	case p.current.Type == token.QUBIT:
		return p.parseQubitDeclStmt()
	case p.current.Type == token.DATASET:
		return p.parseDatasetDecl()
	case p.current.Type == token.INFER:
		return p.parseInferStmt()
	case p.current.Type == token.ASSERT:
		return p.parseAssertStmt()
	case p.current.Type == token.MEMORY:
		return p.parseMemoryDecl()
	case p.current.Type == token.REMEMBER:
		return p.parseRememberStmt()
	case p.current.Type == token.GRAPH:
		return p.parseGraphDecl()
	case p.current.Type == token.NODE:
		return p.parseNodeStmt()
	case p.current.Type == token.LINK:
		return p.parseLinkStmt()
	case p.current.Type == token.EXECUTE:
		return p.parseExecuteStmt()
	case p.current.Type == token.ETHICAL_BOUNDARY:
		return p.parseEthicalBoundaryDecl()
	case p.current.Type == token.ENFORCE:
		return p.parseEnforceStmt()
	case p.current.Type == token.DECOHERENCE:
		return p.parseDecoherenceStmt()
	case p.current.Type == token.VAULT:
		return p.parseVaultDecl()
	case p.current.Type == token.SEAL:
		return p.parseSealStmt()
	case p.current.Type == token.UNSEAL:
		return p.parseUnsealStmt()
	case p.current.Type == token.WIPE:
		return p.parseWipeStmt()
	case p.current.Type == token.PROVENANCE:
		return p.parseProvenanceDecl()
	case p.current.Type == token.CERTIFY:
		return p.parseCertifyStmt()
	case p.current.Type == token.MEMORY_PROV:
		return p.parseMemoryProvenanceDecl()
	case p.current.Type == token.VERIFY_MEMORY:
		return p.parseVerifyMemoryStmt()
	case p.current.Type == token.AUDIT_MEMORY:
		return p.parseAuditMemoryStmt()
	case p.current.Type == token.CORRUPT_SHARE:
		return p.parseCorruptShareStmt()
	case p.current.Type == token.LICENSE:
		return p.parseLicenseDecl()
	case p.current.Type == token.TRANSFER:
		return p.parseTransferStmt()
	case p.current.Type == token.DISTILL:
		return p.parseDistillStmt()
	case p.current.Type == token.REASONING_CHAIN:
		return p.parseReasoningChainDecl()
	case p.current.Type == token.REASON:
		return p.parseReasonStepStmt()
	case p.current.Type == token.VERIFY_CHAIN:
		return p.parseVerifyChainStmt()
	case p.current.Type == token.AUDIT_REASONING:
		return p.parseAuditReasoningStmt()
	case p.current.Type == token.CORRUPT_REASONING_STEP:
		return p.parseCorruptReasoningStepStmt()
	case p.current.Type == token.MODALITY:
		return p.parseModalityDecl()
	case p.current.Type == token.CROSS_MODAL_REASON:
		return p.parseCrossModalReasonStmt()
	case p.current.Type == token.QEC_MEMORY:
		return p.parseQECMemoryDecl()
	case p.current.Type == token.SYNDROME:
		return p.parseSyndromeStmt()
	case p.current.Type == token.VERIFY_SYNDROME:
		return p.parseVerifySyndromeStmt()
	case p.current.Type == token.CORRECT:
		return p.parseCorrectStmt()
	case p.current.Type == token.CORRUPT_SYNDROME:
		return p.parseCorruptSyndromeStmt()
	case p.current.Type == token.CONSOLIDATE:
		return p.parseConsolidateStmt()
	case p.current.Type == token.RECALL_TOKEN:
		return p.parseRecallTokenDecl()
	case p.current.Type == token.GRANT:
		return p.parseGrantStmt()
	case p.current.Type == token.REVOKE:
		return p.parseRevokeStmt()
	case p.current.Type == token.TOOL_REGISTER:
		return p.parseToolRegisterDecl()
	case p.current.Type == token.TOOL_CALL:
		return p.parseToolCallStmt(nil)
	case p.current.Type == token.CERTIFIED_REASON:
		return p.parseCertifiedReasonStmt()
	case p.current.Type == token.ABSTAIN_THRESHOLD:
		return p.parseAbstainThresholdDecl()
	case p.current.Type == token.ABSTAIN_CHECK:
		return p.parseAbstainCheckStmt()
	case p.current.Type == token.SANDBOX_CREATE:
		return p.parseSandboxCreateStmt()
	case p.current.Type == token.SANDBOX_RUN:
		return p.parseSandboxRunStmt()
	case p.current.Type == token.SANDBOX_VERIFY:
		return p.parseSandboxVerifyStmt()
	case p.current.Type == token.AT_GPU:
		return p.parseGPUAnnotatedStmt()
	case p.current.Type == token.GPU_MATMUL:
		return p.parseGPUMatmulStmt(false)
	case p.current.Type == token.GPU_SOFTMAX:
		return p.parseGPUSoftmaxStmt(false)
	case p.current.Type == token.GPU_LAYERNORM:
		return p.parseGPULayernormStmt(false)
	case p.current.Type == token.AGKE_ATTEST:
		return p.parseAGKEAttestStmt()
	case p.current.Type == token.AGKE_VERIFY:
		return p.parseAGKEVerifyStmt()
	case p.current.Type == token.CORRUPT_GPU_DIGEST:
		return p.parseCorruptGPUDigestStmt()
	case p.current.Type == token.MULTI_DEVICE:
		return p.parseMultiDeviceAnnotatedStmt()
	case p.current.Type == token.PIPELINE_CLAIM:
		return p.parsePipelineClaimStmt(false)
	case p.current.Type == token.PIPELINE_RUN:
		return p.parsePipelineRunStmt()
	case p.current.Type == token.CDAGA_RECORD:
		return p.parseCDAGARecordStmt()
	case p.current.Type == token.CDAGA_ATTEST:
		return p.parseCDAGAAttestStmt()
	case p.current.Type == token.CDAGA_VERIFY:
		return p.parseCDAGAVerifyStmt()
	case p.current.Type == token.CORRUPT_CDAGA:
		return p.parseCorruptCDAGAStmt()
	case p.current.Type == token.DIST_INIT:
		return p.parseDistributedAnnotatedStmt()
	case p.current.Type == token.ALL_REDUCE:
		return p.parseAllReduceStmt()
	case p.current.Type == token.PSERVER_INIT:
		return p.parsePserverInitStmt()
	case p.current.Type == token.PSERVER_AGG:
		return p.parsePserverAggregateStmt()
	case p.current.Type == token.BARAT_SUBMIT:
		return p.parseBaratSubmitStmt()
	case p.current.Type == token.BARAT_AGGREGATE:
		return p.parseBaratAggregateStmt()
	case p.current.Type == token.BARAT_VERIFY:
		return p.parseBaratVerifyStmt()
	case p.current.Type == token.CORRUPT_BARAT:
		return p.parseCorruptBaratStmt()
	case p.current.Type == token.QHAL_BACKEND:
		return p.parseQHALBackendAnnotation()
	case p.current.Type == token.QASM_EXPORT:
		return p.parseQASMExportStmt()
	case p.current.Type == token.EVOLVE:
		return p.parseEvolveStmt()
	case p.current.Type == token.EVOLVE_VERIFY:
		return p.parseEvolveVerifyStmt()
	case p.current.Type == token.EVOLVE_HISTORY:
		return p.parseEvolveHistoryStmt()
	case p.current.Type == token.CORRUPT_EVOLVE:
		return p.parseCorruptEvolveStmt()
	case p.current.Type == token.TRANSPILE_TO:
		return p.parseTranspileToStmt()
	case p.current.Type == token.EVOLVE_POLICY:
		return p.parseEvolvePolicyStmt()
	case p.current.Type == token.EVOLVE_BOUNDED:
		return p.parseEvolveBoundedStmt()
	case p.current.Type == token.EEA_VERIFY:
		return p.parseEEAVerifyStmt()
	case p.current.Type == token.CORRUPT_EEA:
		return p.parseCorruptEEAStmt()
	case p.current.Type == token.QECD_DECODE:
		return p.parseQECDDecodeStmt()
	case p.current.Type == token.QECD_VERIFY:
		return p.parseQECDVerifyStmt()
	case p.current.Type == token.QECD_APPLY:
		return p.parseQECDApplyStmt()
	case p.current.Type == token.AAD_CERT:
		return p.parseAADCertStmt()
	case p.current.Type == token.AAD_ATTEST:
		return p.parseAADAttestStmt()
	case p.current.Type == token.META_FIT:
		return p.parseMetaFitStmt()
	case p.current.Type == token.META_ADAPT:
		return p.parseMetaAdaptStmt()
	case p.current.Type == token.META_VERIFY:
		return p.parseMetaVerifyStmt()
	case p.current.Type == token.CORRUPT_AAD_ENV:
		return p.parseCorruptAADEnvStmt()
	case p.current.Type == token.RECORD:
		return p.parseRecordDecl()
	case p.current.Type == token.VARIANT:
		return p.parseVariantDecl()
	case p.current.Type == token.MATCH:
		return p.parseMatchStmt()
	case p.current.Type == token.RSCA_ATTEST:
		return p.parseRSCAAttestStmt()
	case p.current.Type == token.RSCA_VERIFY:
		return p.parseRSCAVerifyStmt()
	case p.current.Type == token.CORRUPT_RSCA:
		return p.parseCorruptRSCAStmt()
	case p.current.Type == token.CTSA_ATTEST:
		return p.parseCTSAAttestStmt()
	case p.current.Type == token.CTSA_VERIFY:
		return p.parseCTSAVerifyStmt()
	case p.current.Type == token.CORRUPT_CTSA:
		return p.parseCorruptCTSAStmt()
	case isTypeToken(p.current.Type) || (p.current.Type == token.IDENT && p.peeked.Type == token.IDENT):
		return p.parseBindStmt()
	case p.current.Type == token.EMIT:
		return p.parseEmitStmt()
	case p.current.Type == token.OBSERVE:
		return p.parseObserveStmt()
	case p.current.Type == token.YIELD_KW:
		return p.parseYieldStmt()
	case p.current.Type == token.WHEN:
		return p.parseWhenStmt()
	case p.current.Type == token.ITERATE:
		return p.parseIterateStmt()
	case p.current.Type == token.CYCLE:
		return p.parseCycleStmt()
	case p.current.Type == token.SUPERPOSE:
		return p.parseSuperposeStmt()
	case p.current.Type == token.ENTANGLE:
		return p.parseEntangleStmt()
	case p.current.Type == token.COLLAPSE:
		return p.parseCollapseStmt()
	case p.current.Type == token.LEARN || p.current.Type == token.TRAIN:
		return p.parseLearnDecl()
	case p.current.Type == token.DATASET:
		return p.parseDatasetDecl()
	case p.current.Type == token.NORMALIZE:
		return p.parseNormalizeStmt()
	case p.current.Type == token.SHUFFLE:
		return p.parseShuffleStmt()
	case p.current.Type == token.BATCH:
		return p.parseBatchStmt()
	case p.current.Type == token.ENERGY:
		return p.parseEnergyDecl()
	case p.current.Type == token.MINIMIZE:
		return p.parseMinimizeStmt()
	case p.current.Type == token.TILE:
		return p.parseTileStmt()
	case p.current.Type == token.HALT:
		pos := p.curPos()
		p.nextToken()
		p.expect(token.STMT_END)
		return &ast.HaltStmt{Pos_: pos}
	case p.current.Type == token.SKIP:
		pos := p.curPos()
		p.nextToken()
		p.expect(token.STMT_END)
		return &ast.SkipStmt{Pos_: pos}
	default:
		return p.parseExprOrAssignStmt()
	}
}

func (p *Parser) parseBindStmt() *ast.BindStmt {
	pos := p.curPos()
	btype := p.parseType()
	name := p.current.Literal
	p.nextToken()
	p.expect(token.BIND)
	val := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.BindStmt{
		Pos_:  pos,
		Type:  btype,
		Name:  name,
		Value: val,
	}
}

func (p *Parser) parseEmitStmt() *ast.EmitStmt {
	pos := p.curPos()
	p.expect(token.EMIT)
	val := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.EmitStmt{
		Pos_:  pos,
		Value: val,
	}
}

func (p *Parser) parseObserveStmt() *ast.ObserveStmt {
	pos := p.curPos()
	p.expect(token.OBSERVE)
	val := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.ObserveStmt{
		Pos_:  pos,
		Value: val,
	}
}

func (p *Parser) parseYieldStmt() *ast.YieldStmt {
	pos := p.curPos()
	p.expect(token.YIELD_KW)
	var val ast.Expr
	if p.current.Type != token.STMT_END && p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
		val = p.parseExpression(LOWEST)
	}
	p.expect(token.STMT_END)
	return &ast.YieldStmt{
		Pos_:  pos,
		Value: val,
	}
}

func (p *Parser) parseWhenStmt() *ast.WhenStmt {
	pos := p.curPos()
	p.expect(token.WHEN)
	cond := p.parseExpression(LOWEST)
	thenBlock := p.parseBlock()
	var elseBlock *ast.Block

	if p.current.Type == token.OTHERWISE {
		p.nextToken() // consume otherwise
		elseBlock = p.parseBlock()
	}

	return &ast.WhenStmt{
		Pos_:      pos,
		Condition: cond,
		Then:      thenBlock,
		Else:      elseBlock,
	}
}

func (p *Parser) parseIterateStmt() *ast.IterateStmt {
	pos := p.curPos()
	p.expect(token.ITERATE)
	varTok := p.expect(token.IDENT)
	p.expect(token.MEMBER) // ∈
	col := p.parseExpression(LOWEST)
	body := p.parseBlock()

	return &ast.IterateStmt{
		Pos_:       pos,
		Var:        varTok.Literal,
		Collection: col,
		Body:       body,
	}
}

func (p *Parser) parseCycleStmt() *ast.CycleStmt {
	pos := p.curPos()
	p.expect(token.CYCLE)
	cond := p.parseExpression(LOWEST)
	body := p.parseBlock()

	return &ast.CycleStmt{
		Pos_:      pos,
		Condition: cond,
		Body:      body,
	}
}

func (p *Parser) parseSuperposeStmt() *ast.SuperposeStmt {
	pos := p.curPos()
	p.expect(token.SUPERPOSE)
	q := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.SuperposeStmt{
		Pos_:  pos,
		Qubit: q,
	}
}

func (p *Parser) parseEntangleStmt() *ast.EntangleStmt {
	pos := p.curPos()
	p.expect(token.ENTANGLE)
	q1 := p.parseExpression(LOWEST)
	if p.current.Type == token.SEP {
		p.nextToken()
	}
	q2 := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.EntangleStmt{
		Pos_: pos,
		Q1:   q1,
		Q2:   q2,
	}
}

func (p *Parser) parseCollapseStmt() *ast.CollapseStmt {
	pos := p.curPos()
	p.expect(token.COLLAPSE)
	q := p.parseExpression(LOWEST)
	var into ast.Expr
	if p.current.Type == token.BIND {
		p.nextToken() // consume ↦
		into = p.parseExpression(LOWEST)
	}
	p.expect(token.STMT_END)
	return &ast.CollapseStmt{
		Pos_:  pos,
		Qubit: q,
		Into:  into,
	}
}

func (p *Parser) parseQubitDeclStmt() *ast.QubitDecl {
	pos := p.curPos()
	p.expect(token.QUBIT)
	var names []string
	for {
		nameTok := p.expect(token.IDENT)
		names = append(names, nameTok.Literal)
		if p.current.Type == token.SEP {
			p.nextToken() // consume ⸴
		} else {
			break
		}
	}
	p.expect(token.STMT_END)
	return &ast.QubitDecl{Pos_: pos, Names: names}
}

func (p *Parser) parseExprOrAssignStmt() ast.Stmt {
	pos := p.curPos()
	expr := p.parseExpression(LOWEST)

	if ident, ok := expr.(*ast.Ident); ok && p.current.Type == token.BIND {
		p.nextToken()
		rhs := p.parseExpression(LOWEST)
		p.expect(token.STMT_END)
		return &ast.BindStmt{
			Pos_:  pos,
			Type:  nil,
			Name:  ident.Name,
			Value: rhs,
		}
	}

	if isAssignOp(p.current.Type) {
		op := p.current.Type
		p.nextToken()
		rhs := p.parseExpression(LOWEST)
		p.expect(token.STMT_END)
		return &ast.AssignStmt{
			Pos_: pos,
			LHS:  expr,
			Op:   op,
			RHS:  rhs,
		}
	}

	p.expect(token.STMT_END)
	return &ast.ExprStmt{
		Pos_: pos,
		Expr: expr,
	}
}

func (p *Parser) parseDatasetDecl() *ast.DatasetDecl {
	pos := p.curPos()
	p.nextToken() // consume 'dataset'
	nameTok := p.expect(token.IDENT)
	p.expect(token.BIND)
	p.expect(token.LOAD)
	closeTok := token.GROUP_CLOSE
	if p.current.Type == token.INDEX_OPEN {
		closeTok = token.INDEX_CLOSE
		p.nextToken()
	} else {
		p.expect(token.GROUP_OPEN)
	}
	pathExpr := p.parseExpression(LOWEST)
	p.expect(closeTok)
	p.expect(token.STMT_END)
	return &ast.DatasetDecl{
		Pos_: pos,
		Name: nameTok.Literal,
		Path: pathExpr,
	}
}

func (p *Parser) parseNormalizeStmt() *ast.NormalizeStmt {
	pos := p.curPos()
	p.nextToken() // consume 'normalize'
	dExpr := p.parseExpression(CALL_PREC)
	mode := p.current.Type
	if mode != token.MINMAX && mode != token.ZSCORE {
		p.errorf(p.curPos(), "expected minmax or zscore, got %s", p.current.Literal)
	} else {
		p.nextToken()
	}
	p.expect(token.STMT_END)
	return &ast.NormalizeStmt{
		Pos_:    pos,
		Dataset: dExpr,
		Mode:    mode,
	}
}

func (p *Parser) parseShuffleStmt() *ast.ShuffleStmt {
	pos := p.curPos()
	p.nextToken() // consume 'shuffle'
	dExpr := p.parseExpression(CALL_PREC)
	p.expect(token.STMT_END)
	return &ast.ShuffleStmt{
		Pos_:    pos,
		Dataset: dExpr,
	}
}

func (p *Parser) parseBatchStmt() *ast.BatchStmt {
	pos := p.curPos()
	p.nextToken() // consume 'batch'
	dExpr := p.parseExpression(CALL_PREC)
	if p.current.Type == token.IDENT && p.current.Literal == "size" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}
	sizeExpr := p.parseExpression(CALL_PREC)
	var body *ast.Block
	switch p.current.Type {
	case token.BLOCK_OPEN:
		body = p.parseBlock()
	case token.STMT_END:
		p.nextToken()
	}
	return &ast.BatchStmt{
		Pos_:    pos,
		Dataset: dExpr,
		Size:    sizeExpr,
		Body:    body,
	}
}

func (p *Parser) parseEnergyDecl() *ast.EnergyDecl {
	pos := p.curPos()
	p.nextToken() // consume 'energy'
	nameTok := p.expect(token.IDENT)
	body := p.parseBlock()
	return &ast.EnergyDecl{
		Pos_: pos,
		Name: nameTok.Literal,
		Fn:   body,
	}
}

func (p *Parser) parseMinimizeStmt() *ast.MinimizeStmt {
	pos := p.curPos()
	p.nextToken() // consume 'minimize'
	gridExpr := p.parseExpression(CALL_PREC)
	p.expect(token.ENERGY)
	energyExpr := p.parseExpression(CALL_PREC)
	if p.current.Type == token.IDENT && p.current.Literal == "iterations" {
		p.nextToken()
	}
	p.expect(token.BIND)
	iterExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.MinimizeStmt{
		Pos_:       pos,
		Grid:       gridExpr,
		Energy:     energyExpr,
		Iterations: iterExpr,
	}
}

func (p *Parser) parseTileStmt() *ast.TileStmt {
	pos := p.curPos()
	p.nextToken() // consume 'tile'
	gridExpr := p.parseExpression(CALL_PREC)
	patternExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.TileStmt{
		Pos_:    pos,
		Grid:    gridExpr,
		Pattern: patternExpr,
	}
}

// ---------------------------------------------------------------------
// Expressions (Pratt Parser)
// ---------------------------------------------------------------------

func (p *Parser) parseExpression(precedence int) ast.Expr {
	left := p.parsePrefix()
	if left == nil {
		return nil
	}

	for p.current.Type != token.STMT_END && p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF && precedence < precedenceOf(p.current.Type) {
		switch p.current.Type {
		case token.GROUP_OPEN:
			left = p.parseCall(left)
		case token.INDEX_OPEN:
			left = p.parseIndex(left)
		case token.MEMBER_ACCESS, token.DOT:
			left = p.parseMember(left)
		case token.RANGE:
			pos := p.curPos()
			p.nextToken()
			end := p.parseExpression(LOWEST)
			left = &ast.RangeExpr{
				Pos_:  pos,
				Start: left,
				End:   end,
			}
		default:
			left = p.parseInfix(left)
		}
	}

	return left
}

func (p *Parser) parsePrefix() ast.Expr {
	pos := p.curPos()
	switch p.current.Type {
	case token.IDENT, token.AFFIRM, token.DENY, token.TOP, token.BOTTOM, token.LOGIC, token.RESULT_T, token.ERROR_T:
		lit := p.current.Literal
		p.nextToken()
		return &ast.Ident{Pos_: pos, Name: lit}
	case token.INT:
		lit := p.current.Literal
		p.nextToken()
		val := parseIntLiteral(lit)
		return &ast.IntLit{Pos_: pos, Value: val}
	case token.FLOAT:
		lit := p.current.Literal
		p.nextToken()
		val, isImag := parseFloatLiteral(lit)
		return &ast.FloatLit{Pos_: pos, Value: val, Imaginary: isImag}
	case token.STRING:
		lit := p.current.Literal
		p.nextToken()
		return &ast.StringLit{Pos_: pos, Value: lit}
	case token.TENSOR:
		return p.parseTensorLit()
	case token.NEG, token.ODIFF, token.ADJOINT, token.SUM, token.PRODUCT:
		op := p.current.Type
		p.nextToken()
		right := p.parseExpression(PREFIX)
		return &ast.UnaryExpr{
			Pos_:  pos,
			Op:    op,
			Right: right,
		}
	case token.GROUP_OPEN:
		p.nextToken()
		inner := p.parseExpression(LOWEST)
		p.expect(token.GROUP_CLOSE)
		return &ast.GroupExpr{
			Pos_:  pos,
			Inner: inner,
		}
	case token.RECORD_INIT:
		return p.parseRecordInitExpr()
	case token.VARIANT_INIT:
		return p.parseVariantInitExpr()
	// collapse as expression: collapse q0  (returns int64 0 or 1)
	case token.COLLAPSE:
		p.nextToken()
		q := p.parseExpression(LOWEST)
		return &ast.CollapseExpr{Pos_: pos, Qubit: q}
	// measure as expression: measure q0 (alias for collapse)
	case token.MEASURE:
		p.nextToken()
		q := p.parseExpression(LOWEST)
		return &ast.CollapseExpr{Pos_: pos, Qubit: q}
	// concept literal in expression position: concept 613
	case token.CONCEPT:
		p.nextToken()
		if p.current.Type == token.INT {
			val := parseIntLiteral(p.current.Literal)
			p.nextToken()
			return &ast.ConceptLit{Pos_: pos, Value: val}
		}
		// concept followed by IDENT — treat as identifier
		lit := p.current.Literal
		p.nextToken()
		return &ast.Ident{Pos_: pos, Name: lit}
	// decompose⦅expr⦆
	case token.DECOMPOSE:
		p.nextToken()
		p.expect(token.GROUP_OPEN)
		inner := p.parseExpression(LOWEST)
		p.expect(token.GROUP_CLOSE)
		return &ast.DecomposeExpr{Pos_: pos, Input: inner}
	case token.RECALL:
		p.nextToken()
		memName := p.expect(token.IDENT).Literal
		closeTok := token.INDEX_CLOSE
		if p.current.Type == token.GROUP_OPEN {
			closeTok = token.GROUP_CLOSE
			p.nextToken()
		} else {
			p.expect(token.INDEX_OPEN)
		}
		key := p.parseExpression(LOWEST)
		p.expect(closeTok)
		return &ast.RecallExpr{Pos_: pos, Memory: &ast.Ident{Pos_: pos, Name: memName}, Key: key}
	case token.RECALL_SCOPED:
		p.nextToken()
		memName := p.expect(token.IDENT).Literal
		closeTok := token.INDEX_CLOSE
		if p.current.Type == token.GROUP_OPEN {
			closeTok = token.GROUP_CLOSE
			p.nextToken()
		} else {
			p.expect(token.INDEX_OPEN)
		}
		key := p.parseExpression(LOWEST)
		p.expect(closeTok)
		if p.current.Type == token.WITH || (p.current.Type == token.IDENT && (p.current.Literal == "with" || p.current.Literal == "using")) {
			p.nextToken()
		}
		tokName := p.expect(token.IDENT).Literal
		return &ast.ScopedRecallExpr{
			Pos_:   pos,
			Memory: &ast.Ident{Pos_: pos, Name: memName},
			Key:    key,
			Token:  &ast.Ident{Pos_: pos, Name: tokName},
		}
	case token.TOOL_CALL:
		p.nextToken()
		toolName := p.expect(token.IDENT).Literal
		p.expect(token.GROUP_OPEN)
		var args []ast.Expr
		for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
			arg := p.parseExpression(LOWEST)
			if arg != nil {
				args = append(args, arg)
			}
			if p.current.Type == token.SEP {
				p.nextToken()
			} else {
				break
			}
		}
		p.expect(token.GROUP_CLOSE)
		var audit ast.Expr
		if p.current.Type == token.TOOL_AUDIT || (p.current.Type == token.IDENT && p.current.Literal == "audit") {
			p.nextToken()
			audit = p.parseExpression(CALL_PREC)
		}
		return &ast.ToolCallExpr{
			Pos_:     pos,
			ToolName: toolName,
			Args:     args,
			Audit:    audit,
		}
	case token.ABSTAIN_CHECK:
		p.nextToken()
		var chain ast.Expr
		if p.current.Type == token.IDENT || p.current.Type == token.LOGIC {
			chain = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
			p.nextToken()
		} else {
			chain = p.parseExpression(CALL_PREC)
		}
		return &ast.AbstainCheckExpr{
			Pos_:  pos,
			Chain: chain,
		}
	case token.SANDBOX_VERIFY:
		p.nextToken()
		sbName := p.expect(token.IDENT).Literal
		return &ast.SandboxVerifyExpr{
			Pos_:    pos,
			Sandbox: sbName,
		}
	case token.GPU_MATMUL:
		p.nextToken()
		p.expect(token.GROUP_OPEN)
		left := p.parseExpression(LOWEST)
		p.expect(token.SEP)
		right := p.parseExpression(LOWEST)
		p.expect(token.GROUP_CLOSE)
		return &ast.GPUMatmulExpr{
			Pos_:  pos,
			Left:  left,
			Right: right,
		}
	case token.GPU_SOFTMAX:
		p.nextToken()
		p.expect(token.GROUP_OPEN)
		src := p.parseExpression(LOWEST)
		p.expect(token.GROUP_CLOSE)
		return &ast.GPUSoftmaxExpr{
			Pos_:   pos,
			Source: src,
		}
	case token.GPU_LAYERNORM:
		p.nextToken()
		p.expect(token.GROUP_OPEN)
		src := p.parseExpression(LOWEST)
		p.expect(token.GROUP_CLOSE)
		return &ast.GPULayernormExpr{
			Pos_:   pos,
			Source: src,
		}
	case token.AGKE_VERIFY:
		p.nextToken()
		var chain ast.Expr
		if p.current.Type == token.IDENT || p.current.Type == token.LOGIC {
			chain = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
			p.nextToken()
		} else {
			chain = p.parseExpression(CALL_PREC)
		}
		return &ast.AGKEVerifyExpr{
			Pos_:  pos,
			Chain: chain,
		}
	case token.TRACE:
		p.nextToken()
		graphName := p.expect(token.IDENT).Literal
		closeTok := token.INDEX_CLOSE
		if p.current.Type == token.GROUP_OPEN {
			closeTok = token.GROUP_CLOSE
			p.nextToken()
		} else {
			p.expect(token.INDEX_OPEN)
		}
		start := p.parseExpression(LOWEST)
		p.expect(token.SEP)
		depth := p.parseExpression(LOWEST)
		p.expect(closeTok)
		return &ast.TraceExpr{Pos_: pos, Graph: &ast.Ident{Pos_: pos, Name: graphName}, Start: start, Depth: depth}
	case token.EXECUTE:
		p.nextToken()
		procName := p.expect(token.IDENT).Literal
		p.expect(token.GROUP_OPEN)
		var args []ast.Expr
		for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
			arg := p.parseExpression(LOWEST)
			if arg != nil {
				args = append(args, arg)
			}
			if p.current.Type == token.SEP {
				p.nextToken()
			} else {
				break
			}
		}
		p.expect(token.GROUP_CLOSE)
		return &ast.ExecuteExpr{Pos_: pos, Procedure: &ast.Ident{Pos_: pos, Name: procName}, Args: args}
	case token.VERIFY_TRAIN:
		p.nextToken()
		closeTok := token.INDEX_CLOSE
		if p.current.Type == token.GROUP_OPEN {
			closeTok = token.GROUP_CLOSE
			p.nextToken()
		} else {
			p.expect(token.INDEX_OPEN)
		}
		cert := p.parseExpression(LOWEST)
		p.expect(closeTok)
		return &ast.VerifyTrainingExpr{Pos_: pos, Certificate: cert}
	case token.QUASI_NEG:
		p.nextToken()
		return &ast.QuasiLit{Pos_: pos, Value: -1}
	case token.QUASI_ZERO:
		p.nextToken()
		return &ast.QuasiLit{Pos_: pos, Value: 0}
	case token.QUASI_POS:
		p.nextToken()
		return &ast.QuasiLit{Pos_: pos, Value: 1}
	default:
		p.errorf(pos, "unexpected token in expression: %s (%q)", p.current.Type, p.current.Literal)
		p.nextToken()
		return nil
	}
}

func (p *Parser) parseInfix(left ast.Expr) ast.Expr {
	pos := p.curPos()
	op := p.current.Type
	prec := precedenceOf(op)
	p.nextToken()
	right := p.parseExpression(prec)

	if op == token.ANNOT {
		scoreVal := 1.0
		if fl, ok := left.(*ast.FloatLit); ok {
			scoreVal = fl.Value
		} else if il, ok := left.(*ast.IntLit); ok {
			scoreVal = float64(il.Value)
		}
		return &ast.ConfidentLit{
			Pos_:  pos,
			Value: right,
			Score: scoreVal,
			State: types.KnownState,
		}
	}

	return &ast.BinaryExpr{
		Pos_:  pos,
		Op:    op,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseCall(fn ast.Expr) ast.Expr {
	pos := p.curPos()
	p.expect(token.GROUP_OPEN)
	args := []ast.Expr{}

	for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
		arg := p.parseExpression(LOWEST)
		if arg != nil {
			args = append(args, arg)
		}
		if p.current.Type == token.SEP {
			p.nextToken()
		} else {
			break
		}
	}

	p.expect(token.GROUP_CLOSE)
	return &ast.CallExpr{
		Pos_: pos,
		Func: fn,
		Args: args,
	}
}

func (p *Parser) parseIndex(obj ast.Expr) ast.Expr {
	pos := p.curPos()
	p.expect(token.INDEX_OPEN)
	idx := p.parseExpression(LOWEST)
	p.expect(token.INDEX_CLOSE)

	return &ast.IndexExpr{
		Pos_:   pos,
		Object: obj,
		Index:  idx,
	}
}

func (p *Parser) parseMember(obj ast.Expr) ast.Expr {
	pos := p.curPos()
	if p.current.Type == token.DOT {
		p.expect(token.DOT)
	} else {
		p.expect(token.MEMBER_ACCESS) // ∸
	}
	memName := p.current.Literal
	p.nextToken()

	return &ast.FieldAccessExpr{
		Pos_:   pos,
		Object: obj,
		Field:  memName,
	}
}

// ---------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------

func (p *Parser) parseTensorLit() *ast.TensorLit {
	pos := p.curPos()
	p.expect(token.TENSOR)
	return p.parseTensorLitInner(pos)
}

func (p *Parser) parseTensorLitInner(pos ast.Position) *ast.TensorLit {
	p.expect(token.INDEX_OPEN)
	var items []ast.Expr
	hasSubTensors := false
	hasScalars := false

	for p.current.Type != token.INDEX_CLOSE && p.current.Type != token.EOF {
		if p.current.Type == token.INDEX_OPEN || p.current.Type == token.TENSOR {
			if hasScalars {
				p.errorf(p.curPos(), "cannot mix scalar and tensor in tensor literal")
			}
			hasSubTensors = true
			if p.current.Type == token.TENSOR {
				p.nextToken()
			}
			subPos := p.curPos()
			subLit := p.parseTensorLitInner(subPos)
			items = append(items, subLit)
		} else {
			if hasSubTensors {
				p.errorf(p.curPos(), "cannot mix scalar and tensor in tensor literal")
			}
			hasScalars = true
			expr := p.parseExpression(LOWEST)
			if expr != nil {
				items = append(items, expr)
			}
		}

		if p.current.Type == token.SEP {
			p.nextToken()
		} else {
			break
		}
	}
	p.expect(token.INDEX_CLOSE)

	dims := []int{len(items)}
	if hasSubTensors && len(items) > 0 {
		firstSub, ok := items[0].(*ast.TensorLit)
		if ok {
			subDims := firstSub.Dims
			for i := 1; i < len(items); i++ {
				sub, ok := items[i].(*ast.TensorLit)
				if !ok || len(sub.Dims) != len(subDims) {
					p.errorf(pos, "different inner sizes in tensor literal")
					break
				}
				for d := range subDims {
					if sub.Dims[d] != subDims[d] {
						p.errorf(pos, "different inner sizes in tensor literal")
						break
					}
				}
			}
			dims = append([]int{len(items)}, subDims...)
		}
	}

	return &ast.TensorLit{
		Pos_:  pos,
		Items: items,
		Dims:  dims,
	}
}

func (p *Parser) parseType() ast.Type {
	pos := p.curPos()
	switch p.current.Type {
	case token.TENSOR:
		p.nextToken()
		if p.current.Type == token.INDEX_OPEN {
			p.errorf(pos, "tensor literal invalid in type position; use tensor⟨type⟩⟬dims⟭")
			return &ast.TensorType{Pos_: pos}
		}
		if p.current.Type != token.LANGLE {
			return &ast.TensorType{Pos_: pos, Elem: nil, Dims: nil}
		}
		p.expect(token.LANGLE)
		elemType := p.parseType()
		p.expect(token.RANGLE)
		var dims []int
		if p.current.Type == token.INDEX_OPEN {
			p.nextToken() // consume ⟬
			for p.current.Type != token.INDEX_CLOSE && p.current.Type != token.EOF {
				if p.current.Type == token.INT {
					val := parseIntLiteral(p.current.Literal)
					dims = append(dims, int(val))
					p.nextToken()
				} else {
					p.errorf(p.curPos(), "expected integer dimension in tensor type, got %s", p.current.Type)
					p.nextToken()
				}
				if p.current.Type == token.SEP {
					p.nextToken()
				} else {
					break
				}
			}
			p.expect(token.INDEX_CLOSE)
		}
		return &ast.TensorType{
			Pos_: pos,
			Elem: elemType,
			Dims: dims,
		}
	case token.CONFIDENT:
		p.nextToken()
		p.expect(token.LANGLE)
		elemType := p.parseType()
		p.expect(token.RANGLE)
		return &ast.NamedType{Pos_: pos, Name: fmt.Sprintf("confident⟨%s⟩", elemType)}
	case token.EXPLORE:
		p.nextToken()
		p.expect(token.LANGLE)
		elemType := p.parseType()
		p.expect(token.RANGLE)
		return &ast.NamedType{Pos_: pos, Name: fmt.Sprintf("explore⟨%s⟩", elemType)}
	case token.CONCEPT:
		p.nextToken()
		return &ast.PrimitiveType{Pos_: pos, Name: "concept"}
	case token.IDENT:
		name := p.current.Literal
		p.nextToken()
		return &ast.NamedType{Pos_: pos, Name: name}
	default:
		if isTypeToken(p.current.Type) {
			name := p.current.Literal
			p.nextToken()
			return &ast.PrimitiveType{Pos_: pos, Name: name}
		}
		p.errorf(pos, "expected type, got %s (%q)", p.current.Type, p.current.Literal)
		p.nextToken()
		return &ast.PrimitiveType{Pos_: pos, Name: "unknown"}
	}
}

// ---------------------------------------------------------------------
// Helper Functions
// ---------------------------------------------------------------------

func isTypeToken(t token.Type) bool {
	switch t {
	case token.INT8, token.INT16, token.INT32, token.INT64,
		token.UINT8, token.UINT16, token.UINT32, token.UINT64,
		token.FLOAT32, token.FLOAT64, token.COMPLEX64, token.COMPLEX128,
		token.LOGIC, token.TEXT, token.TENSOR, token.QUBIT, token.NEURON,
		token.CONFIDENT, token.EXPLORE, token.CONCEPT, token.QUASI,
		token.MEMORY, token.GRAPH, token.PROCEDURE,
		token.ERROR_T, token.RESULT_T:
		return true
	default:
		return false
	}
}

func isAssignOp(t token.Type) bool {
	switch t {
	case token.ASSIGN, token.BIND_ADD, token.BIND_SUB, token.BIND_MUL, token.BIND_DIV:
		return true
	default:
		return false
	}
}

func parseIntLiteral(lit string) int64 {
	clean := strings.ReplaceAll(lit, "_", "")
	val, err := strconv.ParseInt(clean, 0, 64)
	if err != nil {
		return 0
	}
	return val
}

func parseFloatLiteral(lit string) (float64, bool) {
	clean := strings.ReplaceAll(lit, "_", "")
	isImag := false
	if strings.HasSuffix(clean, "i") || strings.HasSuffix(clean, "j") {
		isImag = true
		clean = clean[:len(clean)-1]
	}
	val, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, isImag
	}
	return val, isImag
}

func (p *Parser) parseProcedureDecl() *ast.ProcedureDecl {
	pos := p.curPos()
	p.expect(token.PROCEDURE)
	nameTok := p.expect(token.IDENT)
	decl := &ast.ProcedureDecl{
		Pos_:   pos,
		Name:   nameTok.Literal,
		Params: []*ast.Param{},
	}
	p.expect(token.GROUP_OPEN)
	for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
		param := p.parseParam()
		if param != nil {
			decl.Params = append(decl.Params, param)
		}
		if p.current.Type == token.SEP {
			p.nextToken()
		} else {
			break
		}
	}
	p.expect(token.GROUP_CLOSE)
	if p.current.Type == token.YIELDS {
		p.nextToken()
		decl.RetType = p.parseType()
	}
	decl.Body = p.parseBlock()
	return decl
}

func (p *Parser) parseAssertStmt() *ast.AssertStmt {
	pos := p.curPos()
	p.nextToken() // consume 'assert'
	valExpr := p.parseExpression(CALL_PREC)
	// comparison operator (e.g. ≽)
	p.nextToken()
	threshExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.AssertStmt{
		Pos_:      pos,
		Value:     valExpr,
		Threshold: threshExpr,
	}
}

func (p *Parser) parseInferStmt() *ast.InferStmt {
	pos := p.curPos()
	p.nextToken() // consume 'infer'
	modelExpr := p.parseExpression(CALL_PREC)
	p.expect(token.ON)
	inputExpr := p.parseExpression(CALL_PREC)
	p.expect(token.BIND)
	intoTok := p.expect(token.IDENT)
	p.expect(token.STMT_END)
	return &ast.InferStmt{
		Pos_:  pos,
		Model: modelExpr,
		Input: inputExpr,
		Into:  intoTok.Literal,
	}
}

func (p *Parser) parseMemoryDecl() *ast.MemoryDecl {
	pos := p.curPos()
	p.nextToken() // consume 'memory'
	nameTok := p.expect(token.IDENT)
	var capExpr ast.Expr
	if p.current.Type == token.INDEX_OPEN || p.current.Type == token.GROUP_OPEN {
		closeTok := token.INDEX_CLOSE
		if p.current.Type == token.GROUP_OPEN {
			closeTok = token.GROUP_CLOSE
		}
		p.nextToken()
		capExpr = p.parseExpression(LOWEST)
		p.expect(closeTok)
	}
	p.expect(token.STMT_END)
	return &ast.MemoryDecl{
		Pos_:     pos,
		Name:     nameTok.Literal,
		Capacity: capExpr,
	}
}

func (p *Parser) parseRememberStmt() *ast.RememberStmt {
	pos := p.curPos()
	p.nextToken() // consume 'remember'
	memName := p.expect(token.IDENT).Literal
	closeTok := token.INDEX_CLOSE
	if p.current.Type == token.GROUP_OPEN {
		closeTok = token.GROUP_CLOSE
		p.nextToken()
	} else {
		p.expect(token.INDEX_OPEN)
	}
	key := p.parseExpression(LOWEST)
	p.expect(closeTok)
	p.expect(token.BIND)
	val := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.RememberStmt{
		Pos_:   pos,
		Memory: &ast.Ident{Pos_: pos, Name: memName},
		Key:    key,
		Value:  val,
	}
}

func (p *Parser) parseGraphDecl() *ast.GraphDecl {
	pos := p.curPos()
	p.nextToken() // consume 'graph'
	nameTok := p.expect(token.IDENT)
	p.expect(token.STMT_END)
	return &ast.GraphDecl{
		Pos_: pos,
		Name: nameTok.Literal,
	}
}

func (p *Parser) parseNodeStmt() *ast.NodeStmt {
	pos := p.curPos()
	p.nextToken() // consume 'node'
	graphName := p.expect(token.IDENT).Literal
	closeTok := token.INDEX_CLOSE
	if p.current.Type == token.GROUP_OPEN {
		closeTok = token.GROUP_CLOSE
		p.nextToken()
	} else {
		p.expect(token.INDEX_OPEN)
	}
	idExpr := p.parseExpression(CALL_PREC)
	var dataExpr ast.Expr
	if p.current.Type == token.BIND {
		p.nextToken()
		dataExpr = p.parseExpression(LOWEST)
	}
	p.expect(closeTok)
	p.expect(token.STMT_END)
	return &ast.NodeStmt{
		Pos_:  pos,
		Graph: &ast.Ident{Pos_: pos, Name: graphName},
		ID:    idExpr,
		Data:  dataExpr,
	}
}

func (p *Parser) parseLinkStmt() *ast.LinkStmt {
	pos := p.curPos()
	p.nextToken() // consume 'link'
	graphName := p.expect(token.IDENT).Literal
	closeTok := token.INDEX_CLOSE
	if p.current.Type == token.GROUP_OPEN {
		closeTok = token.GROUP_CLOSE
		p.nextToken()
	} else {
		p.expect(token.INDEX_OPEN)
	}
	fromExpr := p.parseExpression(LOWEST)
	p.expect(token.SEP)
	toExpr := p.parseExpression(LOWEST)
	p.expect(token.SEP)
	weightExpr := p.parseExpression(LOWEST)
	p.expect(closeTok)
	p.expect(token.STMT_END)
	return &ast.LinkStmt{
		Pos_:   pos,
		Graph:  &ast.Ident{Pos_: pos, Name: graphName},
		From:   fromExpr,
		To:     toExpr,
		Weight: weightExpr,
	}
}

func (p *Parser) parseExecuteStmt() *ast.ExecuteStmt {
	pos := p.curPos()
	p.nextToken() // consume 'execute'
	procName := p.expect(token.IDENT).Literal
	p.expect(token.GROUP_OPEN)
	var args []ast.Expr
	for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
		arg := p.parseExpression(LOWEST)
		if arg != nil {
			args = append(args, arg)
		}
		if p.current.Type == token.SEP {
			p.nextToken()
		} else {
			break
		}
	}
	p.expect(token.GROUP_CLOSE)
	p.expect(token.STMT_END)
	return &ast.ExecuteStmt{
		Pos_:      pos,
		Procedure: &ast.Ident{Pos_: pos, Name: procName},
		Args:      args,
	}
}

func (p *Parser) parseEthicalBoundaryDecl() *ast.EthicalBoundaryDecl {
	pos := p.curPos()
	p.nextToken() // consume 'ethical_boundary'

	name := ""
	if p.current.Type == token.IDENT {
		name = p.current.Literal
		p.nextToken()
	}

	p.expect(token.BLOCK_OPEN) // ⟦

	var rules []*ast.EthicalRule
	for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
		rulePos := p.curPos()
		if p.current.Type == token.RULE {
			p.nextToken() // consume 'rule'
		}
		ruleName := ""
		if p.current.Type == token.IDENT || p.current.Type == token.ACCESS || p.current.Type == token.STRING {
			ruleName = p.current.Literal
			p.nextToken()
		}
		if p.current.Type == token.BIND { // ↦
			p.nextToken()
		}
		valExpr := p.parseExpression(LOWEST)
		penalty := 1.0
		if lit, ok := valExpr.(*ast.FloatLit); ok {
			penalty = lit.Value
		} else if lit, ok := valExpr.(*ast.IntLit); ok {
			penalty = float64(lit.Value)
		}
		p.expect(token.STMT_END) // ·
		rules = append(rules, &ast.EthicalRule{
			Pos_:      rulePos,
			Name:      ruleName,
			Condition: valExpr,
			Penalty:   penalty,
		})
	}
	p.expect(token.BLOCK_CLOSE) // ⟧
	return &ast.EthicalBoundaryDecl{
		Pos_:  pos,
		Name:  name,
		Rules: rules,
	}
}

func (p *Parser) parseEnforceStmt() *ast.EnforceStmt {
	pos := p.curPos()
	p.nextToken() // consume 'enforce'

	boundary := p.parseExpression(LOWEST)

	if p.current.Type == token.ON {
		p.nextToken() // consume 'on'
	}

	action := p.parseExpression(LOWEST)
	p.expect(token.STMT_END) // ·

	return &ast.EnforceStmt{
		Pos_:     pos,
		Boundary: boundary,
		Action:   action,
	}
}

func (p *Parser) parseDecoherenceStmt() *ast.DecoherenceStmt {
	pos := p.curPos()
	p.nextToken() // consume 'decoherence'

	var target ast.Expr
	if p.current.Type != token.STMT_END {
		target = p.parseExpression(LOWEST)
	}
	p.expect(token.STMT_END) // ·

	return &ast.DecoherenceStmt{
		Pos_:   pos,
		Target: target,
	}
}

func (p *Parser) parseVaultDecl() *ast.VaultDecl {
	pos := p.curPos()
	p.nextToken() // consume 'vault'

	name := ""
	if p.current.Type == token.IDENT {
		name = p.current.Literal
		p.nextToken()
	}

	var boundary ast.Expr
	if p.current.Type == token.WITH {
		p.nextToken()
		if p.current.Type == token.IDENT {
			boundary = &ast.Ident{
				Pos_: p.curPos(),
				Name: p.current.Literal,
			}
			p.nextToken()
		} else {
			boundary = p.parseExpression(CALL_PREC)
		}
	}

	var netRef ast.Expr
	if p.current.Type == token.INDEX_OPEN {
		p.nextToken()
		netRef = p.parseExpression(LOWEST)
		p.expect(token.INDEX_CLOSE)
	}

	var thresh ast.Expr
	if p.current.Type == token.IDENT && (p.current.Literal == "threshold" || p.current.Literal == "shares") {
		p.nextToken()
		if p.current.Type == token.BIND {
			p.nextToken()
		}
		thresh = p.parseExpression(LOWEST)
	}
	p.expect(token.STMT_END) // ·

	return &ast.VaultDecl{
		Pos_:       pos,
		Name:       name,
		Boundary:   boundary,
		NetworkRef: netRef,
		Threshold:  thresh,
	}
}

func (p *Parser) parseSealStmt() *ast.SealStmt {
	pos := p.curPos()
	p.nextToken() // consume 'seal'

	vaultExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END) // ·

	return &ast.SealStmt{
		Pos_:  pos,
		Vault: vaultExpr,
	}
}

func (p *Parser) parseUnsealStmt() *ast.UnsealStmt {
	pos := p.curPos()
	p.nextToken() // consume 'unseal'

	vaultExpr := p.parseExpression(LOWEST)
	if p.current.Type == token.WITH {
		p.nextToken()
	}
	boundary := p.parseExpression(LOWEST)
	p.expect(token.STMT_END) // ·

	return &ast.UnsealStmt{
		Pos_:     pos,
		Vault:    vaultExpr,
		Boundary: boundary,
	}
}

func (p *Parser) parseWipeStmt() *ast.WipeStmt {
	pos := p.curPos()
	p.nextToken() // consume 'wipe'

	vaultExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END) // ·

	return &ast.WipeStmt{
		Pos_:  pos,
		Vault: vaultExpr,
	}
}

func (p *Parser) parseProvenanceDecl() *ast.ProvenanceDecl {
	pos := p.curPos()
	p.nextToken() // consume 'provenance'

	name := ""
	if p.current.Type == token.IDENT {
		name = p.current.Literal
		p.nextToken()
	}

	var netRef ast.Expr
	if p.current.Type == token.INDEX_OPEN {
		p.nextToken()
		netRef = p.parseExpression(LOWEST)
		p.expect(token.INDEX_CLOSE)
	}
	p.expect(token.STMT_END) // ·

	return &ast.ProvenanceDecl{
		Pos_:       pos,
		Name:       name,
		NetworkRef: netRef,
	}
}

func (p *Parser) parseCertifyStmt() *ast.CertifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'certify'

	provExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END) // ·

	return &ast.CertifyStmt{
		Pos_:       pos,
		Provenance: provExpr,
	}
}

func (p *Parser) parseMemoryProvenanceDecl() *ast.MemoryProvenanceDecl {
	pos := p.curPos()
	p.nextToken() // consume 'memory_prov'

	name := ""
	if p.current.Type == token.IDENT {
		name = p.current.Literal
		p.nextToken()
	}

	var boundary ast.Expr
	if p.current.Type == token.WITH {
		p.nextToken()
		if p.current.Type == token.IDENT {
			boundary = &ast.Ident{
				Pos_: p.curPos(),
				Name: p.current.Literal,
			}
			p.nextToken()
		} else {
			boundary = p.parseExpression(CALL_PREC)
		}
	}

	var capExpr ast.Expr
	if p.current.Type == token.BLOCK_OPEN {
		p.nextToken() // consume '⟦'
		if p.current.Type == token.IDENT && p.current.Literal == "capacity" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			capExpr = p.parseExpression(LOWEST)
			if p.current.Type == token.STMT_END {
				p.nextToken()
			}
		}
		p.expect(token.BLOCK_CLOSE) // '⟧'
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.MemoryProvenanceDecl{
		Pos_:     pos,
		Name:     name,
		Boundary: boundary,
		Capacity: capExpr,
	}
}

func (p *Parser) parseVerifyMemoryStmt() *ast.VerifyMemoryStmt {
	pos := p.curPos()
	p.nextToken() // consume 'verify_memory'
	mem := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.VerifyMemoryStmt{Pos_: pos, Mem: mem}
}

func (p *Parser) parseAuditMemoryStmt() *ast.AuditMemoryStmt {
	pos := p.curPos()
	p.nextToken() // consume 'audit_memory'
	mem := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.AuditMemoryStmt{Pos_: pos, Mem: mem}
}

func (p *Parser) parseCorruptShareStmt() *ast.CorruptShareStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_memory_share'

	memName := ""
	if p.current.Type == token.IDENT {
		memName = p.current.Literal
		p.nextToken()
	}

	closeTok := token.INDEX_CLOSE
	if p.current.Type == token.GROUP_OPEN {
		closeTok = token.GROUP_CLOSE
		p.nextToken()
	} else {
		p.expect(token.INDEX_OPEN)
	}
	key := p.parseExpression(LOWEST)
	p.expect(closeTok)

	if p.current.Type == token.IDENT && p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}
	shareIdx := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.CorruptShareStmt{
		Pos_:       pos,
		Mem:        &ast.Ident{Pos_: pos, Name: memName},
		Key:        key,
		ShareIndex: shareIdx,
	}
}

func (p *Parser) parseLicenseDecl() *ast.LicenseDecl {
	pos := p.curPos()
	p.nextToken() // consume 'license'

	name := ""
	if p.current.Type == token.IDENT {
		name = p.current.Literal
		p.nextToken()
	}

	// optional 'for'
	if p.current.Type == token.IDENT && p.current.Literal == "for" {
		p.nextToken()
	}

	var teacher ast.Expr
	if p.current.Type == token.IDENT {
		teacher = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	// optional 'on'
	if p.current.Type == token.ON || (p.current.Type == token.IDENT && p.current.Literal == "on") {
		p.nextToken()
	}

	var student ast.Expr
	if p.current.Type == token.IDENT {
		student = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	var maxEpochs, maxRate, expiry ast.Expr
	if p.current.Type == token.BLOCK_OPEN {
		p.nextToken() // consume '⟦'
		for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
			if p.current.Type == token.IDENT {
				field := p.current.Literal
				p.nextToken()
				if p.current.Type == token.BIND {
					p.nextToken()
				}
				val := p.parseExpression(LOWEST)
				switch field {
				case "max_epochs":
					maxEpochs = val
				case "max_rate":
					maxRate = val
				case "expiry":
					expiry = val
				}
				if p.current.Type == token.STMT_END {
					p.nextToken()
				}
			} else {
				p.nextToken()
			}
		}
		p.expect(token.BLOCK_CLOSE) // '⟧'
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.LicenseDecl{
		Pos_:      pos,
		Name:      name,
		Teacher:   teacher,
		Student:   student,
		MaxEpochs: maxEpochs,
		MaxRate:   maxRate,
		Expiry:    expiry,
	}
}

func (p *Parser) parseTransferStmt() *ast.TransferStmt {
	pos := p.curPos()
	p.nextToken() // consume 'transfer'

	var lic ast.Expr
	if p.current.Type == token.IDENT {
		lic = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	// optional 'from'
	if p.current.Type == token.IDENT && p.current.Literal == "from" {
		p.nextToken()
	}

	var teacher ast.Expr
	if p.current.Type == token.IDENT {
		teacher = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	// optional 'to'
	if p.current.Type == token.IDENT && p.current.Literal == "to" {
		p.nextToken()
	}

	var student ast.Expr
	if p.current.Type == token.IDENT {
		student = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	// optional 'layers'
	if p.current.Type == token.IDENT && p.current.Literal == "layers" {
		p.nextToken()
	}

	var layers []ast.Expr
	if p.current.Type == token.INDEX_OPEN || p.current.Type == token.GROUP_OPEN {
		closeTok := token.INDEX_CLOSE
		if p.current.Type == token.GROUP_OPEN {
			closeTok = token.GROUP_CLOSE
		}
		p.nextToken() // consume '⟬'
		for p.current.Type != closeTok && p.current.Type != token.EOF {
			elem := p.parseExpression(LOWEST)
			if elem != nil {
				layers = append(layers, elem)
			}
			if p.current.Type == token.SEP {
				p.nextToken()
			} else {
				break
			}
		}
		p.expect(closeTok)
	}

	p.expect(token.STMT_END)

	return &ast.TransferStmt{
		Pos_:    pos,
		License: lic,
		Teacher: teacher,
		Student: student,
		Layers:  layers,
	}
}

func (p *Parser) parseDistillStmt() *ast.DistillStmt {
	pos := p.curPos()
	p.nextToken() // consume 'distill'

	var lic ast.Expr
	if p.current.Type == token.IDENT {
		lic = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	// optional 'on'
	if p.current.Type == token.ON || (p.current.Type == token.IDENT && p.current.Literal == "on") {
		p.nextToken()
	}

	var student ast.Expr
	if p.current.Type == token.IDENT {
		student = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	// optional 'with'
	var dsX, dsY ast.Expr
	if p.current.Type == token.WITH {
		p.nextToken()
		dsX = p.parseExpression(CALL_PREC)
		if p.current.Type == token.SEP {
			p.nextToken()
			dsY = p.parseExpression(CALL_PREC)
		}
	}

	var rate, epochs ast.Expr
	if p.current.Type == token.BLOCK_OPEN {
		p.nextToken() // consume '⟦'
		for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
			if p.current.Type == token.IDENT {
				field := p.current.Literal
				p.nextToken()
				if p.current.Type == token.BIND {
					p.nextToken()
				}
				val := p.parseExpression(LOWEST)
				switch field {
				case "rate":
					rate = val
				case "epochs":
					epochs = val
				}
				if p.current.Type == token.STMT_END {
					p.nextToken()
				}
			} else {
				p.nextToken()
			}
		}
		p.expect(token.BLOCK_CLOSE) // '⟧'
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.DistillStmt{
		Pos_:     pos,
		License:  lic,
		Student:  student,
		DatasetX: dsX,
		DatasetY: dsY,
		Rate:     rate,
		Epochs:   epochs,
	}
}

func (p *Parser) parseReasoningChainDecl() *ast.ReasoningChainDecl {
	pos := p.curPos()
	p.nextToken() // consume 'reasoning_chain'

	name := ""
	if p.current.Type == token.IDENT || p.current.Type == token.LOGIC {
		name = p.current.Literal
		p.nextToken()
	}

	var boundary ast.Expr
	if p.current.Type == token.WITH {
		p.nextToken()
		if p.current.Type == token.IDENT || p.current.Type == token.LOGIC {
			boundary = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
			p.nextToken()
		}
	}

	var capExpr ast.Expr
	if p.current.Type == token.BLOCK_OPEN {
		p.nextToken() // consume '⟦'
		for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
			if p.current.Type == token.IDENT && p.current.Literal == "capacity" {
				p.nextToken()
				if p.current.Type == token.BIND {
					p.nextToken()
				}
				capExpr = p.parseExpression(LOWEST)
				if p.current.Type == token.STMT_END {
					p.nextToken()
				}
			} else {
				p.nextToken()
			}
		}
		p.expect(token.BLOCK_CLOSE) // '⟧'
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.ReasoningChainDecl{
		Pos_:     pos,
		Name:     name,
		Boundary: boundary,
		Capacity: capExpr,
	}
}

func (p *Parser) parseReasonStepStmt() *ast.ReasonStepStmt {
	pos := p.curPos()
	p.nextToken() // consume 'reason'

	var chain ast.Expr
	// If the next token is an IDENT and NOT a reason op keyword or "from", treat as chain identifier
	if p.current.Type == token.IDENT &&
		p.current.Literal != "deduce" && p.current.Literal != "induce" &&
		p.current.Literal != "abduce" && p.current.Literal != "verify" &&
		p.current.Literal != "from" {
		chain = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	op := token.DEDUCE
	switch p.current.Type {
	case token.DEDUCE, token.INDUCE, token.ABDUCE, token.VERIFY:
		op = p.current.Type
		p.nextToken()
	case token.IDENT:
		switch p.current.Literal {
		case "deduce":
			op = token.DEDUCE
			p.nextToken()
		case "induce":
			op = token.INDUCE
			p.nextToken()
		case "abduce":
			op = token.ABDUCE
			p.nextToken()
		case "verify":
			op = token.VERIFY
			p.nextToken()
		}
	}

	// optional 'from'
	if p.current.Type == token.IDENT && p.current.Literal == "from" {
		p.nextToken()
	}

	inputExpr := p.parseExpression(CALL_PREC)

	// optional 'to'
	if p.current.Type == token.IDENT && p.current.Literal == "to" {
		p.nextToken()
	}

	outputExpr := p.parseExpression(CALL_PREC)

	// optional 'with'
	if p.current.Type == token.WITH {
		p.nextToken()
	}

	confExpr := p.parseExpression(LOWEST)

	p.expect(token.STMT_END)

	return &ast.ReasonStepStmt{
		Pos_:       pos,
		Chain:      chain,
		Operation:  op,
		InputExpr:  inputExpr,
		OutputExpr: outputExpr,
		Confidence: confExpr,
	}
}

func (p *Parser) parseVerifyChainStmt() *ast.VerifyChainStmt {
	pos := p.curPos()
	p.nextToken() // consume 'verify_chain'
	chain := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.VerifyChainStmt{Pos_: pos, Chain: chain}
}

func (p *Parser) parseAuditReasoningStmt() *ast.AuditReasoningStmt {
	pos := p.curPos()
	p.nextToken() // consume 'audit_reasoning'
	chain := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)
	return &ast.AuditReasoningStmt{Pos_: pos, Chain: chain}
}

func (p *Parser) parseCorruptReasoningStepStmt() *ast.CorruptReasoningStepStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_reasoning_step'

	chainName := ""
	if p.current.Type == token.IDENT || p.current.Type == token.LOGIC {
		chainName = p.current.Literal
		p.nextToken()
	}

	closeTok := token.INDEX_CLOSE
	if p.current.Type == token.GROUP_OPEN {
		closeTok = token.GROUP_CLOSE
		p.nextToken()
	} else {
		p.expect(token.INDEX_OPEN)
	}
	stepIdx := p.parseExpression(LOWEST)
	p.expect(closeTok)

	if p.current.Type == token.IDENT && p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}
	atVal := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.CorruptReasoningStepStmt{
		Pos_:      pos,
		Chain:     &ast.Ident{Pos_: pos, Name: chainName},
		StepIndex: stepIdx,
		AtVal:     atVal,
	}
}

func (p *Parser) parseModalityDecl() *ast.ModalityDecl {
	pos := p.curPos()
	p.nextToken() // consume 'modality'

	name := ""
	if p.current.Type == token.IDENT || p.current.Type == token.LOGIC {
		name = p.current.Literal
		p.nextToken()
	}

	if p.current.Type == token.IDENT && p.current.Literal == "as" {
		p.nextToken()
	}

	kind := p.current.Type
	switch p.current.Type {
	case token.TENSOR_INPUT, token.TEXT_INPUT, token.CONCEPT_INPUT:
		kind = p.current.Type
		p.nextToken()
	case token.IDENT:
		switch p.current.Literal {
		case "tensor_input":
			kind = token.TENSOR_INPUT
			p.nextToken()
		case "text_input":
			kind = token.TEXT_INPUT
			p.nextToken()
		case "concept_input":
			kind = token.CONCEPT_INPUT
			p.nextToken()
		default:
			p.errorf(p.curPos(), "unknown modality kind %s", p.current.Literal)
		}
	default:
		p.errorf(p.curPos(), "expected modality kind, got %s", p.current.Type)
	}

	if p.current.Type == token.IDENT && p.current.Literal == "from" {
		p.nextToken()
	}

	source := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.ModalityDecl{
		Pos_:   pos,
		Name:   name,
		Kind:   kind,
		Source: source,
	}
}

func (p *Parser) parseCrossModalReasonStmt() *ast.CrossModalReasonStmt {
	pos := p.curPos()
	p.nextToken() // consume 'cross_modal_reason'

	chain := p.parseExpression(CALL_PREC)

	op := token.DEDUCE
	switch p.current.Type {
	case token.DEDUCE, token.INDUCE, token.ABDUCE, token.VERIFY:
		op = p.current.Type
		p.nextToken()
	case token.IDENT:
		switch p.current.Literal {
		case "deduce":
			op = token.DEDUCE
			p.nextToken()
		case "induce":
			op = token.INDUCE
			p.nextToken()
		case "abduce":
			op = token.ABDUCE
			p.nextToken()
		case "verify":
			op = token.VERIFY
			p.nextToken()
		}
	}

	if p.current.Type == token.IDENT && p.current.Literal == "from" {
		p.nextToken()
	}

	var modalities []ast.Expr
	closeTok := token.INDEX_CLOSE
	switch p.current.Type {
	case token.INDEX_OPEN:
		closeTok = token.INDEX_CLOSE
		p.nextToken()
	case token.GROUP_OPEN:
		closeTok = token.GROUP_CLOSE
		p.nextToken()
	}

	for p.current.Type != closeTok && p.current.Type != token.EOF {
		mod := p.parseExpression(CALL_PREC)
		if mod != nil {
			modalities = append(modalities, mod)
		}
		if p.current.Type == token.SEP {
			p.nextToken()
		} else {
			break
		}
	}
	p.expect(closeTok)

	if p.current.Type == token.IDENT && p.current.Literal == "to" {
		p.nextToken()
	}

	outputExpr := p.parseExpression(CALL_PREC)

	if p.current.Type == token.WITH {
		p.nextToken()
	} else if p.current.Type == token.IDENT && p.current.Literal == "with" {
		p.nextToken()
	}

	confExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.CrossModalReasonStmt{
		Pos_:       pos,
		Chain:      chain,
		Operation:  op,
		Modalities: modalities,
		OutputExpr: outputExpr,
		Confidence: confExpr,
	}
}

func (p *Parser) parseQECMemoryDecl() *ast.QECMemoryDecl {
	pos := p.curPos()
	p.nextToken() // consume 'qec_memory'

	name := ""
	if p.current.Type == token.IDENT || p.current.Type == token.LOGIC {
		name = p.current.Literal
		p.nextToken()
	}

	if p.current.Type == token.IDENT && p.current.Literal == "as" {
		p.nextToken()
	}

	codeType := token.SURFACE_CODE
	switch p.current.Type {
	case token.SURFACE_CODE, token.QLDPC_CODE:
		codeType = p.current.Type
		p.nextToken()
	case token.IDENT:
		switch p.current.Literal {
		case "surface_code":
			codeType = token.SURFACE_CODE
			p.nextToken()
		case "qldpc_code":
			codeType = token.QLDPC_CODE
			p.nextToken()
		}
	}

	if p.current.Type == token.IDENT && p.current.Literal == "distance" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}
	distExpr := p.parseExpression(CALL_PREC)

	var boundaryExpr ast.Expr
	if p.current.Type == token.WITH || (p.current.Type == token.IDENT && p.current.Literal == "with") {
		p.nextToken()
		boundaryExpr = p.parseExpression(CALL_PREC)
	}

	if p.current.Type == token.BLOCK_OPEN {
		p.nextToken()
		if p.current.Type == token.BLOCK_CLOSE {
			p.nextToken()
		}
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.QECMemoryDecl{
		Pos_:     pos,
		Name:     name,
		CodeType: codeType,
		Distance: distExpr,
		Boundary: boundaryExpr,
	}
}

func (p *Parser) parseSyndromeStmt() *ast.SyndromeStmt {
	pos := p.curPos()
	p.nextToken() // consume 'syndrome'

	qecMem := p.parseExpression(CALL_PREC)

	into := ""
	if p.current.Type == token.IDENT && p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		} else {
			p.expect(token.IDENT)
		}
	}

	p.expect(token.STMT_END)

	return &ast.SyndromeStmt{
		Pos_:   pos,
		QECMem: qecMem,
		Into:   into,
	}
}

func (p *Parser) parseVerifySyndromeStmt() *ast.VerifySyndromeStmt {
	pos := p.curPos()
	p.nextToken() // consume 'verify_syndrome'

	qecMem := p.parseExpression(CALL_PREC)

	if p.current.Type == token.WITH || (p.current.Type == token.IDENT && p.current.Literal == "with") {
		p.nextToken()
	}

	synExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.VerifySyndromeStmt{
		Pos_:     pos,
		QECMem:   qecMem,
		Syndrome: synExpr,
	}
}

func (p *Parser) parseCorrectStmt() *ast.CorrectStmt {
	pos := p.curPos()
	p.nextToken() // consume 'correct'

	qecMem := p.parseExpression(CALL_PREC)

	if p.current.Type == token.WITH || (p.current.Type == token.IDENT && p.current.Literal == "with") {
		p.nextToken()
	}

	synExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.CorrectStmt{
		Pos_:     pos,
		QECMem:   qecMem,
		Syndrome: synExpr,
	}
}

func (p *Parser) parseCorruptSyndromeStmt() *ast.CorruptSyndromeStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_syndrome'

	qecMemName := ""
	if p.current.Type == token.IDENT {
		qecMemName = p.current.Literal
		p.nextToken()
	}

	p.expect(token.INDEX_OPEN)
	synExpr := p.parseExpression(LOWEST)
	p.expect(token.INDEX_CLOSE)

	if p.current.Type == token.IDENT && p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}
	atVal := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.CorruptSyndromeStmt{
		Pos_:     pos,
		QECMem:   &ast.Ident{Pos_: pos, Name: qecMemName},
		Syndrome: synExpr,
		AtVal:    atVal,
	}
}

func (p *Parser) parseConsolidateStmt() *ast.ConsolidateStmt {
	pos := p.curPos()
	p.nextToken() // consume 'consolidate'

	if p.current.Type == token.EPISODIC || (p.current.Type == token.IDENT && p.current.Literal == "episodic") {
		p.nextToken()
	}

	epMem := p.parseExpression(CALL_PREC)

	if p.current.Type == token.IDENT && p.current.Literal == "into" {
		p.nextToken()
	}

	if p.current.Type == token.SEMANTIC || (p.current.Type == token.IDENT && p.current.Literal == "semantic") {
		p.nextToken()
	}

	semMem := p.parseExpression(CALL_PREC)

	if p.current.Type == token.WITH || (p.current.Type == token.IDENT && p.current.Literal == "with") {
		p.nextToken()
	}

	if p.current.Type == token.IDENT && p.current.Literal == "threshold" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}

	thresh := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.ConsolidateStmt{
		Pos_:          pos,
		EpisodicMem:   epMem,
		SemanticMem:   semMem,
		MinConfidence: thresh,
	}
}

func (p *Parser) parseRecallTokenDecl() *ast.RecallTokenDecl {
	pos := p.curPos()
	p.nextToken() // consume 'recall_token'

	name := ""
	if p.current.Type == token.IDENT {
		name = p.current.Literal
		p.nextToken()
	}

	if p.current.Type == token.IDENT && p.current.Literal == "for" {
		p.nextToken()
	}

	mem := p.parseExpression(CALL_PREC)

	if (p.current.Type == token.ON || (p.current.Type == token.IDENT && p.current.Literal == "on")) ||
		(p.current.Type == token.IDENT && p.current.Literal == "to") {
		p.nextToken()
	}

	recipient := p.parseExpression(CALL_PREC)

	var keys []ast.Expr
	var expiry ast.Expr

	if p.current.Type == token.BLOCK_OPEN {
		p.nextToken() // consume '⟦'
		for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
			if p.current.Type == token.IDENT {
				field := p.current.Literal
				p.nextToken()
				if p.current.Type == token.BIND {
					p.nextToken()
				}
				switch field {
				case "keys":
					if p.current.Type == token.INDEX_OPEN {
						p.nextToken()
						for p.current.Type != token.INDEX_CLOSE && p.current.Type != token.EOF {
							elem := p.parseExpression(LOWEST)
							if elem != nil {
								keys = append(keys, elem)
							}
							if p.current.Type == token.SEP {
								p.nextToken()
							} else {
								break
							}
						}
						p.expect(token.INDEX_CLOSE)
					}
				case "expiry":
					expiry = p.parseExpression(LOWEST)
				default:
					p.parseExpression(LOWEST)
				}
				if p.current.Type == token.STMT_END {
					p.nextToken()
				}
			} else {
				p.nextToken()
			}
		}
		p.expect(token.BLOCK_CLOSE) // '⟧'
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.RecallTokenDecl{
		Pos_:      pos,
		Name:      name,
		Memory:    mem,
		Recipient: recipient,
		Keys:      keys,
		Expiry:    expiry,
	}
}

func (p *Parser) parseGrantStmt() *ast.GrantStmt {
	pos := p.curPos()
	p.nextToken() // consume 'grant'

	tok := p.parseExpression(CALL_PREC)
	var recipient ast.Expr
	if p.current.Type == token.IDENT && p.current.Literal == "to" {
		p.nextToken()
		recipient = p.parseExpression(CALL_PREC)
	}
	p.expect(token.STMT_END)

	return &ast.GrantStmt{
		Pos_:      pos,
		Token:     tok,
		Recipient: recipient,
	}
}

func (p *Parser) parseRevokeStmt() *ast.RevokeStmt {
	pos := p.curPos()
	p.nextToken() // consume 'revoke'

	tok := p.parseExpression(CALL_PREC)
	p.expect(token.STMT_END)

	return &ast.RevokeStmt{
		Pos_:  pos,
		Token: tok,
	}
}

func (p *Parser) parseToolRegisterDecl() *ast.ToolRegisterDecl {
	pos := p.curPos()
	p.nextToken() // consume 'tool_register'

	name := ""
	if p.current.Type == token.IDENT {
		name = p.current.Literal
		p.nextToken()
	}

	if p.current.Type == token.WITH || (p.current.Type == token.IDENT && p.current.Literal == "with") {
		p.nextToken()
	}

	boundary := p.parseExpression(CALL_PREC)

	var preCond ast.Expr
	var postCond ast.Expr

	for (p.current.Type == token.PRE_CONDITION || (p.current.Type == token.IDENT && p.current.Literal == "pre")) ||
		(p.current.Type == token.POST_CONDITION || (p.current.Type == token.IDENT && p.current.Literal == "post")) {
		isPre := p.current.Type == token.PRE_CONDITION || (p.current.Type == token.IDENT && p.current.Literal == "pre")
		p.nextToken()
		if p.current.Type == token.BLOCK_OPEN {
			p.nextToken()
			condExpr := p.parseExpression(LOWEST)
			for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
				p.nextToken()
			}
			p.expect(token.BLOCK_CLOSE)
			if isPre {
				preCond = condExpr
			} else {
				postCond = condExpr
			}
		} else {
			condExpr := p.parseExpression(CALL_PREC)
			if isPre {
				preCond = condExpr
			} else {
				postCond = condExpr
			}
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.ToolRegisterDecl{
		Pos_:          pos,
		Name:          name,
		Boundary:      boundary,
		PreCondition:  preCond,
		PostCondition: postCond,
	}
}

func (p *Parser) parseToolCallStmt(target ast.Expr) *ast.ToolCallStmt {
	pos := p.curPos()
	p.nextToken() // consume 'tool_call'

	name := p.expect(token.IDENT).Literal
	p.expect(token.GROUP_OPEN)
	var args []ast.Expr
	for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
		arg := p.parseExpression(LOWEST)
		if arg != nil {
			args = append(args, arg)
		}
		if p.current.Type == token.SEP {
			p.nextToken()
		} else {
			break
		}
	}
	p.expect(token.GROUP_CLOSE)

	var audit ast.Expr
	if p.current.Type == token.TOOL_AUDIT || (p.current.Type == token.IDENT && p.current.Literal == "audit") {
		p.nextToken()
		audit = p.parseExpression(LOWEST)
	}
	p.expect(token.STMT_END)

	return &ast.ToolCallStmt{
		Pos_:     pos,
		Target:   target,
		ToolName: name,
		Args:     args,
		Audit:    audit,
	}
}

func (p *Parser) parseCertifiedReasonStmt() *ast.CertifiedReasonStmt {
	pos := p.curPos()
	p.nextToken() // consume 'certified_reason'

	var chain ast.Expr
	if (p.current.Type == token.IDENT || p.current.Type == token.LOGIC) &&
		p.current.Literal != "deduce" && p.current.Literal != "induce" &&
		p.current.Literal != "abduce" && p.current.Literal != "verify" &&
		p.current.Literal != "from" {
		chain = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	op := token.DEDUCE
	switch p.current.Type {
	case token.DEDUCE, token.INDUCE, token.ABDUCE, token.VERIFY:
		op = p.current.Type
		p.nextToken()
	case token.IDENT:
		switch p.current.Literal {
		case "deduce":
			op = token.DEDUCE
			p.nextToken()
		case "induce":
			op = token.INDUCE
			p.nextToken()
		case "abduce":
			op = token.ABDUCE
			p.nextToken()
		case "verify":
			op = token.VERIFY
			p.nextToken()
		}
	}

	if p.current.Type == token.IDENT && p.current.Literal == "from" {
		p.nextToken()
	}

	inputExpr := p.parseExpression(CALL_PREC)

	if p.current.Type == token.IDENT && p.current.Literal == "to" {
		p.nextToken()
	}

	outputExpr := p.parseExpression(CALL_PREC)

	if p.current.Type == token.WITNESS || (p.current.Type == token.IDENT && p.current.Literal == "witness") {
		p.nextToken()
	}

	closeTok := token.Type("")
	if p.current.Type == token.GROUP_OPEN {
		closeTok = token.GROUP_CLOSE
		p.nextToken()
	} else if p.current.Type == token.INDEX_OPEN {
		closeTok = token.INDEX_CLOSE
		p.nextToken()
	}

	witnessExpr := p.parseExpression(CALL_PREC)

	if closeTok != "" && (p.current.Type == token.GROUP_CLOSE || p.current.Type == token.INDEX_CLOSE) {
		p.nextToken()
	}

	if p.current.Type == token.IDENT && p.current.Literal == "confidence" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	} else if p.current.Type == token.WITH {
		p.nextToken()
	}

	confExpr := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      chain,
		Kind:       op,
		InputExpr:  inputExpr,
		OutputExpr: outputExpr,
		Witness:    witnessExpr,
		Confidence: confExpr,
	}
}

func (p *Parser) parseAbstainThresholdDecl() *ast.AbstainThresholdDecl {
	pos := p.curPos()
	p.nextToken() // consume 'abstain_threshold'

	var chain ast.Expr
	if (p.current.Type == token.IDENT || p.current.Type == token.LOGIC) && p.current.Literal != "mode" {
		chain = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	}

	if p.current.Type == token.BIND {
		p.nextToken()
	}

	thresholdExpr := p.parseExpression(CALL_PREC)

	mode := token.REFUSE
	if p.current.Type == token.IDENT && p.current.Literal == "mode" {
		p.nextToken()
	}

	switch p.current.Type {
	case token.RETRY:
		mode = token.RETRY
		p.nextToken()
	case token.ESCALATE:
		mode = token.ESCALATE
		p.nextToken()
	case token.REFUSE:
		mode = token.REFUSE
		p.nextToken()
	case token.IDENT:
		switch p.current.Literal {
		case "retry":
			mode = token.RETRY
			p.nextToken()
		case "escalate":
			mode = token.ESCALATE
			p.nextToken()
		case "refuse":
			mode = token.REFUSE
			p.nextToken()
		}
	}

	var maxRetries ast.Expr
	if p.current.Type == token.IDENT && p.current.Literal == "max_retries" {
		p.nextToken()
		if p.current.Type == token.BIND {
			p.nextToken()
		}
		maxRetries = p.parseExpression(CALL_PREC)
	}

	p.expect(token.STMT_END)

	return &ast.AbstainThresholdDecl{
		Pos_:       pos,
		Chain:      chain,
		Threshold:  thresholdExpr,
		Mode:       mode,
		MaxRetries: maxRetries,
	}
}

func (p *Parser) parseAbstainCheckStmt() *ast.AbstainCheckStmt {
	pos := p.curPos()
	p.nextToken() // consume 'abstain_check'

	var chain ast.Expr
	if p.current.Type == token.IDENT || p.current.Type == token.LOGIC {
		chain = &ast.Ident{Pos_: p.curPos(), Name: p.current.Literal}
		p.nextToken()
	} else {
		chain = p.parseExpression(CALL_PREC)
	}

	intoVar := ""
	if p.current.Type == token.IDENT && p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT || p.current.Type == token.LOGIC {
			intoVar = p.current.Literal
			p.nextToken()
		}
	}
	p.expect(token.STMT_END)

	return &ast.AbstainCheckStmt{
		Pos_:  pos,
		Chain: chain,
		Into:  intoVar,
	}
}

func (p *Parser) parseSandboxCreateStmt() *ast.SandboxCreateStmt {
	pos := p.curPos()
	p.nextToken() // consume 'sandbox_create'

	name := p.expect(token.IDENT).Literal

	if p.current.Type == token.WITH || (p.current.Type == token.IDENT && p.current.Literal == "with") {
		p.nextToken()
	}

	boundary := p.expect(token.IDENT).Literal

	var capExpr ast.Expr
	if p.current.Type == token.BLOCK_OPEN {
		p.nextToken()
		for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
			if p.current.Literal == "capacity" {
				p.nextToken()
				if p.current.Type == token.BIND || p.current.Type == token.ASSIGN {
					p.nextToken()
				}
				capExpr = p.parseExpression(CALL_PREC)
				if p.current.Type == token.STMT_END {
					p.nextToken()
				}
			} else {
				p.nextToken()
			}
		}
		p.expect(token.BLOCK_CLOSE)
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.SandboxCreateStmt{
		Pos_:     pos,
		Name:     name,
		Boundary: boundary,
		Capacity: capExpr,
	}
}

func (p *Parser) parseSandboxRunStmt() *ast.SandboxRunStmt {
	pos := p.curPos()
	p.nextToken() // consume 'sandbox_run'

	sbName := p.expect(token.IDENT).Literal

	p.expect(token.GROUP_OPEN)
	toolName := p.expect(token.IDENT).Literal

	p.expect(token.GROUP_OPEN)
	var args []ast.Expr
	for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
		arg := p.parseExpression(LOWEST)
		if arg != nil {
			args = append(args, arg)
		}
		if p.current.Type == token.SEP {
			p.nextToken()
		} else {
			break
		}
	}
	p.expect(token.GROUP_CLOSE) // close inner call
	p.expect(token.GROUP_CLOSE) // close outer sandbox_run

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.SandboxRunStmt{
		Pos_:     pos,
		Sandbox:  sbName,
		ToolName: toolName,
		Args:     args,
	}
}

func (p *Parser) parseSandboxVerifyStmt() *ast.SandboxVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'sandbox_verify'

	sbName := p.expect(token.IDENT).Literal

	intoVar := ""
	if p.current.Literal == "into" {
		p.nextToken()
		intoVar = p.expect(token.IDENT).Literal
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.SandboxVerifyStmt{
		Pos_:    pos,
		Sandbox: sbName,
		Into:    intoVar,
	}
}

func (p *Parser) parseGPUAnnotatedStmt() ast.Stmt {
	p.nextToken() // consume '@gpu'
	switch p.current.Type {
	case token.GPU_MATMUL:
		return p.parseGPUMatmulStmt(true)
	case token.GPU_SOFTMAX:
		return p.parseGPUSoftmaxStmt(true)
	case token.GPU_LAYERNORM:
		return p.parseGPULayernormStmt(true)
	default:
		stmt := p.parseStatement()
		return stmt
	}
}

func (p *Parser) parseGPUMatmulStmt(isAnnotated bool) *ast.GPUMatmulStmt {
	pos := p.curPos()
	p.nextToken() // consume 'gpu_matmul'

	hasGroup := false
	if p.current.Type == token.GROUP_OPEN {
		hasGroup = true
		p.nextToken()
	}

	left := p.parseExpression(LOWEST)
	p.expect(token.SEP)
	right := p.parseExpression(LOWEST)

	if hasGroup {
		p.expect(token.GROUP_CLOSE)
	}

	if p.current.Type == token.IDENT && p.current.Literal == "into" {
		p.nextToken()
	}

	target := p.parseExpression(CALL_PREC)

	var audit ast.Expr
	if (p.current.Type == token.IDENT && p.current.Literal == "audit") || p.current.Type == token.TOOL_AUDIT {
		p.nextToken()
		audit = p.parseExpression(CALL_PREC)
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.GPUMatmulStmt{
		Pos_:        pos,
		IsAnnotated: isAnnotated,
		Left:        left,
		Right:       right,
		Target:      target,
		Audit:       audit,
	}
}

func (p *Parser) parseGPUSoftmaxStmt(isAnnotated bool) *ast.GPUSoftmaxStmt {
	pos := p.curPos()
	p.nextToken() // consume 'gpu_softmax'

	hasGroup := false
	if p.current.Type == token.GROUP_OPEN {
		hasGroup = true
		p.nextToken()
	}

	src := p.parseExpression(LOWEST)

	if hasGroup {
		p.expect(token.GROUP_CLOSE)
	}

	if p.current.Type == token.IDENT && p.current.Literal == "into" {
		p.nextToken()
	}

	target := p.parseExpression(CALL_PREC)

	var audit ast.Expr
	if (p.current.Type == token.IDENT && p.current.Literal == "audit") || p.current.Type == token.TOOL_AUDIT {
		p.nextToken()
		audit = p.parseExpression(CALL_PREC)
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.GPUSoftmaxStmt{
		Pos_:        pos,
		IsAnnotated: isAnnotated,
		Source:      src,
		Target:      target,
		Audit:       audit,
	}
}

func (p *Parser) parseGPULayernormStmt(isAnnotated bool) *ast.GPULayernormStmt {
	pos := p.curPos()
	p.nextToken() // consume 'gpu_layernorm'

	hasGroup := false
	if p.current.Type == token.GROUP_OPEN {
		hasGroup = true
		p.nextToken()
	}

	src := p.parseExpression(LOWEST)

	if hasGroup {
		p.expect(token.GROUP_CLOSE)
	}

	if p.current.Type == token.IDENT && p.current.Literal == "into" {
		p.nextToken()
	}

	target := p.parseExpression(CALL_PREC)

	var audit ast.Expr
	if (p.current.Type == token.IDENT && p.current.Literal == "audit") || p.current.Type == token.TOOL_AUDIT {
		p.nextToken()
		audit = p.parseExpression(CALL_PREC)
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.GPULayernormStmt{
		Pos_:        pos,
		IsAnnotated: isAnnotated,
		Source:      src,
		Target:      target,
		Audit:       audit,
	}
}

func (p *Parser) parseAGKEAttestStmt() *ast.AGKEAttestStmt {
	pos := p.curPos()
	p.nextToken() // consume 'agke_attest'

	chain := p.parseExpression(CALL_PREC)
	var kernel, inTensor, outTensor ast.Expr

	if p.current.Type == token.GROUP_OPEN {
		p.nextToken()
		kernel = p.parseExpression(LOWEST)
		p.expect(token.SEP)
		inTensor = p.parseExpression(LOWEST)
		p.expect(token.SEP)
		outTensor = p.parseExpression(LOWEST)
		p.expect(token.GROUP_CLOSE)
	} else {
		for p.current.Type != token.STMT_END && p.current.Type != token.EOF {
			lit := p.current.Literal
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			val := p.parseExpression(LOWEST)
			switch lit {
			case "kernel":
				kernel = val
			case "input":
				inTensor = val
			case "output":
				outTensor = val
			}
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.AGKEAttestStmt{
		Pos_:         pos,
		Chain:        chain,
		KernelName:   kernel,
		InputTensor:  inTensor,
		OutputTensor: outTensor,
	}
}

func (p *Parser) parseAGKEVerifyStmt() *ast.AGKEVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'agke_verify'

	chain := p.parseExpression(CALL_PREC)
	var target ast.Expr
	if p.current.Type == token.IDENT && p.current.Literal == "into" {
		p.nextToken()
		target = p.parseExpression(CALL_PREC)
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.AGKEVerifyStmt{
		Pos_:   pos,
		Chain:  chain,
		Target: target,
	}
}

func (p *Parser) parseCorruptGPUDigestStmt() *ast.CorruptGPUDigestStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_gpu_digest'

	chain := p.parseExpression(CALL_PREC)
	if p.current.Type == token.IDENT && p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}
	offset := p.parseExpression(LOWEST)

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.CorruptGPUDigestStmt{
		Pos_:   pos,
		Chain:  chain,
		Offset: offset,
	}
}

func (p *Parser) parseMultiDeviceAnnotatedStmt() ast.Stmt {
	p.nextToken() // consume '@multi_device'
	if p.current.Type == token.PIPELINE_CLAIM {
		return p.parsePipelineClaimStmt(true)
	}
	return p.parseStatement()
}

func (p *Parser) parsePipelineClaimStmt(annotated bool) *ast.PipelineClaimStmt {
	pos := p.curPos()
	p.nextToken() // consume 'pipeline_claim'

	stages := 1
	if p.current.Type == token.INT {
		val, _ := strconv.Atoi(p.current.Literal)
		stages = val
		p.nextToken()
	}
	if p.current.Literal == "stages" {
		p.nextToken()
	}
	microBatch := 1
	if p.current.Literal == "micro_batch" {
		p.nextToken()
		if p.current.Type == token.INT {
			val, _ := strconv.Atoi(p.current.Literal)
			microBatch = val
			p.nextToken()
		}
	}
	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}
	return &ast.PipelineClaimStmt{
		Pos_:        pos,
		IsAnnotated: annotated,
		Stages:      stages,
		MicroBatch:  microBatch,
		Into:        into,
	}
}

func (p *Parser) parsePipelineRunStmt() *ast.PipelineRunStmt {
	pos := p.curPos()
	p.nextToken() // consume 'pipeline_run'
	plan := ""
	if p.current.Type == token.IDENT {
		plan = p.current.Literal
		p.nextToken()
	}
	if p.current.Type == token.ON || p.current.Literal == "on" {
		p.nextToken()
	}
	input := p.parseExpression(CALL_PREC)
	if p.current.Literal == "into" {
		p.nextToken()
	}
	output := p.parseExpression(CALL_PREC)
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}
	return &ast.PipelineRunStmt{
		Pos_:   pos,
		Plan:   plan,
		Input:  input,
		Output: output,
	}
}

func (p *Parser) parseCDAGARecordStmt() *ast.CDAGARecordStmt {
	pos := p.curPos()
	p.nextToken() // consume 'cdaga_record'

	if p.current.Literal == "from" {
		p.nextToken()
	}
	var grads []ast.Expr
	if p.current.Type == token.INDEX_OPEN {
		p.nextToken() // consume ⟬
		for p.current.Type != token.INDEX_CLOSE && p.current.Type != token.EOF {
			grad := p.parseExpression(LOWEST)
			if grad != nil {
				grads = append(grads, grad)
			}
			if p.current.Type == token.SEP {
				p.nextToken()
			} else {
				break
			}
		}
		if p.current.Type == token.INDEX_CLOSE {
			p.nextToken()
		}
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}
	return &ast.CDAGARecordStmt{
		Pos_:     pos,
		GradList: grads,
		Into:     into,
	}
}

func (p *Parser) parseCDAGAAttestStmt() *ast.CDAGAAttestStmt {
	pos := p.curPos()
	p.nextToken() // consume 'cdaga_attest'

	record := ""
	if p.current.Type == token.IDENT {
		record = p.current.Literal
		p.nextToken()
	}
	if p.current.Literal == "to" {
		p.nextToken()
	}
	chain := ""
	if p.current.Type == token.IDENT {
		chain = p.current.Literal
		p.nextToken()
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}
	return &ast.CDAGAAttestStmt{
		Pos_:   pos,
		Record: record,
		Chain:  chain,
	}
}

func (p *Parser) parseCDAGAVerifyStmt() *ast.CDAGAVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'cdaga_verify'

	record := ""
	if p.current.Type == token.IDENT {
		record = p.current.Literal
		p.nextToken()
	}
	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}
	return &ast.CDAGAVerifyStmt{
		Pos_:   pos,
		Record: record,
		Into:   into,
	}
}

func (p *Parser) parseCorruptCDAGAStmt() *ast.CorruptCDAGAStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_cdaga'

	record := ""
	if p.current.Type == token.IDENT {
		record = p.current.Literal
		p.nextToken()
	}
	if p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}
	offset := p.parseExpression(LOWEST)
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}
	return &ast.CorruptCDAGAStmt{
		Pos_:   pos,
		Record: record,
		Offset: offset,
	}
}

func (p *Parser) parseDistributedAnnotatedStmt() ast.Stmt {
	p.nextToken() // consume '@distributed'
	return p.parseStatement()
}

func (p *Parser) parseAllReduceStmt() *ast.AllReduceStmt {
	pos := p.curPos()
	p.nextToken() // consume 'all_reduce'

	var grads []ast.Expr
	var target ast.Expr
	var into string

	if p.current.Literal == "from" {
		p.nextToken()
		if p.current.Type == token.INDEX_OPEN {
			p.nextToken()
			for p.current.Type != token.INDEX_CLOSE && p.current.Type != token.EOF {
				g := p.parseExpression(LOWEST)
				if g != nil {
					grads = append(grads, g)
				}
				if p.current.Type == token.SEP {
					p.nextToken()
				} else {
					break
				}
			}
			if p.current.Type == token.INDEX_CLOSE {
				p.nextToken()
			}
		}
	} else if p.current.Type != token.STMT_END && p.current.Literal != "into" {
		target = p.parseExpression(LOWEST)
	}

	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.AllReduceStmt{
		Pos_:     pos,
		GradList: grads,
		Target:   target,
		Into:     into,
	}
}

func (p *Parser) parsePserverInitStmt() *ast.PserverInitStmt {
	pos := p.curPos()
	p.nextToken() // consume 'pserver_init'

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	var workers ast.Expr
	if p.current.Literal == "with" {
		p.nextToken()
		workers = p.parseExpression(LOWEST)
		if p.current.Literal == "workers" {
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.PserverInitStmt{
		Pos_:    pos,
		Into:    into,
		Workers: workers,
	}
}

func (p *Parser) parsePserverAggregateStmt() *ast.PserverAggregateStmt {
	pos := p.curPos()
	p.nextToken() // consume 'pserver_aggregate'

	pserver := ""
	if p.current.Type == token.IDENT {
		pserver = p.current.Literal
		p.nextToken()
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.PserverAggregateStmt{
		Pos_:    pos,
		Pserver: pserver,
	}
}

func (p *Parser) parseBaratSubmitStmt() *ast.BaratSubmitStmt {
	pos := p.curPos()
	p.nextToken() // consume 'barat_submit'

	var worker ast.Expr
	grp := ""
	var grad ast.Expr

	if p.current.Literal == "from" {
		p.nextToken()
		grad = p.parseExpression(LOWEST)
		if p.current.Literal == "into" {
			p.nextToken()
			if p.current.Type == token.IDENT {
				grp = p.current.Literal
				p.nextToken()
			}
		}
	} else {
		worker = p.parseExpression(LOWEST)
		if p.current.Literal == "into" {
			p.nextToken()
			if p.current.Type == token.IDENT {
				grp = p.current.Literal
				p.nextToken()
			}
		}
		if p.current.Literal == "with" {
			p.nextToken()
			grad = p.parseExpression(LOWEST)
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.BaratSubmitStmt{
		Pos_:     pos,
		Worker:   worker,
		Group:    grp,
		Gradient: grad,
	}
}

func (p *Parser) parseBaratAggregateStmt() *ast.BaratAggregateStmt {
	pos := p.curPos()
	p.nextToken() // consume 'barat_aggregate'

	grp := ""
	if p.current.Type == token.IDENT {
		grp = p.current.Literal
		p.nextToken()
	}

	intoGrad := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			intoGrad = p.current.Literal
			p.nextToken()
		}
	}

	intoCert := ""
	if p.current.Literal == "cert" || p.current.Literal == "with" {
		p.nextToken()
		if p.current.Literal == "into" {
			p.nextToken()
		}
		if p.current.Type == token.IDENT {
			intoCert = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.BaratAggregateStmt{
		Pos_:     pos,
		Group:    grp,
		IntoGrad: intoGrad,
		IntoCert: intoCert,
	}
}

func (p *Parser) parseBaratVerifyStmt() *ast.BaratVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'barat_verify'

	cert := ""
	if p.current.Type == token.IDENT {
		cert = p.current.Literal
		p.nextToken()
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.BaratVerifyStmt{
		Pos_: pos,
		Cert: cert,
		Into: into,
	}
}

func (p *Parser) parseCorruptBaratStmt() *ast.CorruptBaratStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_barat'

	cert := ""
	if p.current.Type == token.IDENT {
		cert = p.current.Literal
		p.nextToken()
	}

	if p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}

	offset := p.parseExpression(LOWEST)

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.CorruptBaratStmt{
		Pos_:   pos,
		Cert:   cert,
		Offset: offset,
	}
}

func (p *Parser) parseQHALBackendAnnotation() ast.Stmt {
	pos := p.curPos()
	p.nextToken() // consume '@qhal_backend'
	kind := "sim"
	if p.current.Type == token.IDENT {
		kind = p.current.Literal
		p.nextToken()
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}
	return &ast.QHALBackendAnnotation{
		Pos_: pos,
		Kind: kind,
	}
}

func (p *Parser) parseQASMExportStmt() *ast.QASMExportStmt {
	pos := p.curPos()
	p.nextToken() // consume 'qasm_export'

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	var bufLen ast.Expr
	if p.current.Literal == "with" {
		p.nextToken()
		if p.current.Literal == "size" {
			p.nextToken()
		}
		bufLen = p.parseExpression(LOWEST)
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.QASMExportStmt{
		Pos_:   pos,
		Into:   into,
		BufLen: bufLen,
	}
}

func (p *Parser) parseEvolveStmt() *ast.EvolveStmt {
	pos := p.curPos()
	p.nextToken() // consume 'evolve'

	var fromState ast.Expr
	if p.current.Literal == "from" {
		p.nextToken()
		fromState = p.parseExpression(LOWEST)
	}

	var toState ast.Expr
	if p.current.Literal == "to" {
		p.nextToken()
		toState = p.parseExpression(LOWEST)
	}

	var reason ast.Expr
	if p.current.Literal == "with" {
		p.nextToken()
		if p.current.Literal == "reason" {
			p.nextToken()
		}
		reason = p.parseExpression(LOWEST)
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.EvolveStmt{
		Pos_:      pos,
		FromState: fromState,
		ToState:   toState,
		Reason:    reason,
		Into:      into,
	}
}

func (p *Parser) parseEvolveVerifyStmt() *ast.EvolveVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'evolve_verify'

	prop := ""
	if p.current.Type == token.IDENT {
		prop = p.current.Literal
		p.nextToken()
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.EvolveVerifyStmt{
		Pos_:      pos,
		Proposals: prop,
		Into:      into,
	}
}

func (p *Parser) parseEvolveHistoryStmt() *ast.EvolveHistoryStmt {
	pos := p.curPos()
	p.nextToken() // consume 'evolve_history'

	if p.current.Literal == "to" {
		p.nextToken()
	}

	chain := ""
	if p.current.Type == token.IDENT {
		chain = p.current.Literal
		p.nextToken()
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.EvolveHistoryStmt{
		Pos_:  pos,
		Chain: chain,
	}
}

func (p *Parser) parseCorruptEvolveStmt() *ast.CorruptEvolveStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_evolve'

	prop := ""
	if p.current.Type == token.IDENT {
		prop = p.current.Literal
		p.nextToken()
	}

	if p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}

	offset := p.parseExpression(LOWEST)

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.CorruptEvolveStmt{
		Pos_:     pos,
		Proposal: prop,
		Offset:   offset,
	}
}

func (p *Parser) parseTranspileToStmt() *ast.TranspileToStmt {
	pos := p.curPos()
	p.nextToken() // consume 'transpile_to'

	target := ""
	if p.current.Type == token.STRING || p.current.Type == token.IDENT {
		target = p.current.Literal
		p.nextToken()
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	var bufLen ast.Expr
	if p.current.Literal == "with" {
		p.nextToken()
		if p.current.Literal == "size" {
			p.nextToken()
		}
		bufLen = p.parseExpression(LOWEST)
	}

	var opts *ast.TranspileOpts
	if p.current.Literal == "options" || p.current.Type == token.TRANSPILE_OPTS {
		p.nextToken()
		opts = &ast.TranspileOpts{}
		if p.current.Type == token.BLOCK_OPEN {
			p.nextToken()
			for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
				if p.current.Literal == "optimize" {
					opts.Optimize = true
					p.nextToken()
				} else if p.current.Literal == "reduce_depth" {
					opts.ReduceDepth = true
					p.nextToken()
				} else {
					p.nextToken()
				}
				if p.current.Type == token.STMT_END {
					p.nextToken()
				}
			}
			if p.current.Type == token.BLOCK_CLOSE {
				p.nextToken()
			}
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.TranspileToStmt{
		Pos_:    pos,
		Target:  target,
		Into:    into,
		BufLen:  bufLen,
		Options: opts,
	}
}

func (p *Parser) parseEvolvePolicyStmt() *ast.EvolvePolicyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'evolve_policy'

	kind := ""
	if p.current.Type == token.IDENT || p.current.Type == token.STRING {
		kind = p.current.Literal
		p.nextToken()
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	var initialRate ast.Expr
	var totalBudget ast.Expr
	if p.current.Literal == "with" {
		p.nextToken()
	}
	if p.current.Literal == "rate" {
		p.nextToken()
		initialRate = p.parseExpression(LOWEST)
	}
	if p.current.Literal == "budget" {
		p.nextToken()
		totalBudget = p.parseExpression(LOWEST)
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.EvolvePolicyStmt{
		Pos_:        pos,
		Kind:        kind,
		Into:        into,
		InitialRate: initialRate,
		TotalBudget: totalBudget,
	}
}

func (p *Parser) parseEvolveBoundedStmt() *ast.EvolveBoundedStmt {
	pos := p.curPos()
	p.nextToken() // consume 'evolve_bounded'

	policy := ""
	if p.current.Type == token.IDENT {
		policy = p.current.Literal
		p.nextToken()
	}

	var fromState ast.Expr
	if p.current.Literal == "from" {
		p.nextToken()
		fromState = p.parseExpression(LOWEST)
	}

	var toState ast.Expr
	if p.current.Literal == "to" {
		p.nextToken()
		toState = p.parseExpression(LOWEST)
	}

	var reason ast.Expr
	if p.current.Literal == "with" {
		p.nextToken()
		if p.current.Literal == "reason" {
			p.nextToken()
			reason = p.parseExpression(LOWEST)
		}
	} else if p.current.Literal == "reason" {
		p.nextToken()
		reason = p.parseExpression(LOWEST)
	}

	var maxFlips ast.Expr
	if p.current.Literal == "max_flips" {
		p.nextToken()
		maxFlips = p.parseExpression(LOWEST)
	}

	var maxNewSyms ast.Expr
	if p.current.Literal == "max_new_symbols" {
		p.nextToken()
		maxNewSyms = p.parseExpression(LOWEST)
	}

	intoProp := ""
	intoEnv := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			intoProp = p.current.Literal
			p.nextToken()
		}
		if p.current.Type == token.IDENT {
			intoEnv = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.EvolveBoundedStmt{
		Pos_:          pos,
		Policy:        policy,
		FromState:     fromState,
		ToState:       toState,
		Reason:        reason,
		MaxBitFlips:   maxFlips,
		MaxNewSymbols: maxNewSyms,
		IntoProposal:  intoProp,
		IntoEnvelope:  intoEnv,
	}
}

func (p *Parser) parseEEAVerifyStmt() *ast.EEAVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'eea_verify'

	env := ""
	if p.current.Type == token.IDENT {
		env = p.current.Literal
		p.nextToken()
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.EEAVerifyStmt{
		Pos_:     pos,
		Envelope: env,
		Into:     into,
	}
}

func (p *Parser) parseCorruptEEAStmt() *ast.CorruptEEAStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_eea'

	env := ""
	if p.current.Type == token.IDENT {
		env = p.current.Literal
		p.nextToken()
	}

	if p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}

	offset := p.parseExpression(LOWEST)

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.CorruptEEAStmt{
		Pos_:     pos,
		Envelope: env,
		Offset:   offset,
	}
}

func (p *Parser) parseQECDDecodeStmt() *ast.QECDDecodeStmt {
	pos := p.curPos()
	p.nextToken() // consume 'qecd_decode'

	if p.current.Literal == "distance" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken() // consume ↦
	}
	dist := p.parseExpression(LOWEST)

	synIn := ""
	if p.current.Literal == "from" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			synIn = p.current.Literal
			p.nextToken()
		}
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.QECDDecodeStmt{
		Pos_:       pos,
		Distance:   dist,
		SyndromeIn: synIn,
		Into:       into,
	}
}

func (p *Parser) parseQECDVerifyStmt() *ast.QECDVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'qecd_verify'

	res := ""
	if p.current.Type == token.IDENT {
		res = p.current.Literal
		p.nextToken()
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.QECDVerifyStmt{
		Pos_:   pos,
		Result: res,
		Into:   into,
	}
}

func (p *Parser) parseQECDApplyStmt() *ast.QECDApplyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'qecd_apply'

	res := ""
	if p.current.Type == token.IDENT {
		res = p.current.Literal
		p.nextToken()
	}

	stateVec := ""
	if p.current.Type == token.ON || p.current.Literal == "on" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			stateVec = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.QECDApplyStmt{
		Pos_:     pos,
		Result:   res,
		StateVec: stateVec,
	}
}

func (p *Parser) parseAADCertStmt() *ast.AADCertStmt {
	pos := p.curPos()
	p.nextToken() // consume 'aad_cert'

	if p.current.Type == token.WITH || p.current.Literal == "with" {
		p.nextToken()
	}

	var model, trainSet, maxSyn, maxCorr ast.Expr
	into := ""

	for p.current.Type != token.STMT_END && p.current.Type != token.EOF {
		lit := p.current.Literal
		if lit == "model" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			model = p.parseExpression(LOWEST)
		} else if lit == "train" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			trainSet = p.parseExpression(LOWEST)
		} else if lit == "max_syndrome" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			maxSyn = p.parseExpression(LOWEST)
		} else if lit == "max_corr_dist" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			maxCorr = p.parseExpression(LOWEST)
		} else if lit == "into" {
			p.nextToken()
			if p.current.Type == token.IDENT {
				into = p.current.Literal
				p.nextToken()
			}
		} else {
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.AADCertStmt{
		Pos_:        pos,
		Model:       model,
		TrainSet:    trainSet,
		MaxSyndrome: maxSyn,
		MaxCorrDist: maxCorr,
		Into:        into,
	}
}

func (p *Parser) parseAADAttestStmt() *ast.AADAttestStmt {
	pos := p.curPos()
	p.nextToken() // consume 'aad_attest'

	cert := ""
	decode := ""
	into := ""

	if p.current.Type == token.IDENT && p.peeked.Literal != "↦" && p.peeked.Type != token.BIND {
		cert = p.current.Literal
		p.nextToken()
	}

	for p.current.Type != token.STMT_END && p.current.Type != token.EOF {
		lit := p.current.Literal
		if lit == "cert" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.IDENT {
				cert = p.current.Literal
				p.nextToken()
			}
		} else if lit == "decode" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.IDENT {
				decode = p.current.Literal
				p.nextToken()
			}
		} else if lit == "into" {
			p.nextToken()
			if p.current.Type == token.IDENT {
				into = p.current.Literal
				p.nextToken()
			}
		} else {
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.AADAttestStmt{
		Pos_:   pos,
		Cert:   cert,
		Decode: decode,
		Into:   into,
	}
}

func (p *Parser) parseMetaFitStmt() *ast.MetaFitStmt {
	pos := p.curPos()
	p.nextToken() // consume 'meta_fit'

	into := ""
	cert := ""

	for p.current.Type != token.STMT_END && p.current.Type != token.EOF {
		lit := p.current.Literal
		if lit == "into" {
			p.nextToken()
			if p.current.Type == token.IDENT {
				into = p.current.Literal
				p.nextToken()
			}
		} else if lit == "cert" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.IDENT {
				cert = p.current.Literal
				p.nextToken()
			}
		} else {
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.MetaFitStmt{
		Pos_: pos,
		Into: into,
		Cert: cert,
	}
}

func (p *Parser) parseMetaAdaptStmt() *ast.MetaAdaptStmt {
	pos := p.curPos()
	p.nextToken() // consume 'meta_adapt'

	state := ""
	if p.current.Type == token.IDENT {
		state = p.current.Literal
		p.nextToken()
	}

	env := ""
	if p.current.Type == token.WITH || p.current.Literal == "with" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			env = p.current.Literal
			p.nextToken()
		}
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.MetaAdaptStmt{
		Pos_:     pos,
		State:    state,
		Envelope: env,
		Into:     into,
	}
}

func (p *Parser) parseMetaVerifyStmt() *ast.MetaVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'meta_verify'

	state := ""
	if p.current.Type == token.IDENT {
		state = p.current.Literal
		p.nextToken()
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.MetaVerifyStmt{
		Pos_:  pos,
		State: state,
		Into:  into,
	}
}

func (p *Parser) parseCorruptAADEnvStmt() *ast.CorruptAADEnvStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_aad_env'

	env := ""
	if p.current.Type == token.IDENT {
		env = p.current.Literal
		p.nextToken()
	}

	if p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}

	offset := p.parseExpression(LOWEST)

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.CorruptAADEnvStmt{
		Pos_:   pos,
		Env:    env,
		Offset: offset,
	}
}

// ---------------------------------------------------------------------
// M8.1: Record, Variant, Match & M8.2: RSCA Parsers
// ---------------------------------------------------------------------

func (p *Parser) parseRecordDecl() *ast.RecordDeclStmt {
	pos := p.curPos()
	p.nextToken() // consume 'record'

	name := ""
	if p.current.Type == token.IDENT {
		name = p.current.Literal
		p.nextToken()
	}

	p.expect(token.BLOCK_OPEN) // ⟦
	var fields []*ast.FieldDecl
	for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
		if p.current.Type == token.BLOCK_CLOSE || p.current.Type == token.EOF {
			break
		}
		fPos := p.curPos()
		fName := p.current.Literal
		p.nextToken()

		if p.current.Type == token.BIND {
			p.nextToken() // ↦
		}
		fType := "text"
		if p.current.Type != token.STMT_END && p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
			fType = p.current.Literal
			p.nextToken()
		}
		if p.current.Type == token.STMT_END {
			p.nextToken()
		}
		fields = append(fields, &ast.FieldDecl{
			Pos_: fPos,
			Name: fName,
			Type: fType,
		})
	}
	p.expect(token.BLOCK_CLOSE) // ⟧

	return &ast.RecordDeclStmt{
		Pos_:   pos,
		Name:   name,
		Fields: fields,
	}
}

func (p *Parser) parseVariantDecl() *ast.VariantDeclStmt {
	pos := p.curPos()
	p.nextToken() // consume 'variant'

	name := ""
	if p.current.Type == token.IDENT {
		name = p.current.Literal
		p.nextToken()
	}

	p.expect(token.BLOCK_OPEN) // ⟦
	var cases []*ast.VariantCase
	for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
		if p.current.Type == token.PIPE {
			p.nextToken() // consume |
		}
		if p.current.Type == token.BLOCK_CLOSE || p.current.Type == token.EOF {
			break
		}
		cPos := p.curPos()
		cName := p.current.Literal
		p.nextToken()

		var fields []*ast.FieldDecl
		if p.current.Type == token.GROUP_OPEN {
			p.nextToken() // consume ( or ⦅
			for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
				fPos := p.curPos()
				fName := p.current.Literal
				p.nextToken()

				if p.current.Type == token.BIND {
					p.nextToken() // ↦
				}
				fType := "text"
				if p.current.Type != token.SEP && p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
					fType = p.current.Literal
					p.nextToken()
				}
				fields = append(fields, &ast.FieldDecl{
					Pos_: fPos,
					Name: fName,
					Type: fType,
				})
				if p.current.Type == token.SEP {
					p.nextToken() // consume ⸴
				} else {
					break
				}
			}
			p.expect(token.GROUP_CLOSE) // ) or ⦆
		}
		if p.current.Type == token.STMT_END {
			p.nextToken()
		}

		cases = append(cases, &ast.VariantCase{
			Pos_:   cPos,
			Name:   cName,
			Fields: fields,
		})
	}
	p.expect(token.BLOCK_CLOSE) // ⟧

	return &ast.VariantDeclStmt{
		Pos_:  pos,
		Name:  name,
		Cases: cases,
	}
}

func (p *Parser) parseMatchStmt() *ast.MatchStmt {
	pos := p.curPos()
	p.nextToken() // consume 'match'

	subj := p.parseExpression(LOWEST)

	p.expect(token.BLOCK_OPEN) // ⟦
	var arms []*ast.MatchArm
	var otherwiseBlock *ast.Block

	for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
		if p.current.Type == token.OTHERWISE {
			p.nextToken() // consume 'otherwise'
			if p.current.Type == token.IMPLIES {
				p.nextToken() // consume ⟹ or ⇒
			}
			var stmts []ast.Stmt
			for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF && p.current.Type != token.PIPE {
				stmt := p.parseStatement()
				if stmt != nil {
					stmts = append(stmts, stmt)
				}
			}
			otherwiseBlock = &ast.Block{
				Pos_:       p.curPos(),
				Statements: stmts,
			}
			continue
		}

		if p.current.Type == token.PIPE {
			p.nextToken() // consume |
		}
		if p.current.Type == token.BLOCK_CLOSE || p.current.Type == token.EOF {
			break
		}

		aPos := p.curPos()
		cName := p.current.Literal
		p.nextToken()

		var bindings []string
		if p.current.Type == token.GROUP_OPEN {
			p.nextToken() // consume ( or ⦅
			for p.current.Type != token.GROUP_CLOSE && p.current.Type != token.EOF {
				if p.current.Type == token.IDENT {
					bindings = append(bindings, p.current.Literal)
					p.nextToken()
				}
				if p.current.Type == token.SEP {
					p.nextToken() // consume ⸴
				} else {
					break
				}
			}
			p.expect(token.GROUP_CLOSE) // ) or ⦆
		}

		if p.current.Type == token.IMPLIES {
			p.nextToken() // consume ⟹ or ⇒
		}

		var armStmts []ast.Stmt
		for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF && p.current.Type != token.PIPE && p.current.Type != token.OTHERWISE {
			stmt := p.parseStatement()
			if stmt != nil {
				armStmts = append(armStmts, stmt)
			}
		}

		arms = append(arms, &ast.MatchArm{
			Pos_:     aPos,
			CaseName: cName,
			Bindings: bindings,
			Body: &ast.Block{
				Pos_:       aPos,
				Statements: armStmts,
			},
		})
	}
	p.expect(token.BLOCK_CLOSE) // ⟧

	return &ast.MatchStmt{
		Pos_:      pos,
		Subject:   subj,
		Arms:      arms,
		Otherwise: otherwiseBlock,
	}
}

func (p *Parser) parseRecordInitExpr() *ast.RecordInitExpr {
	pos := p.curPos()
	p.nextToken() // consume 'record_init'

	recName := ""
	if p.current.Type == token.IDENT {
		recName = p.current.Literal
		p.nextToken()
	}

	p.expect(token.BLOCK_OPEN) // ⟦
	var fields []*ast.RecordFieldInit
	for p.current.Type != token.BLOCK_CLOSE && p.current.Type != token.EOF {
		if p.current.Type == token.BLOCK_CLOSE || p.current.Type == token.EOF {
			break
		}
		fPos := p.curPos()
		fName := p.current.Literal
		p.nextToken()

		if p.current.Type == token.BIND {
			p.nextToken() // ↦
		}
		val := p.parseExpression(LOWEST)
		fields = append(fields, &ast.RecordFieldInit{
			Pos_:  fPos,
			Field: fName,
			Value: val,
		})
		if p.current.Type == token.SEP || p.current.Type == token.STMT_END {
			p.nextToken()
		}
	}
	p.expect(token.BLOCK_CLOSE) // ⟧

	return &ast.RecordInitExpr{
		Pos_:       pos,
		RecordName: recName,
		Fields:     fields,
	}
}

func (p *Parser) parseVariantInitExpr() *ast.VariantInitExpr {
	pos := p.curPos()
	p.nextToken() // consume 'variant_init'

	vName := ""
	if p.current.Type == token.IDENT {
		vName = p.current.Literal
		p.nextToken()
	}

	p.expect(token.GROUP_OPEN) // ⦅
	cName := ""
	if p.current.Type != token.GROUP_CLOSE && p.current.Type != token.BIND {
		cName = p.current.Literal
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken() // ↦
	}
	val := p.parseExpression(LOWEST)
	var values []ast.Expr
	if val != nil {
		values = append(values, val)
	}
	for p.current.Type == token.SEP {
		p.nextToken()
		v := p.parseExpression(LOWEST)
		if v != nil {
			values = append(values, v)
		}
	}
	p.expect(token.GROUP_CLOSE) // ⦆

	return &ast.VariantInitExpr{
		Pos_:        pos,
		VariantName: vName,
		CaseName:    cName,
		Value:       val,
		Values:      values,
	}
}

func (p *Parser) parseRSCAAttestStmt() *ast.RSCAAttestStmt {
	pos := p.curPos()
	p.nextToken() // consume 'rsca_attest'

	stage := 0
	source := ""
	compiler := ""
	output := ""
	into := ""

	for p.current.Type != token.STMT_END && p.current.Type != token.EOF {
		lit := p.current.Literal
		if lit == "stage" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.INT {
				stage = int(parseIntLiteral(p.current.Literal))
				p.nextToken()
			}
		} else if lit == "from" {
			p.nextToken()
		} else if lit == "source" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.IDENT || p.current.Type == token.STRING {
				source = p.current.Literal
				p.nextToken()
			}
		} else if lit == "compiler" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.IDENT || p.current.Type == token.STRING {
				compiler = p.current.Literal
				p.nextToken()
			}
		} else if lit == "output" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.IDENT || p.current.Type == token.STRING {
				output = p.current.Literal
				p.nextToken()
			}
		} else if lit == "into" {
			p.nextToken()
			if p.current.Type == token.IDENT {
				into = p.current.Literal
				p.nextToken()
			}
		} else {
			p.nextToken()
		}
	}
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.RSCAAttestStmt{
		Pos_:     pos,
		Stage:    stage,
		Source:   source,
		Compiler: compiler,
		Output:   output,
		Into:     into,
	}
}

func (p *Parser) parseRSCAVerifyStmt() *ast.RSCAVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'rsca_verify'

	rec := ""
	if p.current.Type == token.IDENT {
		rec = p.current.Literal
		p.nextToken()
	}

	into := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			into = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.RSCAVerifyStmt{
		Pos_:   pos,
		Record: rec,
		Into:   into,
	}
}

func (p *Parser) parseCorruptRSCAStmt() *ast.CorruptRSCAStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_rsca'

	rec := ""
	if p.current.Type == token.IDENT {
		rec = p.current.Literal
		p.nextToken()
	}

	if p.current.Literal == "at" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}

	offset := p.parseExpression(LOWEST)
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.CorruptRSCAStmt{
		Pos_:   pos,
		Record: rec,
		Offset: offset,
	}
}

func (p *Parser) parseCTSAAttestStmt() *ast.CTSAAttestStmt {
	pos := p.curPos()
	p.nextToken() // consume 'ctsa_attest'

	astName := ""
	envName := ""
	rulesName := ""
	intoName := ""

	for p.current.Type != token.STMT_END && p.current.Type != token.EOF {
		lit := p.current.Literal
		if lit == "ast" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.IDENT || p.current.Type == token.STRING {
				astName = p.current.Literal
				p.nextToken()
			}
		} else if lit == "env" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.IDENT || p.current.Type == token.STRING {
				envName = p.current.Literal
				p.nextToken()
			}
		} else if lit == "rules" {
			p.nextToken()
			if p.current.Type == token.BIND {
				p.nextToken()
			}
			if p.current.Type == token.IDENT || p.current.Type == token.STRING {
				rulesName = p.current.Literal
				p.nextToken()
			}
		} else if lit == "into" {
			p.nextToken()
			if p.current.Type == token.IDENT {
				intoName = p.current.Literal
				p.nextToken()
			}
		} else {
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.CTSAAttestStmt{
		Pos_:  pos,
		Ast:   astName,
		Env:   envName,
		Rules: rulesName,
		Into:  intoName,
	}
}

func (p *Parser) parseCTSAVerifyStmt() *ast.CTSAVerifyStmt {
	pos := p.curPos()
	p.nextToken() // consume 'ctsa_verify'

	certName := ""
	if p.current.Type == token.IDENT {
		certName = p.current.Literal
		p.nextToken()
	}

	intoName := ""
	if p.current.Literal == "into" {
		p.nextToken()
		if p.current.Type == token.IDENT {
			intoName = p.current.Literal
			p.nextToken()
		}
	}

	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.CTSAVerifyStmt{
		Pos_: pos,
		Cert: certName,
		Into: intoName,
	}
}

func (p *Parser) parseCorruptCTSAStmt() *ast.CorruptCTSAStmt {
	pos := p.curPos()
	p.nextToken() // consume 'corrupt_ctsa'

	certName := ""
	if p.current.Type == token.IDENT {
		certName = p.current.Literal
		p.nextToken()
	}

	if p.current.Literal == "field" {
		p.nextToken()
	}
	if p.current.Type == token.BIND {
		p.nextToken()
	}

	fieldExpr := p.parseExpression(LOWEST)
	if p.current.Type == token.STMT_END {
		p.nextToken()
	}

	return &ast.CorruptCTSAStmt{
		Pos_:  pos,
		Cert:  certName,
		Field: fieldExpr,
	}
}
