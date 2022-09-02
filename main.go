//go:build js
// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/sealdice/dicescript"
)

func newVM(name string) *js.Object {
	attrs := map[string]*dicescript.VMValue{}
	vm := dicescript.NewVM()
	vm.ValueStoreNameFunc = func(name string, v *dicescript.VMValue) {
		attrs[name] = v
	}
	vm.ValueLoadNameFunc = func(name string) *dicescript.VMValue {
		if val, ok := attrs[name]; ok {
			return val
		}
		return nil
	}

	return js.MakeFullWrapper(vm)
}

func main() {
	js.Global.Set("dice", map[string]interface{}{
		"newVM":        newVM,
		"vmNewInt64":   js.MakeWrapper(dicescript.VMValueNewInt64),
		"vmNewFloat64": js.MakeWrapper(dicescript.VMValueNewFloat64),
		"vmNewStr":     js.MakeWrapper(dicescript.VMValueNewStr),
		"help":         js.MakeWrapper("此项目的js绑定: https://github.com/sealdice/dicescript"),
	})

	//js.Module.Set("newVM", dicescript.NewVM)
	//js.Module.Set("Context", dicescript.Context{})
}
