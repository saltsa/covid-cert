package main

import (
	"syscall/js"

	log "github.com/sirupsen/logrus"
)

func verifyJSData(this js.Value, args []js.Value) interface{} {
	log.Info("verify JS data")

	if len(args) < 1 {
		log.Info("args len is zero")
		return false
	}
	res := doValidation([]byte(args[0].String()))
	log.Infof("output: %v", res)
	return res
}

func registerCallbacks() {
	js.Global().Set("goVerify", js.FuncOf(verifyJSData))
}

func main() {
	c := make(chan struct{}, 0)

	log.Info("WASM Go Initialized")
	// register functions
	registerCallbacks()
	<-c
}
