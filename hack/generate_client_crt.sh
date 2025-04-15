#!/bin/bash

openssl genpkey -algorithm RSA -out conf/client.key -pkeyopt rsa_keygen_bits:2048
openssl req -new -key conf/client.key -out conf/client.csr -subj "/CN=admin" \
  -addext "subjectAltName=IP:127.0.0.1"
openssl x509 -req -in conf/client.csr -CA conf/apiserver.crt -CAkey conf/apiserver.key -CAcreateserial \
  -out conf/client.crt -days 365 -sha256 \
  -extfile <(printf "subjectAltName=IP:127.0.0.1")