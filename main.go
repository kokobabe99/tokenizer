package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenType string

const (
	// keywords (lowercase)
	KW_PKG       TokenType = "KW_PKG"
	KW_IMP       TokenType = "KW_IMP"
	KW_DEF       TokenType = "KW_DEF"
	KW_VAR       TokenType = "KW_VAR"
	KW_CONS      TokenType = "KW_CONS"
	KW_TYPE      TokenType = "KW_TYPE"
	KW_STRUCT    TokenType = "KW_STRUCT"
	KW_INTERFACE TokenType = "KW_INTERFACE"
	KW_MAPPING   TokenType = "KW_MAPPING"
	KW_CHANNEL   TokenType = "KW_CHANNEL"
	KW_J         TokenType = "KW_J"
	KW_SELECT    TokenType = "KW_SELECT"
	KW_LATER     TokenType = "KW_LATER"
	KW_RET       TokenType = "KW_RET"
	KW_IF        TokenType = "KW_IF"
	KW_ELSE      TokenType = "KW_ELSE"
	KW_SWITCH    TokenType = "KW_SWITCH"
	KW_CASE      TokenType = "KW_CASE"
	KW_FALL      TokenType = "KW_FALL"
	KW_FR        TokenType = "KW_FR"
	KW_RANGE     TokenType = "KW_RANGE"
	KW_BREAK     TokenType = "KW_BREAK"
	KW_CONTINUE  TokenType = "KW_CONTINUE"
	KW_JOTO      TokenType = "KW_JOTO"
	KW_DFT       TokenType = "KW_DFT"
	KW_PANIC     TokenType = "KW_PANIC"
	KW_RECOVER   TokenType = "KW_RECOVER" // also accepts "recovery"

	// identifiers & literals & type names
	IDENT      TokenType = "IDENT"
	INT_LIT    TokenType = "INT_LIT"
	FLOAT_LIT  TokenType = "FLOAT_LIT"
	STRING_LIT TokenType = "STRING_LIT"
	CHAR_LIT   TokenType = "CHAR_LIT"
	TYPE_NAME  TokenType = "TYPE_NAME"

	// punctuation / operators
	LPAREN TokenType = "LPAREN" // (
	RPAREN TokenType = "RPAREN" // )
	LBRACE TokenType = "LBRACE" // {
	RBRACE TokenType = "RBRACE" // }
	LBRACK TokenType = "LBRACK" // [
	RBRACK TokenType = "RBRACK" // ]
	COMMA  TokenType = "COMMA"  // ,
	SEMI   TokenType = "SEMI"   // ;
	COLON  TokenType = "COLON"  // :
	DOT    TokenType = "DOT"    // .

	ASSIGN  TokenType = "ASSIGN"  // =
	DECL    TokenType = "DECL"    // :=
	PLUS    TokenType = "PLUS"    // +
	MINUS   TokenType = "MINUS"   // -
	STAR    TokenType = "STAR"    // *
	SLASH   TokenType = "SLASH"   // /
	PERCENT TokenType = "PERCENT" // %
	LT      TokenType = "LT"      // <
	GT      TokenType = "GT"      // >
	LE      TokenType = "LE"      // <=
	GE      TokenType = "GE"      // >=
	EQ      TokenType = "EQ"      // ==
	NE      TokenType = "NE"      // !=
	ANDAND  TokenType = "ANDAND"  // &&
	OROR    TokenType = "OROR"    // ||
	BAND    TokenType = "BAND"    // &
	BOR     TokenType = "BOR"     // |
	BXOR    TokenType = "BXOR"    // ^
	SHL     TokenType = "SHL"     // <<
	SHR     TokenType = "SHR"     // >>
	ADDEQ   TokenType = "ADDEQ"   // +=
	SUBEQ   TokenType = "SUBEQ"   // -=
	MULEQ   TokenType = "MULEQ"   // *=
	DIVEQ   TokenType = "DIVEQ"   // /=
	MODEQ   TokenType = "MODEQ"   // %=
	ANDEQ   TokenType = "ANDEQ"   // &=
	OREQ    TokenType = "OREQ"    // |=
	XOREQ   TokenType = "XOREQ"   // ^=
	SHLEQ   TokenType = "SHLEQ"   // <<=
	SHREQ   TokenType = "SHREQ"   // >>=

	CH_SEND TokenType = "CH_SEND" // <-
	BANG    TokenType = "BANG"    // !
)

var keywords = map[string]TokenType{
	"pkg": KW_PKG, "imp": KW_IMP, "def": KW_DEF, "var": KW_VAR, "cons": KW_CONS, "type": KW_TYPE,
	"struct": KW_STRUCT, "interface": KW_INTERFACE, "mapping": KW_MAPPING, "channel": KW_CHANNEL,
	"j": KW_J, "select": KW_SELECT, "later": KW_LATER, "ret": KW_RET, "if": KW_IF, "else": KW_ELSE,
	"switch": KW_SWITCH, "case": KW_CASE, "fall": KW_FALL, "fr": KW_FR, "range": KW_RANGE,
	"break": KW_BREAK, "continue": KW_CONTINUE, "joto": KW_JOTO, "dft": KW_DFT,
	"panic": KW_PANIC, "recover": KW_RECOVER, "recovery": KW_RECOVER,
}

var typeNames = map[string]struct{}{
	"i8": {}, "i16": {}, "i32": {}, "i64": {},
	"u8": {}, "u16": {}, "u32": {}, "u64": {},
	"f32": {}, "f64": {}, "bool": {}, "string": {},
}

type Token struct {
	Type     TokenType `json:"type"`
	Lexeme   string    `json:"lexeme"`
	Line     int       `json:"line"`
	Column   int       `json:"col"`
	IntVal   *int64    `json:"intVal,omitempty"`
	FloatVal *float64  `json:"floatVal,omitempty"`
}

type Lexer struct {
	src    []rune
	i      int
	line   int
	col    int
	length int
	tokens []Token
	errors []string
}

func NewLexer(input string) *Lexer {
	rs := []rune(input)
	return &Lexer{
		src: rs, length: len(rs),
		line: 1, col: 1,
	}
}

func (lx *Lexer) peek(n int) rune {
	j := lx.i + n
	if j >= lx.length {
		return 0
	}
	return lx.src[j]
}
func (lx *Lexer) advance() rune {
	if lx.i >= lx.length {
		return 0
	}
	ch := lx.src[lx.i]
	lx.i++
	if ch == '\n' {
		lx.line++
		lx.col = 1
	} else {
		lx.col++
	}
	return ch
}
func (lx *Lexer) add(tt TokenType, lex string, l, c int, iv *int64, fv *float64) {
	lx.tokens = append(lx.tokens, Token{Type: tt, Lexeme: lex, Line: l, Column: c, IntVal: iv, FloatVal: fv})
}
func (lx *Lexer) errorAt(l, c int, msg string) {
	lx.errors = append(lx.errors, fmt.Sprintf("lexical error at %d:%d: %s", l, c, msg))
}

func (lx *Lexer) isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}
func (lx *Lexer) isIdentPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func (lx *Lexer) skipWSAndComments() {
	for {
		ch := lx.peek(0)
		// whitespace
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			lx.advance()
			continue
		}
		// comments
		if ch == '/' {
			n := lx.peek(1)
			// line comment
			if n == '/' {
				for ch != '\n' && ch != 0 {
					ch = lx.advance()
				}
				continue
			}
			// nested block comment
			if n == '*' {
				startLine, startCol := lx.line, lx.col
				lx.advance()
				lx.advance()
				depth := 1
				for depth > 0 {
					c := lx.peek(0)
					if c == 0 {
						lx.errorAt(startLine, startCol, "unterminated block comment")
						return
					}
					if c == '/' && lx.peek(1) == '*' {
						lx.advance()
						lx.advance()
						depth++
						continue
					}
					if c == '*' && lx.peek(1) == '/' {
						lx.advance()
						lx.advance()
						depth--
						continue
					}
					lx.advance()
				}
				continue
			}
		}
		break
	}
}

// ---------- scans ----------
func (lx *Lexer) scanIdentOrKeyword() {
	l, c := lx.line, lx.col
	var b strings.Builder
	for lx.isIdentPart(lx.peek(0)) {
		b.WriteRune(lx.advance())
	}
	lex := b.String()
	low := strings.ToLower(lex)
	if t, ok := keywords[low]; ok {
		lx.add(t, lex, l, c, nil, nil)
		return
	}
	if _, ok := typeNames[lex]; ok {
		lx.add(TYPE_NAME, lex, l, c, nil, nil)
		return
	}
	lx.add(IDENT, lex, l, c, nil, nil)
}

func validUnderscores(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '_' || s[len(s)-1] == '_' {
		return false
	}
	if strings.Contains(s, "__") {
		return false
	}
	bad := []string{"_.", "._", "e_", "_e", "E_", "_E", "x_", "_x", "X_", "_X", "b_", "_b", "B_", "_B", "o_", "_o", "O_", "_O"}
	for _, p := range bad {
		if strings.Contains(s, p) {
			return false
		}
	}
	return true
}

func (lx *Lexer) scanNumber() {
	l, c := lx.line, lx.col
	start := lx.i

	// base-prefixed
	if lx.peek(0) == '0' && (lx.peek(1) == 'x' || lx.peek(1) == 'X' || lx.peek(1) == 'b' || lx.peek(1) == 'B' || lx.peek(1) == 'o' || lx.peek(1) == 'O') {
		base := lx.peek(1)
		lx.advance()
		lx.advance()
		var count int
		for {
			ch := lx.peek(0)
			if ch == '_' || unicode.IsDigit(ch) || (base == 'x' || base == 'X') && ((ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) || (base == 'b' || base == 'B') && (ch == '0' || ch == '1') || (base == 'o' || base == 'O') && (ch >= '0' && ch <= '7') {
				lx.advance()
				count++
			} else {
				break
			}
		}
		body := string(lx.src[start+2 : lx.i])
		if count == 0 || !validUnderscores(body) {
			msg := "invalid numeric literal"
			switch base {
			case 'x', 'X':
				msg = "invalid hex literal"
			case 'b', 'B':
				msg = "invalid binary literal"
			case 'o', 'O':
				msg = "invalid octal literal"
			}
			lx.errorAt(l, c, msg)
			return
		}
		lex := string(lx.src[start:lx.i])
		lx.add(INT_LIT, lex, l, c, nil, nil)
		return
	}

	// decimal / float
	for unicode.IsDigit(lx.peek(0)) || lx.peek(0) == '_' {
		lx.advance()
	}
	isFloat := false
	if lx.peek(0) == '.' && unicode.IsDigit(lx.peek(1)) {
		isFloat = true
		lx.advance()
		for unicode.IsDigit(lx.peek(0)) || lx.peek(0) == '_' {
			lx.advance()
		}
	}
	if lx.peek(0) == 'e' || lx.peek(0) == 'E' {
		isFloat = true
		lx.advance()
		if lx.peek(0) == '+' || lx.peek(0) == '-' {
			lx.advance()
		}
		if !unicode.IsDigit(lx.peek(0)) {
			lx.errorAt(l, c, "invalid float exponent")
			return
		}
		for unicode.IsDigit(lx.peek(0)) || lx.peek(0) == '_' {
			lx.advance()
		}
	}
	lex := string(lx.src[start:lx.i])
	if !validUnderscores(lex) {
		lx.errorAt(l, c, "illegal underscore placement in number")
		return
	}
	if isFloat || strings.ContainsAny(lex, ".eE") {
		lx.add(FLOAT_LIT, lex, l, c, nil, nil)
	} else {
		lx.add(INT_LIT, lex, l, c, nil, nil)
	}
}

func (lx *Lexer) scanString() {
	l, c := lx.line, lx.col
	var b strings.Builder
	b.WriteRune(lx.advance())
	for {
		ch := lx.peek(0)
		if ch == 0 || ch == '\n' {
			lx.errorAt(l, c, "unterminated string literal")
			return
		}
		if ch == '\\' {
			b.WriteRune(lx.advance())
			if lx.peek(0) == 0 || lx.peek(0) == '\n' {
				lx.errorAt(l, c, "unterminated string escape")
				return
			}
			b.WriteRune(lx.advance())
			continue
		}
		b.WriteRune(lx.advance())
		if b.Len() >= 2 {
			// closing quote?
			r, _ := utf8.DecodeLastRuneInString(b.String())
			if r == '"' && lx.peek(-1) != '\\' {
				break
			}
		}
		if lx.peek(0) == '"' {
			b.WriteRune(lx.advance())
			break
		}
	}
	lx.add(STRING_LIT, b.String(), l, c, nil, nil)
}

func (lx *Lexer) scanRawString() {
	l, c := lx.line, lx.col
	var b strings.Builder
	b.WriteRune(lx.advance()) // `
	for {
		ch := lx.peek(0)
		if ch == 0 {
			lx.errorAt(l, c, "unterminated raw string")
			return
		}
		b.WriteRune(lx.advance())
		if ch == '`' {
			break
		}
	}
	lx.add(STRING_LIT, b.String(), l, c, nil, nil)
}

func (lx *Lexer) scanChar() {
	l, c := lx.line, lx.col
	var b strings.Builder
	b.WriteRune(lx.advance()) // '
	ch := lx.peek(0)
	if ch == '\\' {
		b.WriteRune(lx.advance())
		if lx.peek(0) == 0 || lx.peek(0) == '\n' {
			lx.errorAt(l, c, "unterminated char escape")
			return
		}
		b.WriteRune(lx.advance())
	} else {
		if ch == 0 || ch == '\n' || ch == '\'' {
			lx.errorAt(l, c, "empty or invalid char literal")
			return
		}
		b.WriteRune(lx.advance())
	}
	if lx.peek(0) != '\'' {
		lx.errorAt(l, c, "unterminated char literal")
		return
	}
	b.WriteRune(lx.advance())
	lx.add(CHAR_LIT, b.String(), l, c, nil, nil)
}

// ---------- main tokenization step ----------
func (lx *Lexer) nextToken() bool {
	lx.skipWSAndComments()
	ch := lx.peek(0)
	if ch == 0 {
		return false
	}
	l, c := lx.line, lx.col

	if lx.isIdentStart(ch) {
		lx.scanIdentOrKeyword()
		return true
	}
	// numbers
	if unicode.IsDigit(ch) {
		lx.scanNumber()
		return true
	}
	// strings
	if ch == '"' {
		lx.scanString()
		return true
	}
	if ch == '`' {
		lx.scanRawString()
		return true
	}
	// char
	if ch == '\'' {
		lx.scanChar()
		return true
	}

	switch ch {
	case '(':
		lx.advance()
		lx.add(LPAREN, "(", l, c, nil, nil)
	case ')':
		lx.advance()
		lx.add(RPAREN, ")", l, c, nil, nil)
	case '{':
		lx.advance()
		lx.add(LBRACE, "{", l, c, nil, nil)
	case '}':
		lx.advance()
		lx.add(RBRACE, "}", l, c, nil, nil)
	case '[':
		lx.advance()
		lx.add(LBRACK, "[", l, c, nil, nil)
	case ']':
		lx.advance()
		lx.add(RBRACK, "]", l, c, nil, nil)
	case ',':
		lx.advance()
		lx.add(COMMA, ",", l, c, nil, nil)
	case ';':
		lx.advance()
		lx.add(SEMI, ";", l, c, nil, nil)
	case ':':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(DECL, ":=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(COLON, ":", l, c, nil, nil)
		}
	case '.':
		lx.advance()
		lx.add(DOT, ".", l, c, nil, nil)
	case '+':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(ADDEQ, "+=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(PLUS, "+", l, c, nil, nil)
		}
	case '-':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(SUBEQ, "-=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(MINUS, "-", l, c, nil, nil)
		}
	case '*':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(MULEQ, "*=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(STAR, "*", l, c, nil, nil)
		}
	case '/':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(DIVEQ, "/=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(SLASH, "/", l, c, nil, nil)
		}
	case '%':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(MODEQ, "%=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(PERCENT, "%", l, c, nil, nil)
		}
	case '<':
		if lx.peek(1) == '-' {
			lx.advance()
			lx.advance()
			lx.add(CH_SEND, "<-", l, c, nil, nil)
		} else if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(LE, "<=", l, c, nil, nil)
		} else if lx.peek(1) == '<' {
			if lx.peek(2) == '=' {
				lx.advance()
				lx.advance()
				lx.advance()
				lx.add(SHLEQ, "<<=", l, c, nil, nil)
			} else {
				lx.advance()
				lx.advance()
				lx.add(SHL, "<<", l, c, nil, nil)
			}
		} else {
			lx.advance()
			lx.add(LT, "<", l, c, nil, nil)
		}
	case '>':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(GE, ">=", l, c, nil, nil)
		} else if lx.peek(1) == '>' {
			if lx.peek(2) == '=' {
				lx.advance()
				lx.advance()
				lx.advance()
				lx.add(SHREQ, ">>=", l, c, nil, nil)
			} else {
				lx.advance()
				lx.advance()
				lx.add(SHR, ">>", l, c, nil, nil)
			}
		} else {
			lx.advance()
			lx.add(GT, ">", l, c, nil, nil)
		}
	case '=':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(EQ, "==", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(ASSIGN, "=", l, c, nil, nil)
		}
	case '!':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(NE, "!=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(BANG, "!", l, c, nil, nil)
		}
	case '&':
		if lx.peek(1) == '&' {
			lx.advance()
			lx.advance()
			lx.add(ANDAND, "&&", l, c, nil, nil)
		} else if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(ANDEQ, "&=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(BAND, "&", l, c, nil, nil)
		}
	case '|':
		if lx.peek(1) == '|' {
			lx.advance()
			lx.advance()
			lx.add(OROR, "||", l, c, nil, nil)
		} else if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(OREQ, "|=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(BOR, "|", l, c, nil, nil)
		}
	case '^':
		if lx.peek(1) == '=' {
			lx.advance()
			lx.advance()
			lx.add(XOREQ, "^=", l, c, nil, nil)
		} else {
			lx.advance()
			lx.add(BXOR, "^", l, c, nil, nil)
		}
	default:
		lx.errorAt(l, c, fmt.Sprintf("invalid character %q", ch))
		lx.advance()
	}
	return true
}

func (lx *Lexer) LexAll() ([]Token, []string) {
	for lx.nextToken() {
	}
	return lx.tokens, lx.errors
}
func outputFileName(arg string) string {
	if arg == "" || arg == "-" {
		return "stdin_output.txt"
	}
	base := filepath.Base(arg)
	base = strings.ReplaceAll(base, ".", "_") // e.g., main.jl -> main_jl
	return base + "_output.txt"               // -> main_jl_output.txt
}

func main() {
	var (
		data    []byte
		err     error
		srcPath string
	)
	if len(os.Args) > 1 && os.Args[1] != "-" {
		srcPath = os.Args[1]
		data, err = os.ReadFile(srcPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read file error: %v\n", err)
			os.Exit(1)
		}
	} else {
		srcPath = "-"
		data, err = io.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "read stdin error: %v\n", err)
			os.Exit(1)
		}
	}

	lx := NewLexer(string(data))
	toks, errs := lx.LexAll()

	out := struct {
		Tokens []Token  `json:"tokens"`
		Errors []string `json:"errors"`
	}{
		Tokens: toks,
		Errors: errs,
	}

	jsonBytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal json error: %v\n", err)
		os.Exit(1)
	}

	os.Stdout.Write(jsonBytes)
	os.Stdout.Write([]byte("\n"))

	outPath := outputFileName(srcPath)
	if err := os.WriteFile(outPath, jsonBytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write output file error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", outPath)
}
