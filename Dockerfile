FROM golang:1.12.5-alpine3.9 as builder

ARG TEST_FLAG

ARG PROJECT_PATH=github.com/kklipsch/comed_exporter
ARG CGO_ENABLED=0

WORKDIR /go/src/$PROJECT_PATH
COPY . .

RUN go vet ./...
RUN go test $TEST_FLAG -v ./... 

RUN mkdir -p /out
RUN go build -o /out/comed_exporter $PROJECT_PATH

FROM alpine:3.9

EXPOSE 9010

WORKDIR /root/
COPY --from=builder /out/comed_exporter /usr/local/bin/comed_exporter
RUN apk --no-cache add ca-certificates

ENTRYPOINT ["comed_exporter"]
