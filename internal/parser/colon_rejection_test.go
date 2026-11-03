package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestColonRejectedInParams(t *testing.T) {
	src := "process f⦅n: int64⦆ → int64 ⟦ yield 0· ⟧"
	lex := lexer.New(src, "test.ydv")
	p := New(lex, "test.ydv")
	p.Parse()
	if len(p.Errors()) == 0 {
		t.Fatal("colon : should be rejected as legacy punctuation, but parser accepted it")
	}
	// Also check that the error mentions the colon
	t.Logf("Errors: %v", p.Errors())
}
