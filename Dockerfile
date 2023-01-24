#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git build-base
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build -o /go/bin/app -v ./...

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
COPY config.yaml /config.yaml
ENTRYPOINT /app
LABEL Name=gitlab-clickup-connect Version=0.0.1
EXPOSE 8080
