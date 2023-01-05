FROM --platform=$BUILDPLATFORM golang:alpine AS build

LABEL maintainer="hjfu"

ARG TARGETARCH

ENV GO11MODULE=on \
  CGO_ENABLE=on \
  GOOS=linux \
  GOARCH=$TARGETARCH

WORKDIR /app
COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go get -d ./...
RUN go install ./...

RUN go build -mod=mod -o grpc-server .

FROM alpine
WORKDIR /app

COPY --from=build /app/grpc-server ./grpc-server
COPY ./configs.yaml ./configs.yaml
COPY ./certs ./certs

ENTRYPOINT ["/app/grpc-server"]
EXPOSE 50051
