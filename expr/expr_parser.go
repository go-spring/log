// Code generated from expr.g4 by ANTLR 4.13.2. DO NOT EDIT.

package expr // expr
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type exprParser struct {
	*antlr.BaseParser
}

var ExprParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func exprParserInit() {
	staticData := &ExprParserStaticData
	staticData.LiteralNames = []string{
		"", "'{'", "','", "'}'", "'='", "'.'", "'['", "']'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "IDENT", "INDEX", "STRING", "RAW_VALUE",
		"WS",
	}
	staticData.RuleNames = []string{
		"expr", "innerExpr", "fieldAccess", "value",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 12, 46, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 1, 0, 1,
		0, 1, 0, 1, 0, 1, 0, 5, 0, 14, 8, 0, 10, 0, 12, 0, 17, 9, 0, 1, 0, 3, 0,
		20, 8, 0, 3, 0, 22, 8, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1,
		2, 1, 2, 1, 2, 1, 2, 1, 2, 5, 2, 36, 8, 2, 10, 2, 12, 2, 39, 9, 2, 1, 3,
		1, 3, 1, 3, 3, 3, 44, 8, 3, 1, 3, 0, 0, 4, 0, 2, 4, 6, 0, 0, 48, 0, 8,
		1, 0, 0, 0, 2, 25, 1, 0, 0, 0, 4, 29, 1, 0, 0, 0, 6, 43, 1, 0, 0, 0, 8,
		9, 5, 8, 0, 0, 9, 21, 5, 1, 0, 0, 10, 15, 3, 2, 1, 0, 11, 12, 5, 2, 0,
		0, 12, 14, 3, 2, 1, 0, 13, 11, 1, 0, 0, 0, 14, 17, 1, 0, 0, 0, 15, 13,
		1, 0, 0, 0, 15, 16, 1, 0, 0, 0, 16, 19, 1, 0, 0, 0, 17, 15, 1, 0, 0, 0,
		18, 20, 5, 2, 0, 0, 19, 18, 1, 0, 0, 0, 19, 20, 1, 0, 0, 0, 20, 22, 1,
		0, 0, 0, 21, 10, 1, 0, 0, 0, 21, 22, 1, 0, 0, 0, 22, 23, 1, 0, 0, 0, 23,
		24, 5, 3, 0, 0, 24, 1, 1, 0, 0, 0, 25, 26, 3, 4, 2, 0, 26, 27, 5, 4, 0,
		0, 27, 28, 3, 6, 3, 0, 28, 3, 1, 0, 0, 0, 29, 37, 5, 8, 0, 0, 30, 31, 5,
		5, 0, 0, 31, 36, 5, 8, 0, 0, 32, 33, 5, 6, 0, 0, 33, 34, 5, 9, 0, 0, 34,
		36, 5, 7, 0, 0, 35, 30, 1, 0, 0, 0, 35, 32, 1, 0, 0, 0, 36, 39, 1, 0, 0,
		0, 37, 35, 1, 0, 0, 0, 37, 38, 1, 0, 0, 0, 38, 5, 1, 0, 0, 0, 39, 37, 1,
		0, 0, 0, 40, 44, 5, 10, 0, 0, 41, 44, 3, 0, 0, 0, 42, 44, 5, 11, 0, 0,
		43, 40, 1, 0, 0, 0, 43, 41, 1, 0, 0, 0, 43, 42, 1, 0, 0, 0, 44, 7, 1, 0,
		0, 0, 6, 15, 19, 21, 35, 37, 43,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// exprParserInit initializes any static state used to implement exprParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewexprParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func ExprParserInit() {
	staticData := &ExprParserStaticData
	staticData.once.Do(exprParserInit)
}

// NewexprParser produces a new parser instance for the optional input antlr.TokenStream.
func NewexprParser(input antlr.TokenStream) *exprParser {
	ExprParserInit()
	this := new(exprParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &ExprParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "expr.g4"

	return this
}

// exprParser tokens.
const (
	exprParserEOF       = antlr.TokenEOF
	exprParserT__0      = 1
	exprParserT__1      = 2
	exprParserT__2      = 3
	exprParserT__3      = 4
	exprParserT__4      = 5
	exprParserT__5      = 6
	exprParserT__6      = 7
	exprParserIDENT     = 8
	exprParserINDEX     = 9
	exprParserSTRING    = 10
	exprParserRAW_VALUE = 11
	exprParserWS        = 12
)

// exprParser rules.
const (
	exprParserRULE_expr        = 0
	exprParserRULE_innerExpr   = 1
	exprParserRULE_fieldAccess = 2
	exprParserRULE_value       = 3
)

// IExprContext is an interface to support dynamic dispatch.
type IExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENT() antlr.TerminalNode
	AllInnerExpr() []IInnerExprContext
	InnerExpr(i int) IInnerExprContext

	// IsExprContext differentiates from other interfaces.
	IsExprContext()
}

type ExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExprContext() *ExprContext {
	var p = new(ExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = exprParserRULE_expr
	return p
}

func InitEmptyExprContext(p *ExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = exprParserRULE_expr
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = exprParserRULE_expr

	return p
}

func (s *ExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprContext) IDENT() antlr.TerminalNode {
	return s.GetToken(exprParserIDENT, 0)
}

func (s *ExprContext) AllInnerExpr() []IInnerExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IInnerExprContext); ok {
			len++
		}
	}

	tst := make([]IInnerExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IInnerExprContext); ok {
			tst[i] = t.(IInnerExprContext)
			i++
		}
	}

	return tst
}

func (s *ExprContext) InnerExpr(i int) IInnerExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInnerExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInnerExprContext)
}

func (s *ExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(exprListener); ok {
		listenerT.EnterExpr(s)
	}
}

func (s *ExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(exprListener); ok {
		listenerT.ExitExpr(s)
	}
}

func (p *exprParser) Expr() (localctx IExprContext) {
	localctx = NewExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, exprParserRULE_expr)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(8)
		p.Match(exprParserIDENT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(9)
		p.Match(exprParserT__0)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(21)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == exprParserIDENT {
		{
			p.SetState(10)
			p.InnerExpr()
		}
		p.SetState(15)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(11)
					p.Match(exprParserT__1)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(12)
					p.InnerExpr()
				}

			}
			p.SetState(17)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}
		p.SetState(19)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == exprParserT__1 {
			{
				p.SetState(18)
				p.Match(exprParserT__1)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	}
	{
		p.SetState(23)
		p.Match(exprParserT__2)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IInnerExprContext is an interface to support dynamic dispatch.
type IInnerExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	FieldAccess() IFieldAccessContext
	Value() IValueContext

	// IsInnerExprContext differentiates from other interfaces.
	IsInnerExprContext()
}

type InnerExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInnerExprContext() *InnerExprContext {
	var p = new(InnerExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = exprParserRULE_innerExpr
	return p
}

func InitEmptyInnerExprContext(p *InnerExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = exprParserRULE_innerExpr
}

func (*InnerExprContext) IsInnerExprContext() {}

func NewInnerExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InnerExprContext {
	var p = new(InnerExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = exprParserRULE_innerExpr

	return p
}

func (s *InnerExprContext) GetParser() antlr.Parser { return s.parser }

func (s *InnerExprContext) FieldAccess() IFieldAccessContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFieldAccessContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFieldAccessContext)
}

func (s *InnerExprContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *InnerExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InnerExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InnerExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(exprListener); ok {
		listenerT.EnterInnerExpr(s)
	}
}

func (s *InnerExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(exprListener); ok {
		listenerT.ExitInnerExpr(s)
	}
}

func (p *exprParser) InnerExpr() (localctx IInnerExprContext) {
	localctx = NewInnerExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, exprParserRULE_innerExpr)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(25)
		p.FieldAccess()
	}
	{
		p.SetState(26)
		p.Match(exprParserT__3)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(27)
		p.Value()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFieldAccessContext is an interface to support dynamic dispatch.
type IFieldAccessContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIDENT() []antlr.TerminalNode
	IDENT(i int) antlr.TerminalNode
	AllINDEX() []antlr.TerminalNode
	INDEX(i int) antlr.TerminalNode

	// IsFieldAccessContext differentiates from other interfaces.
	IsFieldAccessContext()
}

type FieldAccessContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldAccessContext() *FieldAccessContext {
	var p = new(FieldAccessContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = exprParserRULE_fieldAccess
	return p
}

func InitEmptyFieldAccessContext(p *FieldAccessContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = exprParserRULE_fieldAccess
}

func (*FieldAccessContext) IsFieldAccessContext() {}

func NewFieldAccessContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldAccessContext {
	var p = new(FieldAccessContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = exprParserRULE_fieldAccess

	return p
}

func (s *FieldAccessContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldAccessContext) AllIDENT() []antlr.TerminalNode {
	return s.GetTokens(exprParserIDENT)
}

func (s *FieldAccessContext) IDENT(i int) antlr.TerminalNode {
	return s.GetToken(exprParserIDENT, i)
}

func (s *FieldAccessContext) AllINDEX() []antlr.TerminalNode {
	return s.GetTokens(exprParserINDEX)
}

func (s *FieldAccessContext) INDEX(i int) antlr.TerminalNode {
	return s.GetToken(exprParserINDEX, i)
}

func (s *FieldAccessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldAccessContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldAccessContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(exprListener); ok {
		listenerT.EnterFieldAccess(s)
	}
}

func (s *FieldAccessContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(exprListener); ok {
		listenerT.ExitFieldAccess(s)
	}
}

func (p *exprParser) FieldAccess() (localctx IFieldAccessContext) {
	localctx = NewFieldAccessContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, exprParserRULE_fieldAccess)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(29)
		p.Match(exprParserIDENT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(37)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == exprParserT__4 || _la == exprParserT__5 {
		p.SetState(35)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case exprParserT__4:
			{
				p.SetState(30)
				p.Match(exprParserT__4)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(31)
				p.Match(exprParserIDENT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case exprParserT__5:
			{
				p.SetState(32)
				p.Match(exprParserT__5)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(33)
				p.Match(exprParserINDEX)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(34)
				p.Match(exprParserT__6)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(39)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IValueContext is an interface to support dynamic dispatch.
type IValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING() antlr.TerminalNode
	Expr() IExprContext
	RAW_VALUE() antlr.TerminalNode

	// IsValueContext differentiates from other interfaces.
	IsValueContext()
}

type ValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueContext() *ValueContext {
	var p = new(ValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = exprParserRULE_value
	return p
}

func InitEmptyValueContext(p *ValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = exprParserRULE_value
}

func (*ValueContext) IsValueContext() {}

func NewValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueContext {
	var p = new(ValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = exprParserRULE_value

	return p
}

func (s *ValueContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueContext) STRING() antlr.TerminalNode {
	return s.GetToken(exprParserSTRING, 0)
}

func (s *ValueContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *ValueContext) RAW_VALUE() antlr.TerminalNode {
	return s.GetToken(exprParserRAW_VALUE, 0)
}

func (s *ValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(exprListener); ok {
		listenerT.EnterValue(s)
	}
}

func (s *ValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(exprListener); ok {
		listenerT.ExitValue(s)
	}
}

func (p *exprParser) Value() (localctx IValueContext) {
	localctx = NewValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, exprParserRULE_value)
	p.SetState(43)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case exprParserSTRING:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(40)
			p.Match(exprParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case exprParserIDENT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(41)
			p.Expr()
		}

	case exprParserRAW_VALUE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(42)
			p.Match(exprParserRAW_VALUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
