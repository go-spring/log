grammar Expr;

// ----------------------------------
// Lexer Rules
// ----------------------------------

// Identifiers: type names and field names
IDENT : [a-zA-Z_][a-zA-Z0-9_]* ;

// Integers
INT : [0-9]+ ;

// Strings: double-quoted or single-quoted
STRING : '"' (~["\\] | '\\' .)* '"'
       | '\'' (~['\\] | '\\' .)* '\''
       ;

// Whitespace (skip)
WS : [ \t\r\n]+ -> skip ;

// ----------------------------------
// Parser Rules
// ----------------------------------

// Root expression: a type with a block of fields
expr
    : IDENT '{' innerExpr (',' innerExpr)* '}'   # TypeExpr
    ;

// Field assignment inside a type block
innerExpr
    : fieldAccess '=' value                      # FieldAssign
    ;

// Nested field access: supports dot and array indexing
fieldAccess
    : IDENT ('.' IDENT | '[' INT ']')*          # NestedField
    ;

// Values: string, identifier, integer, or nested expression
value
    : STRING                                    # StringValue
    | IDENT                                     # IdentValue
    | INT                                       # IntValue
    | expr                                      # NestedExpr
    ;
