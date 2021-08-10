package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const pubkeySource = "https://verifier-api.coronacheck.nl/v4/verifier/public_keys"

//go:embed list_of_keys.json
var embedFS embed.FS

var publicKeys = make(map[string]string)

type pubkeyResponse struct {
	Payload   []byte `json:"payload"`
	Signature []byte `json:"signature"`
}

type pubkeyPayload struct {
	CLKeys []interface{}              `json:"cl_keys"`
	EUKeys map[string][]pubkeyContent `json:"eu_keys"`
}

type pubkeyContent struct {
	Subject  string   `json:"subjectPk"`
	KeyUsage []string `json:"keyUsage"`
}

// getHTTPKeys used to download keys from the source. Not used at the moment.
func getHTTPKeys() (io.Reader, error) {
	client := http.DefaultClient

	ret, err := client.Get(pubkeySource)
	if err != nil {
		return nil, err
	}

	defer ret.Body.Close()

	data, err := io.ReadAll(ret.Body)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewReader(data)
	return buf, nil
}

// getFileKeys reads embedded list of keys
func getFileKeys() (io.Reader, error) {

	f, err := embedFS.Open("list_of_keys.json")
	if err != nil {
		return nil, err
	}

	fd, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewReader(fd)

	return buf, nil
}

func parseKeys(body io.Reader) error {

	// first parse the http response
	dec := json.NewDecoder(body)
	var data pubkeyResponse
	err := dec.Decode(&data)
	if err != nil {
		return err
	}

	// then the payload (base64 encoded) inside it
	dec = json.NewDecoder(bytes.NewReader(data.Payload))
	var dataPayload pubkeyPayload
	err = dec.Decode(&dataPayload)
	if err != nil {
		return err
	}

	for key, value := range dataPayload.EUKeys {
		log.Debugf("kid: %s number of keys: %d", key, len(value))
		if len(value) < 1 {
			continue
		}
		publicKeys[key] = value[0].Subject
	}

	return nil
}
