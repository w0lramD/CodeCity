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

package interpreter

import (
	"CodeCity/server/interpreter/object"
	// "fmt"
	"testing"
)

func TestInterpreterSimple(t *testing.T) {
	var tests = []struct {
		desc     string
		src      string
		expected object.Value
	}{
		{"1+1", onePlusOne, object.Number(2)},
		{"2+2", twoPlusTwo, object.Number(4)},
		{"four functions", simpleFourFunction, object.Number(42)},
		{"variable declaration", variableDecl, object.Number(43)},
		{"?: true", condTrue, object.String("then")},
		{"?: false", condFalse, object.String("else")},
		{"if true", ifTrue, object.String("then")},
		{"if false", ifFalse, object.String("else")},
		{"var x=0; x=44; x", simpleAssignment, object.Number(44)},
		{"var o={}; o.foo=45; o.foo", propertyAssignment, object.Number(45)},
		{"var x=45; x++; x++", postincrement, object.Number(46)},
		{"var x=45; ++x; ++x", preincrement, object.Number(47)},
	}

	for _, c := range tests {
		i := New(c.src)
		i.Run()
		if v := i.Value(); v != c.expected {
			t.Errorf("%s == %v (%T)\n(expected %v (%T))",
				c.desc, v, v, c.expected, c.expected)
		}
	}
}

func TestInterpreterObjectExpression(t *testing.T) {
	i := New(objectExpression)
	i.Run()
	v, ok := i.Value().(*object.Object)
	if !ok {
		t.Errorf("{foo: \"bar\", answer: 42} returned type %T "+
			"(expected object.Object)", i.Value())
	}
	if c := object.PropCount(v); c != 2 {
		t.Errorf("{foo: \"bar\", answer: 42} had %d properties "+
			"(expected 2)", c)
	}
	if foo, _ := v.GetProperty("foo"); foo != object.String("bar") {
		t.Errorf("{foo: \"bar\", answer: 42}'s foo == %v (%T) "+
			"(expected \"bar\")", foo, foo)
	}
	if answer, _ := v.GetProperty("answer"); answer != object.Number(42) {
		t.Errorf("{foo: \"bar\", answer: 42}'s answer == %v (%T) "+
			"(expected 42)", answer, answer)
	}
}

const onePlusOne = `{"type":"Program","start":0,"end":5,"body":[{"type":"ExpressionStatement","start":0,"end":5,"expression":{"type":"BinaryExpression","start":0,"end":5,"left":{"type":"Literal","start":0,"end":1,"value":1,"raw":"1"},"operator":"+","right":{"type":"Literal","start":4,"end":5,"value":1,"raw":"1"}}}]}`

const twoPlusTwo = `{"type":"Program","start":0,"end":5,"body":[{"type":"ExpressionStatement","start":0,"end":5,"expression":{"type":"BinaryExpression","start":0,"end":5,"left":{"type":"Literal","start":0,"end":1,"value":2,"raw":"2"},"operator":"+","right":{"type":"Literal","start":4,"end":5,"value":2,"raw":"2"}}}]}`

const sixTimesSeven = `{"type":"Program","start":0,"end":5,"body":[{"type":"ExpressionStatement","start":0,"end":5,"expression":{"type":"BinaryExpression","start":0,"end":5,"left":{"type":"Literal","start":0,"end":1,"value":6,"raw":"6"},"operator":"*","right":{"type":"Literal","start":4,"end":5,"value":7,"raw":"7"}}}]}`

// (3 + 12 / 4) * (10 - 3)
// => 42
const simpleFourFunction = `{"type":"Program","start":0,"end":21,"body":[{"type":"ExpressionStatement","start":0,"end":21,"expression":{"type":"BinaryExpression","start":0,"end":21,"left":{"type":"BinaryExpression","start":0,"end":12,"left":{"type":"Literal","start":1,"end":2,"value":3,"raw":"3"},"operator":"+","right":{"type":"BinaryExpression","start":5,"end":11,"left":{"type":"Literal","start":6,"end":8,"value":12,"raw":"12"},"operator":"/","right":{"type":"Literal","start":9,"end":10,"value":4,"raw":"4"}}},"operator":"*","right":{"type":"BinaryExpression","start":15,"end":21,"left":{"type":"Literal","start":16,"end":18,"value":10,"raw":"10"},"operator":"-","right":{"type":"Literal","start":19,"end":20,"value":3,"raw":"3"}}}}]}`

// var x = 43;
// x
// => 43
const variableDecl = `{"type":"Program","start":0,"end":13,"body":[{"type":"VariableDeclaration","start":0,"end":11,"declarations":[{"type":"VariableDeclarator","start":4,"end":10,"id":{"type":"Identifier","start":4,"end":5,"name":"x"},"init":{"type":"Literal","start":8,"end":10,"value":43,"raw":"43"}}],"kind":"var"},{"type":"ExpressionStatement","start":12,"end":13,"expression":{"type":"Identifier","start":12,"end":13,"name":"x"}}]}`

// true?"then":"else"
// => "then"
const condTrue = `{"type":"Program","start":0,"end":18,"body":[{"type":"ExpressionStatement","start":0,"end":18,"expression":{"type":"ConditionalExpression","start":0,"end":18,"test":{"type":"Literal","start":0,"end":4,"value":true,"raw":"true"},"consequent":{"type":"Literal","start":5,"end":11,"value":"then","raw":"\"then\""},"alternate":{"type":"Literal","start":12,"end":18,"value":"else","raw":"\"else\""}}}]}`

// false?"then":"else"
const condFalse = `{"type":"Program","start":0,"end":19,"body":[{"type":"ExpressionStatement","start":0,"end":19,"expression":{"type":"ConditionalExpression","start":0,"end":19,"test":{"type":"Literal","start":0,"end":5,"value":false,"raw":"false"},"consequent":{"type":"Literal","start":6,"end":12,"value":"then","raw":"\"then\""},"alternate":{"type":"Literal","start":13,"end":19,"value":"else","raw":"\"else\""}}}]}`

// if(true) {
//     "then";
// }
// else {
//     "else";
// }
// => "then"
const ifTrue = `{"type":"Program","start":0,"end":45,"body":[{"type":"IfStatement","start":0,"end":45,"test":{"type":"Literal","start":3,"end":7,"value":true,"raw":"true"},"consequent":{"type":"BlockStatement","start":9,"end":24,"body":[{"type":"ExpressionStatement","start":15,"end":22,"expression":{"type":"Literal","start":15,"end":21,"value":"then","raw":"\"then\""}}]},"alternate":{"type":"BlockStatement","start":30,"end":45,"body":[{"type":"ExpressionStatement","start":36,"end":43,"expression":{"type":"Literal","start":36,"end":42,"value":"else","raw":"\"else\""}}]}}]}`

// if(false) {
//     "then";
// }
// else {
//     "else";
// }
// => "else"
const ifFalse = `{"type":"Program","start":0,"end":46,"body":[{"type":"IfStatement","start":0,"end":46,"test":{"type":"Literal","start":3,"end":8,"value":false,"raw":"false"},"consequent":{"type":"BlockStatement","start":10,"end":25,"body":[{"type":"ExpressionStatement","start":16,"end":23,"expression":{"type":"Literal","start":16,"end":22,"value":"then","raw":"\"then\""}}]},"alternate":{"type":"BlockStatement","start":31,"end":46,"body":[{"type":"ExpressionStatement","start":37,"end":44,"expression":{"type":"Literal","start":37,"end":43,"value":"else","raw":"\"else\""}}]}}]}`

// var x = 0;
// x = 44;
// x
// => 44
const simpleAssignment = `{"type":"Program","start":0,"end":20,"body":[{"type":"VariableDeclaration","start":0,"end":10,"declarations":[{"type":"VariableDeclarator","start":4,"end":9,"id":{"type":"Identifier","start":4,"end":5,"name":"x"},"init":{"type":"Literal","start":8,"end":9,"value":0,"raw":"0"}}],"kind":"var"},{"type":"ExpressionStatement","start":11,"end":18,"expression":{"type":"AssignmentExpression","start":11,"end":17,"operator":"=","left":{"type":"Identifier","start":11,"end":12,"name":"x"},"right":{"type":"Literal","start":15,"end":17,"value":44,"raw":"44"}}},{"type":"ExpressionStatement","start":19,"end":20,"expression":{"type":"Identifier","start":19,"end":20,"name":"x"}}]}`

// var o = {};
// o.foo = 45;
// o.foo
// => 45
const propertyAssignment = `{"type":"Program","start":0,"end":29,"body":[{"type":"VariableDeclaration","start":0,"end":11,"declarations":[{"type":"VariableDeclarator","start":4,"end":10,"id":{"type":"Identifier","start":4,"end":5,"name":"o"},"init":{"type":"ObjectExpression","start":8,"end":10,"properties":[]}}],"kind":"var"},{"type":"ExpressionStatement","start":12,"end":23,"expression":{"type":"AssignmentExpression","start":12,"end":22,"operator":"=","left":{"type":"MemberExpression","start":12,"end":17,"object":{"type":"Identifier","start":12,"end":13,"name":"o"},"property":{"type":"Identifier","start":14,"end":17,"name":"foo"},"computed":false},"right":{"type":"Literal","start":20,"end":22,"value":45,"raw":"45"}}},{"type":"ExpressionStatement","start":24,"end":29,"expression":{"type":"MemberExpression","start":24,"end":29,"object":{"type":"Identifier","start":24,"end":25,"name":"o"},"property":{"type":"Identifier","start":26,"end":29,"name":"foo"},"computed":false}}]}`

// var x = 45;
// x++;
// x++;
// => 46
const postincrement = `{"type":"Program","start":0,"end":21,"body":[{"type":"VariableDeclaration","start":0,"end":11,"declarations":[{"type":"VariableDeclarator","start":4,"end":10,"id":{"type":"Identifier","start":4,"end":5,"name":"x"},"init":{"type":"Literal","start":8,"end":10,"value":45,"raw":"45"}}],"kind":"var"},{"type":"ExpressionStatement","start":12,"end":16,"expression":{"type":"UpdateExpression","start":12,"end":15,"operator":"++","prefix":false,"argument":{"type":"Identifier","start":12,"end":13,"name":"x"}}},{"type":"ExpressionStatement","start":17,"end":21,"expression":{"type":"UpdateExpression","start":17,"end":20,"operator":"++","prefix":false,"argument":{"type":"Identifier","start":17,"end":18,"name":"x"}}}]}`

// var x = 45;
// ++x;
// ++x;
// => 47
const preincrement = `{"type":"Program","start":0,"end":21,"body":[{"type":"VariableDeclaration","start":0,"end":11,"declarations":[{"type":"VariableDeclarator","start":4,"end":10,"id":{"type":"Identifier","start":4,"end":5,"name":"x"},"init":{"type":"Literal","start":8,"end":10,"value":45,"raw":"45"}}],"kind":"var"},{"type":"ExpressionStatement","start":12,"end":16,"expression":{"type":"UpdateExpression","start":12,"end":15,"operator":"++","prefix":true,"argument":{"type":"Identifier","start":14,"end":15,"name":"x"}}},{"type":"ExpressionStatement","start":17,"end":21,"expression":{"type":"UpdateExpression","start":17,"end":20,"operator":"++","prefix":true,"argument":{"type":"Identifier","start":19,"end":20,"name":"x"}}}]}`

// ({foo: "bar", answer: 42})
// => {foo: "bar", answer: 42}
const objectExpression = `{"type":"Program","start":0,"end":26,"body":[{"type":"ExpressionStatement","start":0,"end":26,"expression":{"type":"ObjectExpression","start":0,"end":26,"properties":[{"key":{"type":"Identifier","start":2,"end":5,"name":"foo"},"value":{"type":"Literal","start":7,"end":12,"value":"bar","raw":"\"bar\""},"kind":"init"},{"key":{"type":"Identifier","start":14,"end":20,"name":"answer"},"value":{"type":"Literal","start":22,"end":24,"value":42,"raw":"42"},"kind":"init"}]}}]}`