package lexer

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/token"
)

func TestNextToken_UnicodeOperators(t *testing.T) {
	for _, op := range token.AllOperators() {
		lex := New(op.Unicode, "test.ydv")
		tok := lex.NextToken()
		if tok.Type != op.Type {
			t.Fatalf("operator %s expected Type %v, got %v", op.Unicode, op.Type, tok.Type)
		}
		if tok.Literal != op.Unicode {
			t.Fatalf("operator %s expected Literal %q, got %q", op.Unicode, op.Unicode, tok.Literal)
		}
		eof := lex.NextToken()
		if eof.Type != token.EOF {
			t.Fatalf("operator %s expected EOF next, got %v", op.Unicode, eof.Type)
		}
	}
}

func TestNextToken_LaTeXFallbacks(t *testing.T) {
	for _, op := range token.AllOperators() {
		lex := New(op.Fallback, "test.ydv")
		tok := lex.NextToken()
		if tok.Type != op.Type {
			t.Fatalf("fallback %s expected Type %v, got %v", op.Fallback, op.Type, tok.Type)
		}
		if tok.Literal != op.Fallback {
			t.Fatalf("fallback %s expected Literal %q, got %q", op.Fallback, op.Fallback, tok.Literal)
		}
		eof := lex.NextToken()
		if eof.Type != token.EOF {
			t.Fatalf("fallback %s expected EOF next, got %v", op.Fallback, eof.Type)
		}
	}
}

func TestFallbackAndUnicodeEquivalent(t *testing.T) {
	unicodeInput := "1 ⊕ 2 ⊗ 3"
	fallbackInput := `1 \oplus 2 \otimes 3`

	lUni := New(unicodeInput, "test.ydv")
	lFb := New(fallbackInput, "test.ydv")

	for {
		tUni := lUni.NextToken()
		tFb := lFb.NextToken()

		if tUni.Type != tFb.Type {
			t.Fatalf("stream mismatch: unicode got %v, fallback got %v", tUni.Type, tFb.Type)
		}

		if tUni.Type == token.EOF {
			break
		}
	}
}

func TestNextToken_Keywords(t *testing.T) {
	kws := token.AllKeywords()
	for kw, expectedType := range kws {
		lex := New(kw, "test.ydv")
		tok := lex.NextToken()
		if tok.Type != expectedType {
			t.Fatalf("keyword %q expected Type %v, got %v", kw, expectedType, tok.Type)
		}
		if tok.Literal != kw {
			t.Fatalf("keyword %q expected Literal %q, got %q", kw, kw, tok.Literal)
		}
	}
}

func TestNextToken_NoLegacyOperators(t *testing.T) {
	legacy := []string{
		"+", "-", "*", "/", "%",
		"==", "!=", "<", ">", "<=", ">=",
		"&&", "||", "!", "=",
		"->", "=>", "@",
	}

	for _, op := range legacy {
		lex := New(op, "test.ydv")
		tok := lex.NextToken()
		if tok.Type != token.ILLEGAL {
			t.Fatalf("legacy operator %q must be ILLEGAL, got %v (%q)", op, tok.Type, tok.Literal)
		}
	}
}

func TestNextToken_Integers(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"0xFF", "0xFF"},
		{"0b1010", "0b1010"},
		{"0o755", "0o755"},
		{"1_000_000", "1_000_000"},
	}

	for _, c := range cases {
		lex := New(c.input, "test.ydv")
		tok := lex.NextToken()
		if tok.Type != token.INT {
			t.Fatalf("expected INT for %q, got %v", c.input, tok.Type)
		}
		if tok.Literal != c.expected {
			t.Fatalf("expected literal %q, got %q", c.expected, tok.Literal)
		}
	}
}

func TestNextToken_Floats(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"3.14", "3.14"},
		{"1e10", "1e10"},
		{"2.5e-3", "2.5e-3"},
		{"1_000.5", "1_000.5"},
		{"3.14i", "3.14i"},
	}

	for _, c := range cases {
		lex := New(c.input, "test.ydv")
		tok := lex.NextToken()
		if tok.Type != token.FLOAT {
			t.Fatalf("expected FLOAT for %q, got %v", c.input, tok.Type)
		}
		if tok.Literal != c.expected {
			t.Fatalf("expected literal %q, got %q", c.expected, tok.Literal)
		}
	}
}

func TestNextToken_Strings(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello\nworld"`, `hello\nworld`},
		{`"\x41\u0915"`, `\x41\u0915`},
		{`"""multi\nline"""`, `multi\nline`},
	}

	for _, c := range cases {
		lex := New(c.input, "test.ydv")
		tok := lex.NextToken()
		if tok.Type != token.STRING {
			t.Fatalf("expected STRING for %s, got %v", c.input, tok.Type)
		}
		if tok.Literal != c.expected {
			t.Fatalf("expected literal %q, got %q", c.expected, tok.Literal)
		}
	}
}

func TestNextToken_Comments(t *testing.T) {
	input := `
	// यह एक सिंगल लाइन कमेंट है
	42
	/* यह एक मल्टी-लाइन कमेंट है
	   /* नेस्टेड ब्लॉक कमेंट */
	   जारी */
	100
	`
	lex := New(input, "test.ydv")
	tok1 := lex.NextToken()
	if tok1.Type != token.INT || tok1.Literal != "42" {
		t.Fatalf("expected INT 42 after line comment, got %v (%q)", tok1.Type, tok1.Literal)
	}

	tok2 := lex.NextToken()
	if tok2.Type != token.INT || tok2.Literal != "100" {
		t.Fatalf("expected INT 100 after nested block comment, got %v (%q)", tok2.Type, tok2.Literal)
	}

	tok3 := lex.NextToken()
	if tok3.Type != token.EOF {
		t.Fatalf("expected EOF, got %v", tok3.Type)
	}
}

func TestNextToken_LineColTracking(t *testing.T) {
	input := "10\n  ⊕\n    \\otimes"
	lex := New(input, "test.ydv")

	t1 := lex.NextToken()
	if t1.Line != 1 || t1.Col != 1 {
		t.Fatalf("t1 pos expected 1:1, got %d:%d", t1.Line, t1.Col)
	}

	t2 := lex.NextToken()
	if t2.Line != 2 || t2.Col != 3 {
		t.Fatalf("t2 pos expected 2:3, got %d:%d", t2.Line, t2.Col)
	}

	t3 := lex.NextToken()
	if t3.Line != 3 || t3.Col != 5 {
		t.Fatalf("t3 pos expected 3:5, got %d:%d", t3.Line, t3.Col)
	}
}

func TestNextToken_UTF8Strings(t *testing.T) {
	input := `"नमस्ते"`
	lex := New(input, "test.ydv")
	tok := lex.NextToken()
	if tok.Type != token.STRING || tok.Literal != "नमस्ते" {
		t.Fatalf("expected string with 'नमस्ते', got %v (%q)", tok.Type, tok.Literal)
	}
}

func TestNextToken_EOF(t *testing.T) {
	lex := New("", "empty.ydv")
	tok := lex.NextToken()
	if tok.Type != token.EOF {
		t.Fatalf("expected EOF for empty input, got %v", tok.Type)
	}
}

func TestFullProgram(t *testing.T) {
	prog := `process main⦅⦆ → int32 ⟦
    int32 x ↦ 1 ⊕ 2 ⊗ 3·
    emit x·
    yield 0·
⟧`

	expected := []token.Type{
		token.PROCESS,
		token.IDENT,
		token.GROUP_OPEN,
		token.GROUP_CLOSE,
		token.YIELDS,
		token.INT32,
		token.BLOCK_OPEN,
		token.INT32,
		token.IDENT,
		token.BIND,
		token.INT,
		token.OSUM,
		token.INT,
		token.OTIM,
		token.INT,
		token.STMT_END,
		token.EMIT,
		token.IDENT,
		token.STMT_END,
		token.YIELD_KW,
		token.INT,
		token.STMT_END,
		token.BLOCK_CLOSE,
		token.EOF,
	}

	lex := New(prog, "prog.ydv")
	for i, exp := range expected {
		tok := lex.NextToken()
		if tok.Type != exp {
			t.Fatalf("[%d] expected token type %v, got %v (literal: %q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestFullProgram_Fallback(t *testing.T) {
	prog := `process main\group\egroup \yields int32 \block
    int32 x \bind 1 \oplus 2 \otimes 3\stmt
    emit x\stmt
    yield 0\stmt
\eblock`

	expected := []token.Type{
		token.PROCESS,
		token.IDENT,
		token.GROUP_OPEN,
		token.GROUP_CLOSE,
		token.YIELDS,
		token.INT32,
		token.BLOCK_OPEN,
		token.INT32,
		token.IDENT,
		token.BIND,
		token.INT,
		token.OSUM,
		token.INT,
		token.OTIM,
		token.INT,
		token.STMT_END,
		token.EMIT,
		token.IDENT,
		token.STMT_END,
		token.YIELD_KW,
		token.INT,
		token.STMT_END,
		token.BLOCK_CLOSE,
		token.EOF,
	}

	lex := New(prog, "prog.ydv")
	for i, exp := range expected {
		tok := lex.NextToken()
		if tok.Type != exp {
			t.Fatalf("[%d] expected token type %v, got %v (literal: %q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestBannedOperatorsRejected(t *testing.T) {
	bannedOps := []string{"+", "-", "*", "/", "==", "!=", "<", ">", "&&", "||", "!", "="}
	for _, op := range bannedOps {
		lex := New(op, "test.ydv")
		tok := lex.NextToken()
		if tok.Type != token.ILLEGAL {
			t.Fatalf("expected ILLEGAL for banned operator %q, got %v", op, tok.Type)
		}
	}
}

func TestBannedKeywordsNotTokenized(t *testing.T) {
	bannedKws := []string{"if", "for", "while", "fn", "return", "let", "var", "mut", "print", "class"}
	for _, kw := range bannedKws {
		lex := New(kw, "test.ydv")
		tok := lex.NextToken()
		if tok.Type != token.IDENT {
			t.Fatalf("expected IDENT for legacy keyword %q, got %v", kw, tok.Type)
		}
	}
}

func TestAtAnnotationEdgeCases(t *testing.T) {
	l1 := New("@unknown", "test.ydv")
	tok1 := l1.NextToken()
	if tok1.Type != token.ILLEGAL || tok1.Literal != "@unknown" {
		t.Fatalf("expected ILLEGAL @unknown, got %+v", tok1)
	}

	l2 := New("@ ", "test.ydv")
	tok2 := l2.NextToken()
	if tok2.Type != token.ILLEGAL || tok2.Literal != "@" {
		t.Fatalf("expected ILLEGAL @, got %+v", tok2)
	}

	l3 := New(`\unknown`, "test.ydv")
	tok3 := l3.NextToken()
	if tok3.Type != token.UNKNOWN {
		t.Fatalf("expected UNKNOWN \\unknown, got %+v", tok3)
	}

	l4 := New("", "test.ydv")
	if l4.peek() != 0 {
		t.Fatalf("expected 0 for peek at EOF")
	}
	l4.advance()
}

func TestLexerCoverageEdgeCases(t *testing.T) {
	cases := []struct {
		input    string
		expected []token.Type
	}{
		{"@gpu @multi_device @distributed @qhal_backend", []token.Type{token.AT_GPU, token.MULTI_DEVICE, token.DIST_INIT, token.QHAL_BACKEND, token.EOF}},
		{"| ||", []token.Type{token.PIPE, token.ILLEGAL, token.EOF}},
		{". ( )", []token.Type{token.DOT, token.GROUP_OPEN, token.GROUP_CLOSE, token.EOF}},
		{"⟹ ⇒ \\implies", []token.Type{token.IMPLIES, token.IMPLIES, token.IMPLIES, token.EOF}},
		{"\\pipe \\record \\variant \\match \\otherwise", []token.Type{token.PIPE, token.RECORD, token.VARIANT, token.MATCH, token.OTHERWISE, token.EOF}},
		{"/* unterminated comment", []token.Type{token.EOF}},
		{"\"unterminated string", []token.Type{token.ILLEGAL, token.EOF}},
		{"\"valid \\n \\t \\r \\\\ \\\" string\"", []token.Type{token.STRING, token.EOF}},
		{"\"bad escape \\z \"", []token.Type{token.STRING, token.EOF}},
		{"0b101 0x1f 1.25e+2 3.4E-1", []token.Type{token.INT, token.INT, token.FLOAT, token.FLOAT, token.EOF}},
		{"\r\n\t  \r", []token.Type{token.EOF}},
		{"// line comment\n42", []token.Type{token.INT, token.EOF}},
		{"@123", []token.Type{token.ILLEGAL, token.INT, token.EOF}},
	}

	for _, tc := range cases {
		l := New(tc.input, "test.ydv")
		for _, exp := range tc.expected {
			tok := l.NextToken()
			if tok.Type != exp {
				t.Fatalf("for input %q expected %v, got %v (%s)", tc.input, exp, tok.Type, tok.Literal)
			}
		}
	}
}

func BenchmarkLexer(b *testing.B) {
	// Construct ~1KB sample of quantum-native Qadav code
	line := "int32 res ↦ a ⊕ b ⊗ c ⊘ d ⊙ e ↑ f·\n"
	sample := strings.Repeat(line, 25) // ~1000 bytes

	b.ReportAllocs()
	b.SetBytes(int64(len(sample)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l := New(sample, "bench.ydv")
		for {
			tok := l.NextToken()
			if tok.Type == token.EOF {
				break
			}
		}
	}
}
