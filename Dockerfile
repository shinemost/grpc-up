FROM golang:alpine AS build

LABEL maintainer="hjfu"

ENV GO11MODULE=on \
  CGO_ENABLE=on \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /app

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go get -d ./...
RUN go install ./...

RUN go build -mod=mod -o grpc-server .

FROM scratch
WORKDIR /app

COPY --from=build /app/grpc-server ./grpc-server
COPY ./configs.yaml ./configs.yaml

ENTRYPOINT ["/app/grpc-server"]
EXPOSE 50051
