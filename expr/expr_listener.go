// Code generated from expr.g4 by ANTLR 4.13.2. DO NOT EDIT.

package expr // expr
import "github.com/antlr4-go/antlr/v4"

// exprListener is a complete listener for a parse tree produced by exprParser.
type exprListener interface {
	antlr.ParseTreeListener

	// EnterExpr is called when entering the expr production.
	EnterExpr(c *ExprContext)

	// EnterInnerExpr is called when entering the innerExpr production.
	EnterInnerExpr(c *InnerExprContext)

	// EnterFieldAccess is called when entering the fieldAccess production.
	EnterFieldAccess(c *FieldAccessContext)

	// EnterValue is called when entering the value production.
	EnterValue(c *ValueContext)

	// ExitExpr is called when exiting the expr production.
	ExitExpr(c *ExprContext)

	// ExitInnerExpr is called when exiting the innerExpr production.
	ExitInnerExpr(c *InnerExprContext)

	// ExitFieldAccess is called when exiting the fieldAccess production.
	ExitFieldAccess(c *FieldAccessContext)

	// ExitValue is called when exiting the value production.
	ExitValue(c *ValueContext)
}
