# EU certificate verification in browser

Application uses browser Camera or Barcode API to take a picture. This version only supports
Certificates signed by "The Social Insurance Institution of Finland" public key.

## Installation

As the Javascript uses Camera API, it needs TLS. Generate certs:

```
go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost -ecdsa-curve P256
```

Compile:

```
./compile.sh
```

## Running

Run the serving http server which serves application to browser from `static` directory:

```
./http-server
```

Application is available at `https://localhost:8080`.

To run Webassembly version with NodeJS:

```
cd static
node wasm_exec.js main.wasm
```

To run Go compiled code directly:

```
./covid-cert
```

## TODO

Support multiple countries. Get public keys from:
* https://verifier-api.coronacheck.nl/v4/verifier/public_keys (EU)
* https://covid-status.service.nhsx.nhs.uk/pubkeys/keys.json (UK)