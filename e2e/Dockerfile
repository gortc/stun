ARG CI_GO_VERSION
FROM golang:${CI_GO_VERSION}

ADD . /go/src/gortc.io/stun

WORKDIR /go/src/gortc.io/stun/e2e

RUN go install .

CMD ["e2e"]

