#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CERT_DIR="$ROOT_DIR/.local-certs"
CA_KEY="$CERT_DIR/rootCA.key"
CA_CERT="$CERT_DIR/rootCA.pem"
CA_SERIAL="$CERT_DIR/rootCA.srl"
SERVER_KEY="$CERT_DIR/xpressgo.home.arpa.key"
SERVER_CSR="$CERT_DIR/xpressgo.home.arpa.csr"
SERVER_CERT="$CERT_DIR/xpressgo.home.arpa.crt"
OPENSSL_CONFIG="$ROOT_DIR/infra/local-https/openssl.cnf"

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl is required but was not found in PATH" >&2
  exit 1
fi

if [[ ! -f "$OPENSSL_CONFIG" ]]; then
  echo "missing OpenSSL config: $OPENSSL_CONFIG" >&2
  exit 1
fi

mkdir -p "$CERT_DIR"

if [[ ! -f "$CA_KEY" || ! -f "$CA_CERT" ]]; then
  openssl genrsa -out "$CA_KEY" 2048
  openssl req -x509 -new -nodes -key "$CA_KEY" -sha256 -days 3650 \
    -out "$CA_CERT" \
    -subj "/C=UZ/ST=Tashkent/L=Tashkent/O=Xpressgo Local/OU=Development/CN=Xpressgo Local Root CA"
fi

if [[ ! -f "$SERVER_KEY" || ! -f "$SERVER_CERT" ]]; then
  openssl genrsa -out "$SERVER_KEY" 2048
  openssl req -new -key "$SERVER_KEY" -out "$SERVER_CSR" -config "$OPENSSL_CONFIG"
  openssl x509 -req -in "$SERVER_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial \
    -out "$SERVER_CERT" -days 825 -sha256 -extensions req_ext -extfile "$OPENSSL_CONFIG"
fi

rm -f "$CA_SERIAL"

printf '%s\n' "Local CA: $CA_CERT"
printf '%s\n' "Server cert: $SERVER_CERT"
printf '%s\n' "Server key: $SERVER_KEY"
