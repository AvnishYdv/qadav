package runtime

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/qadav-lang/qadav/internal/llvm"
)

//go:embed qadav_rt.c qadav_rt_linux.c qadav_tensor.c qadav_tensor.h qadav_nn.c qadav_nn.h qadav_quantum.c qadav_quantum.h qadav_prime.c qadav_prime.h qadav_io.c qadav_io.h qadav_dataset.c qadav_dataset.h qadav_quasinary.c qadav_quasinary.h qadav_cir.c qadav_cir.h qadav_agi.c qadav_agi.h qadav_sha3.c qadav_sha3.h qadav_merkle.c qadav_merkle.h qadav_lamport.c qadav_lamport.h qadav_selfcheck.c qadav_selfcheck.h qadav_qeeh.c qadav_qeeh.h qadav_qemwv.c qadav_qemwv.h qadav_mctp.c qadav_mctp.h qadav_qemp.c qadav_qemp.h qadav_qecmt.c qadav_qecmt.h qadav_qera.c qadav_qera.h qadav_qecma.c qadav_qecma.h qadav_qec.c qadav_qec.h qadav_consolidate.c qadav_consolidate.h qadav_scoped.c qadav_scoped.h qadav_tools.c qadav_tools.h qadav_certificate.c qadav_certificate.h qadav_abstain.c qadav_abstain.h qadav_sandbox.c qadav_sandbox.h qadav_gpu.c qadav_gpu.h qadav_agke.c qadav_agke.h qadav_multigpu.c qadav_multigpu.h qadav_cdaga.c qadav_cdaga.h qadav_distributed.c qadav_distributed.h qadav_barat.c qadav_barat.h qadav_qhal.c qadav_qhal.h qadav_sea.c qadav_sea.h qadav_transpile.c qadav_transpile.h qadav_eea.c qadav_eea.h qadav_qecdecoder.c qadav_qecdecoder.h qadav_aad.c qadav_aad.h qadav_record.c qadav_record.h qadav_variant.c qadav_variant.h qadav_rsca.c qadav_rsca.h
var srcFiles embed.FS

// Sources returns the primary runtime C source file name and content for the current platform.
func Sources() (string, []byte, error) {
	return SourcesForOS(runtime.GOOS)
}

// SourcesForOS returns the runtime C source file name and content for a given OS.
func SourcesForOS(goos string) (string, []byte, error) {
	filename := "qadav_rt.c"
	if goos != "windows" {
		filename = "qadav_rt_linux.c"
	}
	content, err := srcFiles.ReadFile(filename)
	if err != nil {
		return "", nil, fmt.Errorf("reading embedded runtime source %s: %w", filename, err)
	}
	return filename, content, nil
}

// Hash returns a stable SHA-256 hex string of the current platform runtime, tensor, neural network, quantum, and prime sources.
func Hash() string {
	_, content, err := Sources()
	if err != nil {
		return "unknown"
	}
	tC, _ := srcFiles.ReadFile("qadav_tensor.c")
	tH, _ := srcFiles.ReadFile("qadav_tensor.h")
	nnC, _ := srcFiles.ReadFile("qadav_nn.c")
	nnH, _ := srcFiles.ReadFile("qadav_nn.h")
	qC, _ := srcFiles.ReadFile("qadav_quantum.c")
	qH, _ := srcFiles.ReadFile("qadav_quantum.h")
	pC, _ := srcFiles.ReadFile("qadav_prime.c")
	pH, _ := srcFiles.ReadFile("qadav_prime.h")
	ioC, _ := srcFiles.ReadFile("qadav_io.c")
	ioH, _ := srcFiles.ReadFile("qadav_io.h")
	dsC, _ := srcFiles.ReadFile("qadav_dataset.c")
	dsH, _ := srcFiles.ReadFile("qadav_dataset.h")
	qyC, _ := srcFiles.ReadFile("qadav_quasinary.c")
	qyH, _ := srcFiles.ReadFile("qadav_quasinary.h")
	cirC, _ := srcFiles.ReadFile("qadav_cir.c")
	cirH, _ := srcFiles.ReadFile("qadav_cir.h")
	agiC, _ := srcFiles.ReadFile("qadav_agi.c")
	agiH, _ := srcFiles.ReadFile("qadav_agi.h")
	sha3C, _ := srcFiles.ReadFile("qadav_sha3.c")
	sha3H, _ := srcFiles.ReadFile("qadav_sha3.h")
	merkleC, _ := srcFiles.ReadFile("qadav_merkle.c")
	merkleH, _ := srcFiles.ReadFile("qadav_merkle.h")
	lamportC, _ := srcFiles.ReadFile("qadav_lamport.c")
	lamportH, _ := srcFiles.ReadFile("qadav_lamport.h")
	selfcheckC, _ := srcFiles.ReadFile("qadav_selfcheck.c")
	selfcheckH, _ := srcFiles.ReadFile("qadav_selfcheck.h")

	qeehC, _ := srcFiles.ReadFile("qadav_qeeh.c")
	qeehH, _ := srcFiles.ReadFile("qadav_qeeh.h")
	qemwvC, _ := srcFiles.ReadFile("qadav_qemwv.c")
	qemwvH, _ := srcFiles.ReadFile("qadav_qemwv.h")
	mctpC, _ := srcFiles.ReadFile("qadav_mctp.c")
	mctpH, _ := srcFiles.ReadFile("qadav_mctp.h")
	qempC, _ := srcFiles.ReadFile("qadav_qemp.c")
	qempH, _ := srcFiles.ReadFile("qadav_qemp.h")
	qecmtC, _ := srcFiles.ReadFile("qadav_qecmt.c")
	qecmtH, _ := srcFiles.ReadFile("qadav_qecmt.h")
	qeraC, _ := srcFiles.ReadFile("qadav_qera.c")
	qeraH, _ := srcFiles.ReadFile("qadav_qera.h")
	qecmaC, _ := srcFiles.ReadFile("qadav_qecma.c")
	qecmaH, _ := srcFiles.ReadFile("qadav_qecma.h")
	qecC, _ := srcFiles.ReadFile("qadav_qec.c")
	qecH, _ := srcFiles.ReadFile("qadav_qec.h")
	consC, _ := srcFiles.ReadFile("qadav_consolidate.c")
	consH, _ := srcFiles.ReadFile("qadav_consolidate.h")
	scopedC, _ := srcFiles.ReadFile("qadav_scoped.c")
	scopedH, _ := srcFiles.ReadFile("qadav_scoped.h")
	toolsC, _ := srcFiles.ReadFile("qadav_tools.c")
	toolsH, _ := srcFiles.ReadFile("qadav_tools.h")
	certC, _ := srcFiles.ReadFile("qadav_certificate.c")
	certH, _ := srcFiles.ReadFile("qadav_certificate.h")
	abstainC, _ := srcFiles.ReadFile("qadav_abstain.c")
	abstainH, _ := srcFiles.ReadFile("qadav_abstain.h")
	sbC, _ := srcFiles.ReadFile("qadav_sandbox.c")
	sbH, _ := srcFiles.ReadFile("qadav_sandbox.h")
	gpuC, _ := srcFiles.ReadFile("qadav_gpu.c")
	gpuH, _ := srcFiles.ReadFile("qadav_gpu.h")
	agkeC, _ := srcFiles.ReadFile("qadav_agke.c")
	agkeH, _ := srcFiles.ReadFile("qadav_agke.h")
	mgpuC, _ := srcFiles.ReadFile("qadav_multigpu.c")
	mgpuH, _ := srcFiles.ReadFile("qadav_multigpu.h")
	cdagaC, _ := srcFiles.ReadFile("qadav_cdaga.c")
	cdagaH, _ := srcFiles.ReadFile("qadav_cdaga.h")
	distC, _ := srcFiles.ReadFile("qadav_distributed.c")
	distH, _ := srcFiles.ReadFile("qadav_distributed.h")
	baratC, _ := srcFiles.ReadFile("qadav_barat.c")
	baratH, _ := srcFiles.ReadFile("qadav_barat.h")
	qhalC, _ := srcFiles.ReadFile("qadav_qhal.c")
	qhalH, _ := srcFiles.ReadFile("qadav_qhal.h")
	seaC, _ := srcFiles.ReadFile("qadav_sea.c")
	seaH, _ := srcFiles.ReadFile("qadav_sea.h")
	transpileC, _ := srcFiles.ReadFile("qadav_transpile.c")
	transpileH, _ := srcFiles.ReadFile("qadav_transpile.h")
	eeaC, _ := srcFiles.ReadFile("qadav_eea.c")
	eeaH, _ := srcFiles.ReadFile("qadav_eea.h")
	qecdC, _ := srcFiles.ReadFile("qadav_qecdecoder.c")
	qecdH, _ := srcFiles.ReadFile("qadav_qecdecoder.h")
	aadC, _ := srcFiles.ReadFile("qadav_aad.c")
	aadH, _ := srcFiles.ReadFile("qadav_aad.h")
	recC, _ := srcFiles.ReadFile("qadav_record.c")
	recH, _ := srcFiles.ReadFile("qadav_record.h")
	varC, _ := srcFiles.ReadFile("qadav_variant.c")
	varH, _ := srcFiles.ReadFile("qadav_variant.h")
	rscaC, _ := srcFiles.ReadFile("qadav_rsca.c")
	rscaH, _ := srcFiles.ReadFile("qadav_rsca.h")

	h := sha256.New()
	h.Write(content)
	h.Write(tC)
	h.Write(tH)
	h.Write(nnC)
	h.Write(nnH)
	h.Write(qC)
	h.Write(qH)
	h.Write(pC)
	h.Write(pH)
	h.Write(ioC)
	h.Write(ioH)
	h.Write(dsC)
	h.Write(dsH)
	h.Write(qyC)
	h.Write(qyH)
	h.Write(cirC)
	h.Write(cirH)
	h.Write(agiC)
	h.Write(agiH)
	h.Write(sha3C)
	h.Write(sha3H)
	h.Write(merkleC)
	h.Write(merkleH)
	h.Write(lamportC)
	h.Write(lamportH)
	h.Write(selfcheckC)
	h.Write(selfcheckH)
	h.Write(qeehC)
	h.Write(qeehH)
	h.Write(qemwvC)
	h.Write(qemwvH)
	h.Write(mctpC)
	h.Write(mctpH)
	h.Write(qempC)
	h.Write(qempH)
	h.Write(qecmtC)
	h.Write(qecmtH)
	h.Write(qeraC)
	h.Write(qeraH)
	h.Write(qecmaC)
	h.Write(qecmaH)
	h.Write(qecC)
	h.Write(qecH)
	h.Write(consC)
	h.Write(consH)
	h.Write(scopedC)
	h.Write(scopedH)
	h.Write(toolsC)
	h.Write(toolsH)
	h.Write(certC)
	h.Write(certH)
	h.Write(abstainC)
	h.Write(abstainH)
	h.Write(sbC)
	h.Write(sbH)
	h.Write(gpuC)
	h.Write(gpuH)
	h.Write(agkeC)
	h.Write(agkeH)
	h.Write(mgpuC)
	h.Write(mgpuH)
	h.Write(cdagaC)
	h.Write(cdagaH)
	h.Write(distC)
	h.Write(distH)
	h.Write(baratC)
	h.Write(baratH)
	h.Write(qhalC)
	h.Write(qhalH)
	h.Write(seaC)
	h.Write(seaH)
	h.Write(transpileC)
	h.Write(transpileH)
	h.Write(eeaC)
	h.Write(eeaH)
	h.Write(qecdC)
	h.Write(qecdH)
	h.Write(aadC)
	h.Write(aadH)
	h.Write(recC)
	h.Write(recH)
	h.Write(varC)
	h.Write(varH)
	h.Write(rscaC)
	h.Write(rscaH)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)[:16]
}

// DefaultCacheDir returns the default cache directory for runtime build artifacts.
func DefaultCacheDir() string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".qadav", "runtime")
	}
	return filepath.Join(os.TempDir(), "qadav", "runtime")
}

// Build compiles all runtime modules (base + tensor + neural network + quantum + prime) into object files
// and returns all paths: []string{rtObj, tensorObj, nnObj, quantumObj, primeObj}.
func Build(tc *llvm.Toolchain, cacheDir string) ([]string, error) {
	if tc == nil {
		var err error
		tc, err = llvm.Detect()
		if err != nil {
			return nil, fmt.Errorf("detecting toolchain for runtime build: %w", err)
		}
	}

	if cacheDir == "" {
		cacheDir = DefaultCacheDir()
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating runtime cache directory: %w", err)
	}

	hashStr := Hash()
	rtObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_rt_%s.o", hashStr))
	tensorObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_tensor_%s.o", hashStr))
	nnObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_nn_%s.o", hashStr))
	quantumObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_quantum_%s.o", hashStr))
	primeObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_prime_%s.o", hashStr))
	ioObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_io_%s.o", hashStr))
	datasetObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_dataset_%s.o", hashStr))
	quasiObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_quasinary_%s.o", hashStr))
	cirObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_cir_%s.o", hashStr))
	agiObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_agi_%s.o", hashStr))
	sha3ObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_sha3_%s.o", hashStr))
	merkleObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_merkle_%s.o", hashStr))
	lamportObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_lamport_%s.o", hashStr))
	selfcheckObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_selfcheck_%s.o", hashStr))
	qeehObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_qeeh_%s.o", hashStr))
	qemwvObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_qemwv_%s.o", hashStr))
	mctpObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_mctp_%s.o", hashStr))
	qempObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_qemp_%s.o", hashStr))
	qecmtObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_qecmt_%s.o", hashStr))
	qeraObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_qera_%s.o", hashStr))
	qecmaObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_qecma_%s.o", hashStr))
	qecObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_qec_%s.o", hashStr))
	consObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_consolidate_%s.o", hashStr))
	scopedObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_scoped_%s.o", hashStr))
	toolsObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_tools_%s.o", hashStr))
	certObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_certificate_%s.o", hashStr))
	abstainObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_abstain_%s.o", hashStr))
	sbObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_sandbox_%s.o", hashStr))
	gpuObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_gpu_%s.o", hashStr))
	agkeObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_agke_%s.o", hashStr))
	mgpuObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_multigpu_%s.o", hashStr))
	cdagaObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_cdaga_%s.o", hashStr))
	distObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_distributed_%s.o", hashStr))
	baratObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_barat_%s.o", hashStr))
	qhalObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_qhal_%s.o", hashStr))
	seaObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_sea_%s.o", hashStr))
	transpileObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_transpile_%s.o", hashStr))
	eeaObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_eea_%s.o", hashStr))
	qecdObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_qecdecoder_%s.o", hashStr))
	aadObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_aad_%s.o", hashStr))
	recObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_record_%s.o", hashStr))
	varObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_variant_%s.o", hashStr))
	rscaObjPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_rsca_%s.o", hashStr))

	if entries, err := srcFiles.ReadDir("."); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".h") {
				if hContent, err := srcFiles.ReadFile(entry.Name()); err == nil {
					_ = os.WriteFile(filepath.Join(cacheDir, entry.Name()), hContent, 0644)
				}
			}
		}
	}
	if consHContent, err := srcFiles.ReadFile("qadav_consolidate.h"); err == nil {
		_ = os.WriteFile(filepath.Join(cacheDir, "qadav_consolidate.h"), consHContent, 0644)
	}
	if consCContent, err := srcFiles.ReadFile("qadav_consolidate.c"); err == nil {
		_ = os.WriteFile(filepath.Join(cacheDir, "qadav_consolidate.c"), consCContent, 0644)
	}

	// Ensure Windows import library for kernel32.dll is generated in cacheDir
	if runtime.GOOS == "windows" {
		libPath := filepath.Join(cacheDir, "kernel32.lib")
		defPath := filepath.Join(cacheDir, "kernel32.def")
		needsGen := false
		if _, err := os.Stat(libPath); os.IsNotExist(err) {
			needsGen = true
		} else if existingDef, err := os.ReadFile(defPath); err != nil || !strings.Contains(string(existingDef), "LoadLibraryA") {
			needsGen = true
		}
		if needsGen {
			if err := generateKernel32Lib(tc, cacheDir, libPath); err != nil {
				return nil, fmt.Errorf("generating kernel32.lib: %w", err)
			}
		}
	}

	// 1. Build primary OS runtime object
	if _, err := os.Stat(rtObjPath); os.IsNotExist(err) {
		filename, content, err := Sources()
		if err != nil {
			return nil, err
		}

		srcPath := filepath.Join(cacheDir, fmt.Sprintf("qadav_rt_%s.c", hashStr))
		if err := os.WriteFile(srcPath, content, 0644); err != nil {
			return nil, fmt.Errorf("writing runtime source: %w", err)
		}

		cmd := exec.Command(tc.ClangPath, "-c", srcPath, "-o", rtObjPath, "-O2")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("compiling runtime %s: %w\noutput: %s", filename, err, string(out))
		}
	}

	// 2-14. Build modular C runtime objects
	modules := []struct {
		name    string
		objPath string
	}{
		{"qadav_tensor", tensorObjPath},
		{"qadav_nn", nnObjPath},
		{"qadav_quantum", quantumObjPath},
		{"qadav_prime", primeObjPath},
		{"qadav_io", ioObjPath},
		{"qadav_dataset", datasetObjPath},
		{"qadav_quasinary", quasiObjPath},
		{"qadav_cir", cirObjPath},
		{"qadav_agi", agiObjPath},
		{"qadav_sha3", sha3ObjPath},
		{"qadav_merkle", merkleObjPath},
		{"qadav_lamport", lamportObjPath},
		{"qadav_selfcheck", selfcheckObjPath},
		{"qadav_qeeh", qeehObjPath},
		{"qadav_qemwv", qemwvObjPath},
		{"qadav_mctp", mctpObjPath},
		{"qadav_qemp", qempObjPath},
		{"qadav_qecmt", qecmtObjPath},
		{"qadav_qera", qeraObjPath},
		{"qadav_qecma", qecmaObjPath},
		{"qadav_qec", qecObjPath},
		{"qadav_consolidate", consObjPath},
		{"qadav_scoped", scopedObjPath},
		{"qadav_tools", toolsObjPath},
		{"qadav_certificate", certObjPath},
		{"qadav_abstain", abstainObjPath},
		{"qadav_sandbox", sbObjPath},
		{"qadav_gpu", gpuObjPath},
		{"qadav_agke", agkeObjPath},
		{"qadav_multigpu", mgpuObjPath},
		{"qadav_cdaga", cdagaObjPath},
		{"qadav_distributed", distObjPath},
		{"qadav_barat", baratObjPath},
		{"qadav_qhal", qhalObjPath},
		{"qadav_sea", seaObjPath},
		{"qadav_transpile", transpileObjPath},
		{"qadav_eea", eeaObjPath},
		{"qadav_qecdecoder", qecdObjPath},
		{"qadav_aad", aadObjPath},
		{"qadav_record", recObjPath},
		{"qadav_variant", varObjPath},
		{"qadav_rsca", rscaObjPath},
	}

	for _, m := range modules {
		if err := compileCModule(tc, cacheDir, hashStr, m.name, m.objPath); err != nil {
			return nil, err
		}
	}

	return []string{
		rtObjPath, tensorObjPath, nnObjPath, quantumObjPath, primeObjPath,
		ioObjPath, datasetObjPath, quasiObjPath, cirObjPath, agiObjPath,
		sha3ObjPath, merkleObjPath, lamportObjPath, selfcheckObjPath, qeehObjPath,
		qemwvObjPath, mctpObjPath, qempObjPath, qecmtObjPath, qeraObjPath,
		qecmaObjPath, qecObjPath, consObjPath, scopedObjPath, toolsObjPath,
		certObjPath, abstainObjPath, sbObjPath, gpuObjPath, agkeObjPath,
		mgpuObjPath, cdagaObjPath, distObjPath, baratObjPath,
		qhalObjPath, seaObjPath, transpileObjPath, eeaObjPath,
		qecdObjPath, aadObjPath,
		recObjPath, varObjPath, rscaObjPath,
	}, nil
}

func compileCModule(tc *llvm.Toolchain, cacheDir, hashStr, baseName, objPath string) error {
	if _, err := os.Stat(objPath); !os.IsNotExist(err) {
		return nil
	}

	hName := fmt.Sprintf("%s.h", baseName)
	if hContent, err := srcFiles.ReadFile(hName); err == nil {
		hPath := filepath.Join(cacheDir, hName)
		if err := os.WriteFile(hPath, hContent, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", hName, err)
		}
	}

	cName := fmt.Sprintf("%s.c", baseName)
	cContent, err := srcFiles.ReadFile(cName)
	if err != nil {
		return fmt.Errorf("reading embedded %s: %w", cName, err)
	}

	cPath := filepath.Join(cacheDir, fmt.Sprintf("%s_%s.c", baseName, hashStr))
	if err := os.WriteFile(cPath, cContent, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", cName, err)
	}

	cmd := exec.Command(tc.ClangPath, "-c", cPath, "-I", cacheDir, "-o", objPath, "-O2", "-fno-builtin")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compiling %s: %w\noutput: %s", cName, err, string(out))
	}
	return nil
}

func generateKernel32Lib(tc *llvm.Toolchain, cacheDir, libPath string) error {
	dlltool := "llvm-dlltool"
	if tc.ClangPath != "" {
		cand := filepath.Join(filepath.Dir(tc.ClangPath), "llvm-dlltool.exe")
		if _, err := os.Stat(cand); err == nil {
			dlltool = cand
		}
	}

	defPath := filepath.Join(cacheDir, "kernel32.def")
	defContent := `LIBRARY kernel32.dll
EXPORTS
    GetStdHandle
    WriteFile
    ExitProcess
    HeapAlloc
    HeapFree
    GetProcessHeap
    CreateFileA
    ReadFile
    CloseHandle
    VirtualProtect
    GetModuleHandleA
    GetModuleFileNameA
    GetFileSize
    LoadLibraryA
    GetProcAddress
    FreeLibrary
`
	if err := os.WriteFile(defPath, []byte(defContent), 0644); err != nil {
		return err
	}

	cmd := exec.Command(dlltool, "-m", "i386:x86-64", "-d", defPath, "-l", libPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("llvm-dlltool error: %w\noutput: %s", err, string(out))
	}

	return nil
}
