package dicescript

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func vmValueEqual(vm *Context, aKey string, bValue *VMValue) bool {
	return valueEqual(vm.attrs.MustLoad(aKey), bValue)
}

func simpleExecute(t *testing.T, expr string, ret *VMValue) *Context {
	vm := NewVM()
	err := vm.Run(expr)
	if err != nil {
		fmt.Println(vm.GetAsmText())
		t.Errorf("VM Error: %s, %s", expr, err.Error())
		return vm
	}
	if !valueEqual(vm.Ret, ret) {
		fmt.Println(vm.GetAsmText())
		t.Errorf("not equal: %s %s", ret.ToString(), vm.Ret.ToString())
	}
	return vm
}

func TestSimpleRun(t *testing.T) {
	simpleExecute(t, "1+1", ni(2))
	simpleExecute(t, "2.0+1", nf(3))
	simpleExecute(t, ".5+1", nf(1.5))
}

func TestStr(t *testing.T) {
	simpleExecute(t, `""`, ns(""))
	simpleExecute(t, `''`, ns(""))
	simpleExecute(t, "``", ns(""))
	simpleExecute(t, "\x1e\x1e", ns(""))

	simpleExecute(t, "'123'", ns("123"))
	simpleExecute(t, "'12' + '3' ", ns("123"))
	simpleExecute(t, "`12{3}` ", ns("123"))
	simpleExecute(t, "`12{'3'}` ", ns("123"))
	simpleExecute(t, "`12{% 3 %}` ", ns("123"))
	simpleExecute(t, `"123"`, ns("123"))
	simpleExecute(t, "\x1e"+"12{% 3 %}"+"\x1e", ns("123"))

	simpleExecute(t, `"12\n3"`, ns("12\n3"))
	simpleExecute(t, `"12\r3"`, ns("12\r3"))
	simpleExecute(t, `"12\f3"`, ns("12\f3"))
	simpleExecute(t, `"12\t3"`, ns("12\t3"))
	simpleExecute(t, `"12\\3"`, ns("12\\3"))

	simpleExecute(t, `'12\n3'`, ns("12\n3"))
	simpleExecute(t, `'12\r3'`, ns("12\r3"))
	simpleExecute(t, `'12\f3'`, ns("12\f3"))
	simpleExecute(t, `'12\t3'`, ns("12\t3"))
	simpleExecute(t, `'12\\3'`, ns("12\\3"))

	simpleExecute(t, "\x1e"+`12\n3`+"\x1e", ns("12\n3"))
	simpleExecute(t, "\x1e"+`12\r3`+"\x1e", ns("12\r3"))
	simpleExecute(t, "\x1e"+`12\f3`+"\x1e", ns("12\f3"))
	simpleExecute(t, "\x1e"+`12\t3`+"\x1e", ns("12\t3"))
	simpleExecute(t, "\x1e"+`12\\3`+"\x1e", ns("12\\3"))

	simpleExecute(t, "`"+`12\n3`+"`", ns("12\n3"))
	simpleExecute(t, "`"+`12\r3`+"`", ns("12\r3"))
	simpleExecute(t, "`"+`12\f3`+"`", ns("12\f3"))
	simpleExecute(t, "`"+`12\t3`+"`", ns("12\t3"))
	simpleExecute(t, "`"+`12\\3`+"`", ns("12\\3"))

	// TODO: FIX
	//simpleExecute(t, `"12\"3"`, ns(`12"3`))
	//simpleExecute(t, `"\""`, ns(`"`))
	//simpleExecute(t, `"\r"`, ns("\r"))
}

func TestEmptyInput(t *testing.T) {
	vm := NewVM()
	err := vm.Run("")
	if err == nil {
		t.Errorf("VM Error")
	}
}

func TestDice(t *testing.T) {
	// 语法可用性测试(并不做验算)
	simpleExecute(t, "4d1", ni(4))
	simpleExecute(t, "4D1", ni(4))

	simpleExecute(t, "4d1k", ni(1))
	simpleExecute(t, "4d1k1", ni(1))
	simpleExecute(t, "4d1kh", ni(1))
	simpleExecute(t, "4d1kh1", ni(1))

	simpleExecute(t, "4d1q", ni(1))
	simpleExecute(t, "4d1q1", ni(1))
	simpleExecute(t, "4d1kl(1)", ni(1))
	simpleExecute(t, "4d1kl1", ni(1))

	simpleExecute(t, "4d1dl", ni(3))
	simpleExecute(t, "4d1dl1", ni(3))

	simpleExecute(t, "4d1dl", ni(3))
	simpleExecute(t, "4d1dl1", ni(3))

	simpleExecute(t, "4d1dh", ni(3))
	simpleExecute(t, "4d1dh1", ni(3))

	// min max
	simpleExecute(t, "d20min20", ni(20))
	simpleExecute(t, "d20min30", ni(30)) // 与fvtt行为一致
	simpleExecute(t, "d20max1", ni(1))
	simpleExecute(t, "d20min30max1", ni(30)) // 同fvtt
	simpleExecute(t, "4d20k1min20", ni(20))

	// 优势
	simpleExecute(t, "d1优势", ni(1))
	simpleExecute(t, "d1劣势", ni(1))

	// 算力上限
	vm := NewVM()
	err := vm.Run("30001d20")
	if err == nil {
		t.Errorf("VM Error")
	}

	// 这种情况报个错如何？
	simpleExecute(t, "4d1k5", ni(4))
}

func TestDiceNoSpaceForModifier(t *testing.T) {
	vm := NewVM()
	err := vm.Run("3d1 k2")
	if assert.NoError(t, err) {
		// 注: 如果读取为3d1k2，值为2为错，读取3d1剩余文本k2为对
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}
}

func TestUnsupportedOperandType(t *testing.T) {
	vm := NewVM()
	err := vm.Run("2 % 3.1")
	if assert.Error(t, err) {
		// VM Error: 这两种类型无法使用 mod 算符连接: int64, float64
		assert.Equal(t, err.Error(), "这两种类型无法使用 mod 算符连接: int64, float64")
	}
}

func TestValueStore1(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a=1")
	if err != nil {
		// 未设置 ValueStoreNameFunc，无法储存变量
		t.Errorf("VM Error: %s", err.Error())
	}

	err = vm.Run("a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("bbb")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}
}

func TestValueStore(t *testing.T) {
	vm := NewVM()
	err := vm.Run("测试=1")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}

	vm = NewVM()
	err = vm.Run("测试   =   1")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}

	vm = NewVM()
	err = vm.Run("测试")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}

	vm = NewVM()
	err = vm.Run("CC")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}
	assert.True(t, valueEqual(vm.Ret, VMValueNewUndefined()))

	// 栈指针bug(两个变量实际都指向了栈的某一个位置，导致值相同)
	vm = NewVM()
	err = vm.Run("b=1;d=2")
	if err != nil {
		t.Errorf("VM Error: %s", err.Error())
	}
	assert.True(t, vmValueEqual(vm, "b", ni(1)))
	assert.True(t, vmValueEqual(vm, "d", ni(2)))
}

func TestIf(t *testing.T) {
	vm := NewVM()
	err := vm.Run("if 0 { a = 2 } else if 2 { b = 1 } c= 1; ;;;;; d= 2;b")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "b", ni(1)))
	assert.True(t, vmValueEqual(vm, "c", ni(1)))
	assert.True(t, vmValueEqual(vm, "d", ni(2)))

	_, exists := vm.attrs.Load("a")
	assert.True(t, !exists)
}

//

func TestStatementLines(t *testing.T) {
	vm := NewVM()
	err := vm.Run("i = 0 ;; i = 3")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))

	vm = NewVM()
	err = vm.Run("i = 0 ;    ;  ; i = 3")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))

	vm = NewVM()
	err = vm.Run("i = 0 if 1 { i = 3 }")
	assert.NoError(t, err)
	assert.Equal(t, "if 1 { i = 3 }", vm.RestInput)

	vm = NewVM()
	err = vm.Run("i = 0; if 1 { i = 3 }")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))

	vm = NewVM()
	err = vm.Run("i = 0   ;if 1 { i = 3 }")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))

	vm = NewVM()
	err = vm.Run("i = 0;  if 1 { i = 3 }")
	assert.NoError(t, err)
	assert.True(t, vmValueEqual(vm, "i", ni(3)))
}

func TestKeywords(t *testing.T) {
	vm := NewVM()
	err := vm.Run("while123")
	assert.NoError(t, err)
	assert.True(t, vm.RestInput == "")

	keywords := []string{
		"while", "if", "else", "continue", "break", "func",
	}

	suffixBad := []string{
		"", "=", "#", ";", "=1", " ", " =1", "!", "\"", "%", "^", "&", "*", "(", ")", "/", "+", "-", ".", ".aa",
		"[", "]", "[1]", ":", "<", ">", "?",
	}

	for _, i := range keywords {
		for _, j := range suffixBad {
			vm := NewVM()
			err = vm.Run(i + j)
			assert.Errorf(t, err, i+j)
		}
	}

	vm = NewVM()
	err = vm.Run("return 1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestWhile(t *testing.T) {
	vm := NewVM()
	err := vm.Run("i = 0; while i<5 { i=i+1 }")
	assert.NoError(t, err)
	assert.True(t, vm.NumOpCount < 100)

	vm = NewVM()
	err = vm.Run("i = 0; while 1 {  }")
	assert.Error(t, err) // 算力上限

	vm = NewVM()
	err = vm.Run("i = 0; while 1 {}")
	assert.Error(t, err) // 算力上限

	vm = NewVM()
	err = vm.Run("i = 0; while1 {}")
	assert.True(t, vm.RestInput == "{}", vm.RestInput)
}

func TestWhileContinueBreak(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = 0; while a < 5 { a = a+1; continue; a=a+10 }; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(5)))
	}

	vm = NewVM()
	err = vm.Run("a = 0; while a < 5 { a = a+1; a=a+10; continue }; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(11)))
	}

	vm = NewVM()
	err = vm.Run("a = 1; while a < 5 { break; a = a+1; a=a+10 }; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("a = 1; while a < 5 { a = a+1; break; a=a+10 }; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}
}

func TestLineBreak(t *testing.T) {
	vm := NewVM()
	err := vm.Run("if 1 {} 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("1; if 1 {} 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("1; if 1 {}; 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}
}

func TestItemSetBug(t *testing.T) {
	// 由于言诺在2022/9/9提交，此用例之前的输出内容为[3,3,3]
	vm := NewVM()
	err := vm.Run("a = [0,0,0]; i=0; while i<3 { a[i] = i+1; i=i+1 }  a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(1), ni(2), ni(3))))
	}
}

func TestCompareExpr(t *testing.T) {
	tests := []struct {
		expr  string
		value *VMValue
	}{
		{"1>0", ni(1)},
		{"1>=0", ni(1)},
		{"1==0", ni(0)},
		{"1==1", ni(1)},
		{"1<0", ni(0)},
		{"1<=0", ni(0)},
		{"1!=0", ni(1)},

		// 带空格
		{"1 > 0", ni(1)},

		// 中断
		{"5＝+2", ni(5)},
	}

	for _, i := range tests {
		vm := NewVM()
		err := vm.Run(i.expr)
		assert.NoError(t, err, i.expr)
		assert.True(t, valueEqual(vm.Ret, i.value), i.expr)
	}
}

func TestTernary(t *testing.T) {
	vm := NewVM()
	err := vm.Run("1 == 1 ? 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("1 == 1 ? 2 : 3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("1 != 1 ? 2 : 3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	vm.attrs.Store("a", ni(1))
	err = vm.Run("a == 1 ? 'A', a == 2 ? 'B'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("A")))
	}

	vm = NewVM()
	vm.attrs.Store("a", ni(2))
	err = vm.Run("a == 1 ? 'A', a == 2 ? 'B', a == 3 ? 'C'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("B")))
	}

	vm = NewVM()
	vm.attrs.Store("a", ni(3))
	err = vm.Run("a == 1 ? 'A', a == 2 ? 'B', a == 3 ? 'C'")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("C")))
	}
}

func TestUnary(t *testing.T) {
	vm := NewVM()
	err := vm.Run("-1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(-1)))
	}

	vm = NewVM()
	err = vm.Run("--1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("-+1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(-1)))
	}

	vm = NewVM()
	err = vm.Run("+-1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(-1)))
	}

	vm = NewVM()
	err = vm.Run("-1.3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(-1.3)))
	}

	vm = NewVM()
	err = vm.Run("-'123'")
	assert.Error(t, err)
}

func TestRest(t *testing.T) {
	vm := NewVM()
	err := vm.Run("1 2")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
		assert.True(t, vm.RestInput == "2")
	}
}

func TestRecursion(t *testing.T) {
	vm := NewVM()
	err := vm.Run("&a = a + 1")
	assert.NoError(t, err)

	err = vm.Run("a")
	assert.Error(t, err) // 算力上限
}

func TestArray(t *testing.T) {
	vm := NewVM()
	err := vm.Run("[1,2,3]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(1), ni(2), ni(3))))
	}

	vm = NewVM()
	err = vm.Run("[1,3,2]kh")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("[1.2,2,3]kh")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,2.2,3]kh")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,3,2]kl")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("[2,3,1]kl")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("[1,3.1,2.1]kl")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(1)))
	}

	vm = NewVM()
	err = vm.Run("[4.1,3.1,1]kl")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, nf(1)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3][1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3][-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3][-4]")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("[1,2,3][4]")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("a = [1,2,3]; a[1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("b[1]")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("b[0][0]")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("[[1]][0][0]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("([[2]])[0][0]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("a = [0,0,0]; a[0] = 1; a[0]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("a[0] = 1")
	assert.Error(t, err)
}

func TestArrayMethod(t *testing.T) {
	vm := NewVM()
	err := vm.Run("[1,2,3].sum()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(6)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3].len()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3].pop()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("[1,2,3].shift()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("a = [1,2,3]; a.push(4); a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, na(ni(1), ni(2), ni(3), ni(4))))
	}
}

func TestReturn(t *testing.T) {
	vm := NewVM()
	err := vm.Run("func test(n) { return 1; 2 }; test(11)")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestComputed(t *testing.T) {
	vm := NewVM()
	err := vm.Run("&a = d1+2; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("&a = []+2; a")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("&a = undefined; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewUndefined()))
	}
}

func TestComputed2(t *testing.T) {
	vm := NewVM()
	err := vm.Run("&a = d1 + this.x")
	assert.NoError(t, err)

	err = vm.Run("&a.x = 2")
	assert.NoError(t, err)

	err = vm.Run("a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	//vm = NewVM()
	err = vm.Run("&a.x")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}
}

func TestFunction(t *testing.T) {
	vm := NewVM()
	err := vm.Run("func a() { 123 }; a()")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(123)))
	}

	vm = NewVM()
	err = vm.Run("func a(d,b,c) { return this.b }; a(1,2,3)")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}

	vm = NewVM()
	err = vm.Run("func a(d,b,c) { this.b }; a(1,2)")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("func a(d,b,c) { this.b }; a(1,2,3,4,5)")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("func a() { 2 / 0 }; a()")
	assert.Error(t, err)
}

func TestFunctionRecursion(t *testing.T) {
	vm := NewVM()
	err := vm.Run(`
func foo(n) {
	if (n < 2) {
		return foo(n + 1)
	}
	return 123
}
foo(1)
`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(123)))
	}
}

func TestFunctionFib(t *testing.T) {
	vm := NewVM()
	err := vm.Run(`func fib(n) {
  this.n == 0 ? 0,
  this.n == 1 ? 1,
  this.n == 2 ? 1,
   1 ? fib(this.n-1)+fib(this.n-2)
}
fib(11)
`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(89)))
	}
}

func TestBytecodeToString(t *testing.T) {
	ops := []ByteCode{
		{TypePushIntNumber, int64(1)},
		{TypePushFloatNumber, float64(1.2)},
		{TypePushString, "abc"},

		{TypeAdd, nil},
		{TypeSubtract, nil},
		{TypeMultiply, nil},
		{TypeDivide, nil},
		{TypeModulus, nil},
		{TypeExponentiation, nil},
		{TypeNullCoalescing, nil},

		{TypeCompLT, nil},
		{TypeCompLE, nil},
		{TypeCompEQ, nil},
		{TypeCompNE, nil},
		{TypeCompGE, nil},
		{TypeCompGT, nil},

		{TypeLogicAnd, nil},
		{TypeLogicOr, nil},

		{TypeNop, nil},

		{TypeBitwiseAnd, nil},
		{TypeBitwiseOr, nil},

		{TypeDiceInit, nil},
		{TypeDiceSetTimes, nil},
		{TypeDiceSetKeepLowNum, nil},
		{TypeDiceSetKeepHighNum, nil},
		{TypeDiceSetDropLowNum, nil},
		{TypeDiceSetDropHighNum, nil},
		{TypeDiceSetMin, nil},
		{TypeDiceSetMax, nil},

		{TypeJmp, int64(0)},
		{TypeJe, int64(0)},
		{TypeJne, int64(0)},
	}

	for _, i := range ops {
		if i.CodeString() == "" {
			t.Errorf("Not work: %d", i.T)
		}
	}
}

func TestWriteCodeOverflow(t *testing.T) {
	vm := NewVM()
	vm.Run("")
	for i := 0; i < 8193; i++ {
		vm.parser.WriteCode(TypeNop, nil)
	}
	if !vm.parser.checkStackOverflow() {
		t.Errorf("Failed")
	}
}

func TestGetASM(t *testing.T) {
	vm := NewVM()
	vm.Run("1+1")
	vm.GetAsmText()
}

func TestSliceGet(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = [1,2,3,4]")
	assert.NoError(t, err)

	err = vm.Run("a[:]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(1), ni(2), ni(3), ni(4))))
	}

	err = vm.Run("a[:2]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(1), ni(2))))
	}

	err = vm.Run("a[0:-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(1), ni(2), ni(3))))
	}

	err = vm.Run("a[-3:-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(2), ni(3))))
	}

	err = vm.Run("a[-3:-1:1]")
	assert.Error(t, err)
	// 尚不支持分片步长
	//if assert.NoError(t, err) {
	//	assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(2), ni(3))))
	//}

	err = vm.Run("a[-3:-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(2), ni(3))))
	}

	err = vm.Run("b = a[-3:-1]; b[0] = 9; a[-3:-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(2), ni(3))))
	}

	vm = NewVM()
	err = vm.Run("'12345'[2:3]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("3")))
	}
}

func TestSliceSet(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = [1,2,3,4]")
	assert.NoError(t, err)

	err = vm.Run("a[:] = [1,2,3]; a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(1), ni(2), ni(3))))
	}

	err = vm.Run("a = [1,2,3]; a[:1] = [4,5];a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(4), ni(5), ni(2), ni(3))))
	}

	err = vm.Run("a = [1,2,3]; a[2:] = [4,5];a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(1), ni(2), ni(4), ni(5))))
	}
}

func TestRange(t *testing.T) {
	vm := NewVM()
	err := vm.Run("[1..4]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(1), ni(2), ni(3), ni(4))))
	}

	vm = NewVM()
	err = vm.Run("[4..1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, VMValueNewArray(ni(4), ni(3), ni(2), ni(1))))
	}
}

func TestDictExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = {'a': 1}")
	if assert.NoError(t, err) {
	}

	err = vm.Run("a.a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	err = vm.Run("a['a']")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestDictExpr2(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = {'a': 1,}")
	if assert.NoError(t, err) {
	}
	err = vm.Run("a.a")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("c = 'c'; a = {c:1,'b':3}")
	if assert.NoError(t, err) {
	}
	err = vm.Run("a.c")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestIdExpr(t *testing.T) {
	vm := NewVM()
	vm.attrs.Store("a:b", ni(3))
	err := vm.Run("a:b") // 如果读到a 余下a:b即为错误
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}
}

func TestStringExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("\x1e xxx \x1e")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns(" xxx ")))
	}
}

func TestContinuousDiceExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("10d1d1")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(10)))
	}
}

func TestCrash1(t *testing.T) {
	// 一种崩溃，崩溃条件是第二次调用vm.Run且第二次的tokens少于第一次的
	vm := NewVM()
	err := vm.Run("aa + 2//asd")
	if assert.Error(t, err) {
		err := vm.Run("/")
		if assert.Error(t, err) {
			assert.True(t, strings.Index(err.Error(), "parse error near") != -1)
		}
	}
}

func TestDiceWodExpr(t *testing.T) {
	vm := NewVM()
	vm.Flags.EnableDiceWoD = true
	err := vm.Run("8a11m10k1")
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
		assert.True(t, valueEqual(vm.Ret, ni(8)))
	}

	vm = NewVM()
	vm.Flags.EnableDiceWoD = true
	err = vm.Run("20001a11m10k1")
	assert.Error(t, err)

	vm = NewVM()
	vm.Flags.EnableDiceWoD = true
	err = vm.Run("8a1m10k1")
	assert.Error(t, err)

	vm = NewVM()
	vm.Flags.EnableDiceWoD = true
	err = vm.Run("8a11m0k1")
	assert.Error(t, err)

	vm = NewVM()
	vm.Flags.EnableDiceWoD = true
	err = vm.Run("8a11m10k0")
	assert.Error(t, err)
}

func TestDiceDoubleCrossExpr(t *testing.T) {
	// 没有很好的测试用例
	vm := NewVM()
	vm.Flags.EnableDiceDoubleCross = true
	err := vm.Run("10c11m10")
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
		assert.True(t, vm.Ret.MustReadInt() <= 10)
	}

	vm = NewVM()
	vm.Flags.EnableDiceDoubleCross = true
	err = vm.Run("20001c11m10")
	assert.Error(t, err)

	vm = NewVM()
	vm.Flags.EnableDiceDoubleCross = true
	err = vm.Run("10c1m10")
	assert.Error(t, err)

	vm = NewVM()
	vm.Flags.EnableDiceDoubleCross = true
	err = vm.Run("10c11m0")
	assert.Error(t, err)
}

func TestDiceFlagWodMacroExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("// #EnableDiceWoD true")
	if assert.NoError(t, err) {
		err := vm.Run("10a11")
		if assert.NoError(t, err) {
			assert.Equal(t, "", vm.RestInput)

			err := vm.Run("// #EnableDiceWoD false")
			if assert.NoError(t, err) {
				err := vm.Run("10a11")
				if assert.NoError(t, err) {
					assert.Equal(t, "a11", vm.RestInput)
				}
			}
		}
	}
}

func TestDiceFlagCoCMacroExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("// #EnableDiceCoC true")
	if assert.NoError(t, err) {
		err := vm.Run("b2")
		if assert.NoError(t, err) {
			assert.Equal(t, "", vm.RestInput)
			assert.Equal(t, VMTypeInt, vm.Ret.TypeId)

			err := vm.Run("// #EnableDiceCoC false")
			if assert.NoError(t, err) {
				err := vm.Run("b2")
				if assert.NoError(t, err) {
					assert.Equal(t, VMTypeUndefined, vm.Ret.TypeId)
				}
			}
		}
	}
}

func TestDiceFlagFateMacroExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("// #EnableDiceFate true")
	if assert.NoError(t, err) {
		err := vm.Run("f")
		if assert.NoError(t, err) {
			assert.Equal(t, "", vm.RestInput)
			assert.Equal(t, VMTypeInt, vm.Ret.TypeId)

			err := vm.Run("// #EnableDiceFate false")
			if assert.NoError(t, err) {
				err := vm.Run("f")
				if assert.NoError(t, err) {
					assert.Equal(t, VMTypeUndefined, vm.Ret.TypeId)
				}
			}
		}
	}
}
func TestDiceFlagDoubleCrossMacroExpr(t *testing.T) {
	vm := NewVM()
	err := vm.Run("// #EnableDiceDoubleCross true")
	if assert.NoError(t, err) {
		err := vm.Run("2c5")
		if assert.NoError(t, err) {
			assert.Equal(t, "", vm.RestInput)
			assert.Equal(t, VMTypeInt, vm.Ret.TypeId)

			err := vm.Run("// #EnableDiceDoubleCross false")
			if assert.NoError(t, err) {
				err := vm.Run("2c5")
				if assert.NoError(t, err) {
					assert.Equal(t, "c5", vm.RestInput)
				}
			}
		}
	}
}

func TestComment(t *testing.T) {
	vm := NewVM()
	err := vm.Run("// test\na = 1;\na")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestDiceAndSpaceBug(t *testing.T) {
	// 一个错误的代码逻辑: 部分算符后需要跟sp1，导致f +1可以工作，但f+1不行
	// 但也不能让 f1 被解析为f，剩余文本1
	vm := NewVM()
	vm.Flags.EnableDiceFate = true
	err := vm.Run("f +1")
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeInt, vm.Ret.TypeId)
	}

	err = vm.Run("f+1")
	if assert.NoError(t, err) {
		assert.Equal(t, VMTypeInt, vm.Ret.TypeId)
	}

	err = vm.Run("f1")
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
		assert.Equal(t, VMTypeUndefined, vm.Ret.TypeId)
	}
}

func TestDiceAndSpaceBug2(t *testing.T) {
	// 其他版本
	tests := [][]string{
		{"b +1", "b+1", "bX"},
		{"p +1", "p+1", "pX"},
		{"a10 +1", "a10+1", "a10x"},
		{"1c5 +1", "1c5+1", "x"},
	}

	for _, i := range tests {
		e1, e2, e3 := i[0], i[1], i[2]
		vm := NewVM()
		vm.Flags.EnableDiceCoC = true
		vm.Flags.EnableDiceWoD = true
		vm.Flags.EnableDiceDoubleCross = true
		err := vm.Run(e1)
		if assert.NoError(t, err) {
			assert.Equal(t, VMTypeInt, vm.Ret.TypeId)
		}

		err = vm.Run(e2)
		if assert.NoError(t, err) {
			assert.Equal(t, VMTypeInt, vm.Ret.TypeId)
		}

		err = vm.Run(e3)
		if assert.NoError(t, err) {
			assert.Equal(t, "", vm.RestInput)
			assert.Equal(t, VMTypeUndefined, vm.Ret.TypeId)
		}
	}

	vm := NewVM()
	vm.Flags.EnableDiceDoubleCross = true
	err := vm.Run("1c5d")
	if assert.NoError(t, err) {
		assert.Equal(t, "d", vm.RestInput)
	}

	vm = NewVM()
	vm.Flags.EnableDiceWoD = true
	err = vm.Run("2a10x")
	if assert.NoError(t, err) {
		assert.Equal(t, "x", vm.RestInput)
	}
}

func TestBitwisePrecedence(t *testing.T) {
	vm := NewVM()
	err := vm.Run("1|2&4")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	vm = NewVM()
	err = vm.Run("(1|2)&4")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(0)))
	}
}

func TestLogicOp(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = [1,2]; 5 || a.push(3); a ")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, na(ni(1), ni(2))))
	}

	vm = NewVM()
	err = vm.Run("a = [1,2]; 5 && a.push(3); a ")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, na(ni(1), ni(2), ni(3))))
	}
}

func TestFuncAbs(t *testing.T) {
	vm := NewVM()
	err := vm.Run("abs(-1)")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}
}

func TestLogicAnd(t *testing.T) {
	vm := NewVM()
	err := vm.Run("1 && 2 && 3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(3)))
	}

	vm = NewVM()
	err = vm.Run("1 && 0 && 3")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(0)))
	}
}

func TestStackOverFlow(t *testing.T) {
	vm := NewVM()
	err := vm.Run("while 1 { 2 }")
	assert.Error(t, err)
}

func TestSliceUnicode(t *testing.T) {
	vm := NewVM()
	err := vm.Run("'中文测试'[1:3]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("文测")))
	}

	err = vm.Run("'中文测试'[-3:3]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("文测")))
	}
}

func TestDiceExprError(t *testing.T) {
	vm := NewVM()
	err := vm.Run("(-1)d5")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("('xxx')d5")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("3d(-10)")
	assert.Error(t, err)

	vm = NewVM()
	err = vm.Run("3d('xx')")
	assert.Error(t, err)
}

func TestDiceDH_DL(t *testing.T) {
	reResult := regexp.MustCompile(`\{(\d+) (\d+) \| (\d+)}`)

	vm := NewVM()
	for {
		err := vm.Run("3d1000dh1")
		if assert.NoError(t, err) {
			m := reResult.FindStringSubmatch(vm.Detail)
			a1, _ := strconv.ParseInt(m[1], 10, 64)
			a2, _ := strconv.ParseInt(m[2], 10, 64)
			a3, _ := strconv.ParseInt(m[3], 10, 64)
			if a1 != a2 && a2 != a3 {
				// 三个输出数字不等，符合测试条件
				assert.True(t, a3 > a2 && a3 > a1)
				break
			}
		}
	}

	vm = NewVM()
	for {
		err := vm.Run("3d1000dl1")
		if assert.NoError(t, err) {
			m := reResult.FindStringSubmatch(vm.Detail)
			a1, _ := strconv.ParseInt(m[1], 10, 64)
			a2, _ := strconv.ParseInt(m[2], 10, 64)
			a3, _ := strconv.ParseInt(m[3], 10, 64)
			if a1 != a2 && a2 != a3 {
				// 三个输出数字不等，符合测试条件
				assert.True(t, a3 < a2 && a3 < a1)
				break
			}
		}
	}
}

func TestDiceExprIndexBug(t *testing.T) {
	// 12.1 于言诺发现，如 2d(3d1) 会被错误计算为 9[2d(3d1)=9=3+3+3,3d1=3]
	// 经查原因为Dice字节指令执行时，并未将骰子栈正确出栈
	reResult := regexp.MustCompile(`2d\(3d1\)=(\d+)=(\d+)\+(\d+),`)

	vm := NewVM()
	err := vm.Run("2d(3d1)")

	if assert.NoError(t, err) {
		assert.True(t, reResult.MatchString(vm.Detail))
	}
}

func TestStringGetItem(t *testing.T) {
	vm := NewVM()
	err := vm.Run("a = '测试'; a[1]")

	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("试")))
	}

	err = vm.Run("a = '测试'; a[-1]")
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ns("试")))
	}
}

func TestDiceExprKlBug(t *testing.T) {
	// 12.6 云陌发现，2d5kld4时有概率中间过程这样子：1[2d5kld4=1={1 | 2 3 4},2d5kl=4]
	// 原因也是骰子栈未正确出栈
	// 这个有一定运气成分(虽然很小)，所以跑5次

	for i := 0; i < 5; i++ {
		vm := NewVM()
		err := vm.Run("(1d1000kl)d1")

		if assert.NoError(t, err) {
			assert.False(t, strings.Contains(vm.Detail, "|"))
		}
	}
}

func TestIfElseExprBug1(t *testing.T) {
	// 12.6 于言诺 else后面必须跟一个空格
	vm := NewVM()
	err := vm.Run("if true {} else{}")

	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
	}

	vm = NewVM()
	err = vm.Run("if true {} elseif 1{}")

	if assert.NoError(t, err) {
		// 注: 这里 elseif 会被当做变量 所以这里读到是undefined
		assert.Equal(t, "1{}", vm.RestInput)
	}
}

func TestBlockExprBug(t *testing.T) {
	// 12.7 木落
	vm := NewVM()
	err := vm.Run("if 1 {} 1 2 3 4 5")

	if assert.NoError(t, err) {
		assert.Equal(t, "2 3 4 5", vm.RestInput)
	}
}

func TestWhileExprBug(t *testing.T) {
	// 12.7 云陌
	// 故障原因是第二次解析while时，第一次的没有出栈，因此又被处理了一遍，这个会引起程序崩溃
	vm := NewVM()
	err := vm.Run(`i = 1; while i < 2 {continue}`)
	assert.Error(t, err) // 算力超出

	err = vm.Run(`while i < 2 {i=i+1}`)
	if assert.NoError(t, err) {
		assert.Equal(t, "", vm.RestInput)
	}
}

func TestNameDetailBug(t *testing.T) {
	// "a = 1;a   " 时，过程为 "a = 1;1[a    =1]"，不应有空格
	vm := NewVM()
	err := vm.Run(`a = 1;a   `)
	if assert.NoError(t, err) {
		// TODO: 后面的空格
		assert.Equal(t, "a = 1;1[a=1]   ", vm.Detail)
	}
}

func TestLogicOrBug(t *testing.T) {
	// (0||0)+1 报错，原因是生成的代码里最后有一个jmp 1，跳过了1的push，导致栈里只有一个值
	vm := NewVM()
	err := vm.Run(`(0||0)+1`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(1)))
	}

	err = vm.Run(`(0||1)+1`)
	if assert.NoError(t, err) {
		assert.True(t, valueEqual(vm.Ret, ni(2)))
	}
}
