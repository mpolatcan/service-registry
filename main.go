package main

import (
	"./core"
)

func main() {
	var registry = &core.Registry{}

	registry.StartRegistry()
}