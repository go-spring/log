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
		"", "", "", "", "", "", "", "", "IDENT", "INDEX", "STRING", "INTEGER",
		"FLOAT", "WS",
	}
	staticData.RuleNames = []string{
		"T__0", "T__1", "T__2", "T__3", "T__4", "T__5", "T__6", "IDENT", "INDEX",
		"STRING", "INTEGER", "FLOAT", "DIGIT", "LETTER", "HEX_DIGIT", "WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 13, 138, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 1, 0, 1, 0, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 4, 1, 4, 1, 5,
		1, 5, 1, 6, 1, 6, 1, 7, 1, 7, 5, 7, 50, 8, 7, 10, 7, 12, 7, 53, 9, 7, 1,
		8, 4, 8, 56, 8, 8, 11, 8, 12, 8, 57, 1, 9, 1, 9, 1, 9, 1, 9, 5, 9, 64,
		8, 9, 10, 9, 12, 9, 67, 9, 9, 1, 9, 1, 9, 1, 10, 3, 10, 72, 8, 10, 1, 10,
		4, 10, 75, 8, 10, 11, 10, 12, 10, 76, 1, 10, 1, 10, 1, 10, 1, 10, 4, 10,
		83, 8, 10, 11, 10, 12, 10, 84, 3, 10, 87, 8, 10, 1, 11, 3, 11, 90, 8, 11,
		1, 11, 4, 11, 93, 8, 11, 11, 11, 12, 11, 94, 1, 11, 1, 11, 4, 11, 99, 8,
		11, 11, 11, 12, 11, 100, 3, 11, 103, 8, 11, 1, 11, 1, 11, 4, 11, 107, 8,
		11, 11, 11, 12, 11, 108, 3, 11, 111, 8, 11, 1, 11, 1, 11, 3, 11, 115, 8,
		11, 1, 11, 4, 11, 118, 8, 11, 11, 11, 12, 11, 119, 3, 11, 122, 8, 11, 1,
		12, 1, 12, 1, 13, 1, 13, 1, 14, 1, 14, 3, 14, 130, 8, 14, 1, 15, 4, 15,
		133, 8, 15, 11, 15, 12, 15, 134, 1, 15, 1, 15, 0, 0, 16, 1, 1, 3, 2, 5,
		3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19, 10, 21, 11, 23, 12, 25,
		0, 27, 0, 29, 0, 31, 13, 1, 0, 10, 3, 0, 65, 90, 95, 95, 97, 122, 4, 0,
		48, 57, 65, 90, 95, 95, 97, 122, 1, 0, 48, 57, 2, 0, 34, 34, 92, 92, 8,
		0, 34, 34, 47, 47, 92, 92, 98, 98, 102, 102, 110, 110, 114, 114, 116, 116,
		2, 0, 43, 43, 45, 45, 2, 0, 69, 69, 101, 101, 2, 0, 65, 90, 97, 122, 2,
		0, 65, 70, 97, 102, 3, 0, 9, 10, 13, 13, 32, 32, 153, 0, 1, 1, 0, 0, 0,
		0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0,
		0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0,
		0, 0, 19, 1, 0, 0, 0, 0, 21, 1, 0, 0, 0, 0, 23, 1, 0, 0, 0, 0, 31, 1, 0,
		0, 0, 1, 33, 1, 0, 0, 0, 3, 35, 1, 0, 0, 0, 5, 37, 1, 0, 0, 0, 7, 39, 1,
		0, 0, 0, 9, 41, 1, 0, 0, 0, 11, 43, 1, 0, 0, 0, 13, 45, 1, 0, 0, 0, 15,
		47, 1, 0, 0, 0, 17, 55, 1, 0, 0, 0, 19, 59, 1, 0, 0, 0, 21, 86, 1, 0, 0,
		0, 23, 89, 1, 0, 0, 0, 25, 123, 1, 0, 0, 0, 27, 125, 1, 0, 0, 0, 29, 129,
		1, 0, 0, 0, 31, 132, 1, 0, 0, 0, 33, 34, 5, 123, 0, 0, 34, 2, 1, 0, 0,
		0, 35, 36, 5, 125, 0, 0, 36, 4, 1, 0, 0, 0, 37, 38, 5, 44, 0, 0, 38, 6,
		1, 0, 0, 0, 39, 40, 5, 61, 0, 0, 40, 8, 1, 0, 0, 0, 41, 42, 5, 46, 0, 0,
		42, 10, 1, 0, 0, 0, 43, 44, 5, 91, 0, 0, 44, 12, 1, 0, 0, 0, 45, 46, 5,
		93, 0, 0, 46, 14, 1, 0, 0, 0, 47, 51, 7, 0, 0, 0, 48, 50, 7, 1, 0, 0, 49,
		48, 1, 0, 0, 0, 50, 53, 1, 0, 0, 0, 51, 49, 1, 0, 0, 0, 51, 52, 1, 0, 0,
		0, 52, 16, 1, 0, 0, 0, 53, 51, 1, 0, 0, 0, 54, 56, 7, 2, 0, 0, 55, 54,
		1, 0, 0, 0, 56, 57, 1, 0, 0, 0, 57, 55, 1, 0, 0, 0, 57, 58, 1, 0, 0, 0,
		58, 18, 1, 0, 0, 0, 59, 65, 5, 34, 0, 0, 60, 64, 8, 3, 0, 0, 61, 62, 5,
		92, 0, 0, 62, 64, 7, 4, 0, 0, 63, 60, 1, 0, 0, 0, 63, 61, 1, 0, 0, 0, 64,
		67, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 65, 66, 1, 0, 0, 0, 66, 68, 1, 0, 0,
		0, 67, 65, 1, 0, 0, 0, 68, 69, 5, 34, 0, 0, 69, 20, 1, 0, 0, 0, 70, 72,
		7, 5, 0, 0, 71, 70, 1, 0, 0, 0, 71, 72, 1, 0, 0, 0, 72, 74, 1, 0, 0, 0,
		73, 75, 3, 25, 12, 0, 74, 73, 1, 0, 0, 0, 75, 76, 1, 0, 0, 0, 76, 74, 1,
		0, 0, 0, 76, 77, 1, 0, 0, 0, 77, 87, 1, 0, 0, 0, 78, 79, 5, 48, 0, 0, 79,
		80, 5, 120, 0, 0, 80, 82, 1, 0, 0, 0, 81, 83, 3, 29, 14, 0, 82, 81, 1,
		0, 0, 0, 83, 84, 1, 0, 0, 0, 84, 82, 1, 0, 0, 0, 84, 85, 1, 0, 0, 0, 85,
		87, 1, 0, 0, 0, 86, 71, 1, 0, 0, 0, 86, 78, 1, 0, 0, 0, 87, 22, 1, 0, 0,
		0, 88, 90, 7, 5, 0, 0, 89, 88, 1, 0, 0, 0, 89, 90, 1, 0, 0, 0, 90, 110,
		1, 0, 0, 0, 91, 93, 3, 25, 12, 0, 92, 91, 1, 0, 0, 0, 93, 94, 1, 0, 0,
		0, 94, 92, 1, 0, 0, 0, 94, 95, 1, 0, 0, 0, 95, 102, 1, 0, 0, 0, 96, 98,
		5, 46, 0, 0, 97, 99, 3, 25, 12, 0, 98, 97, 1, 0, 0, 0, 99, 100, 1, 0, 0,
		0, 100, 98, 1, 0, 0, 0, 100, 101, 1, 0, 0, 0, 101, 103, 1, 0, 0, 0, 102,
		96, 1, 0, 0, 0, 102, 103, 1, 0, 0, 0, 103, 111, 1, 0, 0, 0, 104, 106, 5,
		46, 0, 0, 105, 107, 3, 25, 12, 0, 106, 105, 1, 0, 0, 0, 107, 108, 1, 0,
		0, 0, 108, 106, 1, 0, 0, 0, 108, 109, 1, 0, 0, 0, 109, 111, 1, 0, 0, 0,
		110, 92, 1, 0, 0, 0, 110, 104, 1, 0, 0, 0, 111, 121, 1, 0, 0, 0, 112, 114,
		7, 6, 0, 0, 113, 115, 7, 5, 0, 0, 114, 113, 1, 0, 0, 0, 114, 115, 1, 0,
		0, 0, 115, 117, 1, 0, 0, 0, 116, 118, 3, 25, 12, 0, 117, 116, 1, 0, 0,
		0, 118, 119, 1, 0, 0, 0, 119, 117, 1, 0, 0, 0, 119, 120, 1, 0, 0, 0, 120,
		122, 1, 0, 0, 0, 121, 112, 1, 0, 0, 0, 121, 122, 1, 0, 0, 0, 122, 24, 1,
		0, 0, 0, 123, 124, 2, 48, 57, 0, 124, 26, 1, 0, 0, 0, 125, 126, 7, 7, 0,
		0, 126, 28, 1, 0, 0, 0, 127, 130, 3, 25, 12, 0, 128, 130, 7, 8, 0, 0, 129,
		127, 1, 0, 0, 0, 129, 128, 1, 0, 0, 0, 130, 30, 1, 0, 0, 0, 131, 133, 7,
		9, 0, 0, 132, 131, 1, 0, 0, 0, 133, 134, 1, 0, 0, 0, 134, 132, 1, 0, 0,
		0, 134, 135, 1, 0, 0, 0, 135, 136, 1, 0, 0, 0, 136, 137, 6, 15, 0, 0, 137,
		32, 1, 0, 0, 0, 20, 0, 51, 57, 63, 65, 71, 76, 84, 86, 89, 94, 100, 102,
		108, 110, 114, 119, 121, 129, 134, 1, 6, 0, 0,
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
	ExprLexerT__0    = 1
	ExprLexerT__1    = 2
	ExprLexerT__2    = 3
	ExprLexerT__3    = 4
	ExprLexerT__4    = 5
	ExprLexerT__5    = 6
	ExprLexerT__6    = 7
	ExprLexerIDENT   = 8
	ExprLexerINDEX   = 9
	ExprLexerSTRING  = 10
	ExprLexerINTEGER = 11
	ExprLexerFLOAT   = 12
	ExprLexerWS      = 13
)
