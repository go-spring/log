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

// String literal, supports only double quotes for simplicity
STRING
    : '"' ( ~["\\] | '\\' ["\\/bfnrt] )* '"'
    ;

// Integer literal
INTEGER
    : ('+' | '-')? DIGIT+ | '0x' HEX_DIGIT+
    ;

// Floating-point number
FLOAT
    : ('+' | '-')? ( DIGIT+ ('.' DIGIT+)? | '.' DIGIT+ ) (('E' | 'e') ('+'|'-')? DIGIT+ )?
    ;

// Fragments
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

// value: clear and conflict-free definition
value
    : STRING                 // must use string for paths, special symbols, etc.
    | IDENT                  // symbolic identifiers like 'debug', 'info'
    | KW_TRUE | KW_FALSE     // boolean literals
    | INTEGER | FLOAT        // numeric literals
    | expr                   // nested expressions
    ;
