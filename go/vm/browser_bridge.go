package vm

import (
	"github.com/123654lkj/zero/go/browser"
	"github.com/123654lkj/zero/go/value"
	"github.com/go-rod/rod/lib/proto"
)

var browserManager = browser.NewManager()

func (vm *VM) registerBrowserBuiltins() {
	vm.registerBuiltin("web_launch", builtinWebLaunch)
	vm.registerBuiltin("web_open", builtinWebOpen)
	vm.registerBuiltin("web_text", builtinWebText)
	vm.registerBuiltin("web_close", builtinWebClose)
}

func builtinWebLaunch(args []value.Value) value.Value {
	if len(args) != 1 || !args[0].IsMap() {
		panic("web_launch needs a config map")
	}
	m := args[0].AsMap()
	cfg := browser.DefaultConfig()
	if v, ok := m["headless"]; ok && v.IsBool() {
		cfg.Headless = v.AsBool()
	}
	if v, ok := m["stealth"]; ok && v.IsBool() {
		cfg.Stealth = v.AsBool()
	}
	b, err := browserManager.Launch(cfg)
	if err != nil {
		panic("web_launch: " + err.Error())
	}
	return value.StringValue(b.ID)
}

func builtinWebOpen(args []value.Value) value.Value {
	if len(args) != 2 || !args[0].IsString() || !args[1].IsString() {
		panic("web_open: need browser_id and url")
	}
	ref := browserManager.Find(args[0].AsString())
	if ref == nil {
		panic("web_open: browser not found")
	}
	page, err := ref.Browser.Page(proto.TargetCreateTarget{URL: args[1].AsString()})
	if err != nil {
		panic("web_open: " + err.Error())
	}
	_ = page
	return value.NilValue()
}

func builtinWebText(args []value.Value) value.Value {
	panic("web_text: not yet implemented")
}

func builtinWebClose(args []value.Value) value.Value {
	browserManager.CloseAll()
	return value.NilValue()
}