FROM golang:1.12

COPY . /go/src/gortc.io/stun

RUN go test gortc.io/stun
