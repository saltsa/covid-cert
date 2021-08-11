// +build js,wasm

package cvcert

import (
	"syscall/js"

	log "github.com/sirupsen/logrus"
)

// verifyJSData returns JS object which contains information about the
// certificate
func verifyJSData(this js.Value, args []js.Value) interface{} {
	log.Info("verify JS data")

	if len(args) < 1 {
		log.Info("args len is zero")
		return false
	}
	res := DoValidation([]byte(args[0].String()))
	log.Infof("output: %v", res)
	if res == nil {
		return map[string]interface{}{
			"error": "certificate validation vailed, see console log",
		}
	}
	return res
}

func RegisterCallbacks() bool {
	js.Global().Set("goVerify", js.FuncOf(verifyJSData))
	return true
}
