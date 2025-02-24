package main

import (
	"quickstart/route"

	"github.com/mingzaily/maltose/frame/m"
)

func main() {
	// adapter, err := mcfg.NewAdapterFile()
	// if err != nil {
	// 	panic(err)
	// }
	// adapter.SetFileName("dev")
	// m.Config().SetAdapter(adapter)

	s := m.Server()
	route.Build(s)
	s.Run()
}
