package core

type Ast interface {
	Eval(env *BsEnv) BsValue
	ShortName() string
}

type AstFunCall struct {
	fun  Ast
	args []Ast
}

func (node AstFunCall) ShortName() string { return "procedure" }

type AstIdent struct {
	name string
}

func (node AstIdent) ShortName() string { return "name" }

type AstLiteral struct {
	value BsValue
}

func (node AstLiteral) ShortName() string { return node.value.PrettyPrint() }

type AstAssign struct {
	lvalue Ast
	rvalue Ast
}

func (node AstAssign) ShortName() string { return "assignment" }

type AstIfStmnt struct {
	cond       Ast
	if_block   []Ast
	else_block []Ast
}

func (node AstIfStmnt) ShortName() string { return "if statement" }

type AstLoop struct {
	cond  Ast
	block []Ast
	// todo: loops in python are weirder than this
	else_block []Ast
}

func (node AstLoop) ShortName() string { return "loop" }

type AstBreak struct {
	returns Ast
}

func (node AstBreak) ShortName() string { return "break" }

type AstFuncDef struct {
	name   AstIdent
	params []AstIdent
	body   []Ast
}

func (node AstFuncDef) ShortName() string { return "procedure definition for '" + node.name.name + "'" }

type AstReturns struct {
	expr Ast
}

func (node AstReturns) ShortName() string { return "returns" }
