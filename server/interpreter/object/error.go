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

// ErrorMsg is a (name, description)-tuple used to report error
// conditions inside the JavaScript interpreter.  It's not actually a
// JS Error object, but might get turned into one if appropriate.
type ErrorMsg struct {
	Name    string
	Message string
}

// *ErrorMsg must satisfy error.
var _ error = (*ErrorMsg)(nil)

func (this ErrorMsg) Error() string {
	return this.Name + ": " + this.Message
}