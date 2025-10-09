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
    : '"' (~["\\] | '\\' .)* '"'   // Double-quoted string with escape support
    | '\'' (~['\\] | '\\' .)* '\'' // Single-quoted string with escape support
    ;

// Whitespace (spaces, tabs, newlines) are skipped
WS : [ \t\r\n]+ -> skip ;

// ----------------------------------
// Parser Rules
// ----------------------------------

// Main expression: a type with optional key-value pairs
expr
    : IDENT '{' (innerExpr (',' innerExpr)*)? '}'   # TypeExpr
    ;

// Key-value assignment: field = value
innerExpr
    : fieldAccess '=' value                          # FieldAssign
    ;

// Field access supports nested fields via dot notation or array indices
fieldAccess
    : IDENT ('.' IDENT | '[' INDEX ']')*            # NestedField
    ;

// Value can be:
// 1. A string literal
// 2. A nested expression
// 3. A raw value (non-whitespace, non-special characters)
value
    : STRING                                       # StringValue
    | expr                                         # NestedExpr
    | ~[ \t\r\n{}=,]+                              # RawValue
    ;
