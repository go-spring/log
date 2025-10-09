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
		"", "'{'", "'}'", "','", "'='", "'.'", "'['", "']'", "'true'", "'false'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "", "", "", "KW_TRUE", "KW_FALSE", "IDENT", "INDEX",
		"STRING", "INTEGER", "FLOAT", "WS",
	}
	staticData.RuleNames = []string{
		"T__0", "T__1", "T__2", "T__3", "T__4", "T__5", "T__6", "KW_TRUE", "KW_FALSE",
		"IDENT", "INDEX", "STRING", "INTEGER", "FLOAT", "DIGIT", "LETTER", "HEX_DIGIT",
		"WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 15, 153, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 1, 0, 1, 0, 1, 1, 1, 1, 1, 2, 1, 2,
		1, 3, 1, 3, 1, 4, 1, 4, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 7,
		1, 7, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 5, 9, 65, 8, 9, 10,
		9, 12, 9, 68, 9, 9, 1, 10, 4, 10, 71, 8, 10, 11, 10, 12, 10, 72, 1, 11,
		1, 11, 1, 11, 1, 11, 5, 11, 79, 8, 11, 10, 11, 12, 11, 82, 9, 11, 1, 11,
		1, 11, 1, 12, 3, 12, 87, 8, 12, 1, 12, 4, 12, 90, 8, 12, 11, 12, 12, 12,
		91, 1, 12, 1, 12, 1, 12, 1, 12, 4, 12, 98, 8, 12, 11, 12, 12, 12, 99, 3,
		12, 102, 8, 12, 1, 13, 3, 13, 105, 8, 13, 1, 13, 4, 13, 108, 8, 13, 11,
		13, 12, 13, 109, 1, 13, 1, 13, 4, 13, 114, 8, 13, 11, 13, 12, 13, 115,
		3, 13, 118, 8, 13, 1, 13, 1, 13, 4, 13, 122, 8, 13, 11, 13, 12, 13, 123,
		3, 13, 126, 8, 13, 1, 13, 1, 13, 3, 13, 130, 8, 13, 1, 13, 4, 13, 133,
		8, 13, 11, 13, 12, 13, 134, 3, 13, 137, 8, 13, 1, 14, 1, 14, 1, 15, 1,
		15, 1, 16, 1, 16, 3, 16, 145, 8, 16, 1, 17, 4, 17, 148, 8, 17, 11, 17,
		12, 17, 149, 1, 17, 1, 17, 0, 0, 18, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11,
		6, 13, 7, 15, 8, 17, 9, 19, 10, 21, 11, 23, 12, 25, 13, 27, 14, 29, 0,
		31, 0, 33, 0, 35, 15, 1, 0, 10, 3, 0, 65, 90, 95, 95, 97, 122, 4, 0, 48,
		57, 65, 90, 95, 95, 97, 122, 1, 0, 48, 57, 2, 0, 34, 34, 92, 92, 8, 0,
		34, 34, 47, 47, 92, 92, 98, 98, 102, 102, 110, 110, 114, 114, 116, 116,
		2, 0, 43, 43, 45, 45, 2, 0, 69, 69, 101, 101, 2, 0, 65, 90, 97, 122, 2,
		0, 65, 70, 97, 102, 3, 0, 9, 10, 13, 13, 32, 32, 168, 0, 1, 1, 0, 0, 0,
		0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0,
		0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0,
		0, 0, 19, 1, 0, 0, 0, 0, 21, 1, 0, 0, 0, 0, 23, 1, 0, 0, 0, 0, 25, 1, 0,
		0, 0, 0, 27, 1, 0, 0, 0, 0, 35, 1, 0, 0, 0, 1, 37, 1, 0, 0, 0, 3, 39, 1,
		0, 0, 0, 5, 41, 1, 0, 0, 0, 7, 43, 1, 0, 0, 0, 9, 45, 1, 0, 0, 0, 11, 47,
		1, 0, 0, 0, 13, 49, 1, 0, 0, 0, 15, 51, 1, 0, 0, 0, 17, 56, 1, 0, 0, 0,
		19, 62, 1, 0, 0, 0, 21, 70, 1, 0, 0, 0, 23, 74, 1, 0, 0, 0, 25, 101, 1,
		0, 0, 0, 27, 104, 1, 0, 0, 0, 29, 138, 1, 0, 0, 0, 31, 140, 1, 0, 0, 0,
		33, 144, 1, 0, 0, 0, 35, 147, 1, 0, 0, 0, 37, 38, 5, 123, 0, 0, 38, 2,
		1, 0, 0, 0, 39, 40, 5, 125, 0, 0, 40, 4, 1, 0, 0, 0, 41, 42, 5, 44, 0,
		0, 42, 6, 1, 0, 0, 0, 43, 44, 5, 61, 0, 0, 44, 8, 1, 0, 0, 0, 45, 46, 5,
		46, 0, 0, 46, 10, 1, 0, 0, 0, 47, 48, 5, 91, 0, 0, 48, 12, 1, 0, 0, 0,
		49, 50, 5, 93, 0, 0, 50, 14, 1, 0, 0, 0, 51, 52, 5, 116, 0, 0, 52, 53,
		5, 114, 0, 0, 53, 54, 5, 117, 0, 0, 54, 55, 5, 101, 0, 0, 55, 16, 1, 0,
		0, 0, 56, 57, 5, 102, 0, 0, 57, 58, 5, 97, 0, 0, 58, 59, 5, 108, 0, 0,
		59, 60, 5, 115, 0, 0, 60, 61, 5, 101, 0, 0, 61, 18, 1, 0, 0, 0, 62, 66,
		7, 0, 0, 0, 63, 65, 7, 1, 0, 0, 64, 63, 1, 0, 0, 0, 65, 68, 1, 0, 0, 0,
		66, 64, 1, 0, 0, 0, 66, 67, 1, 0, 0, 0, 67, 20, 1, 0, 0, 0, 68, 66, 1,
		0, 0, 0, 69, 71, 7, 2, 0, 0, 70, 69, 1, 0, 0, 0, 71, 72, 1, 0, 0, 0, 72,
		70, 1, 0, 0, 0, 72, 73, 1, 0, 0, 0, 73, 22, 1, 0, 0, 0, 74, 80, 5, 34,
		0, 0, 75, 79, 8, 3, 0, 0, 76, 77, 5, 92, 0, 0, 77, 79, 7, 4, 0, 0, 78,
		75, 1, 0, 0, 0, 78, 76, 1, 0, 0, 0, 79, 82, 1, 0, 0, 0, 80, 78, 1, 0, 0,
		0, 80, 81, 1, 0, 0, 0, 81, 83, 1, 0, 0, 0, 82, 80, 1, 0, 0, 0, 83, 84,
		5, 34, 0, 0, 84, 24, 1, 0, 0, 0, 85, 87, 7, 5, 0, 0, 86, 85, 1, 0, 0, 0,
		86, 87, 1, 0, 0, 0, 87, 89, 1, 0, 0, 0, 88, 90, 3, 29, 14, 0, 89, 88, 1,
		0, 0, 0, 90, 91, 1, 0, 0, 0, 91, 89, 1, 0, 0, 0, 91, 92, 1, 0, 0, 0, 92,
		102, 1, 0, 0, 0, 93, 94, 5, 48, 0, 0, 94, 95, 5, 120, 0, 0, 95, 97, 1,
		0, 0, 0, 96, 98, 3, 33, 16, 0, 97, 96, 1, 0, 0, 0, 98, 99, 1, 0, 0, 0,
		99, 97, 1, 0, 0, 0, 99, 100, 1, 0, 0, 0, 100, 102, 1, 0, 0, 0, 101, 86,
		1, 0, 0, 0, 101, 93, 1, 0, 0, 0, 102, 26, 1, 0, 0, 0, 103, 105, 7, 5, 0,
		0, 104, 103, 1, 0, 0, 0, 104, 105, 1, 0, 0, 0, 105, 125, 1, 0, 0, 0, 106,
		108, 3, 29, 14, 0, 107, 106, 1, 0, 0, 0, 108, 109, 1, 0, 0, 0, 109, 107,
		1, 0, 0, 0, 109, 110, 1, 0, 0, 0, 110, 117, 1, 0, 0, 0, 111, 113, 5, 46,
		0, 0, 112, 114, 3, 29, 14, 0, 113, 112, 1, 0, 0, 0, 114, 115, 1, 0, 0,
		0, 115, 113, 1, 0, 0, 0, 115, 116, 1, 0, 0, 0, 116, 118, 1, 0, 0, 0, 117,
		111, 1, 0, 0, 0, 117, 118, 1, 0, 0, 0, 118, 126, 1, 0, 0, 0, 119, 121,
		5, 46, 0, 0, 120, 122, 3, 29, 14, 0, 121, 120, 1, 0, 0, 0, 122, 123, 1,
		0, 0, 0, 123, 121, 1, 0, 0, 0, 123, 124, 1, 0, 0, 0, 124, 126, 1, 0, 0,
		0, 125, 107, 1, 0, 0, 0, 125, 119, 1, 0, 0, 0, 126, 136, 1, 0, 0, 0, 127,
		129, 7, 6, 0, 0, 128, 130, 7, 5, 0, 0, 129, 128, 1, 0, 0, 0, 129, 130,
		1, 0, 0, 0, 130, 132, 1, 0, 0, 0, 131, 133, 3, 29, 14, 0, 132, 131, 1,
		0, 0, 0, 133, 134, 1, 0, 0, 0, 134, 132, 1, 0, 0, 0, 134, 135, 1, 0, 0,
		0, 135, 137, 1, 0, 0, 0, 136, 127, 1, 0, 0, 0, 136, 137, 1, 0, 0, 0, 137,
		28, 1, 0, 0, 0, 138, 139, 2, 48, 57, 0, 139, 30, 1, 0, 0, 0, 140, 141,
		7, 7, 0, 0, 141, 32, 1, 0, 0, 0, 142, 145, 3, 29, 14, 0, 143, 145, 7, 8,
		0, 0, 144, 142, 1, 0, 0, 0, 144, 143, 1, 0, 0, 0, 145, 34, 1, 0, 0, 0,
		146, 148, 7, 9, 0, 0, 147, 146, 1, 0, 0, 0, 148, 149, 1, 0, 0, 0, 149,
		147, 1, 0, 0, 0, 149, 150, 1, 0, 0, 0, 150, 151, 1, 0, 0, 0, 151, 152,
		6, 17, 0, 0, 152, 36, 1, 0, 0, 0, 20, 0, 66, 72, 78, 80, 86, 91, 99, 101,
		104, 109, 115, 117, 123, 125, 129, 134, 136, 144, 149, 1, 6, 0, 0,
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
	ExprLexerT__0     = 1
	ExprLexerT__1     = 2
	ExprLexerT__2     = 3
	ExprLexerT__3     = 4
	ExprLexerT__4     = 5
	ExprLexerT__5     = 6
	ExprLexerT__6     = 7
	ExprLexerKW_TRUE  = 8
	ExprLexerKW_FALSE = 9
	ExprLexerIDENT    = 10
	ExprLexerINDEX    = 11
	ExprLexerSTRING   = 12
	ExprLexerINTEGER  = 13
	ExprLexerFLOAT    = 14
	ExprLexerWS       = 15
)
