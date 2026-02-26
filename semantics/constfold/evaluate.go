// Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package constfold

import (
	"math"
	"strconv"

	"ballerina-lang-go/model"
)

type constValue interface {
	int64 | float64 | string | bool
}

type literalValue struct {
	val any
}

func newLiteralValue[T constValue](v T) literalValue {
	return literalValue{val: v}
}

func evaluateBinary(lhs, rhs literalValue, op model.OperatorKind) (literalValue, bool) {
	switch l := lhs.val.(type) {
	case string:
		if r, ok := rhs.val.(string); ok {
			return evaluateStringBinary(l, r, op)
		}
	case bool:
		if r, ok := rhs.val.(bool); ok {
			return evaluateBoolBinary(l, r, op)
		}
	case int64:
		switch r := rhs.val.(type) {
		case int64:
			return evaluateIntBinary(l, r, op)
		case float64:
			return evaluateFloatBinary(float64(l), r, op)
		}
	case float64:
		switch r := rhs.val.(type) {
		case float64:
			return evaluateFloatBinary(l, r, op)
		case int64:
			return evaluateFloatBinary(l, float64(r), op)
		}
	}
	return literalValue{}, false
}

func evaluateUnary(val literalValue, op model.OperatorKind) (literalValue, bool) {
	switch op {
	case model.OperatorKind_SUB:
		switch v := val.val.(type) {
		case int64:
			return newLiteralValue(-v), true
		case float64:
			return newLiteralValue(-v), true
		}
	case model.OperatorKind_ADD:
		switch val.val.(type) {
		case int64, float64:
			return val, true
		}
	case model.OperatorKind_BITWISE_COMPLEMENT:
		if v, ok := val.val.(int64); ok {
			return newLiteralValue(^v), true
		}
	case model.OperatorKind_NOT:
		if v, ok := val.val.(bool); ok {
			return newLiteralValue(!v), true
		}
	}
	return literalValue{}, false
}

func evaluateIntBinary(lhs, rhs int64, op model.OperatorKind) (literalValue, bool) {
	switch op {
	case model.OperatorKind_ADD:
		return newLiteralValue(lhs + rhs), true
	case model.OperatorKind_SUB:
		return newLiteralValue(lhs - rhs), true
	case model.OperatorKind_MUL:
		return newLiteralValue(lhs * rhs), true
	case model.OperatorKind_DIV:
		if rhs == 0 {
			return literalValue{}, false
		}
		return newLiteralValue(lhs / rhs), true
	case model.OperatorKind_MOD:
		if rhs == 0 {
			return literalValue{}, false
		}
		return newLiteralValue(lhs % rhs), true
	case model.OperatorKind_BITWISE_AND:
		return newLiteralValue(lhs & rhs), true
	case model.OperatorKind_BITWISE_OR:
		return newLiteralValue(lhs | rhs), true
	case model.OperatorKind_BITWISE_XOR:
		return newLiteralValue(lhs ^ rhs), true
	case model.OperatorKind_BITWISE_LEFT_SHIFT:
		return newLiteralValue(lhs << (uint64(rhs) % 64)), true
	case model.OperatorKind_BITWISE_RIGHT_SHIFT:
		return newLiteralValue(lhs >> (uint64(rhs) % 64)), true
	case model.OperatorKind_BITWISE_UNSIGNED_RIGHT_SHIFT:
		return newLiteralValue(int64(uint64(lhs) >> (uint64(rhs) % 64))), true
	}
	return literalValue{}, false
}

func evaluateFloatBinary(lhs, rhs float64, op model.OperatorKind) (literalValue, bool) {
	switch op {
	case model.OperatorKind_ADD:
		return newLiteralValue(lhs + rhs), true
	case model.OperatorKind_SUB:
		return newLiteralValue(lhs - rhs), true
	case model.OperatorKind_MUL:
		return newLiteralValue(lhs * rhs), true
	case model.OperatorKind_DIV:
		return newLiteralValue(lhs / rhs), true
	case model.OperatorKind_MOD:
		return newLiteralValue(math.Mod(lhs, rhs)), true
	}
	return literalValue{}, false
}

func evaluateStringBinary(lhs, rhs string, op model.OperatorKind) (literalValue, bool) {
	if op == model.OperatorKind_ADD {
		return newLiteralValue(lhs + rhs), true
	}
	return literalValue{}, false
}

func evaluateBoolBinary(lhs, rhs bool, op model.OperatorKind) (literalValue, bool) {
	switch op {
	case model.OperatorKind_AND:
		return newLiteralValue(lhs && rhs), true
	case model.OperatorKind_OR:
		return newLiteralValue(lhs || rhs), true
	}
	return literalValue{}, false
}

func formatFloat(f float64) string {
	if math.IsNaN(f) {
		return "NaN"
	}
	if math.IsInf(f, 1) {
		return "Infinity"
	}
	if math.IsInf(f, -1) {
		return "-Infinity"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}
