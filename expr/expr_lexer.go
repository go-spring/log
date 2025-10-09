// Code generated from Expr.g4 by ANTLR 4.13.2. DO NOT EDIT.

package expr

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type ExprLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var ExprLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func exprlexerLexerInit() {
	staticData := &ExprLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "'{'", "'}'", "','", "'='", "'.'", "'['", "']'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "IDENT", "INDEX", "STRING", "CONTINUOUS_VALUE",
		"WS",
	}
	staticData.RuleNames = []string{
		"T__0", "T__1", "T__2", "T__3", "T__4", "T__5", "T__6", "IDENT", "INDEX",
		"STRING", "CONTINUOUS_VALUE", "WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 12, 74, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 1, 0, 1, 0, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3,
		1, 4, 1, 4, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7, 1, 7, 5, 7, 42, 8, 7, 10, 7,
		12, 7, 45, 9, 7, 1, 8, 4, 8, 48, 8, 8, 11, 8, 12, 8, 49, 1, 9, 1, 9, 1,
		9, 1, 9, 5, 9, 56, 8, 9, 10, 9, 12, 9, 59, 9, 9, 1, 9, 1, 9, 1, 10, 4,
		10, 64, 8, 10, 11, 10, 12, 10, 65, 1, 11, 4, 11, 69, 8, 11, 11, 11, 12,
		11, 70, 1, 11, 1, 11, 0, 0, 12, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 13,
		7, 15, 8, 17, 9, 19, 10, 21, 11, 23, 12, 1, 0, 7, 3, 0, 65, 90, 95, 95,
		97, 122, 4, 0, 48, 57, 65, 90, 95, 95, 97, 122, 1, 0, 48, 57, 2, 0, 34,
		34, 92, 92, 8, 0, 34, 34, 47, 47, 92, 92, 98, 98, 102, 102, 110, 110, 114,
		114, 116, 116, 12, 0, 9, 10, 13, 13, 32, 32, 34, 34, 39, 39, 44, 44, 46,
		46, 61, 61, 91, 91, 93, 93, 123, 123, 125, 125, 3, 0, 9, 10, 13, 13, 32,
		32, 79, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1,
		0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0, 15,
		1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0, 0, 21, 1, 0, 0, 0, 0,
		23, 1, 0, 0, 0, 1, 25, 1, 0, 0, 0, 3, 27, 1, 0, 0, 0, 5, 29, 1, 0, 0, 0,
		7, 31, 1, 0, 0, 0, 9, 33, 1, 0, 0, 0, 11, 35, 1, 0, 0, 0, 13, 37, 1, 0,
		0, 0, 15, 39, 1, 0, 0, 0, 17, 47, 1, 0, 0, 0, 19, 51, 1, 0, 0, 0, 21, 63,
		1, 0, 0, 0, 23, 68, 1, 0, 0, 0, 25, 26, 5, 123, 0, 0, 26, 2, 1, 0, 0, 0,
		27, 28, 5, 125, 0, 0, 28, 4, 1, 0, 0, 0, 29, 30, 5, 44, 0, 0, 30, 6, 1,
		0, 0, 0, 31, 32, 5, 61, 0, 0, 32, 8, 1, 0, 0, 0, 33, 34, 5, 46, 0, 0, 34,
		10, 1, 0, 0, 0, 35, 36, 5, 91, 0, 0, 36, 12, 1, 0, 0, 0, 37, 38, 5, 93,
		0, 0, 38, 14, 1, 0, 0, 0, 39, 43, 7, 0, 0, 0, 40, 42, 7, 1, 0, 0, 41, 40,
		1, 0, 0, 0, 42, 45, 1, 0, 0, 0, 43, 41, 1, 0, 0, 0, 43, 44, 1, 0, 0, 0,
		44, 16, 1, 0, 0, 0, 45, 43, 1, 0, 0, 0, 46, 48, 7, 2, 0, 0, 47, 46, 1,
		0, 0, 0, 48, 49, 1, 0, 0, 0, 49, 47, 1, 0, 0, 0, 49, 50, 1, 0, 0, 0, 50,
		18, 1, 0, 0, 0, 51, 57, 5, 34, 0, 0, 52, 56, 8, 3, 0, 0, 53, 54, 5, 92,
		0, 0, 54, 56, 7, 4, 0, 0, 55, 52, 1, 0, 0, 0, 55, 53, 1, 0, 0, 0, 56, 59,
		1, 0, 0, 0, 57, 55, 1, 0, 0, 0, 57, 58, 1, 0, 0, 0, 58, 60, 1, 0, 0, 0,
		59, 57, 1, 0, 0, 0, 60, 61, 5, 34, 0, 0, 61, 20, 1, 0, 0, 0, 62, 64, 8,
		5, 0, 0, 63, 62, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 65,
		66, 1, 0, 0, 0, 66, 22, 1, 0, 0, 0, 67, 69, 7, 6, 0, 0, 68, 67, 1, 0, 0,
		0, 69, 70, 1, 0, 0, 0, 70, 68, 1, 0, 0, 0, 70, 71, 1, 0, 0, 0, 71, 72,
		1, 0, 0, 0, 72, 73, 6, 11, 0, 0, 73, 24, 1, 0, 0, 0, 7, 0, 43, 49, 55,
		57, 65, 70, 1, 6, 0, 0,
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

// ExprLexerInit initializes any static state used to implement ExprLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewExprLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func ExprLexerInit() {
	staticData := &ExprLexerLexerStaticData
	staticData.once.Do(exprlexerLexerInit)
}

// NewExprLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewExprLexer(input antlr.CharStream) *ExprLexer {
	ExprLexerInit()
	l := new(ExprLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &ExprLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "Expr.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// ExprLexer tokens.
const (
	ExprLexerT__0             = 1
	ExprLexerT__1             = 2
	ExprLexerT__2             = 3
	ExprLexerT__3             = 4
	ExprLexerT__4             = 5
	ExprLexerT__5             = 6
	ExprLexerT__6             = 7
	ExprLexerIDENT            = 8
	ExprLexerINDEX            = 9
	ExprLexerSTRING           = 10
	ExprLexerCONTINUOUS_VALUE = 11
	ExprLexerWS               = 12
)
