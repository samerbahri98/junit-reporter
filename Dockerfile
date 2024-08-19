FROM golang:1.22.5 AS build-stage

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY *.go  ./
COPY tpl/ ./tpl/

RUN CGO_ENABLED=0 GOOS=linux go build -o /junit_reporter

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /junit_reporter /junit_reporter

EXPOSE 3000

USER nonroot:nonroot

ENTRYPOINT ["/junit_reporter"]
