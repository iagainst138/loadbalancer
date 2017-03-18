#!/bin/bash

# Credit: https://gist.github.com/spikebike/2232102

CN=localhost

[ $# -eq 1 ] && CN=$1

mkdir -p certs
rm certs/*

echo "making server cert:"
openssl req -new -nodes -x509 \
    -out certs/server.pem \
    -keyout certs/server.key \
    -days 3650 \
    -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=$CN/emailAddress=test@example.com"

echo -e "\nmakng client cert:"
openssl req -new -nodes -x509 \
    -out certs/client.pem \
    -keyout certs/client.key \
    -days 3650 \
    -subj "/C=DE/ST=NRW/L=Earth/O=Random Company/OU=IT/CN=$CN/emailAddress=test@example.com"
