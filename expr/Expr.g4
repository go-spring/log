grammar Expr;

// ----------------------------------
// Lexer Rules
// ----------------------------------

// Identifier for type names or field names
IDENT : [a-zA-Z_][a-zA-Z0-9_]* ;

// Array index (only non-negative integers)
INDEX : [0-9]+ ;

// String literal, supports single or double quotes with common escape sequences
STRING
    : '"' ( ~["\\] | '\\' ["\\/bfnrt] )* '"'
    | '\'' ( ~['\\] | '\\' ["\\/bfnrt] )* '\''
    ;

// Raw value: non-whitespace, non-special characters
RAW_VALUE : ~[ \t\r\n{}=,]+ ;

// Whitespace (spaces, tabs, newlines) are skipped
WS : [ \t\r\n]+ -> skip ;

// ----------------------------------
// Parser Rules
// ----------------------------------

// Main expression: a type with optional key-value pairs
// Example: TypeName { field1 = "value1", field2 = NestedType { ... }, field3 = rawValue }
expr
    : IDENT '{' innerExprList? '}'
    ;

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
value
    : STRING
    | RAW_VALUE
    | expr
    ;
