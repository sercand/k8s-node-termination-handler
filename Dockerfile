# build stage
FROM golang:latest AS build-env
RUN go get github.com/golang/dep/cmd/dep
COPY . /go/src/github.com/sercand/k8s-node-termination-handler
WORKDIR /go/src/github.com/sercand/k8s-node-termination-handler
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -tags netgo -o node-termination-handler

# final stage
FROM gcr.io/distroless/static:latest
WORKDIR /app
COPY --from=build-env /go/src/github.com/sercand/k8s-node-termination-handler/node-termination-handler /app/
ENTRYPOINT ["./node-termination-handler"]
