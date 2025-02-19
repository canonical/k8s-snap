## Prepare Vault

For the purpose of this guide, we are going to install HashiCorp Vault using
snap and start a Vault server in development mode.

```
sudo snap install vault
vault server -dev &
```

Specify the Vault address through an environment variable:

```
export VAULT_ADDR=http://localhost:8200
```

Enable the PKI secrets engine and set the maximum lease time to 10 years
(87600 hours):

```
vault secrets enable pki
vault secrets tune -max-lease-ttl=87600h pki
```

## Generate the CA certificates

Generate the root CA certificate:

```
vault write pki/root/generate/internal \
    common_name=vault \
    ttl=87600h
```

Generate the intermediate CA certificate. We need the resulting Certificate
Signing Request (CSR) and private key, so for convenience we'll use JSON
formatting and store the output in a file.

```
mkdir myca
vault write pki/intermediate/generate/exported common_name=kubernetes \
    -format=json > myca/intermediate.json
```

Extract the CSR and key to separate files:

```
cat myca/intermediate.json | jq -r '.data.csr' > myca/intermediate.csr
cat myca/intermediate.json | jq -r '.data.private_key' > myca/intermediate.key
```

Sign the intermediate CA using the root CA:

```
vault write -format=json pki/root/sign-intermediate \
    common_name=kubernetes \
    csr=@myca/intermediate.csr \
    ttl=87600h > myca/intermediate-signed.json
```

Extract the resulting intermediate CA certificate:

```
cat myca/intermediate-signed.json | jq -r '.data.ca_chain' \
    > myca/intermediate-chain.crt
cat myca/intermediate-signed.json | jq -r '.data.certificate' \
    > myca/intermediate.crt
```
