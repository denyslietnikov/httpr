FROM golang:1.19 AS builder

WORKDIR /go/src
COPY . .
ARG TARGETARCH
RUN make build TARGETARCH=$TARGETARCH

FROM scratch
WORKDIR /
COPY --from=builder /go/src/bin .
ENTRYPOINT ["./app"]
EXPOSE 8181
EXPOSE 8282
