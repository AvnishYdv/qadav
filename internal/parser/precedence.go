package parser

import "github.com/qadav-lang/qadav/internal/token"

// Precedence levels (higher = binds tighter).
// Order chosen for quantum algebra: ⊗ > ⊕ like * > + in arithmetic.
const (
	_ int = iota
	LOWEST
	RANGE_PREC   // …
	PIPELINE     // ↝
	OR_PREC      // ∨ ⊻
	AND_PREC     // ∧
	EQUALS       // ≡ ≠ ≈
	LESSGREATER  // ≺ ≻ ≼ ≽ ∈ ∉
	SUM_PREC     // ⊕ ⊖
	PRODUCT_PREC // ⊗ ⊘ ⊙
	POWER_PREC   // ↑
	PREFIX       // ¬ ⊖ ⋆
	CALL_PREC    // ⦅⦆
	INDEX_PREC   // ⟬⟭
	MEMBER_PREC  // ∸
)

var precedences = map[token.Type]int{
	token.OSUM:          SUM_PREC,
	token.ODIFF:         SUM_PREC,
	token.OTIM:          PRODUCT_PREC,
	token.ODIV:          PRODUCT_PREC,
	token.OMOD:          PRODUCT_PREC,
	token.MATMUL:        PRODUCT_PREC,
	token.OPOW:          POWER_PREC,
	token.EQUIV:         EQUALS,
	token.NEQ:           EQUALS,
	token.APPROX:        EQUALS,
	token.PREC:          LESSGREATER,
	token.SUCC:          LESSGREATER,
	token.PREQ:          LESSGREATER,
	token.SUEQ:          LESSGREATER,
	token.MEMBER:        LESSGREATER,
	token.NOTMEMBER:     LESSGREATER,
	token.WEDGE:         AND_PREC,
	token.VEE:           OR_PREC,
	token.XVEE:          OR_PREC,
	token.PIPELINE:      PIPELINE,
	token.GROUP_OPEN:    CALL_PREC,
	token.INDEX_OPEN:    INDEX_PREC,
	token.MEMBER_ACCESS: MEMBER_PREC,
	token.RANGE:         RANGE_PREC,
	token.ANNOT:         LESSGREATER,
}

func precedenceOf(t token.Type) int {
	if p, ok := precedences[t]; ok {
		return p
	}
	return LOWEST
}
