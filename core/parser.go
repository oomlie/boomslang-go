package core

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"log"
	"strings"
)

type Parser struct {
	buf    *bufio.Reader
	debug  bool
	tokens []Token
	pos    int
}

func MakeParser(opts *Opts, tokens []Token) *Parser {
	parser := new(Parser)
	parser.debug = opts.debug&DBG_PARSE > 0
	parser.tokens = tokens
	return parser
}

func eof() Token {
	return Token{
		Lex: "",
		Ty:  TOKEN_EOF}
}
func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return eof()
	}
	return p.tokens[p.pos]
}
func (p *Parser) consumeOrFail(ty TokenType) (Token, error) {
	tok := p.peek()
	if tok.Ty != ty {
		return eof(), errors.New(fmt.Sprintf("expected %v, found %v\n", tok.Ty, ty))
	}
	if ty != TOKEN_EOF {
		p.pos += 1
	}
	return tok, nil
}
func (p *Parser) hasTokens() bool {
	return p.peek().Ty != TOKEN_EOF
}

func (p *Parser) consumeLine() []Token {
	begin := p.pos
	for p.peek().Ty != TOKEN_NEWLINE && p.peek().Ty != TOKEN_EOF {
		p.pos += 1
	}
	end := p.pos
	p.pos += 1
	return p.tokens[begin:end]
}

func (p *Parser) Parse() ([]Ast, error) {
	ast := make([]Ast, 0, 50)
	for p.hasTokens() {
		line := p.consumeLine()
		node, err := p.parseStmnt(line)
		if err != nil {
			return nil, err
		}
		ast = append(ast, node)
	}

	if p.debug {
		log.Printf("ast = %v\n", spew.Sdump(ast))
	}

	return ast, nil
}

// magic infix functions
type builtindef struct {
	symbol string
	opname string
}

var infixBuiltins []builtindef = []builtindef{
	{"_super-duper-secret__equals", "equals"},
	{"_super-duper-secret__notequals", "notequals"},
	{"_super-duper-secret__smallerthan", "smallerthan"},
	{"_super-duper-secret__biggerthan", "biggerthan"},
	{"_super-duper-secret__multiply", "multiply"},
	{"_super-duper-secret__divides", "divides"},
	{"_super-duper-secret__plus", "plus"},
	{"_super-duper-secret__minus", "minus"},
}

func (p *Parser) ParseStmnt(words []Token) (Ast, error) {
	ast, err := p.parseStmnt(words);
	return ast, err
}
func (p *Parser) parseStmnt(words []Token) (Ast, error) {
	if p.debug {
		log.Printf("parseStmnt %#v\n", words)
	}
	// strip trailing newline or EOF
	if len(words) > 0 && words[len(words)-1].Ty == TOKEN_NEWLINE {
		if p.debug {
			log.Printf("  stripping newline from list of tokens")
		}
		words = words[:len(words)-1]
	}


	if len(words) == 0 {
		if p.debug {
			log.Printf(" return AstLiteral\n")
		}
		return AstLiteral{value: BsNilVal{}}, nil
	}

	if words[0].Ty == TOKEN_KW_IF {
		return p.parseConditional(words)
	} else if words[0].Ty == TOKEN_KW_WHILE {
		return p.parseLoop(words)
	} else if words[0].Ty == TOKEN_KW_BREAK {
		if len(words) != 1 {
			return nil, parseErr("no tokens after break", words[1])
		}
		node := AstBreak{returns: nil}
		return node, nil
	} else if words[0].Ty == TOKEN_KW_RETURNS {
		var expr Ast = AstLiteral{value: BsNilVal{}}
		if len(words) > 1 {
			expr1, err := p.parseExpr(words[1:])
			if err != nil {
				return nil, err
			}
			expr = expr1
		}
		node := AstReturns{expr: expr}
		return node, nil
	} else if words[0].Ty == TOKEN_KW_BY {
		return p.parseFuncDef(words)
	}

	if left, right, found := Partition(words, TOKEN_KW_IS); found {
		lval, err := p.parseIdent(left)
		if err != nil {
			return nil, err
		}
		rval, err := p.parseExpr(right)
		if err != nil {
			return nil, err
		}
		node := AstAssign{lval, rval}
		return node, nil
	}

	return p.parseExpr(words)
}

func (p *Parser) parseFuncDef(words []Token) (Ast, error) {
	if p.debug {
		log.Printf("parseFuncDef %#v\n", words)
	}
	// assume first token is "BY"
	//  BY name... OF params ... WE MEAN
	nameTokens, remaining, found := Partition(words[1:], TOKEN_KW_OF)
	if !found {
		// todo
		return nil, parseErr("expected keyword 'of' inside definition", words[0])
	}
	paramTokens, remaining, found := Partition(remaining, TOKEN_KW_WE_MEAN)
	if !found {
		// todo
		return nil, parseErr("expected keyword 'we mean' inside definition", words[0])
	}
	if len(remaining) > 0 {
		return nil, parseErr("unexpected tokens after 'we mean'", remaining[0])
	}
	name, err := JoinTokens(nameTokens)
	if err != nil {
		return nil, err
	}
	funcName := AstIdent{name: name}

	// todo: multiplw parameters are weird
	paramName, err := JoinTokens(paramTokens)
	if err != nil {
		return nil, err
	}
	param := AstIdent{name: paramName}

	if p.debug {
		log.Printf("AstFuncDef: paramTokens = %#v, paramName = %#v\n", paramTokens, paramName)
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	node := AstFuncDef{
		name:   funcName,
		params: []AstIdent{param},
		body:   body}
	return node, nil
}

func (p *Parser) parseLoop(words []Token) (Ast, error) {
	if p.debug {
		log.Printf("parseLoop %#v\n", words)
	}
	// we assume our caller knew what they are doing,
	// and just ignore words[0] (it should be FOR or WHILE)
	cond, err := p.parseExpr(words[1:])
	if err != nil {
		return nil, err
	}
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	else_block := []Ast{}
	if p.peek().Ty == TOKEN_KW_OTHERWISE {
		if p.debug {
			log.Printf(" saw otherwise after loop, parsing else block now\n")
		}
		words := p.consumeLine()
		if len(words) != 1 {
			return nil, parseErr("unexpected tokens after otherwise block", words[0])
		}
		else_block, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}
	node := AstLoop{
		cond:       cond,
		block:      block,
		else_block: else_block,
	}
	return node, nil
}

func (p *Parser) parseConditional(words []Token) (Ast, error) {
	if p.debug {
		log.Printf("parseConditional %#v\n", words)
	}
	// we assume our caller knew what they are doing,
	// and just ignore words[0] (it should be IF or OTIF)
	cond, err := p.parseExpr(words[1:])
	if err != nil {
		return nil, err
	}
	if_block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	else_block := []Ast{}
	if p.peek().Ty == TOKEN_KW_OTHERWISE {
		line := p.consumeLine()
		// todo: this could actually be parsed here
		// for now, we expect one tokens: OTHERWISE
		if len(line) != 1 {
			return nil, parseErr("expected newline after 'otherwise' keyword", p.peek())
		}
		else_block, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	} else if p.peek().Ty == TOKEN_KW_OTIF {
		words := p.consumeLine()
		// now we have a normal conditional
		else_node, err := p.parseConditional(words)
		if err != nil {
			return nil, err
		}
		else_block = []Ast{else_node}
	}

	node := AstIfStmnt{
		cond:       cond,
		if_block:   if_block,
		else_block: else_block,
	}
	return node, nil
}

func (p *Parser) parseExpr(words []Token) (Ast, error) {
	if p.debug {
		log.Printf("parseExpr %#v\n", words)
	}

	if left, right, found := Partition(words, TOKEN_KW_OF); found {
		fun, err := p.parseFunHead(left)
		if err != nil {
			return nil, err
		}
		args, err := p.parseArgs(right)
		if err != nil {
			return nil, err
		}
		if p.debug {
			log.Printf(" return AstFunCall\n")
		}
		node := AstFunCall{
			fun:  fun,
			args: args,
		}
		return node, nil
	}

	// builtin infix functions

	for _, infix := range infixBuiltins {

		if left, right, found := PartitionByLexeme(words, infix.opname); found {
			lexpr, err := p.parseExpr(left)
			if err != nil {
				return nil, err
			}
			rexpr, err := p.parseExpr(right)
			if err != nil {
				return nil, err
			}

			if p.debug {
				log.Printf(" return AstFunCall\n")
			}
			node := AstFunCall{
				fun:  AstIdent{name: infix.symbol},
				args: []Ast{lexpr, rexpr},
			}
			return node, nil
		}
	}

	// prefix operators
	if words[0].Ty == TOKEN_KW_THE {
		return p.parseIdent(words)
	}

	// we assume its a function call
	return p.parseFunCall(words)
}

func (p *Parser) parseBlock() ([]Ast, error) {
	if p.debug {
		log.Printf("parseBlock \n")
	}
	p.consumeOrFail(TOKEN_BEGIN_INDENT)

	ast := make([]Ast, 0, 5)
	for p.hasTokens() && p.peek().Ty != TOKEN_END_INDENT {
		line := p.consumeLine()
		n, err := p.parseStmnt(line)
		if err != nil {
			return nil, err
		}
		ast = append(ast, n)
	}
	p.consumeOrFail(TOKEN_END_INDENT)

	return ast, nil
}

func (p *Parser) parseFunCall(words []Token) (Ast, error) {
	if p.debug {
		log.Printf("parseFunCall %#v\n", words)
	}

	if len(words) == 0 {
		return nil, errors.New("need tokens inside function call")
	}
	if len(words) == 1 {
		return p.parseAtom(words[0])
	}

	head, err := p.parseFunHead(words[:1])
	if err != nil {
		return nil, err
	}

	args, err := p.parseArgs(words[1:])
	if err != nil {
		return nil, err
	}
	node := AstFunCall{
		fun:  head,
		args: args,
	}

	if p.debug {
		log.Printf(" return AstFunCall\n")
	}
	return node, nil
}

func (p *Parser) parseFunHead(words []Token) (Ast, error) {
	if p.debug {
		log.Printf("parseFunHead %#v\n", words)
	}
	if len(words) == 0 {
		return nil, parseErr("invocing function needs a parameter", eof())
	}
	if len(words) == 1 {
		if words[0].Ty == TOKEN_WORD {
			// only case where a single bare word can become an identifier: when it is invoked as a function
			node := AstIdent{
				name: words[0].Lex,
			}
			return node, nil
		}
		// must parse this here to avoid infinite depth
		return p.parseAtom(words[0])
	}
	return p.parseExpr(words)
}
func (p *Parser) parseArgs(words []Token) ([]Ast, error) {
	if p.debug {
		log.Printf("parseArgs %#v\n", words)
	}
	if len(words) == 0 {
		// empty list must be handled explicitly here
		// otherwise we get some infinite descent
		return []Ast{}, nil
	}

	// todo: multiple args better
	node, err := p.parseExpr(words)
	if err != nil {
		return nil, err
	}
	list := []Ast{node}
	return list, nil
}

func (p *Parser) parseIdent(words []Token) (Ast, error) {
	if p.debug {
		log.Printf("parseIdent %#v\n", words)
	}

	if words[0].Ty != TOKEN_KW_THE {
		return nil, parseErr(fmt.Sprintf("expected 'the' keyword to begin identifier, found %#v", words[0]), words[0])
	}

	name, err := JoinTokens(words[1:])
	if err != nil {
		return nil, err
	}
	if p.debug {
		log.Printf(" return AstIdent\n")
	}
	return AstIdent{name: name}, nil
}

func (p *Parser) parseAtom(word Token) (Ast, error) {
	if p.debug {
		log.Printf("parseAtom %#v\n", word)
	}

	if word.Ty == TOKEN_NUMBER {
		var value int64
		rc, err := fmt.Sscanf(word.Lex, "%d", &value)
		if err != nil {
			return nil, err
		}
		if rc == 0 {
			return nil, parseErr("not a valid number", word)
		}
		literal := BsIntVal{value: value}
		if p.debug {
			log.Printf(" return AstLiteral\n")
		}
		node := AstLiteral{value: literal}
		return node, nil
	} else if word.Ty == TOKEN_KW_TRUE {
		node := AstLiteral{
			value: BsBooleVal{
				value: true,
			},
		}
		if p.debug {
			log.Printf(" return AstLiteral\n")
		}
		return node, nil
	} else if word.Ty == TOKEN_KW_FALSE {
		node := AstLiteral{
			value: BsBooleVal{
				value: false,
			},
		}
		if p.debug {
			log.Printf(" return AstLiteral\n")
		}
		return node, nil
	} else if word.Ty == TOKEN_TEXT {
		value := BsStrVal{value: word.Lex}
		node := AstLiteral{value: value}
		if p.debug {
			log.Printf(" return AstLiteral\n")
		}
		return node, nil
	} else {
		return nil, parseErr(fmt.Sprintf("unexpected token type %#v inside atom", word), word)
	}
}

// error reporting

func parseErr(msg string, token Token) error {
	return ParseError{
		msg:     msg,
		token:   token,
		context: nil,
	}
}

type ParseError struct {
	msg     string
	token   Token
	context []string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error at %s:%d\n%s",
		e.token.Spn.SourceName, e.token.Spn.Lineno, e.msg,
	)
}

// utilities

func PartitionByLexeme(words []Token, split string) ([]Token, []Token, bool) {
	idx := FindFirst(words, func(tok Token) bool { return tok.Lex == split })
	if idx == -1 {
		return nil, nil, false
	}
	left := words[:idx]
	right := words[idx+1:]
	return left, right, true
}

func Partition(words []Token, split TokenType) ([]Token, []Token, bool) {
	idx := FindFirst(words, func(tok Token) bool { return tok.Ty == split })
	if idx == -1 {
		return nil, nil, false
	}
	left := words[:idx]
	right := words[idx+1:]
	return left, right, true
}

func FindFirst(words []Token, pred func(Token) bool) int {
	for i := range words {
		if pred(words[i]) {
			return i
		}
	}
	return -1
}

func JoinTokens(words []Token) (string, error) {
	b := new(strings.Builder)
	for i, w := range words {
		if i != 0 {
			b.WriteString(" ")
		}
		if w.Ty != TOKEN_WORD {
			return "", parseErr(fmt.Sprintf("expected TOKEN_WORD, found %#v", w.Ty), w)
		}
		b.WriteString(w.Lex)
	}
	return b.String(), nil
}
