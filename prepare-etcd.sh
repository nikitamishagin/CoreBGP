#!/bin/bash
# Script prepares etcd instance for development
openssl genrsa -out ./certs/ca.key 2048
openssl req -x509 -new -nodes -key ./certs/ca.key -sha256 -days 1024 -out ./certs/ca.crt -subj "/CN=etcd-ca"

openssl genrsa -out ./certs/server.key 2048
openssl req -new -key ./certs/server.key -out ./certs/server.csr -subj "/CN=localhost"
openssl x509 -req -in ./certs/server.csr -CA ./certs/ca.crt -CAkey ./certs/ca.key -CAcreateserial -out ./certs/server.crt -days 365 -sha256

openssl genrsa -out ./certs/client.key 2048
openssl req -new -key ./certs/client.key -out ./certs/client.csr -subj "/CN=etcd-client"
openssl x509 -req -in ./certs/client.csr -CA ./certs/ca.crt -CAkey ./certs/ca.key -CAcreateserial -out ./certs/client.crt -days 365 -sha256

docker run -d \
  --network host \
  --name etcd \
  --rm \
  -v ./certs:/etc/ssl/certs \
  quay.io/coreos/etcd:v3.5.17 \
  /usr/local/bin/etcd \
    --data-dir=/etcd-data \
    --advertise-client-urls=http://localhost:2379 \
    --listen-client-urls=http://0.0.0.0:2379 \
    --cert-file=/etc/ssl/certs/server.crt \
    --key-file=/etc/ssl/certs/server.key \
    --client-cert-auth \
    --trusted-ca-file=/etc/ssl/certs/ca.crt
