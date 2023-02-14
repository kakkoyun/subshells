FROM golang:1.20-alpine3.17 as builder
RUN mkdir /.cache && chown nobody:nogroup /.cache && touch -t 202101010000.00 /.cache

ARG VERSION
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download -modcacherw

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY --chown=nobody:nogroup ./pkg ./pkg
COPY --chown=nobody:nogroup ./cmd/subshells/main.go ./cmd/subshells/main.go
COPY --chown=nobody:nogroup ./cmd/infiniteloop/main.go ./cmd/infiniteloop/main.go

RUN mkdir bin
RUN go build -trimpath -ldflags='--X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=kakkoyun' -a -o ./bin/subshells ./cmd/subshells/main.go
RUN go build -trimpath -ldflags='--X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=kakkoyun' -a -o ./bin/infiniteloop ./cmd/infiniteloop/main.go

FROM alpine:3.17

USER nobody

COPY --chown=0:0 --from=builder /app/bin/subshells /bin/subshells
COPY --chown=0:0 --from=builder /app/bin/infiniteloop /bin/infiniteloop

CMD ["/bin/subshells"]
