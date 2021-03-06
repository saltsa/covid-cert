package main

import (
	cvcert "github.com/saltsa/covid-cert"
	log "github.com/sirupsen/logrus"
)

func main() {

	keys, err := cvcert.GetFileKeys()
	if err != nil {
		log.Println(err)
	}

	err = cvcert.ParseKeys(keys)
	if err != nil {
		log.Println(err)
	}

	callbacksRegistered := cvcert.RegisterCallbacks()

	data, err := cvcert.ReadData()
	if err != nil {
		if callbacksRegistered {
			log.Infof("reading data failed (%s), but JS callbacks registered. Entering event loop.", err)
			mainLoop()
		}
		log.Fatalf("reading data failed: %s", err)
	}
	ret := cvcert.DoValidation(data)

	for key, val := range ret {
		log.Printf("%s: %s\n", key, val)
	}
}

func init() {
	// log.SetLevel(log.DebugLevel)
}

// mainLoop just waits for channel close (which never happens). Actions are triggered
// from Javascript (js.go).
func mainLoop() {
	c := make(chan struct{})

	log.Info("WASM Go Initialized")
	// registerCallbacks()
	<-c
}
