#!/bin/bash

openssl req -new -newkey rsa:4096 -days 3650 -nodes -x509 \
    -subj "/C=CN/CN=vpn.shawn.local emailAddress=lucheng0127@outlook.com" \
    -addext "subjectAltName=IP:127.0.0.1" \
    -keyout conf/apiserver.key -out conf/apiserver.crt