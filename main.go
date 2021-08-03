// +build !js

package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	data, err := readData()
	if err != nil {
		log.Fatalln(data)
	}
	ret := doValidation(data)

	for key, val := range ret {
		log.Printf("%s: %s\n", key, val)
	}
}
