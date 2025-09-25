# FileSync

A simple **one-way file synchronization CLI** written in Go.  
It copies files from a **source directory → target directory**, replacing outdated files and (optionally) deleting files missing from the source.

---

## Features
- One-time synchronization (no background watching).
- Copies new files from source to target.
- Updates files in target if size or modification time differ.
- Optionally deletes files from target that are missing in source (`--delete-missing`).
- Preserves directory structure and file modification times.


## Usage

From project root, run:

```bash
go run main.go ./examples/source ./examples/target
```

With deletion enabled:
```bash
go run main.go --delete-missing ./examples/source ./examples/target
```

## Tests
```bash
cd src/filesync
```
If you want to run tests, run:
```bash
go test
```
if you want to run benchmarks, run:
```bash
go test --bench=.
```