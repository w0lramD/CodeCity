#!/usr/bin/env node
/**
 * @license
 * Code City interpreter JS test case generator
 *
 * Copyright 2017 Google Inc.
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

/**
 * @fileoverview Tool to translate JS test case spec into golang test.go file.
 * @author cpcallen@google.com (Christopher Allen)
 */

'use strict';

var acorn = require('../nodejs/acorn.js');
var tests = require('./testcases.js');

console.log('// AUTO-GENERATED BY gentests.js. DO NOT EDIT.\n');
console.log('package interpreter\n');
console.log('var tests = []struct {');
console.log('\tname     string');
console.log('\tast      string');
console.log('\tsrc      string');
console.log('\texpected string');
console.log('}{');

var t = tests.tests;
for (var i = 0; i < t.length; i++) {
  var name = t[i].name;
  // Compact source code by removing line breaks and unneeded spaces &
  // semicolons:
  var src = t[i].src.replace(/\s+/g, ' ').trim()
  src = src.replace(/ ?([{}()\[\],;=+*/<>~|&!?:-]+) ?/g, '$1');
  src = src.replace(/;}/g, '}');
  src = src.replace(/;$/, '');
  // Convert the expected value so it can be parsed by data.NewFromRaw():
  var expt = JSON.stringify(t[i].expected);
  console.log('\t{"%s", %s, `%s`, `%s`},', name, name, src, expt);
}

console.log('}\n');

for (var i = 0; i < t.length; i++) {
  console.log('const %s = `%s`\n', t[i].name,
              JSON.stringify(acorn.parse(t[i].src)));
}
