package token

import (
	"strings"
	"testing"
)

func TestAllOperators(t *testing.T) {
	ops := AllOperators()
	if len(ops) == 0 {
		t.Fatal("no operators registered")
	}

	for _, op := range ops {
		if op.Type == "" {
			t.Fatalf("operator has empty Type: %+v", op)
		}
		if op.Unicode == "" {
			t.Fatalf("operator has empty Unicode: %+v", op)
		}
		if !strings.HasPrefix(op.Fallback, `\`) {
			t.Fatalf("operator fallback must start with backslash: %+v", op)
		}
		if op.Meaning == "" {
			t.Fatalf("operator has empty Meaning: %+v", op)
		}

		// Verify lookup maps match
		tokUni, okUni := LookupUnicode(op.Unicode)
		if !okUni || tokUni != op.Type {
			t.Fatalf("LookupUnicode(%s) = (%v, %v), expected %v", op.Unicode, tokUni, okUni, op.Type)
		}

		tokFb, okFb := LookupFallback(op.Fallback)
		if !okFb || tokFb != op.Type {
			t.Fatalf("LookupFallback(%s) = (%v, %v), expected %v", op.Fallback, tokFb, okFb, op.Type)
		}
	}
}

func TestAllKeywords(t *testing.T) {
	kws := AllKeywords()
	if len(kws) < 70 {
		t.Fatalf("expected at least 70 keywords, got %d", len(kws))
	}
	for kw, typ := range kws {
		if kw == "" || typ == "" {
			t.Fatalf("invalid keyword entry: %q -> %q", kw, typ)
		}
		if LookupIdent(kw) != typ {
			t.Fatalf("LookupIdent(%q) = %q, expected %q", kw, LookupIdent(kw), typ)
		}
	}
}

func TestNoLegacyOperators(t *testing.T) {
	legacyOps := []string{
		"+", "-", "*", "/", "%",
		"==", "!=", "<", ">", "<=", ">=",
		"&&", "||", "!", "=",
		"&", "|", "^", "~", "@",
		"->", "=>", "**",
		"+=", "-=", "*=", "/=", "%=",
	}

	for _, legacy := range legacyOps {
		if _, ok := LookupUnicode(legacy); ok {
			t.Fatalf("legacy operator %q must NOT be recognized as a Qadav Unicode operator", legacy)
		}
		if _, ok := LookupFallback(legacy); ok {
			t.Fatalf("legacy operator %q must NOT be recognized as a Qadav Fallback operator", legacy)
		}
	}
}

func TestNoLegacyKeywords(t *testing.T) {
	banned := []string{
		"if", "else", "elif", "for", "while", "fn", "def", "func", "return",
		"let", "mut", "var", "const", "class", "struct", "interface", "enum",
		"type", "import", "export", "from", "true", "false", "null", "nil",
		"None", "async", "await", "function",
		"int", "float", "bool", "string", "char", "double", "long", "short", "byte",
		"print", "println", "echo", "cout", "console",
		"public", "private", "protected", "static", "final", "abstract",
		"switch", "case", "default", "break", "continue", "goto",
		"try", "catch", "finally", "throw", "throws", "raise", "except",
		"and", "or", "not", "in", "is", "as", "del", "pass", "lambda", "yield_from",
		"package", "go", "chan", "defer", "select",
		"self", "this", "super", "new", "delete",
	}

	for _, b := range banned {
		tok := LookupIdent(b)
		if tok != IDENT {
			t.Fatalf("banned legacy keyword %q returned token type %v, must be IDENT", b, tok)
		}
		if _, found := keywords[b]; found {
			t.Fatalf("banned legacy keyword %q found in keywords map", b)
		}
	}
}

func TestNoLegacyPunctuation(t *testing.T) {
	legacyPunct := []string{"(", ")", "{", "}", "[", "]", ";", ",", ":", "."}
	for _, p := range legacyPunct {
		if _, ok := LookupUnicode(p); ok {
			t.Fatalf("legacy punctuation %q must not be registered in Unicode operators", p)
		}
	}
}

func TestLookupIdent_Keywords(t *testing.T) {
	samples := map[string]Type{
		"process":     PROCESS,
		"yield":       YIELD_KW,
		"flux":        FLUX,
		"when":        WHEN,
		"otherwise":   OTHERWISE,
		"iterate":     ITERATE,
		"cycle":       CYCLE,
		"halt":        HALT,
		"form":        FORM,
		"schema":      SCHEMA,
		"stream":      STREAM,
		"network":     NETWORK,
		"learn":       LEARN,
		"emit":        EMIT,
		"superpose":   SUPERPOSE,
		"entangle":    ENTANGLE,
		"sign":        SIGN,
		"verify":      VERIFY,
		"parallel":    PARALLEL,
		"transformer": TRANSFORMER,
		"attention":   ATTENTION,
		"logic":       LOGIC,
		"text":        TEXT,
	}

	for kw, expected := range samples {
		got := LookupIdent(kw)
		if got != expected {
			t.Fatalf("LookupIdent(%q) = %v, expected %v", kw, got, expected)
		}
	}
}

func TestLookupIdent_NonKeywords(t *testing.T) {
	nonKeywords := []string{"foo", "myvar", "Int32", "NETWORK", "Process", "x1", "_private", "counter"}
	for _, s := range nonKeywords {
		if got := LookupIdent(s); got != IDENT {
			t.Fatalf("LookupIdent(%q) = %v, expected IDENT", s, got)
		}
	}
}

func TestLookupFallback(t *testing.T) {
	checks := map[string]Type{
		`\oplus`:    OSUM,
		`\otimes`:   OTIM,
		`\equiv`:    EQUIV,
		`\bind`:     BIND,
		`\yields`:   YIELDS,
		`\group`:    GROUP_OPEN,
		`\egroup`:   GROUP_CLOSE,
		`\block`:    BLOCK_OPEN,
		`\eblock`:   BLOCK_CLOSE,
		`\bracket`:  INDEX_OPEN,
		`\ebracket`: INDEX_CLOSE,
		`\stmt`:     STMT_END,
		`\sep`:      SEP,
		`\because`:  ANNOT,
		`\dot`:      MEMBER_ACCESS,
		`\had`:      HADAMARD,
		`\cnot`:     CNOT,
	}

	for fb, expected := range checks {
		typ, ok := LookupFallback(fb)
		if !ok || typ != expected {
			t.Fatalf("LookupFallback(%s) = (%v, %v), expected %v", fb, typ, ok, expected)
		}
	}
}

func TestTokenString(t *testing.T) {
	tok := Token{
		Type:    OSUM,
		Literal: "⊕",
		Line:    3,
		Col:     12,
		File:    "test.ydv",
	}
	expected := "OSUM(⊕) @ test.ydv:3:12"
	if tok.String() != expected {
		t.Fatalf("Token.String() = %q, expected %q", tok.String(), expected)
	}
}
