grammar Expr;

// ----------------------------------
// Lexer Rules
// ----------------------------------

// Identifier for type names or field names
IDENT : [a-zA-Z_][a-zA-Z0-9_]* ;

// Array index (only supports integers)
INDEX : [0-9]+ ;

// String literal, supports both single and double quotes
STRING
    : '"' (~["\\] | '\\' .)* '"'
    | '\'' (~['\\] | '\\' .)* '\''
    ;

// Whitespace (spaces, tabs, newlines) are skipped
WS : [ \t\r\n]+ -> skip ;

// ----------------------------------
// Parser Rules
// ----------------------------------

// Main expression: a type with optional key-value pairs
expr
    : IDENT '{' (innerExpr (',' innerExpr)*)? '}'
    ;

// Key-value assignment: field = value
innerExpr
    : fieldAccess '=' value
    ;

// Field access supports nested fields via dot notation or array indices
fieldAccess
    : IDENT ('.' IDENT | '[' INDEX ']')*
    ;

// Value can be:
// 1. A string literal
// 2. A nested expression
// 3. A raw value (non-whitespace, non-special characters)
value
    : STRING
    | expr
    | ~[ \t\r\n{}=,]+
    ;
