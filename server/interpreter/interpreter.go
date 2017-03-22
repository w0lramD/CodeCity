/* Copyright 2017 Google Inc.
 * https://github.com/NeilFraser/CodeCity
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package interpreter implements a JavaScript interpreter.
package interpreter

import (
	"fmt"

	"CodeCity/server/interpreter/ast"
	"CodeCity/server/interpreter/object"
)

// Interpreter implements a JavaScript interpreter.
type Interpreter struct {
	state   state
	value   object.Value
	Verbose bool
}

// New takes a JavaScript program, in the form of an JSON-encoded
// ESTree, and creates a new Interpreter that will execute that
// program.
func New(astJSON string) *Interpreter {
	var this = new(Interpreter)

	tree, err := ast.NewFromJSON(astJSON)
	if err != nil {
		panic(err)
	}
	s := newScope(nil, this)
	// FIXME: insert global names into s
	s.populate(tree)
	this.state = newState(nil, s, tree)
	return this
}

// Step performs the next step in the evaluation of program.  Returns
// true if a step was executed; false if the program has terminated.
func (this *Interpreter) Step() bool {
	if this.state == nil {
		return false
	}
	if this.Verbose {
		fmt.Printf("Next step is a %T\n", this.state)
	}
	this.state = this.state.step()
	return true
}

// Run runs the program to completion.
func (this *Interpreter) Run() {
	for this.Step() {
	}
}

// Value returns the final value computed by the last statement
// expression of the program.
func (this *Interpreter) Value() object.Value {
	return this.value
}

// acceptValue receives values computed by StatementExpressions; the
// last such value accepted is the completion value of the program.
func (this *Interpreter) acceptValue(v object.Value) {
	if this.Verbose {
		fmt.Printf("Interpreter just got %v.\n", v)
	}
	this.value = v
}

/********************************************************************/

// scope implements JavaScript (block) scope; it's basically just a
// mapping of declared variable names to values, with two additions:
//
// - parent is a pointer to the parent scope (if nil then this is the
// global scope)
//
// - interpreter is a pointer to the interpreter that this scope
// belongs to.  It is provided so that stateExpressionStatement can
// send a completion value to the interpreter, which is useful for
// testing purposes now and possibly for eval() later.  This may go
// away if we find a better way to test and decide not to implement
// eval().  It's on scope instead of stateCommon just to reduce the
// number of redundant copies.
//
// FIXME: readonly flag?  Or readonly if parent == nil?
type scope struct {
	vars        map[string]object.Value
	parent      *scope
	interpreter *Interpreter
}

// newScope is a factory for scope objects.  The parent param is a
// pointer to the parent (enclosing scope); it is nil if the scope
// being created is the global scope.  The interpreter param is a
// pointer to the interpreter this scope belongs to.
func newScope(parent *scope, interpreter *Interpreter) *scope {
	return &scope{make(map[string]object.Value), parent, interpreter}
}

// setVar sets the named variable to the specified value, after
// first checking that it exists.
//
// FIXME: this should probably recurse if name is not found in current
// scope - but not when called from stateVariableDeclarator, which
// should never be setting variables other than in the
// immediately-enclosing scope.
func (this *scope) setVar(name string, value object.Value) {
	_, ok := this.vars[name]
	if !ok {
		panic(fmt.Errorf("can't set undeclared variable %v", name))
	}
	this.vars[name] = value
}

// getVar gets the current value of the specified variable, after
// first checking that it exists.
//
// FIXME: this should probably recurse if name is not found in current
// scope.
func (this *scope) getVar(name string) object.Value {
	v, ok := this.vars[name]
	if !ok {
		// FIXME: should probably throw
		panic(fmt.Errorf("can't get undeclared variable %v", name))
	}
	return v
}

func (this *scope) populate(node ast.Node) {
	switch n := node.(type) {

	// The interesting cases:
	case *ast.VariableDeclarator:
		this.vars[n.Id.Name] = object.Undefined{}
	case *ast.FunctionDeclaration:
		// Add name of function to scope; ignore contents.
		this.vars[n.Id.Name] = object.Undefined{}

	// The recursive cases:
	case *ast.BlockStatement:
		for _, s := range n.Body {
			this.populate(s)
		}
	case *ast.CatchClause:
		this.populate(n.Body)
	case *ast.DoWhileStatement:
		this.populate(n.Body.S)
	case *ast.ForInStatement:
		this.populate(n.Left.N)
		this.populate(n.Body.S)
	case *ast.ForStatement:
		this.populate(n.Init.N)
		this.populate(n.Body.S)
	case *ast.IfStatement:
		this.populate(n.Consequent.S)
		this.populate(n.Alternate.S)
	case *ast.LabeledStatement:
		this.populate(n.Body.S)
	case *ast.Program:
		for _, s := range n.Body {
			this.populate(s)
		}
	case *ast.SwitchCase:
		for _, s := range n.Consequent {
			this.populate(s)
		}
	case *ast.SwitchStatement:
		for _, c := range n.Cases {
			this.populate(c)
		}
	case *ast.TryStatement:
		this.populate(n.Block)
		this.populate(n.Handler)
		this.populate(n.Finalizer)
	case *ast.VariableDeclaration:
		for _, d := range n.Declarations {
			this.populate(d)
		}
	case *ast.WhileStatement:
		this.populate(n.Body.S)
	case *ast.WithStatement:
		panic("not implemented")

	// The cases we can ignore because they cannot contain
	// declarations:
	case *ast.ArrayExpression:
	case *ast.AssignmentExpression:
	case *ast.BinaryExpression:
	case *ast.BreakStatement:
	case *ast.CallExpression:
	case *ast.ConditionalExpression:
	case *ast.ContinueStatement:
	case *ast.DebuggerStatement:
	case *ast.EmptyStatement:
	case *ast.ExpressionStatement:
	case *ast.FunctionExpression:
	case *ast.Identifier:
	case *ast.Literal:
	case *ast.LogicalExpression:
	case *ast.MemberExpression:
	case *ast.NewExpression:
	case *ast.ObjectExpression:
	case *ast.Property:
	case *ast.ReturnStatement:
	case *ast.SequenceExpression:
	case *ast.ThisExpression:
	case *ast.ThrowStatement:
	case *ast.UnaryExpression:
	case *ast.UpdateExpression:

	// Just in case:
	default:
		panic(fmt.Errorf("Unrecognized ast.Node type %T", node))
	}
}

/********************************************************************/

// state is the interface implemented by each of the types
// representing different possible next states for the interpreter
// (roughly: one state per ast.Node implementation); each value of
// this type represents a possible state of the computation.
type state interface {
	// step performs the next step in the evaluation of the program, and
	// returns the new state execution state.
	step() state
}

// valueAcceptor is the interface implemented by any object (mostly
// states with subexpressions) that can accept a value.
type valueAcceptor interface {
	// acceptValue receives the value resulting from the evaluation of
	//a child expression.
	/// It is normally called by the
	// subexpression's step method, typically as follows:
	//
	//        // ... compute value to be returned ...
	//        this.parent.acceptValue(value)
	//        return this.parent
	//    }
	acceptValue(object.Value)
}

// newState creates a state object corresponding to the given AST
// node.  The parent parameter represents the state the interpreter
// should return to after evaluating the tree rooted at node.
func newState(parent state, scope *scope, node ast.Node) state {
	var sc = stateCommon{parent, scope}
	switch n := node.(type) {
	case *ast.AssignmentExpression:
		s := stateAssignmentExpression{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.BinaryExpression:
		s := stateBinaryExpression{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.BlockStatement:
		s := stateBlockStatement{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.ConditionalExpression:
		s := stateConditionalExpression{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.EmptyStatement:
		s := stateEmptyStatement{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.ExpressionStatement:
		s := stateExpressionStatement{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.FunctionDeclaration:
		s := stateFunctionDeclaration{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.Identifier:
		s := stateIdentifier{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.IfStatement:
		s := stateIfStatement{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.Literal:
		s := stateLiteral{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.ObjectExpression:
		s := stateObjectExpression{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.Program:
		s := stateBlockStatement{stateCommon: sc}
		s.initFromProgram(n)
		return &s
	case *ast.VariableDeclaration:
		s := stateVariableDeclaration{stateCommon: sc}
		s.init(n)
		return &s
	case *ast.VariableDeclarator:
		s := stateVariableDeclarator{stateCommon: sc}
		s.init(n)
		return &s
	default:
		panic(fmt.Errorf("State for AST node type %T not implemented\n", n))
	}
}

/********************************************************************/

// stateCommon is a struct, intended to be embedded in most or all
// state<NodeType> types, which provides fields common to most/all
// states.
type stateCommon struct {
	// state is the state to return to once evaluation of this state
	// is finished.  (This is "state" rather than "*state" because the
	// interface value already containins a pointer to the actual
	// state<Whatever> object.)
	parent state

	// scope is the symobl table for the innermost scope.
	scope *scope
}

/********************************************************************/

type stateAssignmentExpression struct {
	stateCommon
	op    string
	left  lvalue
	rNode ast.Expression
	right object.Value
}

func (this *stateAssignmentExpression) init(node *ast.AssignmentExpression) {
	this.op = node.Operator
	this.rNode = node.Right
	this.left.init(this.scope, node.Left)
}

func (this *stateAssignmentExpression) step() state {
	if !this.left.ready {
		return this.left.next(this)
	} else if this.right == nil {
		return newState(this, this.scope, ast.Node(this.rNode.E))
	}

	// Do assignment:
	if !this.left.ready {
		panic("lvalue not ready???")
	}
	this.left.set(this.right)

	return this.parent
}

func (this *stateAssignmentExpression) acceptValue(v object.Value) {
	this.right = v
}

/********************************************************************/

type stateBinaryExpression struct {
	stateCommon
	op                  string
	lNode, rNode        ast.Expression
	haveLeft, haveRight bool
	left, right         object.Value
}

func (this *stateBinaryExpression) init(node *ast.BinaryExpression) {
	this.op = node.Operator
	this.lNode = node.Left
	this.rNode = node.Right
	this.haveLeft = false
	this.haveRight = false
}

func (this *stateBinaryExpression) step() state {
	if !this.haveLeft {
		return newState(this, this.scope, ast.Node(this.lNode.E))
	} else if !this.haveRight {
		return newState(this, this.scope, ast.Node(this.rNode.E))
	}

	// FIXME: implement other operators, types

	var v object.Value
	switch this.op {
	case "+":
		v = object.Number(this.left.(object.Number) +
			this.right.(object.Number))
	case "-":
		v = object.Number(this.left.(object.Number) -
			this.right.(object.Number))
	case "*":
		v = object.Number(this.left.(object.Number) *
			this.right.(object.Number))
	case "/":
		v = object.Number(this.left.(object.Number) /
			this.right.(object.Number))
	default:
		panic("not implemented")
	}

	this.parent.(valueAcceptor).acceptValue(v)
	return this.parent

}

func (this *stateBinaryExpression) acceptValue(v object.Value) {
	if this.scope.interpreter.Verbose {
		fmt.Printf("stateBinaryExpression just got %v.\n", v)
	}
	if !this.haveLeft {
		this.left = v
		this.haveLeft = true
	} else if !this.haveRight {
		this.right = v
		this.haveRight = true
	} else {
		panic(fmt.Errorf("too may values"))
	}
}

/********************************************************************/

type stateBlockStatement struct {
	stateCommon
	body  ast.Statements
	value object.Value
	n     int
}

func (this *stateBlockStatement) initFromProgram(node *ast.Program) {
	this.body = node.Body
}

func (this *stateBlockStatement) init(node *ast.BlockStatement) {
	this.body = node.Body
}

func (this *stateBlockStatement) step() state {
	if this.n < len(this.body) {
		s := newState(this, this.scope, (this.body)[this.n])
		this.n++
		return s
	}
	return this.parent
}

/********************************************************************/

type stateConditionalExpression struct {
	stateCommon
	test       ast.Expression
	consequent ast.Expression
	alternate  ast.Expression
	result     bool
	haveResult bool
}

func (this *stateConditionalExpression) init(node *ast.ConditionalExpression) {
	this.test = node.Test
	this.consequent = node.Consequent
	this.alternate = node.Alternate
}

func (this *stateConditionalExpression) step() state {
	if !this.haveResult {
		return newState(this, this.scope, ast.Node(this.test.E))
	}
	if this.result {
		return newState(this.parent, this.scope, this.consequent.E)
	} else {
		return newState(this.parent, this.scope, this.alternate.E)
	}
}

func (this *stateConditionalExpression) acceptValue(v object.Value) {
	if this.scope.interpreter.Verbose {
		fmt.Printf("stateConditionalExpression just got %v.\n", v)
	}
	this.result = object.IsTruthy(v)
	this.haveResult = true
}

/********************************************************************/

type stateEmptyStatement struct {
	stateCommon
}

func (this *stateEmptyStatement) init(node *ast.EmptyStatement) {
}

func (this *stateEmptyStatement) step() state {
	return this.parent
}

/********************************************************************/

type stateExpressionStatement struct {
	stateCommon
	expr ast.Expression
	done bool
}

func (this *stateExpressionStatement) init(node *ast.ExpressionStatement) {
	this.expr = node.Expression
	this.done = false
}

func (this *stateExpressionStatement) step() state {
	if !this.done {
		this.done = true
		return newState(this, this.scope, ast.Node(this.expr.E))
	} else {
		return this.parent
	}
}

// FIXME: this is only needed so a completion value is available in
// the interpreter for test purposes (and possibly for eval); if it
// was not required we could greatly simplify this state and only
// visit it once.
func (this *stateExpressionStatement) acceptValue(v object.Value) {
	if this.scope.interpreter.Verbose {
		fmt.Printf("stateExpressionStatement just got %v.\n", v)
	}
	this.scope.interpreter.acceptValue(v)
}

/********************************************************************/

// Evaluating a function declaration has no effect; the declaration
// has already been hoisted into the enclosing scope.
type stateFunctionDeclaration struct {
	stateCommon
}

func (this *stateFunctionDeclaration) init(node *ast.FunctionDeclaration) {
}

func (this *stateFunctionDeclaration) step() state {
	return this.parent
}

/********************************************************************/

type stateIdentifier struct {
	stateCommon
	name string
}

func (this *stateIdentifier) init(node *ast.Identifier) {
	this.name = node.Name
}

func (this *stateIdentifier) step() state {
	// Note: if we getters/setters and a global scope object (like
	// window), we would have to do a check to see if we need to run a
	// getter.  But we have neither, so this is a straight variable
	// lookup.
	this.parent.(valueAcceptor).acceptValue(this.scope.getVar(this.name))
	return this.parent
}

/********************************************************************/

// This is exactly the same as stateConditionalExpression except for
// the types of consequent and alternate (and the name and node type,
// of course).
type stateIfStatement struct {
	stateCommon
	test       ast.Expression
	consequent ast.Statement
	alternate  ast.Statement
	result     bool
	haveResult bool
}

func (this *stateIfStatement) init(node *ast.IfStatement) {
	this.test = node.Test
	this.consequent = node.Consequent
	this.alternate = node.Alternate
}

func (this *stateIfStatement) step() state {
	if !this.haveResult {
		return newState(this, this.scope, ast.Node(this.test.E))
	}
	if this.result {
		return newState(this.parent, this.scope, this.consequent.S)
	} else {
		return newState(this.parent, this.scope, this.alternate.S)
	}
}

func (this *stateIfStatement) acceptValue(v object.Value) {
	if this.scope.interpreter.Verbose {
		fmt.Printf("stateIfStatement just got %v.\n", v)
	}
	this.result = object.IsTruthy(v)
	this.haveResult = true
}

/********************************************************************/

type stateLiteral struct {
	stateCommon
	value object.Value
}

func (this *stateLiteral) init(node *ast.Literal) {
	this.value = object.PrimitiveFromRaw(node.Raw)
}

func (this *stateLiteral) step() state {
	this.parent.(valueAcceptor).acceptValue(this.value)
	return this.parent
}

/********************************************************************/

type stateObjectExpression struct {
	stateCommon
	props            []*ast.Property
	obj              *object.Object
	n                int
	key              string
	value            object.Value
	gotKey, gotValue bool
}

func (this *stateObjectExpression) init(node *ast.ObjectExpression) {
	this.props = node.Properties
	this.obj = nil
	this.n = 0
}

// FIXME: (maybe) getters and setters not supported.
func (this *stateObjectExpression) step() state {
	if this.obj == nil {
		if this.n != 0 {
			//			panic("lost object under construction!")
		}
		// FIXME: set owner of new object
		this.obj = object.New(nil, object.ObjectProto)
	}
	if this.n < len(this.props) {
		return newState(this, this.scope, this.props[this.n].Value.E)
	} else {
		this.parent.(valueAcceptor).acceptValue(this.obj)
		return this.parent
	}
}

func (this *stateObjectExpression) acceptValue(v object.Value) {
	if this.scope.interpreter.Verbose {
		fmt.Printf("stateObjectExpression just got %v.\n", v)
	}
	var key string
	switch k := this.props[this.n].Key.N.(type) {
	case *ast.Literal:
		v := object.PrimitiveFromRaw(k.Raw)
		key = v.ToString()
	case *ast.Identifier:
		key = k.Name
	}
	this.obj.SetProperty(key, v)
	this.n++
}

/********************************************************************/

type stateVariableDeclaration struct {
	stateCommon
	decls []*ast.VariableDeclarator
}

func (this *stateVariableDeclaration) init(node *ast.VariableDeclaration) {
	this.decls = node.Declarations
	if node.Kind != "var" {
		panic(fmt.Errorf("Unknown VariableDeclaration kind '%v'", node.Kind))
	}
}

func (this *stateVariableDeclaration) step() state {
	// Create a stateVariableDeclarator for every VariableDeclarator
	// that has an Init value, chaining them together so they will
	// execute in left-to-right order.
	var p = this.parent
	for i := len(this.decls) - 1; i >= 0; i-- {
		if this.decls[i].Init.E != nil {
			p = newState(p, this.scope, this.decls[i])
		}
	}
	return p
}

/********************************************************************/

type stateVariableDeclarator struct {
	stateCommon
	name  string
	expr  ast.Expression
	value object.Value
}

func (this *stateVariableDeclarator) init(node *ast.VariableDeclarator) {
	this.name = node.Id.Name
	this.expr = node.Init
	this.value = nil
}

func (this *stateVariableDeclarator) step() state {
	if this.expr.E == nil {
		panic("Why are we bothering to execute an variable declaration" +
			"(that has already been hoisted) that has no initialiser?")
	}
	if this.value == nil {
		return newState(this, this.scope, ast.Node(this.expr.E))
	} else {
		this.scope.setVar(this.name, this.value)
		return this.parent
	}
}

func (this *stateVariableDeclarator) acceptValue(v object.Value) {
	this.value = v
}

/********************************************************************/

// lvalue is an object which encapsulates reading and modification of
// lvalues in assignment and update expressions.  It also acts as an
// interpreter state for the evaluation of lvalue subexpressions.
//
// Usage:
//
//  struct stateFoo {
//      stateCommon
//      lv lvalue
//      ...
//  }
//
//  func (this *stateFoo) init(node *ast.Foo) {
//      this.lv.init(this.scope, node.left)
//      ...
//  }
//
//  func (this *stateFoo) step() state {
//      if(!this.lv.ready) {
//          return this.lv.next(this)
//      }
//      ...
//      lv.set(lv.get() + 1) // or whatever
//      ...
//  }
//
type lvalue struct {
	stateCommon
	name  string
	ready bool
}

func (this *lvalue) init(scope *scope, expr ast.Expression) {
	this.scope = scope

	switch e := expr.E.(type) {
	case *ast.Identifier:
		this.name = e.Name
		this.ready = true
	case *ast.MemberExpression:
		panic("not implemented")
	default:
		panic(fmt.Errorf("%T is not an lvalue", expr.E))
	}
}

func (this *lvalue) next(parent state) state {
	if this.ready {
		// Nothing to do.  Why was this called?
		panic("lvalue already ready")
	}
	panic("not implemented")
}

// get returns the current value of the variable or property denoted
// by the lvalue expression.
func (this *lvalue) get() object.Value {
	if !this.ready {
		panic("lvalue not ready")
	}
	return this.scope.getVar(this.name)
}

// set updates the variable or property denoted
// by the lvalue expression to the given value.
func (this *lvalue) set(value object.Value) {
	if !this.ready {
		panic("lvalue not ready")
	}
	this.scope.setVar(this.name, value)
}

func (this *lvalue) step() state {
	panic("not implemented")
}

func (this *lvalue) acceptValue(v object.Value) {
	panic("not implemented")
}
