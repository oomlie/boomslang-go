package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode"
)

type TokenType string

const (
	TOKEN_NEWLINE      TokenType = "TOKEN_NEWLINE"
	TOKEN_EOF                    = "TOKEN_EOF"
	TOKEN_BEGIN_INDENT           = "TOKEN_BEGIN_INDENT"
	TOKEN_END_INDENT             = "TOKEN_END_INDENT"
	TOKEN_NUMBER                 = "TOKEN_NUMBER"
	TOKEN_WORD                   = "TOKEN_WORD"
	TOKEN_TEXT                   = "TOKEN_TEXT"
	TOKEN_KW_IS                  = "TOKEN_KW_IS"
	TOKEN_KW_THE                 = "TOKEN_KW_THE"
	TOKEN_KW_FALSE               = "TOKEN_KW_FALSE"
	TOKEN_KW_TRUE                = "TOKEN_KW_TRUE"
	TOKEN_KW_THAT                = "TOKEN_KW_THAT"
	TOKEN_KW_OF                  = "TOKEN_KW_OF"
	TOKEN_KW_IF                  = "TOKEN_KW_IF"
	TOKEN_KW_OTHERWISE           = "TOKEN_KW_OTHERWISE"
	TOKEN_KW_OTIF                = "TOKEN_KW_OTIF"
	TOKEN_KW_FOR                 = "TOKEN_KW_FOR"
	TOKEN_KW_WHILE               = "TOKEN_KW_WHILE"
	TOKEN_KW_BREAK               = "TOKEN_KW_BREAK"
	TOKEN_KW_BY                  = "TOKEN_KW_BY"
	TOKEN_KW_WE_MEAN             = "TOKEN_KW_WE_MEAN"
	TOKEN_KW_RETURNS             = "TOKEN_KW_RETURNS"
)

type Source interface {
	Name() string;
	ReadLine() (string, error);
}

type Span struct {
	SourceName string
	Lineno     int
	Begin      int
	End        int
}

type Token struct {
	Ty  TokenType
	Lex string
	Spn Span
}

type Lexer struct {
	debug      bool
	source     Source
	lineno     int
	indent     int
	shiftWidth indentlevel
}

func MakeLexer(opts *Opts, source Source) *Lexer {
	lexer := new(Lexer)
	lexer.source = source	
	lexer.debug = opts.debug&DBG_LEX > 0
	return lexer
}

func (l *Lexer) makeToken(lex string, ty TokenType) Token {
	span := Span{
		SourceName: l.source.Name(),
		Lineno:     l.lineno,
	}
	tok := Token{ty, lex, span}
	return tok
}

func (l *Lexer) Lex() ([]Token, error) {
	tokens := make([]Token, 0, 50)
	for {
		newTokens, err := l.LexLine()
		if err != nil {
			return nil, err
		}
		for _, t := range newTokens {
			tokens = append(tokens, t)
		}
		if tokens[len(tokens)-1].Ty == TOKEN_EOF {
			break
		}
	}
	// emit dedents down to zero
	indentTokens, err := l.handleIndent(indentlevel{})
	if err != nil {
		return nil, err
	}
	for _, tok := range indentTokens {
		tokens = append(tokens, tok)
	}

	if l.debug {
		log.Printf("tokens = %#v\n", tokens)
	}

	return tokens, nil
}
func (l *Lexer) LexLine() ([]Token, error) {
	if l.debug {
		log.Printf("entering lexLine\n")
	}
	l.lineno += 1
	tokens := make([]Token, 0, 7)
	line, err := l.source.ReadLine()
	if err == io.EOF {
		tokens = append(tokens, l.makeToken("", TOKEN_EOF))
		return tokens, nil
	} else if err != nil {
		return tokens,err
	}

	if l.debug {
		log.Printf(" line no [%d] = %#v\n", l.lineno, line)
	}

	// emit indents
	line, indent := TrimIndent(line)
	indentTokens, err := l.handleIndent(indent)
	if err != nil {
		return tokens, err
	}
	for _, tok := range indentTokens {
		tokens = append(tokens, tok)
	}

	// word to token
	line = strings.TrimSpace(line)
	words := strings.Fields(line)
	// not using range so we can consume multi word tokens
	for i := 0; i < len(words); i += 1 {
		word := words[i]
		// dumb way of look ahead
		var nextword string
		if i+1 < len(words) {
			nextword = words[i+1]
		}

		if word == "is" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_IS))
		} else if word == "if" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_IF))
		} else if word == "otherwise" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_OTHERWISE))
		} else if word == "otif" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_OTIF))
		} else if word == "of" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_OF))
		} else if word == "the" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_THE))
		} else if word == "that" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_THAT))
		} else if word == "for" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_FOR))
		} else if word == "while" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_WHILE))
		} else if word == "true" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_TRUE))
		} else if word == "false" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_FALSE))
		} else if word == "break" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_BREAK))
		} else if word == "by" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_BY))
		} else if word == "we" && nextword == "mean" {
			i += 1
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_WE_MEAN))
		} else if word == "returns" {
			tokens = append(tokens, l.makeToken(word, TOKEN_KW_RETURNS))
		} else if word == "text" {
			text := strings.Join(words[i+1:], " ")
			tokens = append(tokens, l.makeToken(text, TOKEN_TEXT))
			break // break out of for loop
		} else if unicode.IsNumber(FirstRune(word)) {
			tokens = append(tokens, l.makeToken(word, TOKEN_NUMBER))
		} else {
			tokens = append(tokens, l.makeToken(word, TOKEN_WORD))
		}

	}

	// emit newline
	tokens = append(tokens, l.makeToken("\n", TOKEN_NEWLINE))

	return tokens, nil
}

func (l *Lexer) handleIndent(newIndent indentlevel) ([]Token, error) {

	if l.debug {
		log.Printf("entering handleIndent, newIdent = %#v, curr = %#v, shiftWidth = %#v\n", newIndent, l.indent, l.shiftWidth)
	}
	curr := l.indent
	if l.shiftWidth.spaces == 0 && l.shiftWidth.tabs == 0 {
		// if its zero, set the shift width for the first time
		l.shiftWidth = newIndent
		if l.debug {
			log.Printf("setting shift widths to %#v\n", l.shiftWidth)
		}
	}

	newLevel, err := translateIndent(newIndent, l.shiftWidth)
	if err != nil {
		return nil, err
	}

	diff := newLevel - curr

	if l.debug {
		log.Printf("newLevel = %d, currLevel = %d diff = %d\n", newLevel, curr, diff)
	}

	tokens := make([]Token, 0, 0)
	if diff > 0 {
		if diff != 1 {
			return nil,errors.New(fmt.Sprintf("can not indent multiple at a time: you tried to indent %d levels", diff))
		}
		tokens = append(tokens, l.makeToken("  ", TOKEN_BEGIN_INDENT))
		l.indent = newLevel
		return tokens, nil
	}

	diff = -diff
	if l.debug {
		log.Printf("dedenting from curr = %d to newLevel = %d\n", curr, newLevel)
	}

	for i := 0; i < diff; i += 1 {
		tokens = append(tokens, l.makeToken("  ", TOKEN_END_INDENT))
	}
	l.indent = newLevel
	return tokens, nil
}

func translateIndent(newIndent indentlevel, shiftWidth indentlevel) (int, error) {
	if shiftWidth.tabs != 0 && shiftWidth.spaces != 0 {
		return 0, errors.New("can not mix tabs and spaces")
	}

	if shiftWidth.tabs > 0 {
		if newIndent.spaces > 0 {
			return 0, errors.New("can not mix tabs and spaces")
		}
		if shiftWidth.tabs != 1 {
			return 0, errors.New("can not have multipletabs in shift width")
		}
		return newIndent.tabs, nil
	}

	if shiftWidth.spaces > 0 {
		if newIndent.tabs > 0 {
			return 0, errors.New("can not mix tabs and spaces")
		}
		if newIndent.spaces%shiftWidth.spaces != 0 {
			return 0, errors.New(fmt.Sprintf("wrong number of spaces in indentation: %d. You are using %d", newIndent.spaces, shiftWidth.spaces))
		}
		return newIndent.spaces / shiftWidth.spaces, nil
	}

	if newIndent.spaces == 0 && newIndent.tabs == 0 {
		return 0, nil
	}

	return 0, errors.New("can have zero shift width (this is likely a compiler bug)")
}
