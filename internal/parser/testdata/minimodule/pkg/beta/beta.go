package beta

import "example.com/minimodule/pkg/gamma"

type Type struct{}

func Foo() {
	helper()
	gamma.Bar()
}

func (Type) Method() {}
