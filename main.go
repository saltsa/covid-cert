package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {

	callbacksRegistered := registerCallbacks()

	data, err := readData()
	if err != nil {
		if callbacksRegistered {
			mainLoop()
		}
		log.Fatalln(data)
	}
	ret := doValidation(data)

	for key, val := range ret {
		log.Printf("%s: %s\n", key, val)
	}
}

func mainLoop() {
	c := make(chan struct{}, 0)

	log.Info("WASM Go Initialized")
	// registerCallbacks()
	<-c
}
