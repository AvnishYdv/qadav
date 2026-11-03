package lexer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/selfhost"
	"github.com/qadav-lang/qadav/internal/token"
)

func TestSelfHostedLexerByteIdentical(t *testing.T) {
	testDir := filepath.Join("..", "..", "selfhost", "tests")
	files, err := filepath.Glob(filepath.Join(testDir, "*.ydv"))
	if err != nil || len(files) == 0 {
		t.Fatalf("failed to find selfhost test files: %v", err)
	}

	for _, file := range files {
		srcBytes, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		src := string(srcBytes)

		// 1. Reference Go Lexer
		l := lexer.New(src, file)
		var goTokens []string
		for {
			tok := l.NextToken()
			goTokens = append(goTokens, tok.String())
			if tok.Type == token.EOF {
				break
			}
		}

		// 2. Self-Hosted Lexer Engine
		shTokens, err := selfhost.LexSource(src, file)
		if err != nil {
			t.Fatalf("selfhost.LexSource failed for %s: %v", file, err)
		}

		if len(goTokens) != len(shTokens) {
			t.Fatalf("token count mismatch for %s: Go=%d, SelfHost=%d", file, len(goTokens), len(shTokens))
		}

		for i := range goTokens {
			tok := shTokens[i]
			if tok.Type == "" || tok.Line <= 0 || tok.Col <= 0 {
				t.Fatalf("invalid token record: %+v", tok)
			}
		}
	}
}

func TestLexerUnicodeRoundtrip(t *testing.T) {
	cases := []struct {
		unicode  string
		fallback string
		expected token.Type
	}{
		{"⊕", `\oplus`, token.OSUM},
		{"⊖", `\ominus`, token.ODIFF},
		{"⊗", `\otimes`, token.OTIM},
		{"⊘", `\oslash`, token.ODIV},
		{"↦", `\bind`, token.BIND},
		{"←", `\assign`, token.ASSIGN},
		{"→", `\yields`, token.YIELDS},
		{"≡", `\equiv`, token.EQUIV},
		{"≠", `\neq`, token.NEQ},
	}

	for _, c := range cases {
		lUni := lexer.New(c.unicode, "roundtrip.ydv")
		tUni := lUni.NextToken()
		if tUni.Type != c.expected {
			t.Fatalf("unicode %s expected %v, got %v", c.unicode, c.expected, tUni.Type)
		}

		lFb := lexer.New(c.fallback, "roundtrip.ydv")
		tFb := lFb.NextToken()
		if tFb.Type != c.expected {
			t.Fatalf("fallback %s expected %v, got %v", c.fallback, c.expected, tFb.Type)
		}
	}
}

func TestLexerErrorRecovery(t *testing.T) {
	inputs := []string{
		"||",
		"@unknown_annotation",
		"@123",
	}

	for _, inp := range inputs {
		l := lexer.New(inp, "error.ydv")
		tok := l.NextToken()
		if tok.Type != token.ILLEGAL {
			t.Fatalf("input %q expected ILLEGAL, got %v", inp, tok.Type)
		}
	}
}
