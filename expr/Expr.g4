grammar Expr;

// ----------------------------------
// Lexer Rules
// ----------------------------------


// Identifier for type names, field names, or symbolic constants
IDENT : [a-zA-Z_][a-zA-Z0-9_]* ;

// Array index (only non-negative integers)
INDEX : [0-9]+ ;

// String literal: double-quoted strings
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

// Main expression: a type name with optional key-value pairs enclosed in braces
// Example: TypeName { field1 = "value1", field2 = NestedType { ... }, field3 = rawValue }
expr
    : IDENT '{' innerExprList? '}'
    ;

// List of key-value assignments inside an expression, optionally comma-separated
// Trailing comma is allowed
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

// Value can be a string, identifier, boolean, numeric literal, or nested expression
value
    : STRING
    | IDENT
    | INTEGER | FLOAT
    | expr
    ;
