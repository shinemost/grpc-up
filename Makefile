proto:
	rm -rf pbs/*.go
	protoc --proto_path=protos --go_out=pbs --go_opt=paths=source_relative \
    --go-grpc_out=pbs --go-grpc_opt=paths=source_relative \
    protos/*.proto


.PHONY: proto