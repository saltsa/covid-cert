# EU certificate verification in browser

Application uses browser Camera or Barcode API to take a picture. This version only supports
Certificates signed by "The Social Insurance Institution of Finland" public key.

## Installation

As the Javascript uses Camera API, it needs TLS. Generate certs:

```
go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost -ecdsa-curve P256
```


## Running

Run the serving http server:

```
./run.sh
```

Application is available at `https://localhost:8080`

## Running command line version

Command line version expects the QR code data to be in file called `data.txt`.
```
./run_local.sh
```
