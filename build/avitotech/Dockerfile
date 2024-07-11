# Build in go
FROM golang:1.22 as build

ENV BIN_FILE /opt/avitotech/avitotech-app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Caching layers with modules
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Building a static binary Go (no dependencies on C API),
# otherwise it won't work in alpine image.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/avitotech/*

# The output is a subtle image
FROM alpine:3.20

LABEL ORGANIZATION="My Company"
LABEL SERVICE="avitotech"
LABEL MAINTAINERS="example@example.com"

ENV BIN_FILE "/opt/avitotech/avitotech-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/avitotech/config.toml
COPY ./build/avitotech/config.toml ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
