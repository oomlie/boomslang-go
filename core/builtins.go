package core

import (
	"fmt"
	"log"
	"reflect"
)

// Methods on this are called to initialize the name space bindings on compiler intrinsics
// This is done to keep all of the definitions for a builtin in 1 spot
type BuiltinRegistry struct{}

// Called to before entering the runtime
func LoadBuiltins(env *BsEnv) {
	registryV := BuiltinRegistry{}
	registryT := reflect.TypeOf(registryV)
	for i := 0; i < registryT.NumMethod(); i += 1 {
		method := registryT.Method(i)
		if env.debug {
			log.Printf("calling %s\n", method.Name)
		}
		regFunc := method.Func.Interface().(func(BuiltinRegistry, *BsEnv))
		regFunc(registryV, env)
	}
}

// ==========================================
//
//	show operator:
//	  prints each argument
//	  and returns nil
type BsBuiltinShow struct{}

func (this BsBuiltinShow) PrettyPrint() string {
	return fmt.Sprintf("<builtin procedure 'show'>")
}
func (this BsBuiltinShow) Call(env *BsEnv, args []BsValue) BsValue {
	for i, arg := range args {
		if i != 0 {
			fmt.Fprintf(env.ostr, " ")
		}
		fmt.Fprintf(env.ostr, "%s", arg.PrettyPrint())
	}
	fmt.Fprintf(env.ostr, "\n")
	return new(BsNilVal)
}
func (r BuiltinRegistry) RegisterShow(env *BsEnv) {
	env.AssignName("show", BsFunVal{thunk: BsBuiltinShow{}})
}

// ==========================================
//
//	debug operator:
//	  prints argument representation
//	  and returns the first
type BsBuiltinDebug struct{}

func (this BsBuiltinDebug) PrettyPrint() string {
	return fmt.Sprintf("<builtin procedure 'debug'>")
}
func (this BsBuiltinDebug) Call(env *BsEnv, args []BsValue) BsValue {
	for _, arg := range args {
		fmt.Fprintf(env.ostr, "%v\n", arg)
	}
	if len(args) == 0 {
		return new(BsNilVal)
	}
	return args[0]
}

func (r BuiltinRegistry) RegisterDebug(env *BsEnv) {
	env.AssignName("debug", BsFunVal{thunk: BsBuiltinDebug{}})
}

// ==========================================
//
//	ask
//	  prints arguments and and pauses for user input
//	  returns result of user input
type BsBuiltinAsk struct{}

func (this BsBuiltinAsk) PrettyPrint() string {
	return fmt.Sprintf("<builtin procedure 'ask'>")
}
func (this BsBuiltinAsk) Call(env *BsEnv, args []BsValue) BsValue {
	for _, arg := range args {
		fmt.Fprintf(env.ostr, "%s ", arg.PrettyPrint())
	}
	var result string
	_, err := fmt.Fscanf(env.istr, "%s", &result)
	if err != nil {
		return BsIoErr{msg: err.Error()}
	}
	return BsStrVal{value: result}
}

func (r BuiltinRegistry) RegisterAsk(env *BsEnv) {
	env.AssignName("ask", BsFunVal{thunk: BsBuiltinAsk{}})
}

// ==========================================
//
//	casts: take an arbitrary object
//	  return that as the correct type
type BsBuiltinCastToInt struct{}

func (this BsBuiltinCastToInt) PrettyPrint() string {
	return fmt.Sprintf("<builtin procedure 'number'>")
}
func (this BsBuiltinCastToInt) Call(env *BsEnv, args []BsValue) BsValue {
	if len(args) != 1 {
		return BsMethodErr{expected: fmt.Sprintf("1 parameter to %s, got %d", this.PrettyPrint(), len(args))}
	}
	switch v := args[0].(type) {
	case BsIntVal:
		return args[0]
	case BsStrVal:
		var value int64
		n, err := fmt.Sscanf(v.value, "%d", &value)
		if err != nil || n == 0 {
			return BsTypeErr{expected: "text with a number", value: v}
		}
		return BsIntVal{value}
	}
	return BsTypeErr{expected: "something I can turn into a number", value: args[0]}
}

func (r BuiltinRegistry) RegisterCastToInt(env *BsEnv) {
	env.AssignName("number", BsFunVal{thunk: BsBuiltinCastToInt{}})
}

// ==========================================
//
//	binary integer operations:
//	  take 2 integers and returns a third
type BsBuiltinIntBinOp struct {
	name string
	op   func(int64, int64) int64
}

func makeIntBinOp(name string, op func(int64, int64) int64) BsFunVal {
	thunk := BsBuiltinIntBinOp{
		name: name,
		op:   op,
	}
	fun := BsFunVal{thunk: thunk}
	return fun
}
func (this BsBuiltinIntBinOp) PrettyPrint() string {
	return fmt.Sprintf("<builtin procedure '%s'>", this.name)
}
func (this BsBuiltinIntBinOp) Call(env *BsEnv, args []BsValue) BsValue {
	if len(args) != 2 {
		return BsMethodErr{expected: "2 parameters"}
	}
	left, ok := args[0].(BsIntVal)
	if !ok {
		return BsTypeErr{expected: "number", value: args[0]}
	}
	right, ok := args[1].(BsIntVal)
	if !ok {
		return BsTypeErr{expected: "number", value: args[1]}
	}
	value := this.op(left.value, right.value)
	return BsIntVal{value: value}
}

func (r BuiltinRegistry) RegisterIntegerBinaryOperations(env *BsEnv) {
	env.AssignName("_super-duper-secret__plus", makeIntBinOp("plus", func(x, y int64) int64 { return x + y }))
	env.AssignName("_super-duper-secret__minus", makeIntBinOp("minus", func(x, y int64) int64 { return x - y }))
	env.AssignName("_super-duper-secret__multiply", makeIntBinOp("multiply", func(x, y int64) int64 { return x * y }))
	env.AssignName("_super-duper-secret__divide", makeIntBinOp("divide", func(x, y int64) int64 { return x / y }))
}

// ==========================================
//
//	binary integer predicates:
//	  take 2 integers and returns a boole
type BsBuiltinIntBinPred struct {
	name string
	op   func(int64, int64) bool
}

func makeIntBinPred(name string, op func(int64, int64) bool) BsFunVal {
	thunk := BsBuiltinIntBinPred{
		name: name,
		op:   op,
	}
	fun := BsFunVal{thunk: thunk}
	return fun
}
func (this BsBuiltinIntBinPred) PrettyPrint() string {
	return fmt.Sprintf("<builtin procedure '%s'>", this.name)
}
func (this BsBuiltinIntBinPred) Call(env *BsEnv, args []BsValue) BsValue {
	if len(args) != 2 {
		return BsMethodErr{expected: "2 parameters"}
	}
	left, ok := args[0].(BsIntVal)
	if !ok {
		return BsTypeErr{expected: "number", value: args[0]}
	}
	right, ok := args[1].(BsIntVal)
	if !ok {
		return BsTypeErr{expected: "number", value: args[1]}
	}
	value := this.op(left.value, right.value)
	return BsBooleVal{value: value}
}

func (r BuiltinRegistry) RegisterIntBinPred(env *BsEnv) {
	env.AssignName("_super-duper-secret__smallerthan", makeIntBinPred("smallerthan", func(x, y int64) bool { return x < y }))
	env.AssignName("_super-duper-secret__biggerthan", makeIntBinPred("biggerthan", func(x, y int64) bool { return x > y }))
	env.AssignName("_super-duper-secret__equals", makeIntBinPred("equals", func(x, y int64) bool { return x == y }))
}
