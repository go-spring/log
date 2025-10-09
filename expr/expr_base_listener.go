// Code generated from expr.g4 by ANTLR 4.13.2. DO NOT EDIT.

package expr // expr
import "github.com/antlr4-go/antlr/v4"

// BaseexprListener is a complete listener for a parse tree produced by exprParser.
type BaseexprListener struct{}

var _ exprListener = &BaseexprListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseexprListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseexprListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseexprListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseexprListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterExpr is called when production expr is entered.
func (s *BaseexprListener) EnterExpr(ctx *ExprContext) {}

// ExitExpr is called when production expr is exited.
func (s *BaseexprListener) ExitExpr(ctx *ExprContext) {}

// EnterInnerExpr is called when production innerExpr is entered.
func (s *BaseexprListener) EnterInnerExpr(ctx *InnerExprContext) {}

// ExitInnerExpr is called when production innerExpr is exited.
func (s *BaseexprListener) ExitInnerExpr(ctx *InnerExprContext) {}

// EnterFieldAccess is called when production fieldAccess is entered.
func (s *BaseexprListener) EnterFieldAccess(ctx *FieldAccessContext) {}

// ExitFieldAccess is called when production fieldAccess is exited.
func (s *BaseexprListener) ExitFieldAccess(ctx *FieldAccessContext) {}

// EnterValue is called when production value is entered.
func (s *BaseexprListener) EnterValue(ctx *ValueContext) {}

// ExitValue is called when production value is exited.
func (s *BaseexprListener) ExitValue(ctx *ValueContext) {}
