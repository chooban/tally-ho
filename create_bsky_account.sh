#!/usr/bin/env sh

set -e

URL=https://pds.tallyho.test
NAME=$1
pds_account_password=$(openssl rand --hex 16)

curl -XPOST $URL/xrpc/com.atproto.server.createAccount \
  -H "Content-Type: application/json" \
  -d '{"handle": "'$NAME.tallyho.test'", "email": "'$NAME@tallyho.test'", "password": "'$pds_account_password'"}' \
  --cacert ./localdev/caddy/caddy/pki/authorities/tallyho/root.crt

echo "\nPassword set to" $pds_account_password
