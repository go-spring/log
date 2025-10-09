grammar Expr;

// ----------------------------------
// Lexer Rules
// ----------------------------------


KW_TRUE  : 'true';
KW_FALSE : 'false';

// Identifier for type names or field names
IDENT : [a-zA-Z_][a-zA-Z0-9_]* ;

// Array index (only non-negative integers)
INDEX : [0-9]+ ;

// String literal, supports single or double quotes with common escape sequences
STRING
    : '"' ( ~["\\] | '\\' ["\\/bfnrt] )* '"'
    ;

// --------------------
// Integer literal
// Decimal integer with optional sign (+/-) or hexadecimal integer prefixed with 0x.
// --------------------
INTEGER
    : ('+' | '-')? DIGIT+ | '0x' HEX_DIGIT+
    ;

// --------------------
// Floating-point number
// Supports decimals and scientific notation (e.g., 1.23e+10)
// --------------------
FLOAT
    : ('+' | '-')? ( DIGIT+ ('.' DIGIT+)? | '.' DIGIT+ ) (('E' | 'e') ('+'|'-')? DIGIT+ )?
    ;

// --------------------
// Fragments (used internally, not emitted as tokens)
// --------------------
fragment DIGIT     : '0'..'9';
fragment LETTER    : 'A'..'Z' | 'a'..'z';
fragment HEX_DIGIT : DIGIT | 'A'..'F' | 'a'..'f';

// Whitespace (spaces, tabs, newlines) are skipped
WS : [ \t\r\n]+ -> skip ;

// ----------------------------------
// Parser Rules
// ----------------------------------

// Root node: entry point of the parser, ensures the entire input is a single expression
root: expr EOF ;

// Main expression: a type with optional key-value pairs
// Example: TypeName { field1 = "value1", field2 = NestedType { ... }, field3 = rawValue }
expr
    : IDENT '{' innerExprList? '}'
    ;

// List of key-value assignments inside an expression, optionally comma-separated
innerExprList
    : innerExpr (',' innerExpr)* ','?
    ;

// Key-value assignment: field = value
innerExpr
    : fieldAccess '=' value
    ;

// Field access supports nested fields via dot notation or array indices
// Examples: foo, foo.bar, foo[0], foo.bar[1].baz
fieldAccess
    : IDENT ('.' IDENT | '[' INDEX ']')*
    ;

// Value can be:
// 1. A string literal
// 2. A nested expression
// 3. A raw value (non-whitespace, non-special characters)
// 4. An identifier (for simple symbolic values)
value
    : STRING
    | IDENT
    | KW_TRUE | KW_FALSE | INTEGER | FLOAT
    | expr
    ;
