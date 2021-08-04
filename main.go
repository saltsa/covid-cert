package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {

	callbacksRegistered := registerCallbacks()

	data, err := readData()
	if err != nil {
		if callbacksRegistered {
			log.Infof("reading data failed (%s), but JS callbacks registered. Entering event loop.", err)
			mainLoop()
		}
		log.Fatalf("reading data failed: %s", err)
	}
	ret := doValidation(data)

	for key, val := range ret {
		log.Printf("%s: %s\n", key, val)
	}
}

// mainLoop just waits for channel close (which never happens). Actions are triggered
// from Javascript (js.go).
func mainLoop() {
	c := make(chan struct{})

	log.Info("WASM Go Initialized")
	// registerCallbacks()
	<-c
}
