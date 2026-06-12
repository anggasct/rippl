package alpha

import "example.com/minimodule/pkg/beta"

var x beta.Type

func Run() {
	beta.Foo()
	var t beta.Type
	t.Method()
	self()
}

func self() {
	self()
}
