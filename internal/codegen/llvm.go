package codegen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/llvm"
	"github.com/qadav-lang/qadav/internal/runtime"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

// Value represents a generated LLVM IR operand.
type Value struct {
	Type  string      // e.g. "i64", "double", "i1", "i8*"
	Ref   string      // e.g. "%1", "42", "@.str.0"
	QType *types.Type // Qadav type from checker
}

// Generator produces LLVM IR text from a type-checked AST.
type Generator struct {
	checker    *checker.Checker
	prog       *ast.Program
	exprTypes  map[ast.Expr]*types.Type
	tempCount  int
	labelCount int
	strCount   int

	// String constants: name -> {content, length}
	stringConstants []stringConstant

	// Shape constants: name -> dims
	shapeConstants []shapeConstant
	shapeCount     int

	// Per-process state
	vars                 map[string]*varSlot
	currentProc          *ast.ProcessDecl
	currentBlock         string
	isTerminated         bool
	activeReasoningChain string
	modalities           map[string]modalityInfo
	abstainPolicies      map[string]string
	sb                   strings.Builder
}

type stringConstant struct {
	name    string
	escaped string
	length  int
}

type shapeConstant struct {
	name string
	dims []int
}

type varSlot struct {
	ptr      string // e.g. "%x.ptr"
	llvmType string // e.g. "i64"
	qtype    *types.Type
}

// New creates an initialized LLVM code generator.
func New(chk *checker.Checker) *Generator {
	g := &Generator{
		checker:         chk,
		exprTypes:       make(map[ast.Expr]*types.Type),
		vars:            make(map[string]*varSlot),
		abstainPolicies: make(map[string]string),
	}
	if chk != nil {
		g.exprTypes = chk.ExprTypes()
	}
	return g
}

func (g *Generator) nextTemp() string {
	g.tempCount++
	return fmt.Sprintf("%%t%d", g.tempCount)
}

func (g *Generator) nextLabel(prefix string) string {
	g.labelCount++
	return fmt.Sprintf("%s_%d", prefix, g.labelCount)
}

func (g *Generator) addStringConstant(raw string) (string, int) {
	name := fmt.Sprintf("@.str.%d", g.strCount)
	g.strCount++

	var escaped strings.Builder
	length := 0
	for i := 0; i < len(raw); i++ {
		b := raw[i]
		switch b {
		case '\n':
			escaped.WriteString(`\0A`)
			length++
		case '\r':
			escaped.WriteString(`\0D`)
			length++
		case '\t':
			escaped.WriteString(`\09`)
			length++
		case '"':
			escaped.WriteString(`\22`)
			length++
		case '\\':
			escaped.WriteString(`\5C`)
			length++
		default:
			if b >= 32 && b <= 126 {
				escaped.WriteByte(b)
			} else {
				fmt.Fprintf(&escaped, `\%02X`, b)
			}
			length++
		}
	}
	escaped.WriteString(`\00`)
	length++ // for null terminator

	g.stringConstants = append(g.stringConstants, stringConstant{
		name:    name,
		escaped: escaped.String(),
		length:  length,
	})

	return name, length
}

func (g *Generator) addShapeConstant(dims []int) string {
	name := fmt.Sprintf("@.shape.%d", g.shapeCount)
	g.shapeCount++
	g.shapeConstants = append(g.shapeConstants, shapeConstant{
		name: name,
		dims: dims,
	})
	return name
}

// Generate translates the Program AST into complete LLVM IR text.
func (g *Generator) Generate(prog *ast.Program) (string, error) {
	g.prog = prog
	var bodyBuilder strings.Builder

	// Generate all process functions first into bodyBuilder
	for _, decl := range prog.Decls {
		if proc, ok := decl.(*ast.ProcessDecl); ok {
			procIR, err := g.genProcess(proc)
			if err != nil {
				return "", err
			}
			bodyBuilder.WriteString(procIR)
			bodyBuilder.WriteString("\n")
		} else if pdecl, ok := decl.(*ast.ProcedureDecl); ok {
			g.sb.Reset()
			g.genProcedureDecl(pdecl)
			bodyBuilder.WriteString(g.sb.String())
			bodyBuilder.WriteString("\n")
		}
	}

	// Now build full module header + runtime globals + body
	var header strings.Builder
	header.WriteString("; ModuleID = 'qadav_module'\n")
	header.WriteString("source_filename = \"qadav_module\"\n\n")

	// Standard I/O runtime declarations (Qadav Runtime - zero libc)
	header.WriteString("declare void @qadav_emit_int(i64)\n")
	header.WriteString("declare void @qadav_emit_float(double)\n")
	header.WriteString("declare void @qadav_emit_text(i8*)\n")
	header.WriteString("declare void @qadav_emit_logic(i32)\n")
	header.WriteString("declare void @qadav_emit_qubit(i8*)\n")
	header.WriteString("declare void @qadav_exit(i32)\n\n")

	// Tensor runtime declarations (zero libc)
	header.WriteString("declare i8* @qadav_tensor_create(i32, i32, i32*)\n")
	header.WriteString("declare void @qadav_tensor_free(i8*)\n")
	header.WriteString("declare void @qadav_tensor_set_i64(i8*, i32, i64)\n")
	header.WriteString("declare i64 @qadav_tensor_get_i64(i8*, i32)\n")
	header.WriteString("declare void @qadav_tensor_set_f64(i8*, i32, double)\n")
	header.WriteString("declare double @qadav_tensor_get_f64(i8*, i32)\n")
	header.WriteString("declare i8* @qadav_tensor_add(i8*, i8*)\n")
	header.WriteString("declare i8* @qadav_tensor_sub(i8*, i8*)\n")
	header.WriteString("declare i8* @qadav_tensor_mul(i8*, i8*)\n")
	header.WriteString("declare i8* @qadav_tensor_div(i8*, i8*)\n")
	header.WriteString("declare i8* @qadav_tensor_add_scalar_f64(i8*, double)\n")
	header.WriteString("declare i8* @qadav_tensor_mul_scalar_f64(i8*, double)\n")
	header.WriteString("declare i8* @qadav_tensor_sub_scalar_f64(i8*, double)\n")
	header.WriteString("declare i8* @qadav_tensor_div_scalar_f64(i8*, double)\n")
	header.WriteString("declare i8* @qadav_tensor_add_scalar_i64(i8*, i64)\n")
	header.WriteString("declare i8* @qadav_tensor_mul_scalar_i64(i8*, i64)\n")
	header.WriteString("declare i8* @qadav_tensor_sub_scalar_i64(i8*, i64)\n")
	header.WriteString("declare i8* @qadav_tensor_div_scalar_i64(i8*, i64)\n")
	header.WriteString("declare i8* @qadav_tensor_matmul(i8*, i8*)\n")
	header.WriteString("declare i64 @qadav_tensor_reduce_sum_i64(i8*)\n")
	header.WriteString("declare double @qadav_tensor_reduce_sum_f64(i8*)\n")
	header.WriteString("declare i64 @qadav_tensor_reduce_prod_i64(i8*)\n")
	header.WriteString("declare double @qadav_tensor_reduce_prod_f64(i8*)\n")
	header.WriteString("declare i8* @qadav_tensor_transpose(i8*)\n")
	header.WriteString("declare void @qadav_tensor_print_i64(i8*)\n")
	header.WriteString("declare void @qadav_tensor_print_f64(i8*)\n\n")

	// Neural network runtime declarations (zero libc)
	header.WriteString("declare i8* @qadav_network_create(i32, i32*, i32*, i32*)\n")
	header.WriteString("declare void @qadav_network_free(i8*)\n")
	header.WriteString("declare i8* @qadav_network_forward(i8*, i8*)\n")
	header.WriteString("declare void @qadav_network_backward(i8*, i8*, i8*)\n")
	header.WriteString("declare void @qadav_network_sgd_step(i8*, double)\n")
	header.WriteString("declare void @qadav_network_adam_step(i8*, double, i32)\n")
	header.WriteString("declare double @qadav_loss_mse(i8*, i8*)\n")
	header.WriteString("declare double @qadav_loss_cross_entropy(i8*, i8*)\n")
	header.WriteString("declare double @qadav_train_step(i8*, i8*, i8*, i32, i32, double, i32)\n")
	header.WriteString("declare void @qadav_train_epochs(i8*, i8*, i8*, i32, i32, double, i32)\n")
	header.WriteString("declare double @qadav_sqrt(double)\n")
	header.WriteString("declare double @qadav_exp(double)\n")
	header.WriteString("declare double @qadav_log(double)\n\n")

	// Quantum runtime declarations (zero libc, xorshift64 PRNG — NOT cryptographic)
	header.WriteString("declare i8* @qadav_get_qstate()\n")
	header.WriteString("declare i8* @qadav_qstate_create(i32)\n")
	header.WriteString("declare void @qadav_qstate_free(i8*)\n")
	header.WriteString("declare i32 @qadav_qubit_alloc(i8*)\n")
	header.WriteString("declare void @qadav_gate_h(i8*, i32)\n")
	header.WriteString("declare void @qadav_gate_x(i8*, i32)\n")
	header.WriteString("declare void @qadav_gate_y(i8*, i32)\n")
	header.WriteString("declare void @qadav_gate_z(i8*, i32)\n")
	header.WriteString("declare void @qadav_gate_s(i8*, i32)\n")
	header.WriteString("declare void @qadav_gate_t(i8*, i32)\n")
	header.WriteString("declare void @qadav_gate_cnot(i8*, i32, i32)\n")
	header.WriteString("declare i32 @qadav_measure(i8*, i32)\n")
	header.WriteString("declare double @qadav_probability(i8*, i32)\n")
	header.WriteString("declare void @qadav_qstate_print(i8*)\n\n")

	// Prime-symbolic runtime declarations (zero libc)
	header.WriteString("declare i32 @qadav_is_prime(i64)\n")
	header.WriteString("declare i32 @qadav_factorize(i64, i64*, i32)\n")
	header.WriteString("declare i64 @qadav_encode_concept(i64*, i32)\n")
	header.WriteString("declare void @qadav_print_concept(i64)\n\n")

	// Dataset runtime declarations (zero libc)
	header.WriteString("declare i8* @qadav_dataset_load(i8*)\n")
	header.WriteString("declare void @qadav_dataset_free(i8*)\n")
	header.WriteString("declare i8* @qadav_dataset_get(i8*, i8*)\n")
	header.WriteString("declare void @qadav_dataset_normalize_minmax(i8*)\n")
	header.WriteString("declare void @qadav_dataset_normalize_zscore(i8*)\n")
	header.WriteString("declare void @qadav_dataset_shuffle(i8*)\n")
	header.WriteString("declare void @qadav_dataset_print(i8*)\n\n")

	// Quasinary runtime declarations (zero libc)
	header.WriteString("declare i32 @qadav_quasi_add(i64, i64)\n")
	header.WriteString("declare i32 @qadav_quasi_mul(i64, i64)\n")
	header.WriteString("declare i32 @qadav_quasi_neg(i64)\n")
	header.WriteString("declare i8* @qadav_quasi_grid_create(i32, i32)\n")
	header.WriteString("declare i64 @qadav_quasi_grid_get(i8*, i32, i32)\n")
	header.WriteString("declare void @qadav_quasi_grid_set(i8*, i32, i32, i64)\n")
	header.WriteString("declare double @qadav_quasi_energy_ising(i8*)\n")
	header.WriteString("declare double @qadav_quasi_energy_majority(i8*)\n")
	header.WriteString("declare void @qadav_quasi_minimize(i8*, i32)\n")
	header.WriteString("declare void @qadav_quasi_print(i64)\n")
	header.WriteString("declare void @qadav_quasi_grid_print(i8*)\n\n")

	// CIR runtime declarations (zero libc)
	header.WriteString("declare i8* @qadav_confidence_create(double, double, i8)\n")
	header.WriteString("declare void @qadav_confidence_free(i8*)\n")
	header.WriteString("declare void @qadav_confidence_print(i8*)\n")
	header.WriteString("declare i8* @qadav_confidence_add(i8*, i8*)\n")
	header.WriteString("declare i8* @qadav_confidence_mul(i8*, i8*)\n")
	header.WriteString("declare i8* @qadav_confidence_min(i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_confidence_assert(i8*, double)\n\n")

	// AGI native types runtime declarations (zero libc)
	header.WriteString("declare i8* @qadav_memory_create(i64)\n")
	header.WriteString("declare void @qadav_memory_free(i8*)\n")
	header.WriteString("declare void @qadav_memory_remember(i8*, i8*, i64)\n")
	header.WriteString("declare i64 @qadav_memory_recall(i8*, i8*)\n")
	header.WriteString("declare void @qadav_memory_print(i8*)\n")
	header.WriteString("declare i8* @qadav_graph_create()\n")
	header.WriteString("declare void @qadav_graph_free(i8*)\n")
	header.WriteString("declare void @qadav_graph_node(i8*, i8*, i64)\n")
	header.WriteString("declare void @qadav_graph_link(i8*, i8*, i8*, double)\n")
	header.WriteString("declare i8* @qadav_graph_trace(i8*, i8*, i64)\n")
	header.WriteString("declare void @qadav_graph_print(i8*)\n")
	header.WriteString("declare i8* @qadav_procedure_create(i8*)\n")
	header.WriteString("declare void @qadav_procedure_free(i8*)\n")
	header.WriteString("declare i8* @qadav_qeeh_create(i32, double*)\n")
	header.WriteString("declare void @qadav_qeeh_set_rule_name(i8*, i32, i8*)\n")
	header.WriteString("declare i32 @qadav_qeeh_boundary_hash(i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_qeeh_evaluate(i8*, double*)\n")
	header.WriteString("declare i32 @qadav_qeeh_evaluate_action(i8*, i8*)\n")
	header.WriteString("declare void @qadav_qeeh_trigger_decoherence(i8*)\n")
	header.WriteString("declare void @qadav_qeeh_free(i8*)\n\n")

	header.WriteString("declare i8* @qadav_qemwv_create(i32, i32, i8*)\n")
	header.WriteString("declare i32 @qadav_qemwv_seal(i8*, i8**, i32)\n")
	header.WriteString("declare i8** @qadav_qemwv_unseal(i8*, i8*)\n")
	header.WriteString("declare void @qadav_qemwv_wipe(i8*)\n")
	header.WriteString("declare i8* @qadav_qemwv_read_share_unauthorized(i8*, i32)\n")
	header.WriteString("declare void @qadav_qemwv_free(i8*)\n\n")

	header.WriteString("declare i8* @qadav_mctp_init(i32)\n")
	header.WriteString("declare void @qadav_mctp_record_epoch(i8*, i32, i8*, i64, i8*, i64, double)\n")
	header.WriteString("declare void @qadav_mctp_record_epochs(i8*, i32, i8*, i64, i8*, i64, double)\n")
	header.WriteString("declare void @qadav_mctp_certify(i8*)\n")
	header.WriteString("declare i32 @qadav_mctp_verify(i8*, i8*, i64)\n")
	header.WriteString("declare void @qadav_mctp_free(i8*)\n\n")

	// QEMP (Quantum-Entangled Memory Provenance)
	header.WriteString("declare i8* @qadav_qemp_create(i8*, i32)\n")
	header.WriteString("declare i32 @qadav_qemp_remember_text(i8*, i8*, i8*, i8*)\n")
	header.WriteString("declare i8* @qadav_qemp_recall_text(i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_qemp_verify_all(i8*)\n")
	header.WriteString("declare i32 @qadav_qemp_audit(i8*)\n")
	header.WriteString("declare i32 @qadav_qemp_corrupt_share(i8*, i8*, i32)\n")
	header.WriteString("declare void @qadav_qemp_wipe_all(i8*)\n")
	header.WriteString("declare void @qadav_qemp_free(i8*)\n\n")

	// QECMT (Quantum-Entangled Cross-Model Transfer)
	header.WriteString("declare i32 @qadav_qecmt_issue_license(i8*, i8*, i8*, i64, double, i64, i8*)\n")
	header.WriteString("declare i32 @qadav_qecmt_transfer_layer(i8*, i32, i8*, i8*, i32*, i32, i8*)\n")
	header.WriteString("declare i32 @qadav_qecmt_distill(i8*, i32, i8*, i8*, i8*, double, i32)\n")
	header.WriteString("declare i32 @qadav_qecmt_verify_chain(i8*)\n")
	header.WriteString("declare void @qadav_qecmt_free(i8*)\n\n")

	// QERA (Quantum-Entangled Reasoning Attestation - Invention #11)
	header.WriteString("declare i8* @qadav_qera_create(i8*, i32, i8*)\n")
	header.WriteString("declare i32 @qadav_qera_reason(i8*, i32, i8*, i64, i8*, i64, i8*, i64, double, i8*)\n")
	header.WriteString("declare i32 @qadav_qera_verify(i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_qera_audit(i8*)\n")
	header.WriteString("declare i32 @qadav_qera_corrupt_step(i8*, i32, i32)\n")
	header.WriteString("declare void @qadav_qera_free(i8*)\n\n")

	// QECMA (Quantum-Entangled Cross-Modal Attestation - Invention #12)
	header.WriteString("declare i8* @qadav_qecma_create()\n")
	header.WriteString("declare void @qadav_qecma_free(i8*)\n")
	header.WriteString("declare i32 @qadav_qecma_add(i8*, i8, i8*, i64)\n")
	header.WriteString("declare i32 @qadav_qecma_build(i8*)\n")
	header.WriteString("declare i32 @qadav_qecma_get_unified_root(i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_qera_reason_cross_modal(i8*, i32, i8*, i8*, i64, double, i8*)\n\n")

	// QEC-QEMP (Invention #13)
	header.WriteString("declare i8* @qadav_qec_create(i8*, i32, i32, i8*)\n")
	header.WriteString("declare i32 @qadav_qec_measure(i8*, i8*, i64, i64*, i8*)\n")
	header.WriteString("declare i32 @qadav_qec_verify(i8*, i64, i8*)\n")
	header.WriteString("declare i32 @qadav_qec_correct(i8*, i64, i8*)\n")
	header.WriteString("declare i32 @qadav_qec_corrupt_syndrome(i8*, i64, i32)\n")
	header.WriteString("declare void @qadav_qec_free(i8*)\n\n")

	// Memory Consolidation
	header.WriteString("declare i32 @qadav_consolidate_memory(i8*, i8*, double, i8*)\n\n")

	// CB-SRCPP (Invention #14)
	header.WriteString("declare i8* @qadav_scoped_issue(i8*, i8*, i8*, i8**, i32, i64)\n")
	header.WriteString("declare i32 @qadav_scoped_grant(i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_scoped_revoke(i8*)\n")
	header.WriteString("declare i32 @qadav_scoped_verify(i8*, i8*, i8*, i64)\n")
	header.WriteString("declare i32 @qadav_scoped_recall(i8*, i8*, i8*, i8*, i64, i8**, i64*)\n")
	header.WriteString("declare void @qadav_scoped_free(i8*)\n\n")

	// Safe Tool Orchestration (M5.8)
	header.WriteString("declare i32 @qadav_tool_register(i8*, i8*, i32, i32)\n")
	header.WriteString("declare i32 @qadav_tool_call(i8*, i8*, i8*, i8*)\n")
	header.WriteString("declare void @qadav_tools_reset()\n\n")

	// TTPCR (Invention #15) & Calibrated Abstention (M5.9 + M5.10)
	header.WriteString("declare i8* @qadav_alloc(i64)\n")
	header.WriteString("declare void @qadav_free(i8*)\n")
	header.WriteString("declare i32 @qadav_certificate_attach(i8*, i64, i8, i8*, i32, double)\n")
	header.WriteString("declare i32 @qadav_certificate_verify(i8*)\n")
	header.WriteString("declare i32 @qadav_certificate_verify_chain(i8*)\n")
	header.WriteString("declare double @qadav_abstain_compute_confidence(i8*)\n")
	header.WriteString("declare i32 @qadav_abstain_check(i8*, i8*, i32)\n")
	header.WriteString("declare i32 @qadav_abstain_execute(i8*, i8*)\n")
	header.WriteString("declare i8* @qadav_sandbox_create(i8*, i8*, i32)\n")
	header.WriteString("declare i32 @qadav_sandbox_run(i8*, i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_sandbox_verify(i8*, i8*)\n")
	header.WriteString("declare void @qadav_sandbox_quarantine(i8*)\n")
	header.WriteString("declare void @qadav_sandbox_free(i8*)\n")
	header.WriteString("declare i8* @qadav_sandbox_find(i8*)\n")
	header.WriteString("declare i8* @qadav_sandbox_verify_text(i8*)\n\n")

	// DirectCompute GPU Backend (M6.1) & AGKE (M6.2 — Invention #17)
	header.WriteString("declare i32 @qadav_gpu_init()\n")
	header.WriteString("declare i32 @qadav_gpu_is_available()\n")
	header.WriteString("declare i8* @qadav_gpu_tensor_matmul(i8*, i8*)\n")
	header.WriteString("declare i8* @qadav_gpu_tensor_softmax(i8*)\n")
	header.WriteString("declare i8* @qadav_gpu_tensor_layernorm(i8*, double)\n")
	header.WriteString("declare i32 @qadav_gpu_matmul_tensor_raw(i8*, i8*, i8*, i32, i32, i32)\n")
	header.WriteString("declare i32 @qadav_gpu_softmax_tensor_raw(i8*, i8*, i32, i32)\n")
	header.WriteString("declare i32 @qadav_gpu_layernorm_tensor_raw(i8*, i8*, i32, i32, double)\n")
	header.WriteString("declare i32 @qadav_agke_attest(i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_agke_verify(i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_agke_corrupt_record(i32, i32)\n\n")

	// Multi-GPU Pipeline Parallelism (M6.3) & CDAGA (M6.4 — Invention #18)
	header.WriteString("declare i8* @qadav_multigpu_init()\n")
	header.WriteString("declare i32 @qadav_multigpu_build_pipeline(i8*, i32, i32, i32)\n")
	header.WriteString("declare i32 @qadav_multigpu_run_pipeline(i8*, i8*, i8*)\n")
	header.WriteString("declare void @qadav_multigpu_shutdown(i8*)\n")
	header.WriteString("declare i64 @qadav_multigpu_get_device_count()\n")
	header.WriteString("declare i32 @qadav_cdaga_record_aggregation(i8*, i64, i8**, i64*, i32, i8*)\n")
	header.WriteString("declare i32 @qadav_cdaga_verify(i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_cdaga_attest(i8*, i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_cdaga_corrupt(i8*, i32)\n\n")

	// Distributed Training (M6.5) & BARAT (M6.6 — Invention #19)
	header.WriteString("declare i8* @qadav_dist_init(i32)\n")
	header.WriteString("declare i32 @qadav_dist_ring_allreduce(i8*)\n")
	header.WriteString("declare i32 @qadav_dist_pserver_aggregate(i8*)\n")
	header.WriteString("declare void @qadav_dist_shutdown(i8*)\n")
	header.WriteString("declare i64 @qadav_dist_get_worker_count()\n")

	header.WriteString("declare i8* @qadav_barat_create_group(i32)\n")
	header.WriteString("declare void @qadav_barat_destroy_group(i8*)\n")
	header.WriteString("declare i32 @qadav_barat_submit(i8*, i32, i8*, i64)\n")
	header.WriteString("declare i32 @qadav_barat_aggregate(i8*, i8*, i64, i8*)\n")
	header.WriteString("declare i32 @qadav_barat_verify(i8*, i8*, i64)\n")
	header.WriteString("declare i32 @qadav_barat_attest(i8*, i8*, i8*, i64)\n")
	header.WriteString("declare i32 @qadav_barat_corrupt(i8*, i32)\n\n")

	// Quantum Hardware Abstraction Layer (QHAL — M7.1)
	header.WriteString("declare i8* @qhal_create(i32, i32)\n")
	header.WriteString("declare i32 @qhal_add_op(i8*, i8, i32, i32)\n")
	header.WriteString("declare i32 @qhal_export_qasm3(i8*, i8*, i64, i64*)\n")
	header.WriteString("declare i32 @qhal_hash_circuit(i8*)\n")
	header.WriteString("declare void @qhal_shutdown(i8*)\n")

	// Self-Evolution Attestation (SEA — Invention #20, M7.2)
	header.WriteString("declare void @sea_hash_text(i8*, i8*)\n")
	header.WriteString("declare i8* @sea_create()\n")
	header.WriteString("declare i32 @sea_propose(i8*, i8*, i8*, i8*, i8**)\n")
	header.WriteString("declare i32 @sea_verify_proposal(i8*, i8*)\n")
	header.WriteString("declare i32 @sea_verify_history(i8*)\n")
	header.WriteString("declare i32 @sea_attest(i8*, i8*)\n")
	header.WriteString("declare i32 @sea_corrupt(i8*, i32)\n")
	header.WriteString("declare void @sea_free(i8*)\n\n")

	// Hardware Target Transpilation (QHAL Transpiler — M7.3)
	header.WriteString("declare i32 @transpile_circuit(i8*, i32, i8*, i8*, i64, i64*, i32*)\n")
	header.WriteString("declare i32 @transpile_supported_targets(i8*, i64)\n\n")

	// Evolution Operators & EEA (Invention #21, M7.4)
	header.WriteString("declare i8* @eea_create_policy(i32, double, i64)\n")
	header.WriteString("declare i8* @eea_create_envelope()\n")
	header.WriteString("declare i32 @eea_build_envelope_text(i8*, i8*, i8*, i8*, i64, i64, i8*)\n")
	header.WriteString("declare i32 @eea_verify_envelope(i8*)\n")
	header.WriteString("declare i32 @eea_attest_envelope(i8*, i8*)\n")
	header.WriteString("declare i32 @eea_corrupt(i8*, i32)\n")
	header.WriteString("declare void @eea_free_envelope(i8*)\n")
	header.WriteString("declare void @eea_free(i8*)\n\n")

	// Real QEC Decoder (M7.5)
	header.WriteString("declare i8* @qecd_decode(i32, i32, i8*, i32)\n")
	header.WriteString("declare i32 @qecd_verify_decode_attestation(i8*)\n")
	header.WriteString("declare i32 @qecd_apply_corrections(i8*, i8*, i32)\n")
	header.WriteString("declare void @qecd_free(i8*)\n\n")

	// Invention #22: Attested Adaptive Decoder (AAD) & Meta-Learning (M7.6)
	header.WriteString("declare i8* @aad_build_training_cert(i8*, i64, i8*, i64, i64, i64, i64)\n")
	header.WriteString("declare i32 @aad_wrap_decode(i8*, i8*, i8*, i32, i8*)\n")
	header.WriteString("declare i32 @aad_verify_envelope(i8*, i8*)\n")
	header.WriteString("declare i32 @aad_attest_envelope(i8*, i8*)\n")
	header.WriteString("declare i32 @aad_meta_fit(i8*, i8*)\n")
	header.WriteString("declare i32 @aad_meta_adapt(i8*, i8*)\n")
	header.WriteString("declare i32 @aad_meta_verify(i8*)\n")
	header.WriteString("declare i32 @aad_corrupt_envelope(i8*, i32)\n")
	header.WriteString("declare void @aad_free_training_cert(i8*)\n")
	header.WriteString("declare void @aad_free_meta_state(i8*)\n\n")

	// Records & Variants (M8.1)
	header.WriteString("declare i8* @qadav_record_alloc(i64, i32, i32*)\n")
	header.WriteString("declare i32 @qadav_record_set_field(i8*, i32, i8*, i32)\n")
	header.WriteString("declare i8* @qadav_record_get_field(i8*, i32)\n")
	header.WriteString("declare i64 @qadav_record_get_field_i64(i8*, i32)\n")
	header.WriteString("declare void @qadav_record_set_field_i64(i8*, i32, i64)\n")
	header.WriteString("declare i8* @qadav_record_get_field_ptr(i8*, i32)\n")
	header.WriteString("declare void @qadav_record_set_field_ptr(i8*, i32, i8*)\n")
	header.WriteString("declare void @qadav_record_free(i8*)\n")

	header.WriteString("declare i8* @qadav_variant_create(i64, i32, i32, i8*)\n")
	header.WriteString("declare i8* @qadav_variant_create_i64(i64, i32, i64)\n")
	header.WriteString("declare i32 @qadav_variant_get_tag(i8*)\n")
	header.WriteString("declare i8* @qadav_variant_get_payload(i8*)\n")
	header.WriteString("declare i64 @qadav_variant_get_payload_i64(i8*)\n")
	header.WriteString("declare i32 @qadav_variant_match(i8*, i32)\n")
	header.WriteString("declare void @qadav_variant_free(i8*)\n")

	// RSCA (M8.2 - Invention #23)
	header.WriteString("declare i32 @qadav_rsca_record_stage(i64, i8*, i8*, i8*)\n")
	header.WriteString("declare i32 @qadav_rsca_verify_record(i32)\n")
	header.WriteString("declare i32 @qadav_rsca_compare_outputs(i32, i32)\n")
	header.WriteString("declare i32 @qadav_rsca_attest_chain()\n")
	header.WriteString("declare i32 @qadav_rsca_corrupt_record(i32, i32)\n\n")

	// CTSA (M8.3b - Invention #24)
	header.WriteString("declare i32 @qadav_ctsa_attest(i8*, i8*, i8*, i64, i64)\n")
	header.WriteString("declare i32 @qadav_ctsa_verify()\n")
	header.WriteString("declare i32 @qadav_ctsa_corrupt(i32)\n\n")

	// Quantum global state
	header.WriteString(g.generateQuantumGlobal())

	// Network declarations and constructor helpers
	for _, decl := range prog.Decls {
		if net, ok := decl.(*ast.NetworkDecl); ok {
			header.WriteString(g.genNetworkGlobals(net))
		}
	}

	// Generated string constants
	for _, sc := range g.stringConstants {
		fmt.Fprintf(&header, "%s = private unnamed_addr constant [%d x i8] c\"%s\"\n",
			sc.name, sc.length, sc.escaped)
	}

	// Generated shape constants
	for _, sc := range g.shapeConstants {
		dimsLen := len(sc.dims)
		if dimsLen == 0 {
			dimsLen = 1
		}
		var elemStrs []string
		if len(sc.dims) == 0 {
			elemStrs = append(elemStrs, "i32 0")
		} else {
			for _, d := range sc.dims {
				elemStrs = append(elemStrs, fmt.Sprintf("i32 %d", d))
			}
		}
		fmt.Fprintf(&header, "%s = private unnamed_addr constant [%d x i32] [%s]\n",
			sc.name, dimsLen, strings.Join(elemStrs, ", "))
	}
	header.WriteString("\n")

	header.WriteString(bodyBuilder.String())
	return header.String(), nil
}

func (g *Generator) genProcess(proc *ast.ProcessDecl) (string, error) {
	g.currentProc = proc
	g.vars = make(map[string]*varSlot)
	g.sb.Reset()
	g.isTerminated = false
	g.activeReasoningChain = ""

	isMain := proc.Name == "main"
	retLLVMType := "void"
	if isMain {
		retLLVMType = "i32"
	} else if proc.Returns != nil {
		retLLVMType = g.llvmType(g.resolveType(proc.Returns))
	}

	// Function header
	var paramsBuilder strings.Builder
	for i, param := range proc.Params {
		if i > 0 {
			paramsBuilder.WriteString(", ")
		}
		ptype := g.llvmType(g.resolveType(param.Type))
		fmt.Fprintf(&paramsBuilder, "%s %%%s.in", ptype, param.Name)
	}

	fmt.Fprintf(&g.sb, "define %s @\"%s\"(%s) {\n", retLLVMType, proc.Name, paramsBuilder.String())
	g.sb.WriteString("entry:\n")
	g.currentBlock = "entry"

	// Allocate space and store incoming parameters
	for _, param := range proc.Params {
		ptype := g.llvmType(g.resolveType(param.Type))
		ptr := fmt.Sprintf("%%%s.ptr", param.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca %s\n", ptr, ptype)
		fmt.Fprintf(&g.sb, "  store %s %%%s.in, %s* %s\n", ptype, param.Name, ptype, ptr)
		g.vars[param.Name] = &varSlot{
			ptr:      ptr,
			llvmType: ptype,
			qtype:    g.resolveType(param.Type),
		}
	}

	// Generate body statements
	if isMain && g.prog != nil {
		for _, decl := range g.prog.Decls {
			if learn, ok := decl.(*ast.LearnDecl); ok {
				g.genLearnStmt(learn)
			}
		}
	}

	if proc.Body != nil {
		for _, stmt := range proc.Body.Statements {
			if g.isTerminated {
				break
			}
			g.genStmt(stmt)
		}
	}

	// If entry block is not terminated, append default return
	if !g.isTerminated {
		if isMain {
			g.sb.WriteString("  ret i32 0\n")
		} else if retLLVMType == "void" {
			g.sb.WriteString("  ret void\n")
		} else {
			fmt.Fprintf(&g.sb, "  ret %s 0\n", retLLVMType)
		}
		g.isTerminated = true
	}

	g.sb.WriteString("}\n")
	return g.sb.String(), nil
}

func (g *Generator) genStmt(stmt ast.Stmt) {
	if stmt == nil || g.isTerminated {
		return
	}

	switch s := stmt.(type) {
	case *ast.Block:
		for _, inner := range s.Statements {
			if g.isTerminated {
				break
			}
			g.genStmt(inner)
		}

	case *ast.BindStmt:
		val := g.genExpr(s.Value)
		slot, exists := g.vars[s.Name]
		if !exists {
			var qtype *types.Type
			if s.Type != nil {
				qtype = g.resolveType(s.Type)
			} else {
				qtype = val.QType
			}
			llvmT := val.Type
			if qtype != nil {
				llvmT = g.llvmType(qtype)
			}
			ptr := fmt.Sprintf("%%%s.ptr", s.Name)
			fmt.Fprintf(&g.sb, "  %s = alloca %s\n", ptr, llvmT)
			slot = &varSlot{
				ptr:      ptr,
				llvmType: llvmT,
				qtype:    qtype,
			}
			g.vars[s.Name] = slot
		}

		if slot.llvmType == "i8*" && slot.qtype != nil && (slot.qtype.Kind == types.KindConfident || slot.qtype.Kind == types.KindExplore) && val.Type != "i8*" {
			dRef := val.Ref
			if val.Type == "i64" {
				conv := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = sitofp i64 %s to double\n", conv, val.Ref)
				dRef = conv
			}
			scoreTemp := g.nextTemp()
			score := 1.0
			state := int8(0)
			if slot.qtype.Kind == types.KindExplore {
				state = 2
				score = 0.5
			} else if slot.qtype.Confidence > 0.0 {
				score = slot.qtype.Confidence
			}
			fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_confidence_create(double %s, double %f, i8 %d)\n", scoreTemp, dRef, score, state)
			val = Value{Type: "i8*", Ref: scoreTemp, QType: slot.qtype}
		}

		fmt.Fprintf(&g.sb, "  store %s %s, %s* %s\n", val.Type, val.Ref, slot.llvmType, slot.ptr)

	case *ast.AssignStmt:
		val := g.genExpr(s.RHS)
		if ident, ok := s.LHS.(*ast.Ident); ok {
			slot, exists := g.vars[ident.Name]
			if exists {
				if s.Op == token.ASSIGN {
					fmt.Fprintf(&g.sb, "  store %s %s, %s* %s\n", val.Type, val.Ref, slot.llvmType, slot.ptr)
				} else {
					// Compound assignment
					currTemp := g.nextTemp()
					fmt.Fprintf(&g.sb, "  %s = load %s, %s* %s\n", currTemp, slot.llvmType, slot.llvmType, slot.ptr)
					resTemp := g.nextTemp()
					opName := "add"
					if slot.llvmType == "double" || slot.llvmType == "float" {
						opName = "fadd"
					}
					switch s.Op {
					case token.BIND_SUB:
						if slot.llvmType == "double" {
							opName = "fsub"
						} else {
							opName = "sub"
						}
					case token.BIND_MUL:
						if slot.llvmType == "double" {
							opName = "fmul"
						} else {
							opName = "mul"
						}
					case token.BIND_DIV:
						if slot.llvmType == "double" {
							opName = "fdiv"
						} else {
							opName = "sdiv"
						}
					}
					fmt.Fprintf(&g.sb, "  %s = %s %s %s, %s\n", resTemp, opName, slot.llvmType, currTemp, val.Ref)
					fmt.Fprintf(&g.sb, "  store %s %s, %s* %s\n", slot.llvmType, resTemp, slot.llvmType, slot.ptr)
				}
			}
		}

	case *ast.EmitStmt:
		val := g.genExpr(s.Value)
		g.genEmit(val)

	case *ast.ObserveStmt:
		val := g.genExpr(s.Value)
		g.genEmit(val)

	case *ast.YieldStmt:
		if g.currentProc != nil && g.currentProc.Name == "main" {
			if s.Value != nil {
				val := g.genExpr(s.Value)
				retTemp := g.nextTemp()
				switch val.Type {
				case "i64":
					fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", retTemp, val.Ref)
					fmt.Fprintf(&g.sb, "  ret i32 %s\n", retTemp)
				case "i32":
					fmt.Fprintf(&g.sb, "  ret i32 %s\n", val.Ref)
				default:
					g.sb.WriteString("  ret i32 0\n")
				}
			} else {
				g.sb.WriteString("  ret i32 0\n")
			}
		} else {
			if s.Value != nil {
				val := g.genExpr(s.Value)
				fmt.Fprintf(&g.sb, "  ret %s %s\n", val.Type, val.Ref)
			} else {
				g.sb.WriteString("  ret void\n")
			}
		}
		g.isTerminated = true

	case *ast.WhenStmt:
		cond := g.genExpr(s.Condition)
		thenLbl := g.nextLabel("when_then")
		elseLbl := g.nextLabel("when_else")
		mergeLbl := g.nextLabel("when_merge")

		targetElse := mergeLbl
		if s.Else != nil {
			targetElse = elseLbl
		}

		fmt.Fprintf(&g.sb, "  br i1 %s, label %%%s, label %%%s\n", cond.Ref, thenLbl, targetElse)

		// Then block
		fmt.Fprintf(&g.sb, "%s:\n", thenLbl)
		g.isTerminated = false
		g.genStmt(s.Then)
		if !g.isTerminated {
			fmt.Fprintf(&g.sb, "  br label %%%s\n", mergeLbl)
		}

		// Else block (if any)
		if s.Else != nil {
			fmt.Fprintf(&g.sb, "%s:\n", elseLbl)
			g.isTerminated = false
			g.genStmt(s.Else)
			if !g.isTerminated {
				fmt.Fprintf(&g.sb, "  br label %%%s\n", mergeLbl)
			}
		}

		fmt.Fprintf(&g.sb, "%s:\n", mergeLbl)
		g.isTerminated = false

	case *ast.CycleStmt:
		condLbl := g.nextLabel("cycle_cond")
		bodyLbl := g.nextLabel("cycle_body")
		endLbl := g.nextLabel("cycle_end")

		fmt.Fprintf(&g.sb, "  br label %%%s\n", condLbl)

		// Condition block
		fmt.Fprintf(&g.sb, "%s:\n", condLbl)
		g.isTerminated = false
		if s.Condition != nil {
			cond := g.genExpr(s.Condition)
			fmt.Fprintf(&g.sb, "  br i1 %s, label %%%s, label %%%s\n", cond.Ref, bodyLbl, endLbl)
		} else {
			fmt.Fprintf(&g.sb, "  br label %%%s\n", bodyLbl)
		}

		// Body block
		fmt.Fprintf(&g.sb, "%s:\n", bodyLbl)
		g.isTerminated = false
		g.genStmt(s.Body)
		if !g.isTerminated {
			fmt.Fprintf(&g.sb, "  br label %%%s\n", condLbl)
		}

		fmt.Fprintf(&g.sb, "%s:\n", endLbl)
		g.isTerminated = false

	case *ast.ExprStmt:
		g.genExpr(s.Expr)

	case *ast.LearnDecl:
		g.genLearnStmt(s)

	case *ast.SuperposeStmt:
		g.genSuperpose(s)
	case *ast.EntangleStmt:
		g.genEntangle(s)
	case *ast.CollapseStmt:
		g.genCollapseStmt(s)
	case *ast.QubitDecl:
		g.genQubitDecl(s)
	case *ast.DatasetDecl:
		g.genDatasetDecl(s)
	case *ast.NormalizeStmt:
		g.genNormalizeStmt(s)
	case *ast.ShuffleStmt:
		g.genShuffleStmt(s)
	case *ast.BatchStmt:
		g.genBatchStmt(s)
	case *ast.MinimizeStmt:
		g.genMinimizeStmt(s)
	case *ast.AssertStmt:
		g.genAssertStmt(s)
	case *ast.InferStmt:
		g.genInferStmt(s)
	case *ast.PropagateStmt:
		g.genPropagateStmt(s)
	case *ast.MemoryDecl:
		g.genMemoryDecl(s)
	case *ast.RememberStmt:
		g.genRememberStmt(s)
	case *ast.GraphDecl:
		g.genGraphDecl(s)
	case *ast.NodeStmt:
		g.genNodeStmt(s)
	case *ast.LinkStmt:
		g.genLinkStmt(s)
	case *ast.ExecuteStmt:
		g.genExecuteStmt(s)
	case *ast.EthicalBoundaryDecl:
		g.genEthicalBoundaryDecl(s)
	case *ast.EnforceStmt:
		g.genEnforceStmt(s)
	case *ast.DecoherenceStmt:
		g.genDecoherenceStmt(s)
	case *ast.VaultDecl:
		g.genVaultDecl(s)
	case *ast.SealStmt:
		g.genSealStmt(s)
	case *ast.UnsealStmt:
		g.genUnsealStmt(s)
	case *ast.WipeStmt:
		g.genWipeStmt(s)
	case *ast.ProvenanceDecl:
		g.genProvenanceDecl(s)
	case *ast.CertifyStmt:
		g.genCertifyStmt(s)
	case *ast.MemoryProvenanceDecl:
		g.genMemoryProvenanceDecl(s)
	case *ast.VerifyMemoryStmt:
		g.genVerifyMemoryStmt(s)
	case *ast.AuditMemoryStmt:
		g.genAuditMemoryStmt(s)
	case *ast.CorruptShareStmt:
		g.genCorruptShareStmt(s)
	case *ast.LicenseDecl:
		g.genLicenseDecl(s)
	case *ast.TransferStmt:
		g.genTransferStmt(s)
	case *ast.DistillStmt:
		g.genDistillStmt(s)
	case *ast.ReasoningChainDecl:
		g.genReasoningChainDecl(s)
	case *ast.ReasonStepStmt:
		g.genReasonStepStmt(s)
	case *ast.VerifyChainStmt:
		g.genVerifyChainStmt(s)
	case *ast.AuditReasoningStmt:
		g.genAuditReasoningStmt(s)
	case *ast.CorruptReasoningStepStmt:
		g.genCorruptReasoningStepStmt(s)
	case *ast.ModalityDecl:
		g.genModalityDecl(s)
	case *ast.CrossModalReasonStmt:
		g.genCrossModalReasonStmt(s)
	case *ast.QECMemoryDecl:
		g.genQECMemoryDecl(s)
	case *ast.SyndromeStmt:
		g.genSyndromeStmt(s)
	case *ast.VerifySyndromeStmt:
		g.genVerifySyndromeStmt(s)
	case *ast.CorrectStmt:
		g.genCorrectStmt(s)
	case *ast.CorruptSyndromeStmt:
		g.genCorruptSyndromeStmt(s)
	case *ast.ConsolidateStmt:
		g.genConsolidateStmt(s)
	case *ast.RecallTokenDecl:
		g.genRecallTokenDecl(s)
	case *ast.GrantStmt:
		g.genGrantStmt(s)
	case *ast.RevokeStmt:
		g.genRevokeStmt(s)
	case *ast.ToolRegisterDecl:
		g.genToolRegisterDecl(s)
	case *ast.ToolCallStmt:
		g.genToolCallStmt(s)
	case *ast.CertifiedReasonStmt:
		g.genCertifiedReasonStmt(s)
	case *ast.AbstainThresholdDecl:
		g.genAbstainThresholdDecl(s)
	case *ast.AbstainCheckStmt:
		g.genAbstainCheckStmt(s)
	case *ast.SandboxCreateStmt:
		g.genSandboxCreateStmt(s)
	case *ast.SandboxRunStmt:
		g.genSandboxRunStmt(s)
	case *ast.SandboxVerifyStmt:
		g.genSandboxVerifyStmt(s)
	case *ast.GPUMatmulStmt:
		g.genGPUMatmulStmt(s)
	case *ast.GPUSoftmaxStmt:
		g.genGPUSoftmaxStmt(s)
	case *ast.GPULayernormStmt:
		g.genGPULayernormStmt(s)
	case *ast.AGKEAttestStmt:
		g.genAGKEAttestStmt(s)
	case *ast.AGKEVerifyStmt:
		g.genAGKEVerifyStmt(s)
	case *ast.CorruptGPUDigestStmt:
		g.genCorruptGPUDigestStmt(s)
	case *ast.PipelineClaimStmt:
		g.genPipelineClaimStmt(s)
	case *ast.PipelineRunStmt:
		g.genPipelineRunStmt(s)
	case *ast.CDAGARecordStmt:
		g.genCDAGARecordStmt(s)
	case *ast.CDAGAAttestStmt:
		g.genCDAGAAttestStmt(s)
	case *ast.CDAGAVerifyStmt:
		g.genCDAGAVerifyStmt(s)
	case *ast.CorruptCDAGAStmt:
		g.genCorruptCDAGAStmt(s)
	case *ast.AllReduceStmt:
		g.genAllReduceStmt(s)
	case *ast.PserverInitStmt:
		g.genPserverInitStmt(s)
	case *ast.PserverAggregateStmt:
		g.genPserverAggregateStmt(s)
	case *ast.BaratSubmitStmt:
		g.genBaratSubmitStmt(s)
	case *ast.BaratAggregateStmt:
		g.genBaratAggregateStmt(s)
	case *ast.BaratVerifyStmt:
		g.genBaratVerifyStmt(s)
	case *ast.CorruptBaratStmt:
		g.genCorruptBaratStmt(s)
	case *ast.QASMExportStmt:
		g.genQASMExportStmt(s)
	case *ast.EvolveStmt:
		g.genEvolveStmt(s)
	case *ast.EvolveVerifyStmt:
		g.genEvolveVerifyStmt(s)
	case *ast.EvolveHistoryStmt:
		g.genEvolveHistoryStmt(s)
	case *ast.CorruptEvolveStmt:
		g.genCorruptEvolveStmt(s)
	case *ast.TranspileToStmt:
		g.genTranspileToStmt(s)
	case *ast.EvolvePolicyStmt:
		g.genEvolvePolicyStmt(s)
	case *ast.EvolveBoundedStmt:
		g.genEvolveBoundedStmt(s)
	case *ast.EEAVerifyStmt:
		g.genEEAVerifyStmt(s)
	case *ast.CorruptEEAStmt:
		g.genCorruptEEAStmt(s)
	case *ast.QECDDecodeStmt:
		g.genQECDDecodeStmt(s)
	case *ast.QECDVerifyStmt:
		g.genQECDVerifyStmt(s)
	case *ast.QECDApplyStmt:
		g.genQECDApplyStmt(s)
	case *ast.AADCertStmt:
		g.genAADCertStmt(s)
	case *ast.AADAttestStmt:
		g.genAADAttestStmt(s)
	case *ast.MetaFitStmt:
		g.genMetaFitStmt(s)
	case *ast.MetaAdaptStmt:
		g.genMetaAdaptStmt(s)
	case *ast.MetaVerifyStmt:
		g.genMetaVerifyStmt(s)
	case *ast.CorruptAADEnvStmt:
		g.genCorruptAADEnvStmt(s)
	case *ast.RecordDeclStmt, *ast.VariantDeclStmt:
		// Metadata declarations
	case *ast.MatchStmt:
		g.genMatchStmt(s)
	case *ast.RSCAAttestStmt:
		g.genRSCAAttestStmt(s)
	case *ast.RSCAVerifyStmt:
		g.genRSCAVerifyStmt(s)
	case *ast.CorruptRSCAStmt:
		g.genCorruptRSCAStmt(s)
	case *ast.CTSAAttestStmt:
		g.genCTSAAttestStmt(s)
	case *ast.CTSAVerifyStmt:
		g.genCTSAVerifyStmt(s)
	case *ast.CorruptCTSAStmt:
		g.genCorruptCTSAStmt(s)
	case *ast.QHALBackendAnnotation:
		// No-op
	case *ast.TileStmt, *ast.EnergyDecl:
		// No-op
	}
}

func (g *Generator) genStringConst(s string) string {
	globalName, length := g.addStringConstant(s)
	ptrTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [%d x i8], [%d x i8]* %s, i32 0, i32 0\n",
		ptrTemp, length, length, globalName)
	return ptrTemp
}

func (g *Generator) genEmit(val Value) {
	if val.QType != nil && val.QType.IsQuantum() {
		fmt.Fprintf(&g.sb, "  call void @qadav_emit_qubit(i8* null)\n")
		return
	}
	if val.QType != nil && val.QType.Kind == types.KindConcept {
		g.genEmitConcept(val)
		return
	}
	if val.QType != nil && val.QType.Kind == types.KindDataset {
		fmt.Fprintf(&g.sb, "  call void @qadav_dataset_print(i8* %s)\n", val.Ref)
		return
	}
	if val.QType != nil && val.QType.Kind == types.KindQuasi {
		fmt.Fprintf(&g.sb, "  call void @qadav_quasi_print(i64 %s)\n", val.Ref)
		return
	}
	if val.QType != nil && (val.QType.Kind == types.KindConfident || val.QType.Kind == types.KindExplore) {
		fmt.Fprintf(&g.sb, "  call void @qadav_confidence_print(i8* %s)\n", val.Ref)
		return
	}
	if val.QType != nil && val.QType.Kind == types.KindMemory {
		fmt.Fprintf(&g.sb, "  call void @qadav_memory_print(i8* %s)\n", val.Ref)
		return
	}
	if val.QType != nil && val.QType.Kind == types.KindGraph {
		fmt.Fprintf(&g.sb, "  call void @qadav_graph_print(i8* %s)\n", val.Ref)
		return
	}
	if val.QType != nil && val.QType.Kind == types.KindTensor {
		if val.QType.Elem != nil && val.QType.Elem.Kind == types.KindQuasi {
			fmt.Fprintf(&g.sb, "  call void @qadav_quasi_grid_print(i8* %s)\n", val.Ref)
			return
		}
		if val.QType.Elem != nil && val.QType.Elem.Kind == types.KindInt64 {
			fmt.Fprintf(&g.sb, "  call void @qadav_tensor_print_i64(i8* %s)\n", val.Ref)
		} else {
			fmt.Fprintf(&g.sb, "  call void @qadav_tensor_print_f64(i8* %s)\n", val.Ref)
		}
		return
	}

	switch val.Type {
	case "i8*":
		fmt.Fprintf(&g.sb, "  call void @qadav_emit_text(i8* %s)\n", val.Ref)
	case "i1":
		extTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = zext i1 %s to i32\n", extTemp, val.Ref)
		fmt.Fprintf(&g.sb, "  call void @qadav_emit_logic(i32 %s)\n", extTemp)
	case "double", "float":
		dRef := val.Ref
		if val.Type == "float" {
			extTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = fpext float %s to double\n", extTemp, val.Ref)
			dRef = extTemp
		}
		fmt.Fprintf(&g.sb, "  call void @qadav_emit_float(double %s)\n", dRef)
	default:
		// Integer types: extend to i64
		i64Ref := val.Ref
		if val.Type != "i64" {
			extTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = sext %s %s to i64\n", extTemp, val.Type, val.Ref)
			i64Ref = extTemp
		}
		fmt.Fprintf(&g.sb, "  call void @qadav_emit_int(i64 %s)\n", i64Ref)
	}
}

func flattenTensorLit(lit *ast.TensorLit) []ast.Expr {
	var out []ast.Expr
	for _, item := range lit.Items {
		if sub, ok := item.(*ast.TensorLit); ok {
			out = append(out, flattenTensorLit(sub)...)
		} else {
			out = append(out, item)
		}
	}
	return out
}

func (g *Generator) genExpr(expr ast.Expr) Value {
	if expr == nil {
		return Value{Type: "i64", Ref: "0"}
	}

	qtype := g.exprTypes[expr]

	switch e := expr.(type) {
	case *ast.TensorLit:
		elemKind := 10 // types.KindFloat64
		if qtype != nil && qtype.Elem != nil && (qtype.Elem.Kind == types.KindInt64 || qtype.Elem.Kind == types.KindQuasi) {
			elemKind = 4 // types.KindInt64
		}
		dims := e.Dims
		if len(dims) == 0 {
			dims = []int{len(e.Items)}
		}
		shapeGlobal := g.addShapeConstant(dims)
		dimsLen := len(dims)
		if dimsLen == 0 {
			dimsLen = 1
		}
		shapePtr := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [%d x i32], [%d x i32]* %s, i32 0, i32 0\n",
			shapePtr, dimsLen, dimsLen, shapeGlobal)
		tPtr := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_tensor_create(i32 %d, i32 %d, i32* %s)\n",
			tPtr, elemKind, len(dims), shapePtr)

		flat := flattenTensorLit(e)
		for i, item := range flat {
			itemVal := g.genExpr(item)
			if elemKind == 4 {
				iVal := itemVal.Ref
				if itemVal.Type != "i64" {
					extTemp := g.nextTemp()
					fmt.Fprintf(&g.sb, "  %s = sext %s %s to i64\n", extTemp, itemVal.Type, itemVal.Ref)
					iVal = extTemp
				}
				fmt.Fprintf(&g.sb, "  call void @qadav_tensor_set_i64(i8* %s, i32 %d, i64 %s)\n", tPtr, i, iVal)
			} else {
				fVal := itemVal.Ref
				if itemVal.Type != "double" {
					if itemVal.Type == "float" {
						extTemp := g.nextTemp()
						fmt.Fprintf(&g.sb, "  %s = fpext float %s to double\n", extTemp, itemVal.Ref)
						fVal = extTemp
					} else {
						extTemp := g.nextTemp()
						fmt.Fprintf(&g.sb, "  %s = sitofp %s %s to double\n", extTemp, itemVal.Type, itemVal.Ref)
						fVal = extTemp
					}
				}
				fmt.Fprintf(&g.sb, "  call void @qadav_tensor_set_f64(i8* %s, i32 %d, double %s)\n", tPtr, i, fVal)
			}
		}
		if qtype == nil {
			if elemKind == 4 {
				qtype = types.TensorOfShape(types.Int64, dims)
			} else {
				qtype = types.TensorOfShape(types.Float64, dims)
			}
		}
		return Value{Type: "i8*", Ref: tPtr, QType: qtype}

	case *ast.IntLit:
		return Value{
			Type:  "i64",
			Ref:   fmt.Sprintf("%d", e.Value),
			QType: types.Int64,
		}

	case *ast.FloatLit:
		return Value{
			Type:  "double",
			Ref:   fmt.Sprintf("%f", e.Value),
			QType: types.Float64,
		}

	case *ast.StringLit:
		globalName, length := g.addStringConstant(e.Value)
		ptrTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [%d x i8], [%d x i8]* %s, i32 0, i32 0\n",
			ptrTemp, length, length, globalName)
		return Value{
			Type:  "i8*",
			Ref:   ptrTemp,
			QType: types.Text,
		}

	case *ast.QuasiLit:
		return g.genQuasiLit(e)

	case *ast.Ident:
		switch e.Name {
		case "affirm", "⊤":
			return Value{Type: "i1", Ref: "1", QType: types.Logic}
		case "deny", "⊥":
			return Value{Type: "i1", Ref: "0", QType: types.Logic}
		}
		slot, ok := g.vars[e.Name]
		if ok {
			loadTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load %s, %s* %s\n", loadTemp, slot.llvmType, slot.llvmType, slot.ptr)
			return Value{
				Type:  slot.llvmType,
				Ref:   loadTemp,
				QType: slot.qtype,
			}
		}
		return Value{Type: "i64", Ref: "0", QType: qtype}

	case *ast.BinaryExpr:
		left := g.genExpr(e.Left)
		right := g.genExpr(e.Right)

		// Tensor operations
		if (left.QType != nil && left.QType.Kind == types.KindTensor) || (right.QType != nil && right.QType.Kind == types.KindTensor) {
			if e.Op == token.MATMUL {
				resTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_tensor_matmul(i8* %s, i8* %s)\n", resTemp, left.Ref, right.Ref)
				if qtype != nil && qtype.Kind != types.KindTensor {
					if qtype.Kind == types.KindInt64 {
						valTemp := g.nextTemp()
						fmt.Fprintf(&g.sb, "  %s = call i64 @qadav_tensor_get_i64(i8* %s, i32 0)\n", valTemp, resTemp)
						fmt.Fprintf(&g.sb, "  call void @qadav_tensor_free(i8* %s)\n", resTemp)
						return Value{Type: "i64", Ref: valTemp, QType: types.Int64}
					}
					valTemp := g.nextTemp()
					fmt.Fprintf(&g.sb, "  %s = call double @qadav_tensor_get_f64(i8* %s, i32 0)\n", valTemp, resTemp)
					fmt.Fprintf(&g.sb, "  call void @qadav_tensor_free(i8* %s)\n", resTemp)
					return Value{Type: "double", Ref: valTemp, QType: types.Float64}
				}
				return Value{Type: "i8*", Ref: resTemp, QType: qtype}
			}

			// Element-wise ops between two tensors
			if left.QType != nil && left.QType.Kind == types.KindTensor && right.QType != nil && right.QType.Kind == types.KindTensor {
				var fnName string
				switch e.Op {
				case token.OSUM:
					fnName = "@qadav_tensor_add"
				case token.ODIFF:
					fnName = "@qadav_tensor_sub"
				case token.OTIM:
					fnName = "@qadav_tensor_mul"
				case token.ODIV:
					fnName = "@qadav_tensor_div"
				}
				if fnName != "" {
					resTemp := g.nextTemp()
					fmt.Fprintf(&g.sb, "  %s = call i8* %s(i8* %s, i8* %s)\n", resTemp, fnName, left.Ref, right.Ref)
					return Value{Type: "i8*", Ref: resTemp, QType: qtype}
				}
			}

			// Broadcast: Tensor (left) with Scalar (right)
			if left.QType != nil && left.QType.Kind == types.KindTensor {
				isFloat := (left.QType.Elem != nil && left.QType.Elem.Kind == types.KindFloat64) || right.Type == "double" || right.Type == "float"
				if isFloat {
					sRef := right.Ref
					if right.Type != "double" {
						if right.Type == "float" {
							extTemp := g.nextTemp()
							fmt.Fprintf(&g.sb, "  %s = fpext float %s to double\n", extTemp, right.Ref)
							sRef = extTemp
						} else {
							extTemp := g.nextTemp()
							fmt.Fprintf(&g.sb, "  %s = sitofp %s %s to double\n", extTemp, right.Type, right.Ref)
							sRef = extTemp
						}
					}
					var fnName string
					switch e.Op {
					case token.OSUM:
						fnName = "@qadav_tensor_add_scalar_f64"
					case token.ODIFF:
						fnName = "@qadav_tensor_sub_scalar_f64"
					case token.OTIM:
						fnName = "@qadav_tensor_mul_scalar_f64"
					case token.ODIV:
						fnName = "@qadav_tensor_div_scalar_f64"
					}
					if fnName != "" {
						resTemp := g.nextTemp()
						fmt.Fprintf(&g.sb, "  %s = call i8* %s(i8* %s, double %s)\n", resTemp, fnName, left.Ref, sRef)
						return Value{Type: "i8*", Ref: resTemp, QType: qtype}
					}
				} else {
					sRef := right.Ref
					if right.Type != "i64" {
						extTemp := g.nextTemp()
						fmt.Fprintf(&g.sb, "  %s = sext %s %s to i64\n", extTemp, right.Type, right.Ref)
						sRef = extTemp
					}
					var fnName string
					switch e.Op {
					case token.OSUM:
						fnName = "@qadav_tensor_add_scalar_i64"
					case token.ODIFF:
						fnName = "@qadav_tensor_sub_scalar_i64"
					case token.OTIM:
						fnName = "@qadav_tensor_mul_scalar_i64"
					case token.ODIV:
						fnName = "@qadav_tensor_div_scalar_i64"
					}
					if fnName != "" {
						resTemp := g.nextTemp()
						fmt.Fprintf(&g.sb, "  %s = call i8* %s(i8* %s, i64 %s)\n", resTemp, fnName, left.Ref, sRef)
						return Value{Type: "i8*", Ref: resTemp, QType: qtype}
					}
				}
			}

			// Broadcast: Scalar (left) with Tensor (right) for commutative ops
			if right.QType != nil && right.QType.Kind == types.KindTensor {
				isFloat := (right.QType.Elem != nil && right.QType.Elem.Kind == types.KindFloat64) || left.Type == "double" || left.Type == "float"
				if isFloat {
					sRef := left.Ref
					if left.Type != "double" {
						if left.Type == "float" {
							extTemp := g.nextTemp()
							fmt.Fprintf(&g.sb, "  %s = fpext float %s to double\n", extTemp, left.Ref)
							sRef = extTemp
						} else {
							extTemp := g.nextTemp()
							fmt.Fprintf(&g.sb, "  %s = sitofp %s %s to double\n", extTemp, left.Type, left.Ref)
							sRef = extTemp
						}
					}
					var fnName string
					switch e.Op {
					case token.OSUM:
						fnName = "@qadav_tensor_add_scalar_f64"
					case token.OTIM:
						fnName = "@qadav_tensor_mul_scalar_f64"
					}
					if fnName != "" {
						resTemp := g.nextTemp()
						fmt.Fprintf(&g.sb, "  %s = call i8* %s(i8* %s, double %s)\n", resTemp, fnName, right.Ref, sRef)
						return Value{Type: "i8*", Ref: resTemp, QType: qtype}
					}
				} else {
					sRef := left.Ref
					if left.Type != "i64" {
						extTemp := g.nextTemp()
						fmt.Fprintf(&g.sb, "  %s = sext %s %s to i64\n", extTemp, left.Type, left.Ref)
						sRef = extTemp
					}
					var fnName string
					switch e.Op {
					case token.OSUM:
						fnName = "@qadav_tensor_add_scalar_i64"
					case token.OTIM:
						fnName = "@qadav_tensor_mul_scalar_i64"
					}
					if fnName != "" {
						resTemp := g.nextTemp()
						fmt.Fprintf(&g.sb, "  %s = call i8* %s(i8* %s, i64 %s)\n", resTemp, fnName, right.Ref, sRef)
						return Value{Type: "i8*", Ref: resTemp, QType: qtype}
					}
				}
			}
		}

		// Arithmetic
		switch e.Op {
		case token.OSUM:
			if left.QType != nil && left.QType.Kind == types.KindConfident {
				resTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_confidence_add(i8* %s, i8* %s)\n", resTemp, left.Ref, right.Ref)
				return Value{Type: "i8*", Ref: resTemp, QType: left.QType}
			}
			if (left.QType != nil && left.QType.Kind == types.KindQuasi) && (right.QType != nil && right.QType.Kind == types.KindQuasi) {
				resI32 := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_quasi_add(i64 %s, i64 %s)\n", resI32, left.Ref, right.Ref)
				resTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = sext i32 %s to i64\n", resTemp, resI32)
				return Value{Type: "i64", Ref: resTemp, QType: types.Quasi}
			}
			resTemp := g.nextTemp()
			if left.Type == "double" {
				fmt.Fprintf(&g.sb, "  %s = fadd double %s, %s\n", resTemp, left.Ref, right.Ref)
				return Value{Type: "double", Ref: resTemp, QType: types.Float64}
			}
			fmt.Fprintf(&g.sb, "  %s = add i64 %s, %s\n", resTemp, left.Ref, right.Ref)
			return Value{Type: "i64", Ref: resTemp, QType: types.Int64}

		case token.ODIFF:
			resTemp := g.nextTemp()
			if left.Type == "double" {
				fmt.Fprintf(&g.sb, "  %s = fsub double %s, %s\n", resTemp, left.Ref, right.Ref)
				return Value{Type: "double", Ref: resTemp, QType: types.Float64}
			}
			fmt.Fprintf(&g.sb, "  %s = sub i64 %s, %s\n", resTemp, left.Ref, right.Ref)
			return Value{Type: "i64", Ref: resTemp, QType: types.Int64}

		case token.OTIM:
			if left.QType != nil && left.QType.Kind == types.KindConfident {
				resTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_confidence_mul(i8* %s, i8* %s)\n", resTemp, left.Ref, right.Ref)
				return Value{Type: "i8*", Ref: resTemp, QType: left.QType}
			}
			if (left.QType != nil && left.QType.Kind == types.KindQuasi) && (right.QType != nil && right.QType.Kind == types.KindQuasi) {
				resI32 := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_quasi_mul(i64 %s, i64 %s)\n", resI32, left.Ref, right.Ref)
				resTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = sext i32 %s to i64\n", resTemp, resI32)
				return Value{Type: "i64", Ref: resTemp, QType: types.Quasi}
			}
			resTemp := g.nextTemp()
			if left.Type == "double" {
				fmt.Fprintf(&g.sb, "  %s = fmul double %s, %s\n", resTemp, left.Ref, right.Ref)
				return Value{Type: "double", Ref: resTemp, QType: types.Float64}
			}
			fmt.Fprintf(&g.sb, "  %s = mul i64 %s, %s\n", resTemp, left.Ref, right.Ref)
			outQType := types.Int64
			if (left.QType != nil && left.QType.Kind == types.KindConcept) || (right.QType != nil && right.QType.Kind == types.KindConcept) {
				outQType = types.Concept
			}
			return Value{Type: "i64", Ref: resTemp, QType: outQType}

		case token.ODIV:
			resTemp := g.nextTemp()
			if left.Type == "double" {
				fmt.Fprintf(&g.sb, "  %s = fdiv double %s, %s\n", resTemp, left.Ref, right.Ref)
				return Value{Type: "double", Ref: resTemp, QType: types.Float64}
			}
			fmt.Fprintf(&g.sb, "  %s = sdiv i64 %s, %s\n", resTemp, left.Ref, right.Ref)
			return Value{Type: "i64", Ref: resTemp, QType: types.Int64}

		case token.OMOD:
			resTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = srem i64 %s, %s\n", resTemp, left.Ref, right.Ref)
			return Value{Type: "i64", Ref: resTemp, QType: types.Int64}

		// Comparisons
		case token.EQUIV, token.NEQ, token.PREC, token.SUCC, token.PREQ, token.SUEQ:
			resTemp := g.nextTemp()
			if left.Type == "double" {
				pred := "oeq"
				switch e.Op {
				case token.NEQ:
					pred = "one"
				case token.PREC:
					pred = "olt"
				case token.SUCC:
					pred = "ogt"
				case token.PREQ:
					pred = "ole"
				case token.SUEQ:
					pred = "oge"
				}
				fmt.Fprintf(&g.sb, "  %s = fcmp %s double %s, %s\n", resTemp, pred, left.Ref, right.Ref)
			} else {
				pred := "eq"
				switch e.Op {
				case token.NEQ:
					pred = "ne"
				case token.PREC:
					pred = "slt"
				case token.SUCC:
					pred = "sgt"
				case token.PREQ:
					pred = "sle"
				case token.SUEQ:
					pred = "sge"
				}
				fmt.Fprintf(&g.sb, "  %s = icmp %s %s %s, %s\n", resTemp, pred, left.Type, left.Ref, right.Ref)
			}
			return Value{Type: "i1", Ref: resTemp, QType: types.Logic}

		// Logical
		case token.WEDGE:
			resTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = and i1 %s, %s\n", resTemp, left.Ref, right.Ref)
			return Value{Type: "i1", Ref: resTemp, QType: types.Logic}

		case token.VEE:
			resTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = or i1 %s, %s\n", resTemp, left.Ref, right.Ref)
			return Value{Type: "i1", Ref: resTemp, QType: types.Logic}

		case token.XVEE:
			resTemp := g.nextTemp()
			if left.Type == "i1" && right.Type == "i1" {
				fmt.Fprintf(&g.sb, "  %s = xor i1 %s, %s\n", resTemp, left.Ref, right.Ref)
				return Value{Type: "i1", Ref: resTemp, QType: types.Logic}
			}
			fmt.Fprintf(&g.sb, "  %s = xor i64 %s, %s\n", resTemp, left.Ref, right.Ref)
			return Value{Type: "i64", Ref: resTemp, QType: types.Int64}
		}

	case *ast.UnaryExpr:
		val := g.genExpr(e.Right)
		if val.QType != nil && val.QType.Kind == types.KindTensor {
			switch e.Op {
			case token.SUM:
				if val.QType.Elem != nil && val.QType.Elem.Kind == types.KindInt64 {
					resTemp := g.nextTemp()
					fmt.Fprintf(&g.sb, "  %s = call i64 @qadav_tensor_reduce_sum_i64(i8* %s)\n", resTemp, val.Ref)
					return Value{Type: "i64", Ref: resTemp, QType: types.Int64}
				}
				resTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call double @qadav_tensor_reduce_sum_f64(i8* %s)\n", resTemp, val.Ref)
				return Value{Type: "double", Ref: resTemp, QType: types.Float64}

			case token.PRODUCT:
				if val.QType.Elem != nil && val.QType.Elem.Kind == types.KindInt64 {
					resTemp := g.nextTemp()
					fmt.Fprintf(&g.sb, "  %s = call i64 @qadav_tensor_reduce_prod_i64(i8* %s)\n", resTemp, val.Ref)
					return Value{Type: "i64", Ref: resTemp, QType: types.Int64}
				}
				resTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call double @qadav_tensor_reduce_prod_f64(i8* %s)\n", resTemp, val.Ref)
				return Value{Type: "double", Ref: resTemp, QType: types.Float64}

			case token.ADJOINT:
				resTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_tensor_transpose(i8* %s)\n", resTemp, val.Ref)
				return Value{Type: "i8*", Ref: resTemp, QType: qtype}

			case token.ODIFF:
				if val.QType.Elem != nil && val.QType.Elem.Kind == types.KindInt64 {
					resTemp := g.nextTemp()
					fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_tensor_mul_scalar_i64(i8* %s, i64 -1)\n", resTemp, val.Ref)
					return Value{Type: "i8*", Ref: resTemp, QType: qtype}
				}
				resTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_tensor_mul_scalar_f64(i8* %s, double -1.0)\n", resTemp, val.Ref)
				return Value{Type: "i8*", Ref: resTemp, QType: qtype}
			}
		}

		switch e.Op {
		case token.NEG:
			resTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = xor i1 %s, true\n", resTemp, val.Ref)
			return Value{Type: "i1", Ref: resTemp, QType: types.Logic}
		case token.ODIFF:
			resTemp := g.nextTemp()
			if val.Type == "double" {
				fmt.Fprintf(&g.sb, "  %s = fneg double %s\n", resTemp, val.Ref)
				return Value{Type: "double", Ref: resTemp, QType: types.Float64}
			}
			fmt.Fprintf(&g.sb, "  %s = sub %s 0, %s\n", resTemp, val.Type, val.Ref)
			return Value{Type: val.Type, Ref: resTemp, QType: val.QType}
		}

	case *ast.GroupExpr:
		return g.genExpr(e.Inner)

	case *ast.CollapseExpr:
		return g.genCollapseExpr(e)

	case *ast.ConceptLit:
		return g.genConceptLit(e)

	case *ast.DecomposeExpr:
		return g.genDecomposeExpr(e)

	case *ast.IndexExpr:
		objVal := g.genExpr(e.Object)
		idxVal := g.genExpr(e.Index)
		if objVal.QType != nil && objVal.QType.Kind == types.KindMemoryProv {
			resTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_qemp_recall_text(i8* %s, i8* %s)\n", resTemp, objVal.Ref, idxVal.Ref)
			return Value{Type: "i8*", Ref: resTemp, QType: types.Text}
		}
		if objVal.QType != nil && objVal.QType.Kind == types.KindMemory {
			resTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call i64 @qadav_memory_recall(i8* %s, i8* %s)\n", resTemp, objVal.Ref, idxVal.Ref)
			return Value{Type: "i64", Ref: resTemp, QType: types.Int64}
		}

	case *ast.ConfidentLit:
		return g.genConfidentLit(e)

	case *ast.RecallExpr:
		return g.genRecallExpr(e)

	case *ast.TraceExpr:
		return g.genTraceExpr(e)

	case *ast.ExecuteExpr:
		return g.genExecuteExpr(e)

	case *ast.VerifyTrainingExpr:
		return g.genVerifyTrainingExpr(e)

	case *ast.ScopedRecallExpr:
		return g.genScopedRecallExpr(e)

	case *ast.ToolCallExpr:
		return g.genToolCallExpr(e)

	case *ast.AbstainCheckExpr:
		return g.genAbstainCheckExpr(e)
	case *ast.SandboxVerifyExpr:
		return g.genSandboxVerifyExpr(e)
	case *ast.GPUMatmulExpr:
		return g.genGPUMatmulExpr(e)
	case *ast.GPUSoftmaxExpr:
		return g.genGPUSoftmaxExpr(e)
	case *ast.GPULayernormExpr:
		return g.genGPULayernormExpr(e)
	case *ast.AGKEVerifyExpr:
		return g.genAGKEVerifyExpr(e)
	case *ast.RecordInitExpr:
		return g.genRecordInitExpr(e)
	case *ast.VariantInitExpr:
		return g.genVariantInitExpr(e)
	case *ast.FieldAccessExpr:
		return g.genFieldAccessExpr(e)

	case *ast.CallExpr:
		if ident, ok := e.Func.(*ast.Ident); ok {
			if g.checker != nil && g.checker.Networks() != nil {
				if netInfo, ok := g.checker.Networks()[ident.Name]; ok && len(e.Args) == 1 {
					return g.genNetworkForward(netInfo, ident.Name, e.Args[0])
				}
			}
			var argsBuilder strings.Builder
			for i, arg := range e.Args {
				if i > 0 {
					argsBuilder.WriteString(", ")
				}
				argVal := g.genExpr(arg)
				fmt.Fprintf(&argsBuilder, "%s %s", argVal.Type, argVal.Ref)
			}
			resLLVMType := "i64"
			if qtype != nil {
				resLLVMType = g.llvmType(qtype)
			}
			if resLLVMType == "void" {
				fmt.Fprintf(&g.sb, "  call void @\"%s\"(%s)\n", ident.Name, argsBuilder.String())
				return Value{Type: "void", Ref: "", QType: types.Void}
			}
			resTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call %s @\"%s\"(%s)\n", resTemp, resLLVMType, ident.Name, argsBuilder.String())
			return Value{Type: resLLVMType, Ref: resTemp, QType: qtype}
		}
	}

	return Value{Type: "i64", Ref: "0", QType: types.Int64}
}

func (g *Generator) resolveType(t ast.Type) *types.Type {
	if t == nil {
		return types.Void
	}
	switch node := t.(type) {
	case *ast.PrimitiveType:
		if node.Name == "concept" {
			return types.Concept
		}
		if node.Name == "quasi" {
			return types.Quasi
		}
		if node.Name == "dataset" {
			return types.Dataset
		}
		pt := types.FromPrimitiveName(node.Name)
		if pt != nil {
			return pt
		}
		return types.Void
	case *ast.TensorType:
		elem := g.resolveType(node.Elem)
		return types.TensorOfShape(elem, node.Dims)
	case *ast.NamedType:
		if pt := types.FromPrimitiveName(node.Name); pt != nil {
			return pt
		}
		if g.checker != nil && g.checker.RecordDecls() != nil {
			if _, ok := g.checker.RecordDecls()[node.Name]; ok {
				return &types.Type{Kind: types.KindRecord, Name: node.Name}
			}
		}
		if g.checker != nil && g.checker.VariantDecls() != nil {
			if _, ok := g.checker.VariantDecls()[node.Name]; ok {
				return &types.Type{Kind: types.KindVariant, Name: node.Name}
			}
		}
		if strings.HasPrefix(node.Name, "confident⟨") {
			return types.ConfidentOf(types.Float64, 0.95)
		}
		if strings.HasPrefix(node.Name, "explore⟨") {
			return types.ExploreOf(types.Float64)
		}
		return types.Void
	default:
		return types.Void
	}
}

func (g *Generator) llvmType(t *types.Type) string {
	if t == nil {
		return "void"
	}
	switch t.Kind {
	case types.KindInt8, types.KindUint8:
		return "i8"
	case types.KindInt16, types.KindUint16:
		return "i16"
	case types.KindInt32, types.KindUint32:
		return "i32"
	case types.KindInt64, types.KindUint64:
		return "i64"
	case types.KindFloat32:
		return "float"
	case types.KindFloat64:
		return "double"
	case types.KindLogic:
		return "i1"
	case types.KindText, types.KindTensor, types.KindQubitState, types.KindDataset, types.KindMemory, types.KindGraph, types.KindProcedure, types.KindConfident, types.KindExplore, types.KindRecord, types.KindVariant:
		return "i8*"
	case types.KindQubit:
		return "i32" // qubit ID
	case types.KindConcept, types.KindQuasi:
		return "i64" // concept / quasi encoding
	case types.KindVoid:
		return "void"
	default:
		return "i64"
	}
}

// Compile writes the generated LLVM IR to a file and invokes Clang to build an executable,
// linking against the Qadav minimal native runtime.
func (g *Generator) Compile(ctx context.Context, ir string, outputPath string) error {
	tc, err := llvm.Detect()
	if err != nil {
		return fmt.Errorf("detecting llvm toolchain: %w", err)
	}

	dir := filepath.Dir(outputPath)
	base := strings.TrimSuffix(filepath.Base(outputPath), filepath.Ext(outputPath))
	llPath := filepath.Join(dir, base+".ll")

	if err := os.WriteFile(llPath, []byte(ir), 0644); err != nil {
		return fmt.Errorf("writing ll file: %w", err)
	}

	rtObjs, err := runtime.Build(tc, "")
	if err != nil {
		return fmt.Errorf("building qadav runtime: %w", err)
	}

	// Compile IR to binary using Clang with runtime
	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], outputPath, rtObjs[1:]...); err != nil {
		return fmt.Errorf("compiling ir with runtime: %w", err)
	}

	return nil
}
