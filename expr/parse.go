package expr

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

func Parse(data string) (ret map[string]string, err error) {
	if data = strings.TrimSpace(data); data == "" {
		return nil, nil
	}

	e := &ErrorListener{Data: data}

	// Recover from parser panics to provide better error reporting
	defer func() {
		if r := recover(); r != nil {
			ret = nil
			err = fmt.Errorf("[PANIC]: %v\n%s", r, debug.Stack())
			if e.Error != nil {
				err = fmt.Errorf("%w\n%w", e.Error, err)
			}
		}
	}()

	// Step 1: Create lexer and token stream
	input := antlr.NewInputStream(data)
	lexer := NewExprLexer(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(e)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Step 2: Create parser and attach custom error listener
	p := NewExprParser(tokens)
	p.RemoveErrorListeners()
	p.AddErrorListener(e)

	// Step 3: Walk parse tree with custom listener
	l := &ParseTreeListener{
		Result: make(map[string]string),
	}
	antlr.ParseTreeWalkerDefault.Walk(l, p.Expr())

	// Step 4: Return parsed expression or error
	if e.Error != nil {
		return nil, e.Error
	}
	return l.Result, nil
}

// ErrorListener implements a custom ANTLR error listener that records syntax errors.
type ErrorListener struct {
	*antlr.DefaultErrorListener
	Error error
	Data  string
}

// SyntaxError is called by ANTLR when a syntax error occurs.
func (l *ErrorListener) SyntaxError(_ antlr.Recognizer, _ any, line, column int, msg string, e antlr.RecognitionException) {
	if l.Error == nil {
		l.Error = fmt.Errorf("line %d:%d %s << text: %q", line, column, msg, l.Data)
		return
	}
	l.Error = fmt.Errorf("%w\nline %d:%d %s << text: %q", l.Error, line, column, msg, l.Data)
}

// ParseTreeListener walks the parse tree and constructs the expression AST.
type ParseTreeListener struct {
	BaseExprListener
	Result map[string]string
}

func (l *ParseTreeListener) ExitExpr(ctx *ExprContext) {
	l.parseExpr("", ctx)
}

func (l *ParseTreeListener) parseExpr(key string, ctx IExprContext) {
	typeKey := "type"
	if key != "" {
		typeKey = key + ".type"
	}
	l.Result[typeKey] = ctx.IDENT().GetText()
	if x := ctx.InnerExprList(); x != nil {
		for _, innerExpr := range x.AllInnerExpr() {
			l.parseInnerExpr(key, innerExpr)
		}
	}
}

func (l *ParseTreeListener) parseInnerExpr(key string, ctx IInnerExprContext) {
	fieldKey := ctx.FieldAccess().GetText()
	if key != "" {
		fieldKey = key + "." + fieldKey
	}
	switch {
	case ctx.Value().STRING() != nil:
		l.Result[fieldKey], _ = strconv.Unquote(ctx.Value().STRING().GetText())
	case ctx.Value().RAW_VALUE() != nil:
		l.Result[fieldKey] = ctx.Value().RAW_VALUE().GetText()
	case ctx.Value().Expr() != nil:
		l.parseExpr(fieldKey, ctx.Value().Expr())
	default: // for linter
	}
}
