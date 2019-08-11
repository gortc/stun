#!/bin/bash

# Copy hmac.ho and hmac_test.go from current go distribution.
cp -v $GOROOT/src/crypto/hmac/{hmac,hmac_test}.go .
git diff {hmac,hmac_test}.go
