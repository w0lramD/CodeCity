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

package object

import (
	"math"
	"testing"
)

func TestPrimitiveFromRaw(t *testing.T) {
	var tests = []struct {
		raw      string
		expected Value
	}{
		{`true`, Boolean(true)},
		{`false`, Boolean(false)},
		{`undefined`, Undefined{}},
		{`null`, Null{}},
		{`42`, Number(42)},
		{`"Hello, World!"`, String("Hello, World!")},
		// {`'Hello, World!'`, String("Hello, World!")}, // FIXME
		{`"foo'bar\"baz"`, String(`foo'bar"baz`)},
		// {`'foo\'bar"baz'`, String(`foo'bar"baz`)}, // FIXME
	}

	for _, c := range tests {
		if v := NewFromRaw(c.raw); v != c.expected {
			t.Errorf("newFromRaw(%v) == %v (%T)\n(expected %v (%T))",
				c.raw, v, v, c.expected, c.expected)
		}
	}
}

func TestIsTruthy(t *testing.T) {
	var tests = []struct {
		input    Value
		expected bool
	}{
		{Boolean(true), true},
		{Boolean(false), false},
		{Null{}, false},
		{Undefined{}, false},
		{String(""), false},
		{String("foo"), true},
		{String("0"), true},
		{String("false"), true},
		{String("null"), true},
		{String("undefined"), true},
		{Number(0), false},
		{Number(-0), false},
		{Number(0.0), false},
		{Number(-0.0), false},
		{Number(1), true},
		{Number(math.Inf(+1)), true},
		{Number(math.Inf(-1)), true},
		{Number(math.NaN()), false},
		{Number(math.MaxFloat64), true},
		{Number(math.SmallestNonzeroFloat64), true},
	}
	for _, c := range tests {
		if v := c.input.ToBoolean(); v != Boolean(c.expected) {
			t.Errorf("IsTruthy(%v) (%T) == %v", c.input, c.input, v)
		}
	}
}

func TestPrimitivesPrimitiveness(t *testing.T) {
	var prims [5]Value
	prims[0] = Boolean(false)
	prims[1] = Number(42)
	prims[2] = String("Hello, world!")
	prims[3] = Null{}
	prims[4] = Undefined{}

	for i := 0; i < len(prims); i++ {
		if !prims[i].IsPrimitive() {
			t.Errorf("%v.isPrimitive() = false", prims[i])
		}
	}
}

func TestBoolean(t *testing.T) {
	b := Boolean(false)
	if b.Parent() != Value(BooleanProto) {
		t.Errorf("%v.Parent() != BooleanProto", b)
	}
	if b.Parent().Parent() != Value(ObjectProto) {
		t.Errorf("%v.Parent().Parent() != ObjectProto", b)
	}
}

func TestNumber(t *testing.T) {
	n := Number(0)
	if n.Parent() != Value(NumberProto) {
		t.Errorf("%v.Parent() != NumberProto", n)
	}
	if n.Parent().Parent() != Value(ObjectProto) {
		t.Errorf("%v.Parent().Parent() != ObjectProto", n)
	}
}

func TestString(t *testing.T) {
	var s Value = String("")
	if s.Parent() != Value(StringProto) {
		t.Errorf("%v.Parent() != StringProto", s)
	}
	if s.Parent().Parent() != Value(ObjectProto) {
		t.Errorf("%v.Parent().Parent() != ObjectProto", s)
	}
}

func TestStringLength(t *testing.T) {
	v, err := String("").GetProperty("length")
	if v != Number(0) || err != nil {
		t.Errorf("String(\"\").GetProperty(\"length\") == %v, %v"+
			"(expected 0, nil)", v, err)
	}

	v, err = String("Hello, World!").GetProperty("length")
	if v != Number(13) || err != nil {
		t.Errorf("String(\"కోడ్ సిటీ\").GetProperty(\"length\") == %v, %v "+
			"(expected 13, nil)", v, err)
	}

	// "Code City" in Telugu (according to translate.google.com):
	v, err = String("కోడ్ సిటీ").GetProperty("length")
	if v != Number(9) || err != nil {
		t.Errorf("String(\"కోడ్ సిటీ\").GetProperty(\"length\") == %v, %v "+
			"(expected 9, nil)", v, err)
	}

}

func TestNull(t *testing.T) {
	n := Null{}
	if v := n.Type(); v != "object" {
		t.Errorf("Null{}.Type() == %v (expected \"object\")", v)
	}
	if v, e := n.GetProperty("foo"); e == nil {
		t.Errorf("Null{}.GetProperty(\"foo\") == %v, %v "+
			"(expected nil, !nil)", v, e)
	}
}

func TestNullParentPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Null{}.Parent() did not panic")
		}
	}()
	_ = Null{}.Parent()
}

func TestUndefined(t *testing.T) {
}

func TestUndefinedParentPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Undefined{}.Parent() did not panic")
		}
	}()
	_ = Undefined{}.Parent()
}

func TestToString(t *testing.T) {
	var tests = []struct {
		input    Value
		expected string
	}{
		{Boolean(true), "true"},
		{Boolean(false), "false"},
		{Null{}, "null"},
		{Undefined{}, "undefined"},
		{String(""), ""},
		{String("foo"), "foo"},
		{String("\"foo\""), "\"foo\""},
		{Number(0), "0"},
		{Number(math.Copysign(0, -1)), "-0"},
		{Number(math.Inf(+1)), "Infinity"},
		{Number(math.Inf(-1)), "-Infinity"},
		{Number(math.NaN()), "NaN"},
		// FIXME: add test cases for decimal -> scientific notation
		// transition threshold.
	}
	for _, c := range tests {
		if v := c.input.ToString(); v != String(c.expected) {
			t.Errorf("%v.ToString() (input type %T) == %v "+
				"(expected %v)", c.input, c.input, v, c.expected)
		}
	}
}