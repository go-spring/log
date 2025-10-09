grammar Expr;

// ----------------------------------
// Lexer Rules
// ----------------------------------

// 类型名或者字段名（标识符）
IDENT : [a-zA-Z_][a-zA-Z0-9_]* ;

// index （数组下标，只支持普通整数）
INDEX : [0-9]+ ;

// 字符串，双引号或者单引号
STRING : '"' (~["\\] | '\\' .)* '"'
       | '\'' (~['\\] | '\\' .)* '\''
       ;

// 空白字符
WS : [ \t\r\n]+ -> skip ;

// ----------------------------------
// Parser Rules
// ----------------------------------

// 表达式 type{k=v 可以不包含任何kv，需要使用逗号进行分隔}
expr
    : IDENT '{' (innerExpr (',' innerExpr)*)* '}'   # TypeExpr
    ;

// k=v
innerExpr
    : fieldAccess '=' value                      # FieldAssign
    ;

// k，支持多段，点号或者数组形式
fieldAccess
    : IDENT ('.' IDENT | '[' INDEX ']')*          # NestedField
    ;

// 值，字符串，整数，浮点数，字面值（不只是标识符），嵌套的表达式
value
    : STRING                                    # StringValue
    | expr                                      # NestedExpr
    | 我认为只要是不包含空格的内容就行
    ;
