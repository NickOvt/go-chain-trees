# Pprof PDF generation commands

These commands assume:
- You run them from the directory that contains the `.prof` files and the test binary.
- `go` and Graphviz (`dot`) are installed.

## 1) Cast all CPU profiles to PDF

```bash
for f in ./cpu_*.prof; do
  go tool pprof -pdf ./avlhashtree.test "$f" > "${f%.prof}.pdf"
done
```

## 2) Cast all heap profiles to inuse_space PDF

```bash
for f in ./heap_*.prof; do
  go tool pprof -pdf -inuse_space ./avlhashtree.test "$f" > "${f%.prof}_inuse.pdf"
done
```

## 3) Cast all heap profiles to alloc_space PDF

```bash
for f in ./heap_*.prof; do
  go tool pprof -pdf -alloc_space ./avlhashtree.test "$f" > "${f%.prof}_alloc.pdf"
done
```
