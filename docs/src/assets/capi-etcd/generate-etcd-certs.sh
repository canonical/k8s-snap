#!/bin/bash

cfssl gencert --initca=true "$EXT_ETCD_DIR/etcd-root-ca-csr.json" | cfssljson --bare "$EXT_ETCD_DIR/etcd-root-ca"

cfssl gencert \
  --ca "$EXT_ETCD_DIR/etcd-root-ca.pem" \
  --ca-key "$EXT_ETCD_DIR/etcd-root-ca-key.pem" \
  --config "$EXT_ETCD_DIR/etcd-gencert.json" \
  "$EXT_ETCD_DIR/etcd-1-ca-csr.json" | cfssljson --bare "$EXT_ETCD_DIR/etcd-1"

cfssl gencert \
  --ca "$EXT_ETCD_DIR/etcd-root-ca.pem" \
  --ca-key "$EXT_ETCD_DIR/etcd-root-ca-key.pem" \
  --config "$EXT_ETCD_DIR/etcd-gencert.json" \
  "$EXT_ETCD_DIR/etcd-2-ca-csr.json" | cfssljson --bare "$EXT_ETCD_DIR/etcd-2"

cfssl gencert \
  --ca "$EXT_ETCD_DIR/etcd-root-ca.pem" \
  --ca-key "$EXT_ETCD_DIR/etcd-root-ca-key.pem" \
  --config "$EXT_ETCD_DIR/etcd-gencert.json" \
  "$EXT_ETCD_DIR/etcd-3-ca-csr.json" | cfssljson --bare "$EXT_ETCD_DIR/etcd-3"