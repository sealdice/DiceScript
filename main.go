//go:build js
// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/sealdice/dicescript"
)

var scope = map[string]*dicescript.VMValue{}

func newVM(name string) *js.Object {
	player := dicescript.VMValueNewDict(nil)
	scope["player"] = player.V()

	vm := dicescript.NewVM()
	vm.ValueStoreFunc = func(name string, v *dicescript.VMValue) {
		scope[name] = v
	}
	vm.ValueLoadFunc = func(name string) *dicescript.VMValue {
		if val, ok := scope[name]; ok {
			return val
		}
		return nil
	}

	return js.MakeFullWrapper(vm)
}

func main() {
	newDict := func() *dicescript.VMDictValue {
		return dicescript.VMValueNewDict(nil)
	}

	newValueMap := func() *dicescript.ValueMap {
		return &dicescript.ValueMap{}
	}

	js.Global.Set("dice", map[string]interface{}{
		"newVM":        newVM,
		"newValueMap":  newValueMap,
		"vmNewInt64":   js.MakeWrapper(dicescript.VMValueNewInt),
		"vmNewFloat64": js.MakeWrapper(dicescript.VMValueNewFloat),
		"vmNewStr":     js.MakeWrapper(dicescript.VMValueNewStr),
		"vmNewDict":    js.MakeWrapper(newDict),
		"help":         js.MakeWrapper("此项目的js绑定: https://github.com/sealdice/dicescript"),
	})

	//js.Module.Set("newVM", dicescript.NewVM)
	//js.Module.Set("Context", dicescript.Context{})
}
