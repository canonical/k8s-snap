#!/bin/bash

cfssl gencert --initca=true "$CERTS_DIR/etcd-root-ca-csr.json" | cfssljson --bare "$CERTS_DIR/etcd-root-ca"

cfssl gencert \
  --ca "$CERTS_DIR/etcd-root-ca.pem" \
  --ca-key "$CERTS_DIR/etcd-root-ca-key.pem" \
  --config "$CERTS_DIR/etcd-gencert.json" \
  "$CERTS_DIR/etcd-1-ca-csr.json" | cfssljson --bare "$CERTS_DIR/etcd-1"

cfssl gencert \
  --ca "$CERTS_DIR/etcd-root-ca.pem" \
  --ca-key "$CERTS_DIR/etcd-root-ca-key.pem" \
  --config "$CERTS_DIR/etcd-gencert.json" \
  "$CERTS_DIR/etcd-2-ca-csr.json" | cfssljson --bare "$CERTS_DIR/etcd-2"

cfssl gencert \
  --ca "$CERTS_DIR/etcd-root-ca.pem" \
  --ca-key "$CERTS_DIR/etcd-root-ca-key.pem" \
  --config "$CERTS_DIR/etcd-gencert.json" \
  "$CERTS_DIR/etcd-3-ca-csr.json" | cfssljson --bare "$CERTS_DIR/etcd-3"