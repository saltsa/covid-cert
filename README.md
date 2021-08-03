# EU certificate verification in browser

## Installation

As the Javascript uses camera API, it needs TLS. Generate certs:

```
go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost -ecdsa-curve P256
```

Run the serving http server:

```
./run.sh
```