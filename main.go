//go:build js
// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"regexp"
	"strconv"

	ds "github.com/sealdice/dicescript"
)

var scope = map[string]*ds.VMValue{}

func newVM(name string) *js.Object {
	player := ds.VMValueNewDict(nil)
	player.Store("力量", ds.VMValueNewInt(50))
	player.Store("敏捷", ds.VMValueNewInt(60))
	player.Store("智力", ds.VMValueNewInt(70))
	scope["player"] = player.V()

	vm := ds.NewVM()
	//vm.GlobalValueStoreFunc = func(name string, v *ds.VMValue) {
	//	scope[name] = v
	//}

	re := regexp.MustCompile(`^_(\D+)(\d+)$`)
	vm.GlobalValueLoadFunc = func(name string) *ds.VMValue {
		m := re.FindStringSubmatch(name)
		if len(m) > 1 {
			val, _ := strconv.ParseInt(m[2], 10, 64)
			return ds.VMValueNewInt(val)
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
	diceModule := map[string]interface{}{
		"newVMForPlaygournd": newVM,
		"newVM": func() *js.Object {
			vm := ds.NewVM()
			return js.MakeFullWrapper(vm)
		},
		"newConfig": func() *js.Object {
			return js.MakeFullWrapper(&ds.RollConfig{})
		},
		"newValueMap": func() *js.Object {
			return js.MakeFullWrapper(&ds.ValueMap{})
		},
		"vmNewInt": func(i int64) *js.Object {
			return js.MakeFullWrapper(ds.VMValueNewInt(i))
		},
		"vmNewFloat": func(i float64) *js.Object {
			return js.MakeFullWrapper(ds.VMValueNewFloat(i))
		},
		"vmNewStr": func(s string) *js.Object {
			return js.MakeFullWrapper(ds.VMValueNewStr(s))
		},
		//"vmNewArray":    js.MakeWrapper(newArray),
		"vmNewDict": func() *js.Object {
			return js.MakeFullWrapper(ds.VMValueNewDict(nil))
		},
		"help": "此项目的js绑定: https://github.com/sealdice/dice",
	}

	js.Module.Get("exports").Set("ds", diceModule)
}
