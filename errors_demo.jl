pkg main
imp "std/io"

// 1) invalid HEX underscore (should error: "invalid hex literal")
var a: i32 = 0x_1

// 2) invalid BINARY literal (no digits after 0b)
var b: i32 = 0b

// 3) invalid OCTAL literal (no octal digits after 0o)
var c: i32 = 0o

// 4) invalid FLOAT exponent (e/+ but no digits)
var d: f64 = 1e+

// 5) illegal underscore placement in decimal
var e: i32 = 123_

// 6) invalid character
@
