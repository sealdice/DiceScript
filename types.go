/*
  Copyright 2022 fy <fy0748@gmail.com>

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package dicescript

import (
	"math"
	"strconv"
)

type VMValueType int

const (
	VMTypeInt64         VMValueType = 0
	VMTypeFloat64       VMValueType = 1
	VMTypeString        VMValueType = 2
	VMTypeUndefined     VMValueType = 3
	VMTypeNone          VMValueType = 4
	VMTypeComputedValue VMValueType = 5
	VMTypeArray         VMValueType = 6
)

var binOperator = []func(*VMValue, *Context, *VMValue) *VMValue{
	(*VMValue).OpAdd,
	(*VMValue).OpSub,
	(*VMValue).OpMultiply,
	(*VMValue).OpDivide,
	(*VMValue).OpModulus,
	(*VMValue).OpPower,

	(*VMValue).OpCompLT,
	(*VMValue).OpCompLE,
	(*VMValue).OpCompEQ,
	(*VMValue).OpCompNE,
	(*VMValue).OpCompGE,
	(*VMValue).OpCompGT,
}

type RollExtraFlags struct {
	DiceMinMode        bool  // 骰子以最小值结算，用于获取下界
	DiceMaxMode        bool  // 以最大值结算 获取上界
	DisableLoadVarname bool  // 不允许加载变量，这是为了防止遇到 .r XXX 被当做属性读取，而不是“由于XXX，骰出了”
	IgnoreDiv0         bool  // 当div0时暂不报错
	DefaultDiceSideNum int64 // 默认骰子面数
	PrintBytecode      bool  // 执行时打印字节码
}

type Context struct {
	parser         *Parser
	subThread      *Context // 用于执行子句
	subThreadDepth int

	code      []ByteCode
	codeIndex int

	stack []VMValue
	top   int

	NumOpCount int64 // 算力计数
	//CocFlagVarPrefix string // 解析过程中出现，当VarNumber开启时有效，可以是困难极难常规大成功

	Flags RollExtraFlags // 标记
	Error error          // 报错信息

	Ret       *VMValue // 返回值
	RestInput string   // 剩余字符串
	Matched   string   // 匹配的字符串

	ValueStoreNameFunc func(name string, v *VMValue)
	ValueLoadNameFunc  func(name string) *VMValue
}

func (e *Context) Init(stackLength int) {
	e.code = make([]ByteCode, stackLength)
}

type VMValue struct {
	TypeId      VMValueType `json:"typeId"`
	Value       interface{} `json:"value"`
	ExpiredTime int64       `json:"expiredTime"`
}

func (v *VMValue) Clone() *VMValue {
	vNew := &VMValue{TypeId: v.TypeId, Value: v.Value}
	// TODO: 针对特定类型，进行Value的处理，不过大多数时候应该够用
	switch v.TypeId {
	}
	return vNew
}

func (v *VMValue) AsBool() bool {
	switch v.TypeId {
	case VMTypeInt64:
		return v.Value != int64(0)
	case VMTypeString:
		return v.Value != ""
	case VMTypeNone, VMTypeUndefined:
		return false
	//case VMTypeComputedValue:
	//	vd := v.Value.(*VMComputedValueData)
	//	return vd.BaseValue.AsBool()
	default:
		return false
	}
}

func (v *VMValue) ToString() string {
	if v == nil {
		return "NIL"
	}
	if v.Value == nil {
		return "unknown"
	}
	switch v.TypeId {
	case VMTypeInt64:
		return strconv.FormatInt(v.Value.(int64), 10)
	case VMTypeFloat64:
		return strconv.FormatFloat(v.Value.(float64), 'f', 2, 64)
	case VMTypeString:
		return v.Value.(string)
	case VMTypeUndefined:
		return "undefined"
	case VMTypeNone:
		return "null"
	//case VMTypeComputedValue:
	//vd := v.Value.(*VMComputedValueData)
	//return vd.BaseValue.ToString() + "=> (" + vd.Expr + ")"
	default:
		return "a value"
	}
}

func (v *VMValue) ReadInt64() (int64, bool) {
	if v.TypeId == VMTypeInt64 {
		return v.Value.(int64), true
	}
	return 0, false
}

func (v *VMValue) ReadFloat64() (float64, bool) {
	if v.TypeId == VMTypeFloat64 {
		return v.Value.(float64), true
	}
	return 0, false
}

func (v *VMValue) ReadString() (string, bool) {
	if v.TypeId == VMTypeString {
		return v.Value.(string), true
	}
	return "", false
}

func (v *VMValue) OpAdd(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(int64) + v2.Value.(int64)
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := float64(v.Value.(int64)) + v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(float64) + float64(v2.Value.(int64))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := v.Value.(float64) + v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeString:
		switch v2.TypeId {
		case VMTypeString:
			val := v.Value.(string) + v2.Value.(string)
			return VMValueNewStr(val)
		}
	case VMTypeComputedValue:
		// TODO:
	case VMTypeArray:
		// TODO:
	}

	return nil
}

func (v *VMValue) OpSub(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(int64) - v2.Value.(int64)
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := float64(v.Value.(int64)) - v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(float64) - float64(v2.Value.(int64))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := v.Value.(float64) - v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	}

	return nil
}

func (v *VMValue) OpMultiply(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			// TODO: 溢出，均未考虑溢出
			val := v.Value.(int64) * v2.Value.(int64)
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := float64(v.Value.(int64)) * v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(float64) * float64(v2.Value.(int64))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := v.Value.(float64) * v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	}

	return nil
}

func (v *VMValue) OpDivide(ctx *Context, v2 *VMValue) *VMValue {
	// TODO: 被除数为0
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(int64) / v2.Value.(int64)
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := float64(v.Value.(int64)) / v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(float64) / float64(v2.Value.(int64))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := v.Value.(float64) / v2.Value.(float64)
			return VMValueNewFloat64(val)
		}
	}

	return nil
}

func (v *VMValue) OpModulus(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := v.Value.(int64) % v2.Value.(int64)
			return VMValueNewInt64(val)
		}
	}

	return nil
}

func (v *VMValue) OpPower(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := int64(math.Pow(float64(v.Value.(int64)), float64(v2.Value.(int64))))
			return VMValueNewInt64(val)
		case VMTypeFloat64:
			val := math.Pow(float64(v.Value.(int64)), v2.Value.(float64))
			return VMValueNewFloat64(val)
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			val := math.Pow(v.Value.(float64), float64(v2.Value.(int64)))
			return VMValueNewFloat64(val)
		case VMTypeFloat64:
			val := math.Pow(v.Value.(float64), v2.Value.(float64))
			return VMValueNewFloat64(val)
		}
	}

	return nil
}

func boolToVMValue(v bool) *VMValue {
	var val int64
	if v {
		val = 1
	}
	return VMValueNewInt64(val)
}

func (v *VMValue) OpCompLT(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(int64) < v2.Value.(int64))
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) < v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) < float64(v2.Value.(int64)))
		case VMTypeFloat64:
			return boolToVMValue(v.Value.(float64) < v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompLE(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(int64) <= v2.Value.(int64))
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) <= v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) <= float64(v2.Value.(int64)))
		case VMTypeFloat64:
			return boolToVMValue(v.Value.(float64) <= v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompEQ(ctx *Context, v2 *VMValue) *VMValue {
	if v == v2 {
		return VMValueNewInt64(1)
	}
	if v.TypeId == v2.TypeId {
		return boolToVMValue(v.Value == v2.Value)
	}

	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) == v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) == float64(v2.Value.(int64)))
		}
	}

	return VMValueNewInt64(0)
}

func (v *VMValue) OpCompNE(ctx *Context, v2 *VMValue) *VMValue {
	ret := v.OpCompEQ(ctx, v2)
	return boolToVMValue(!ret.AsBool())
}

func (v *VMValue) OpCompGE(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(int64) >= v2.Value.(int64))
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) >= v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) >= float64(v2.Value.(int64)))
		case VMTypeFloat64:
			return boolToVMValue(v.Value.(float64) >= v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpCompGT(ctx *Context, v2 *VMValue) *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(int64) > v2.Value.(int64))
		case VMTypeFloat64:
			return boolToVMValue(float64(v.Value.(int64)) > v2.Value.(float64))
		}
	case VMTypeFloat64:
		switch v2.TypeId {
		case VMTypeInt64:
			return boolToVMValue(v.Value.(float64) > float64(v2.Value.(int64)))
		case VMTypeFloat64:
			return boolToVMValue(v.Value.(float64) > v2.Value.(float64))
		}
	}

	return nil
}

func (v *VMValue) OpPositive() *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		return VMValueNewInt64(v.Value.(int64))
	case VMTypeFloat64:
		return VMValueNewFloat64(v.Value.(float64))
	}
	return nil
}

func (v *VMValue) OpNegation() *VMValue {
	switch v.TypeId {
	case VMTypeInt64:
		return VMValueNewInt64(-v.Value.(int64))
	case VMTypeFloat64:
		return VMValueNewFloat64(-v.Value.(float64))
	}
	return nil
}

func (v *VMValue) GetTypeName() string {
	switch v.TypeId {
	case VMTypeInt64:
		return "int64"
	case VMTypeFloat64:
		return "float64"
	case VMTypeString:
		return "str"
	case VMTypeUndefined:
		return "undefined"
	case VMTypeNone:
		return "none"
	case VMTypeComputedValue:
		return "computed"
	case VMTypeArray:
		return "array"
	}
	return "unknown"
}

func VMValueNewInt64(i int64) *VMValue {
	// TODO: 小整数可以处理为不可变对象，且一直停留在内存中，就像python那样。这可以避免很多内存申请
	return &VMValue{TypeId: VMTypeInt64, Value: i}
}

func VMValueNewFloat64(i float64) *VMValue {
	return &VMValue{TypeId: VMTypeFloat64, Value: i}
}

func VMValueNewStr(s string) *VMValue {
	return &VMValue{TypeId: VMTypeString, Value: s}
}

func VMValueNewUndefined() *VMValue {
	return &VMValue{TypeId: VMTypeUndefined}
}

func VMValueNewNone() *VMValue {
	return &VMValue{TypeId: VMTypeNone}
}
