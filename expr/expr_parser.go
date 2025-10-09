// Code generated from Expr.g4 by ANTLR 4.13.2. DO NOT EDIT.

package expr // Expr
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

type ExprParser struct {
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
		"", "'{'", "'}'", "','", "'='", "'.'", "'['", "']'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "IDENT", "INDEX", "STRING", "RAW_VALUE",
		"WS",
	}
	staticData.RuleNames = []string{
		"expr", "innerExprList", "innerExpr", "fieldAccess", "value",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 12, 49, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 1, 0, 1, 0, 1, 0, 3, 0, 14, 8, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 5, 1,
		21, 8, 1, 10, 1, 12, 1, 24, 9, 1, 1, 1, 3, 1, 27, 8, 1, 1, 2, 1, 2, 1,
		2, 1, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 5, 3, 39, 8, 3, 10, 3, 12,
		3, 42, 9, 3, 1, 4, 1, 4, 1, 4, 3, 4, 47, 8, 4, 1, 4, 0, 0, 5, 0, 2, 4,
		6, 8, 0, 0, 50, 0, 10, 1, 0, 0, 0, 2, 17, 1, 0, 0, 0, 4, 28, 1, 0, 0, 0,
		6, 32, 1, 0, 0, 0, 8, 46, 1, 0, 0, 0, 10, 11, 5, 8, 0, 0, 11, 13, 5, 1,
		0, 0, 12, 14, 3, 2, 1, 0, 13, 12, 1, 0, 0, 0, 13, 14, 1, 0, 0, 0, 14, 15,
		1, 0, 0, 0, 15, 16, 5, 2, 0, 0, 16, 1, 1, 0, 0, 0, 17, 22, 3, 4, 2, 0,
		18, 19, 5, 3, 0, 0, 19, 21, 3, 4, 2, 0, 20, 18, 1, 0, 0, 0, 21, 24, 1,
		0, 0, 0, 22, 20, 1, 0, 0, 0, 22, 23, 1, 0, 0, 0, 23, 26, 1, 0, 0, 0, 24,
		22, 1, 0, 0, 0, 25, 27, 5, 3, 0, 0, 26, 25, 1, 0, 0, 0, 26, 27, 1, 0, 0,
		0, 27, 3, 1, 0, 0, 0, 28, 29, 3, 6, 3, 0, 29, 30, 5, 4, 0, 0, 30, 31, 3,
		8, 4, 0, 31, 5, 1, 0, 0, 0, 32, 40, 5, 8, 0, 0, 33, 34, 5, 5, 0, 0, 34,
		39, 5, 8, 0, 0, 35, 36, 5, 6, 0, 0, 36, 37, 5, 9, 0, 0, 37, 39, 5, 7, 0,
		0, 38, 33, 1, 0, 0, 0, 38, 35, 1, 0, 0, 0, 39, 42, 1, 0, 0, 0, 40, 38,
		1, 0, 0, 0, 40, 41, 1, 0, 0, 0, 41, 7, 1, 0, 0, 0, 42, 40, 1, 0, 0, 0,
		43, 47, 5, 10, 0, 0, 44, 47, 3, 0, 0, 0, 45, 47, 5, 11, 0, 0, 46, 43, 1,
		0, 0, 0, 46, 44, 1, 0, 0, 0, 46, 45, 1, 0, 0, 0, 47, 9, 1, 0, 0, 0, 6,
		13, 22, 26, 38, 40, 46,
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

// ExprParserInit initializes any static state used to implement ExprParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewExprParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func ExprParserInit() {
	staticData := &ExprParserStaticData
	staticData.once.Do(exprParserInit)
}

// NewExprParser produces a new parser instance for the optional input antlr.TokenStream.
func NewExprParser(input antlr.TokenStream) *ExprParser {
	ExprParserInit()
	this := new(ExprParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &ExprParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Expr.g4"

	return this
}

// ExprParser tokens.
const (
	ExprParserEOF       = antlr.TokenEOF
	ExprParserT__0      = 1
	ExprParserT__1      = 2
	ExprParserT__2      = 3
	ExprParserT__3      = 4
	ExprParserT__4      = 5
	ExprParserT__5      = 6
	ExprParserT__6      = 7
	ExprParserIDENT     = 8
	ExprParserINDEX     = 9
	ExprParserSTRING    = 10
	ExprParserRAW_VALUE = 11
	ExprParserWS        = 12
)

// ExprParser rules.
const (
	ExprParserRULE_expr          = 0
	ExprParserRULE_innerExprList = 1
	ExprParserRULE_innerExpr     = 2
	ExprParserRULE_fieldAccess   = 3
	ExprParserRULE_value         = 4
)

// IExprContext is an interface to support dynamic dispatch.
type IExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENT() antlr.TerminalNode
	InnerExprList() IInnerExprListContext

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
	p.RuleIndex = ExprParserRULE_expr
	return p
}

func InitEmptyExprContext(p *ExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ExprParserRULE_expr
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ExprParserRULE_expr

	return p
}

func (s *ExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprContext) IDENT() antlr.TerminalNode {
	return s.GetToken(ExprParserIDENT, 0)
}

func (s *ExprContext) InnerExprList() IInnerExprListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInnerExprListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInnerExprListContext)
}

func (s *ExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.EnterExpr(s)
	}
}

func (s *ExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.ExitExpr(s)
	}
}

func (p *ExprParser) Expr() (localctx IExprContext) {
	localctx = NewExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ExprParserRULE_expr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(10)
		p.Match(ExprParserIDENT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(11)
		p.Match(ExprParserT__0)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(13)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == ExprParserIDENT {
		{
			p.SetState(12)
			p.InnerExprList()
		}

	}
	{
		p.SetState(15)
		p.Match(ExprParserT__1)
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

// IInnerExprListContext is an interface to support dynamic dispatch.
type IInnerExprListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllInnerExpr() []IInnerExprContext
	InnerExpr(i int) IInnerExprContext

	// IsInnerExprListContext differentiates from other interfaces.
	IsInnerExprListContext()
}

type InnerExprListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInnerExprListContext() *InnerExprListContext {
	var p = new(InnerExprListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ExprParserRULE_innerExprList
	return p
}

func InitEmptyInnerExprListContext(p *InnerExprListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ExprParserRULE_innerExprList
}

func (*InnerExprListContext) IsInnerExprListContext() {}

func NewInnerExprListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InnerExprListContext {
	var p = new(InnerExprListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ExprParserRULE_innerExprList

	return p
}

func (s *InnerExprListContext) GetParser() antlr.Parser { return s.parser }

func (s *InnerExprListContext) AllInnerExpr() []IInnerExprContext {
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

func (s *InnerExprListContext) InnerExpr(i int) IInnerExprContext {
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

func (s *InnerExprListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InnerExprListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InnerExprListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.EnterInnerExprList(s)
	}
}

func (s *InnerExprListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.ExitInnerExprList(s)
	}
}

func (p *ExprParser) InnerExprList() (localctx IInnerExprListContext) {
	localctx = NewInnerExprListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, ExprParserRULE_innerExprList)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(17)
		p.InnerExpr()
	}
	p.SetState(22)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(18)
				p.Match(ExprParserT__2)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(19)
				p.InnerExpr()
			}

		}
		p.SetState(24)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(26)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == ExprParserT__2 {
		{
			p.SetState(25)
			p.Match(ExprParserT__2)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
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
	p.RuleIndex = ExprParserRULE_innerExpr
	return p
}

func InitEmptyInnerExprContext(p *InnerExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ExprParserRULE_innerExpr
}

func (*InnerExprContext) IsInnerExprContext() {}

func NewInnerExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InnerExprContext {
	var p = new(InnerExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ExprParserRULE_innerExpr

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
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.EnterInnerExpr(s)
	}
}

func (s *InnerExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.ExitInnerExpr(s)
	}
}

func (p *ExprParser) InnerExpr() (localctx IInnerExprContext) {
	localctx = NewInnerExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, ExprParserRULE_innerExpr)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(28)
		p.FieldAccess()
	}
	{
		p.SetState(29)
		p.Match(ExprParserT__3)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(30)
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
	p.RuleIndex = ExprParserRULE_fieldAccess
	return p
}

func InitEmptyFieldAccessContext(p *FieldAccessContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ExprParserRULE_fieldAccess
}

func (*FieldAccessContext) IsFieldAccessContext() {}

func NewFieldAccessContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldAccessContext {
	var p = new(FieldAccessContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ExprParserRULE_fieldAccess

	return p
}

func (s *FieldAccessContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldAccessContext) AllIDENT() []antlr.TerminalNode {
	return s.GetTokens(ExprParserIDENT)
}

func (s *FieldAccessContext) IDENT(i int) antlr.TerminalNode {
	return s.GetToken(ExprParserIDENT, i)
}

func (s *FieldAccessContext) AllINDEX() []antlr.TerminalNode {
	return s.GetTokens(ExprParserINDEX)
}

func (s *FieldAccessContext) INDEX(i int) antlr.TerminalNode {
	return s.GetToken(ExprParserINDEX, i)
}

func (s *FieldAccessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldAccessContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldAccessContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.EnterFieldAccess(s)
	}
}

func (s *FieldAccessContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.ExitFieldAccess(s)
	}
}

func (p *ExprParser) FieldAccess() (localctx IFieldAccessContext) {
	localctx = NewFieldAccessContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, ExprParserRULE_fieldAccess)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(32)
		p.Match(ExprParserIDENT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(40)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ExprParserT__4 || _la == ExprParserT__5 {
		p.SetState(38)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case ExprParserT__4:
			{
				p.SetState(33)
				p.Match(ExprParserT__4)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(34)
				p.Match(ExprParserIDENT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case ExprParserT__5:
			{
				p.SetState(35)
				p.Match(ExprParserT__5)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(36)
				p.Match(ExprParserINDEX)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(37)
				p.Match(ExprParserT__6)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(42)
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
	p.RuleIndex = ExprParserRULE_value
	return p
}

func InitEmptyValueContext(p *ValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ExprParserRULE_value
}

func (*ValueContext) IsValueContext() {}

func NewValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueContext {
	var p = new(ValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ExprParserRULE_value

	return p
}

func (s *ValueContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueContext) STRING() antlr.TerminalNode {
	return s.GetToken(ExprParserSTRING, 0)
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
	return s.GetToken(ExprParserRAW_VALUE, 0)
}

func (s *ValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.EnterValue(s)
	}
}

func (s *ValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ExprListener); ok {
		listenerT.ExitValue(s)
	}
}

func (p *ExprParser) Value() (localctx IValueContext) {
	localctx = NewValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, ExprParserRULE_value)
	p.SetState(46)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ExprParserSTRING:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(43)
			p.Match(ExprParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ExprParserIDENT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(44)
			p.Expr()
		}

	case ExprParserRAW_VALUE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(45)
			p.Match(ExprParserRAW_VALUE)
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
