# J Language Tokenizer

A lexical analyzer (tokenizer) for the custom **J Language**.

---

## ✅ Features

- Recognizes **all language keywords** (`pkg`, `imp`, `def`, `var`, `if`, `fr`, `select`, `panic`, `recover`, etc.)
- Supports:
    - Identifiers (Unicode allowed)
    - Integer & float literals with underscore rules
    - String literals (`"..."`) and raw strings (`` `...` ``)
    - Character literals (`'a'`, `'\n'`, `'\x41'`)
    - Type indicators (`i32`, `f64`, `bool`, `string`, ...)
    - Operators & delimiters (`==`, `<=`, `:=`, `<-`, etc.)
    - Line comments (`//`)
    - **Nested block comments** (`/* ... /* ... */ ... */`)
- **Lexical error detection** with **line and column number**
- Outputs **JSON** containing all tokens and errors
- Writes result to:
    - **stdout**, and
    - an output file based on source file name
        - e.g. `main.jl → main_jl_output.txt`

---

## ✅ How to Run

### Option 1: Provide a `.jl` source file
```bash
  go run jlang_lexer.go main.jl

```

Output:

- JSON printed to terminal

- File main_jl_output.txt is created 
### Option 2 — From stdin

```bash
    cat main.jl | go run main.go
    go run main.go <
```


Output:

- JSON is printed to terminal

- A file stdin_output.txt is created automatically

Output Format (JSON)

```json

{
  "tokens": [
    {"type":"KW_PKG","lexeme":"pkg","line":1,"col":1},
    {"type":"IDENT","lexeme":"main","line":1,"col":5}
  ],
  "errors": [
    "lexical error at 5:14: invalid hex literal"
  ]
}
```



## ✅ Example Source (main.jl)

```text
    pkg main
imp "std/io"
def main(): i32 {
  later handler()
  var x: i32 = 1_024; if x > 0 { io.Println("ok") } panic "boom"; ret 0
}
def handler() { msg := recover(); if msg != "" { io.Println("recovered:", msg) } }

```