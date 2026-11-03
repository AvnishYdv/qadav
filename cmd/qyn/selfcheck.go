package main

import (
	"fmt"
	"os"

	"github.com/qadav-lang/qadav/internal/selfcheck"
)

func runSelfCheck(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn selfcheck <build|verify|strip> [arguments]")
		return 1
	}

	sub := args[0]
	switch sub {
	case "build":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: qyn selfcheck build <file.ydv> [-o <output>]")
			return 1
		}
		buildArgs := append([]string{args[1], "--selfcheck"}, args[2:]...)
		return runBuild(buildArgs)

	case "verify":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: qyn selfcheck verify <binary.exe>")
			return 1
		}
		binPath := args[1]
		valid, root, err := selfcheck.Verify(binPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "qyn selfcheck verify: %v\n", err)
			return 1
		}
		if !valid {
			fmt.Fprintf(os.Stderr, "INTEGRITY TAMPERED: Merkle root mismatch or invalid Lamport signature!\n")
			return 1
		}
		fmt.Printf("INTEGRITY OK: Merkle root %x verified via Lamport post-quantum signature\n", root)
		return 0

	case "strip":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: qyn selfcheck strip <binary.exe>")
			return 1
		}
		binPath := args[1]
		if err := selfcheck.Strip(binPath); err != nil {
			fmt.Fprintf(os.Stderr, "qyn selfcheck strip: %v\n", err)
			return 1
		}
		fmt.Printf("Stripped self-check sidecar for %s\n", binPath)
		return 0

	default:
		fmt.Fprintf(os.Stderr, "qyn selfcheck: unknown subcommand %q\n", sub)
		return 1
	}
}
