# iochain

`iochain` is a lightweight Go package for building dynamic chains of `io.Writer` and `io.Reader` using stackable, resettable components.

It allows you to:

- Chain multiple `io.WriteCloser`s and write to the top layer.
- Chain multiple `io.ReadCloser`s and read from the top layer.
- Automatically wire each new layer via its `Reset()` method.
- Flush and close the entire stack in the correct order.

Ideal for building pipelines such as:
- `gzip.Writer → bufio.Writer → file`
- `gzip.Reader ← bufio.Reader ← file`

---

## ✨ Features

* Supports `ResettableWriteCloser` and `ResettableReadCloser`
* Correct order of `Flush()` (from top to base)
* Correct order of `Close()` (from top to base)
* Thread-safe via internal mutex

---

## ✍️ Interfaces

```go
type ResettableWriteCloser interface {
    io.WriteCloser
    Reset(io.Writer) error
}

type ResettableReadCloser interface {
    io.ReadCloser
    Reset(io.Reader) error
}

type Flusher interface {
    Flush() error
}
```

---

## ✏️ Example: MultiWriter with gzip

```go
package main


func main() {
    f, _ := os.Create("out.gz")
    mw, _ := iochain.NewWriter(f)

    // gzip.Writer implements Reset(io.Writer)
    gz := gzip.NewWriter(nil)
    _ = mw.AddWriter(gz)

    mw.Write([]byte("Hello World!\n"))
    mw.FlushAndClose()
}
```

---

## 🔎 Example: MultiReader with gzip

```go
package main

import (
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "github.com/pyxsoft/iochain"
)

func main() {
    f, _ := os.Open("out.gz")
    mr, _ := iochain.NewReader(f)

    gz, _ := gzip.NewReader(nil)
    _ = mr.AddReader(gz)

    buf := make([]byte, 1024)
    n, _ := mr.Read(buf)
    fmt.Print(string(buf[:n]))

    mr.Close()
}
```

---

## 🧠 Notes

* Writers and readers must support `Reset()` to be chained.
* `Flush()` calls all flushable layers from **top to base**, ensuring all buffered data is pushed through.
* `Close()` calls all closers from **top to base**, like a proper pipeline teardown.

---

## 📄 License

MIT License. See [LICENSE](LICENSE).

