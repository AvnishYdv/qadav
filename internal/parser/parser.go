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
	if isTypeToken(p.current.Type) {
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
	case isTypeToken(p.current.Type):
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
	nameTok := p.expect(token.IDENT)
	p.expect(token.BIND)
	val := p.parseExpression(LOWEST)
	p.expect(token.STMT_END)

	return &ast.BindStmt{
		Pos_:  pos,
		Type:  btype,
		Name:  nameTok.Literal,
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
		case token.MEMBER_ACCESS:
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
	case token.IDENT, token.AFFIRM, token.DENY, token.TOP, token.BOTTOM:
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
	p.expect(token.MEMBER_ACCESS) // ∸
	memTok := p.expect(token.IDENT)

	return &ast.MemberExpr{
		Pos_:   pos,
		Object: obj,
		Member: memTok.Literal,
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
		token.MEMORY, token.GRAPH, token.PROCEDURE:
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
		if p.current.Type == token.IDENT {
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
