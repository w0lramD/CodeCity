/**
 * @license
 * Copyright 2017 Google LLC
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
 * @fileoverview Saving and restoring the state of the interpreter.
 * @author fraser@google.com (Neil Fraser)
 */
'use strict';

var Interpreter = require('./interpreter');
var IterableWeakMap = require('./iterable_weakmap');
var IterableWeakSet = require('./iterable_weakset');
var net = require('net');
var Node = require('./parser').Node;
var Registry = require('./registry');

var Serializer = {};

/** !Set<!Object> Prototypes of types to be excluded from serialization. */
Serializer.excludeTypes = new Set([net.Socket.prototype, net.Server.prototype]);

/**
 * Per-prototype sets of properties to be excluded from serialization.
 * @type {!Map<!Object,!Array<string>>}
 */
Serializer.pruneProperties = new Map([
  [Interpreter.prototype, [
    'hrStartTime_',
    'previousTime_',
    'runner_',
    'Object',
    'Function',
    'UserFunction',
    'BoundFunction',
    'NativeFunction',
    'OldNativeFunction',
    'Array',
    'Date',
    'RegExp',
    'Error',
    'Arguments',
    'WeakMap',
    'Thread',
    'Box',
    'Server',
  ]],
  [IterableWeakMap.prototype, [
    'refs_',
    'finalisers_',
  ]],
  [IterableWeakSet.prototype, [
    'refs_',
    'map_',
    'finalisers_',
  ]],
]);

/**
 * Deserialize the provided JSON-compatible object into an interpreter.
 * @param {!Object} JSON-compatible object.
 * @param {!Interpreter} intrp JS-Interpreter instance.
 */
Serializer.deserialize = function(json, intrp) {
  function decodeValue(value) {
    if (value && typeof value === 'object') {
      var data;
      if ((data = value['#'])) {
       // Object reference: {'#': 42}
       value = objectList[data];
        if (!value) {
          throw new ReferenceError('Object reference not found: ' + data);
        }
        return value;
      }
      if ((data = value['Number'])) {
        // Special number: {'Number': 'Infinity'}
        return Number(data);
      }
      if ((data = value['Value'])) {
        // Special value: {'Value': 'undefined'}
        if (value['Value'] === 'undefined') {
          return undefined;
        }
      }
    }
    return value;
  }

  if (!Array.isArray(json)) {
    throw new TypeError('Top-level JSON is not a list.');
  }

  // Require native functions to be present.  Can't just create fresh
  // new interpreter instance because client code may want to add
  // custom builtins.
  if (!intrp.global) {
    throw new Error(
        'Interpreter must be initialized prior to deserialization.');
  }

  // Find all native functions to get id => func mappings.
  var functionHash = Object.create(null);
  // Builtins.
  var builtins = Array.from(intrp.builtins.values());
  var implProps = ['impl', 'call', 'construct'];
  for (var i = 0; i < builtins.length; i++) {
    var builtin = builtins[i];
    for (var j = 0; j < implProps.length; j++) {
      var func = builtin[implProps[j]];
      if (func) functionHash[func.id] = func;
    }
  }
  // Step functions.
  for (var stepName in intrp.stepFuncs) {
    var stepFunc = intrp.stepFuncs[stepName];
    functionHash[stepFunc.id] = stepFunc;
  }

  // Get constructors
  var constructors = Serializer.getTypesDeserialize_(intrp);

  // First pass: Create object stubs for every object.  We don't need
  // to (re)create object #0, because that's the interpreter proper.
  var objectList = [intrp];
  for (var i = 1; i < json.length; i++) {
    var jsonObj = json[i];
    var obj;
    var type = jsonObj['type'];
    switch (type) {
      case 'Object':
        obj = {};
        break;
      case 'Function':
        obj = functionHash[jsonObj['id']];
        if (!obj) {
          throw new RangeError('Function ID not found: ' + jsonObj['id']);
        }
        break;
      case 'Array':
        obj = [];
        break;
      case 'Date':
        obj = new Date(jsonObj['data']);
        if (isNaN(obj)) {
          throw new TypeError('Invalid date: ' + jsonObj['data']);
        }
        break;
      case 'RegExp':
        obj = RegExp(jsonObj['source'], jsonObj['flags']);
        break;
      case 'Map':
        obj = new Map;
        break;
      case 'Set':
        obj = new Set;
        break;
      case 'IterableWeakMap':
        obj = new IterableWeakMap;
        break;
      case 'IterableWeakSet':
        obj = new IterableWeakSet;
        break;
      case 'Registry':
        obj = new Registry;
        break;
      case 'State':
        // TODO(cpcallen): this is just a little performance kludge so
        // that the State constructor doesn't need a conditional in it.
        // Find a more general solution to constructors requiring args.
        obj = new Interpreter.State(/** @type {?} */({}),
            /** @type {?} */(undefined));
        break;
      default:
        if (constructors[type]) {
          obj = new constructors[type];
        } else {
          throw new TypeError('Unknown type: ' + jsonObj['type']);
        }
    }
    objectList[i] = obj;
  }
  // Second pass: Populate properties for every object.
  for (var i = 0; i < json.length; i++) {
    var jsonObj = json[i];
    var obj = objectList[i];
    // Set prototype, if specified.
    if (jsonObj['proto']) {
      Object.setPrototypeOf(obj, decodeValue(jsonObj['proto']));
    }
    // Repopulate properties.
    var props = jsonObj['props'];
    if (props) {
      var nonConfigurable = jsonObj['nonConfigurable'] || [];
      var nonEnumerable = jsonObj['nonEnumerable'] || [];
      var nonWritable = jsonObj['nonWritable'] || [];
      var keys = Object.getOwnPropertyNames(props);
      for (var j = 0; j < keys.length; j++) {
        var key = keys[j];
        Object.defineProperty(obj, key,
            {configurable: !nonConfigurable.includes(key),
             enumerable: !nonEnumerable.includes(key),
             writable: !nonWritable.includes(key),
             value: decodeValue(props[key])});
      }
    }
    // Repopulate sets.
    if (obj instanceof Set || obj instanceof IterableWeakSet) {
      var data = jsonObj['data'];
      if (data) {
        for (var j = 0; j < data.length; j++) {
          obj.add(decodeValue(data[j]));
        }
      }
    }
    // Repopulate maps.
    if (obj instanceof Map || obj instanceof IterableWeakMap) {
      var entries = jsonObj['entries'];
      if (entries) {
        for (var j = 0; j < entries.length; j++) {
          var key = decodeValue(entries[j][0]);
          var value = decodeValue(entries[j][1]);
          obj.set(key, value);
        }
      }
    }
    if (jsonObj['isExtensible'] === false) {  // N.B. normally omitted if true.
      Object.preventExtensions(obj);
    }
  }
  // Finally: fixup interpreter state, post-deserialization.
  intrp.postDeserialize();
};

/**
 * Serialize the provided interpreter.
 * @param {!Interpreter} intrp JS-Interpreter instance.
 * @return {!Object} JSON-compatible object.
 */
Serializer.serialize = function(intrp) {
  // First: prepare interpreter for serialization.
  intrp.preSerialize();
  
  function encodeValue(value) {
    if (value && (typeof value === 'object' || typeof value === 'function')) {
      // TODO(cpcallen): For some reason the Closure Compiler thinks
      // value might be null at this point (it can't be), and
      // complains about passing it to Object.getPrototypeOf.  Remove
      // this type-narrowing check once this compiler bug is fixed.
      if (!value) throw new Error();
      if (Serializer.excludeTypes.has(Object.getPrototypeOf(value))) {
        return null;
      }
      var ref = objectRefs.get(value);
      if (ref === undefined) {
        throw new RangeError('Object not found in table.');
      }
      return {'#': ref};
    }
    if (value === undefined) {
      return {'Value': 'undefined'};
    }
    if (typeof value === 'number') {
      if (value === Infinity) {
        return {'Number': 'Infinity'};
      } else if (value === -Infinity) {
        return {'Number': '-Infinity'};
      } else if (Number.isNaN(value)) {
        return {'Number': 'NaN'};
      } else if (Object.is(value, -0)) {
        return {'Number': '-0'};
      }
    }
    return value;
  }

  // Find all objects.
  var objectList = Serializer.getObjectList_(intrp);
  // Build reverse-lookup cache.
  var /** !Map<Object,number> */ objectRefs = new Map();
  for (var i = 0; i < objectList.length; i++) {
    objectRefs.set(objectList[i], i);
  }
  // Get types.
  var types = Serializer.getTypesSerialize_(intrp);
  // Serialize every object.
  var json = [];
  for (var i = 0; i < objectList.length; i++) {
    var jsonObj = Object.create(null);
    json.push(jsonObj);
    var obj = objectList[i];
    // TODO: Add a flag on the '#' prop.  On for debugging, off for production.
    if (true) {
      jsonObj['#'] = i;
    }
    var proto = Object.getPrototypeOf(obj);
    switch (proto) {
      case Object.prototype:
        jsonObj['type'] = 'Object';
        break;
      case Function.prototype:
        jsonObj['type'] = 'Function';
        jsonObj['id'] = obj.id;
        if (!obj.id) {
          throw new Error('Native function has no ID: ' + obj);
        }
        continue;  // No need to index properties.
      case Array.prototype:
        jsonObj['type'] = 'Array';
        break;
      case Date.prototype:
        jsonObj['type'] = 'Date';
        jsonObj['data'] = obj.toJSON();
        continue;  // No need to index properties.
      case RegExp.prototype:
        jsonObj['type'] = 'RegExp';
        jsonObj['source'] = obj.source;
        jsonObj['flags'] = obj.flags;
        continue;  // No need to index properties.
      case Map.prototype:
        jsonObj['type'] = 'Map';
        if (obj.size) {
          jsonObj['entries'] =
              Array.from(/** @type {?} */(obj),function(entry) {
                var key = encodeValue(entry[0]);
                var value = encodeValue(entry[1]);
                return [key, value];
              });
        }
        break;
      case Set.prototype:
        jsonObj['type'] = 'Set';
        if (obj.size) {
          jsonObj['data'] = Array.from(obj.values(), encodeValue);
        }
        break;
      case IterableWeakMap.prototype:
        jsonObj['type'] = 'IterableWeakMap';
        if (obj.size) {
          jsonObj['entries'] =
              Array.from(/** @type {?} */(obj), function(entry) {
                var key = encodeValue(entry[0]);
                var value = encodeValue(entry[1]);
                return [key, value];
              });
        }
        continue;  // Mustn't index internal properties for IterableWeakMap
      case IterableWeakSet.prototype:
        jsonObj['type'] = 'IterableWeakSet';
        if (obj.size) {
          jsonObj['data'] = Array.from(obj.values(), encodeValue);
        }
        continue;  // Mustn't index internal properties for IterableWeakSet
      case Registry.prototype:
        jsonObj['type'] = 'Registry';
        break;
      default:
        var type = types.get(proto);
        if (type) {
          jsonObj['type'] = type;
        } else {
          jsonObj['type'] = Array.isArray(obj) ? 'Array' : 'Object';
          jsonObj['proto'] = encodeValue(proto);
        }
    }
    var props = Object.create(null);
    var nonConfigurable = [];
    var nonEnumerable = [];
    var nonWritable = [];
    var prune = (proto && Serializer.pruneProperties.get(proto)) || [];
    var keys = Object.getOwnPropertyNames(obj);
    for (var j = 0; j < keys.length; j++) {
      var key = keys[j];
      if (prune.includes(key)) continue;
      // Skip [[Socket]] slot on connected objects.
      // TODO(cpcallen): this is pretty kludgy.  Try to find a better way.
      if (obj instanceof intrp.Object && key === 'socket') continue;

      props[key] = encodeValue(obj[key]);
      var descriptor = Object.getOwnPropertyDescriptor(obj, key);
      if (!descriptor.configurable) {
        nonConfigurable.push(key);
      }
      if (!descriptor.enumerable) {
        nonEnumerable.push(key);
      }
      if (!descriptor.writable) {
        nonWritable.push(key);
      }
    }
    if (Object.getOwnPropertyNames(keys).length) {
      jsonObj['props'] = props;
    }
    if (nonConfigurable.length) {
      jsonObj['nonConfigurable'] = nonConfigurable;
    }
    if (nonEnumerable.length) {
      jsonObj['nonEnumerable'] = nonEnumerable;
    }
    if (nonWritable.length) {
      jsonObj['nonWritable'] = nonWritable;
    }
    if (!Object.isExtensible(obj)) {
      jsonObj['isExtensible'] = false;
    }
  }
  return json;
};

/**
 * Recursively search node to find all non-primitives.
 *
 * TODO(cpcallen): use a Registry instead of Array for objectList;
 *     this would allow more readable references by using paths
 *     instead of numerical indices.
 * @param {*} node JavaScript value to search.
 * @return {!Array<!Object>} objectList Array of all objects found via node.
 */
Serializer.getObjectList_ = function(node) {
  var seen = new Set();
  Serializer.objectHunt_(node, seen);
  return Array.from(seen.keys());
}

/**
 * Recursively search node find all non-primitives.
 *
 * @param {*} node JavaScript value to search.
 * @param {!Set<?Object>} seen Set of objects found so far.
 */
Serializer.objectHunt_ = function(node, seen) {
  if (!node || (typeof node !== 'object' && typeof node !== 'function')) {
    // node is primitive.  Nothing to do.
    return;
  }
  var obj = /** @type {!Object} */(node);
  var proto = Object.getPrototypeOf(/** @type {!Object} */(obj));
  if (Serializer.excludeTypes.has(proto) ||
      seen.has(/** @type {!Object} */(obj))) {
    return;
  }
  seen.add(obj);
  if (typeof obj === 'object') {  // Recurse.
    // Properties.
    var prune = (proto && Serializer.pruneProperties.get(proto)) || [];
    var keys = Object.getOwnPropertyNames(obj);
    for (var i = 0; i < keys.length; i++) {
      var key = keys[i];
      if (prune.includes(key)) continue;
      Serializer.objectHunt_(obj[key], seen);
    }
    // Set members.
    if (obj instanceof Set || obj instanceof IterableWeakSet) {
      obj.forEach(function(value) {
        Serializer.objectHunt_(value, seen);
      });
    }
    // Map entries.
    if (obj instanceof Map || obj instanceof IterableWeakMap) {
      obj.forEach(function(value, key) {
        Serializer.objectHunt_(key, seen);
        Serializer.objectHunt_(value, seen);
      });
    }
  }
};

/**
 * Make a map of typename to contructor for each type that might be
 * found while serializing an Interpreter instance.
 * @param {!Interpreter} intrp The interpreter instance being serialized
 *     (needed for inner classes).
 * @return {!Object} A key/value map of typesnames to constructors.
 */
Serializer.getTypesDeserialize_ = function(intrp) {
  return {
    'Interpreter': Interpreter,
    'Scope': Interpreter.Scope,
    'State': Interpreter.State,
    'Thread': Interpreter.Thread,
    'PropertyIterator': Interpreter.PropertyIterator,
    'Source': Interpreter.Source,
    'PseudoObject': intrp.Object,
    'PseudoFunction': intrp.Function,
    'PseudoUserFunction': intrp.UserFunction,
    'PseudoBoundFunction': intrp.BoundFunction,
    'PseudoNativeFunction': intrp.NativeFunction,
    'PseudoOldNativeFunction': intrp.OldNativeFunction,
    'PseudoArray': intrp.Array,
    'PseudoDate': intrp.Date,
    'PseudoRegExp': intrp.RegExp,
    'PseudoError': intrp.Error,
    'PseudoArguments': intrp.Arguments,
    'PseudoWeakMap': intrp.WeakMap,
    'PseudoThread': intrp.Thread,
    'Box': intrp.Box,
    'Server': intrp.Server,
    'Node': Node,
  };
};

/**
 * Make a map of prototype to typename for each of the types that
 * might be found while deserializing an Interpreter instance.
 * @param {!Interpreter} intrp An interpreter instance being
 *     deserialized into (needed for inner classes).
 * @return {!Map} A key/value map of protoytype objects to typesnames.
 */
Serializer.getTypesSerialize_ = function(intrp) {
  var types = Serializer.getTypesDeserialize_(intrp);
  var map = new Map;
  for (var t in types) {
    if (types.hasOwnProperty(t)) {
      map.set(types[t].prototype, t);
    }
  }
  return map;
};

module.exports = Serializer;
