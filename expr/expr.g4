grammar Expr;

// ----------------------------------
// Lexer Rules
// ----------------------------------

// 类型名或者字段名（标识符）
IDENT : [a-zA-Z_][a-zA-Z0-9_]* ;

// 数组下标，只支持整数
INDEX : [0-9]+ ;

// 字符串，支持单引号或双引号
STRING
    : '"' (~["\\] | '\\' .)* '"'
    | '\'' (~['\\] | '\\' .)* '\''
    ;

// 空白字符
WS : [ \t\r\n]+ -> skip ;

// ----------------------------------
// Parser Rules
// ----------------------------------

// 表达式 type{k=v，可以没有任何kv}
expr
    : IDENT '{' (innerExpr (',' innerExpr)*)? '}'   # TypeExpr
    ;

// k=v
innerExpr
    : fieldAccess '=' value                          # FieldAssign
    ;

// 支持多段字段访问，点号或数组形式
fieldAccess
    : IDENT ('.' IDENT | '[' INDEX ']')*            # NestedField
    ;

// 值，可以是字符串，嵌套表达式，或者不包含空格的原始内容
value
    : STRING                                       # StringValue
    | expr                                         # NestedExpr
    | ~[ \t\r\n{}=,]+                              # RawValue：非空白、非特殊字符的内容
    ;
