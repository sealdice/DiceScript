//go:build js
// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	dice "github.com/sealdice/dicescript"
	"regexp"
	"strconv"
)

var scope = map[string]*dice.VMValue{}

func newVM(name string) *js.Object {
	player := dice.VMValueNewDict(nil)
	player.Store("力量", dice.VMValueNewInt(50))
	player.Store("敏捷", dice.VMValueNewInt(60))
	player.Store("智力", dice.VMValueNewInt(70))
	scope["player"] = player.V()

	vm := dice.NewVM()
	//vm.ValueStoreFunc = func(name string, v *dice.VMValue) {
	//	scope[name] = v
	//}

	re := regexp.MustCompile(`^_(\D+)(\d+)$`)
	vm.ValueLoadFunc = func(name string) *dice.VMValue {
		m := re.FindStringSubmatch(name)
		if len(m) > 1 {
			val, _ := strconv.ParseInt(m[2], 10, 64)
			return dice.VMValueNewInt(val)
		}

		if v, exists := player.Load(name); exists {
			return v
		}

		if val, ok := scope[name]; ok {
			return val
		}
		return nil
	}

	return js.MakeFullWrapper(vm)
}

func main() {
	newDict := func() *dice.VMDictValue {
		return dice.VMValueNewDict(nil)
	}

	newValueMap := func() *dice.ValueMap {
		return &dice.ValueMap{}
	}

	js.Global.Set("dice", map[string]interface{}{
		"newVM":        newVM,
		"newValueMap":  newValueMap,
		"vmNewInt64":   js.MakeWrapper(dice.VMValueNewInt),
		"vmNewFloat64": js.MakeWrapper(dice.VMValueNewFloat),
		"vmNewStr":     js.MakeWrapper(dice.VMValueNewStr),
		//"vmNewArray":    js.MakeWrapper(newArray),
		"vmNewDict": js.MakeWrapper(newDict),
		"help":      js.MakeWrapper("此项目的js绑定: https://github.com/sealdice/dice"),
	})

	//js.Module.Set("newVM", dice.NewVM)
	//js.Module.Set("Context", dice.Context{})
}
