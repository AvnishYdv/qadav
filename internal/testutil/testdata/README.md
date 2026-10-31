# Testdata Conventions

This directory contains supporting files for tests in the `internal/testutil` package.

## File Types

- **`.golden`** – Golden (snapshot) files that hold the expected output of a test.  They are compared against the actual output using `AssertSnapshot`.  If a golden file does not exist on the first test run it will be created automatically.
- **`.in`** – Input files used by a test.  Tests can read these files to feed data into the code under test.
- **`.ll.snap`** – LLVM IR snapshot files.  Used with `ValidateLLVMIR` to store the LLVM IR that a test expects to be valid.

## Updating Snapshots

When a legitimate change to the expected output occurs, run the tests with the `-update` flag:

```bash
go test -update ./...
```

This will regenerate all golden files and IR snapshots.

## Adding New Testdata

1. Place the file in a relevant sub‑directory under `testdata/`.
2. Name it with the appropriate extension (`.golden`, `.in`, or `.ll.snap`).
3. Reference it from your test using `filepath.Join("testdata", "mycase.golden")` (or similar).

The `go test` command automatically marks the `testdata` directory as a data source, so the files are available at runtime.

