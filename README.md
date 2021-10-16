# EU health certificate (Covid-19 passport) verification in browser

Application uses browser Camera or Barcode API to take a picture. This version only supports
Certificates signed by "The Social Insurance Institution of Finland" public key.

## Compiling

This app requires Go, install in on macOS by running `brew install go`.

As the Javascript uses Camera API, it needs TLS. Generate certs. Also, fetch the list of public keys
used for verification. Third command copies newest version of `wasm_exec.js` to enable web assembly
usage.

```
go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost -ecdsa-curve P256
curl -o list_of_keys.json https://verifier-api.coronacheck.nl/v4/verifier/public_keys
cp $(go env GOROOT)/misc/wasm/wasm_exec.js static
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

The following command line versions require file called `data.txt` which should include the
QR code data (starting with "HC1:" string).

To run Webassembly version with NodeJS:

```
cd static
node wasm_exec.js main.wasm
```

To run Go compiled code directly:

```
./covid-cert
```

## Public key list

Support multiple countries. Get public keys from:
* https://verifier-api.coronacheck.nl/v4/verifier/public_keys (EU)
* https://covid-status.service.nhsx.nhs.uk/pubkeys/keys.json (UK, not supported atm)
