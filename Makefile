proto:
	rm -rf pbs/*.go
	protoc --proto_path=protos --go_out=pbs --go_opt=paths=source_relative \
    --go-grpc_out=pbs --go-grpc_opt=paths=source_relative \
    protos/*.proto

# GO版本1.15以后，废弃了一般的x509证书，需要采用SAN证书进行通信

# SAN(Subject Alternative Name) 是 SSL 标准 x509 中定义的一个扩展。
# 使用了 SAN 字段的 SSL 证书，可以扩展此证书支持的域名，使得一个证书可以支持多个不同域名的解析。
genSANCert:
	rm -rf certs/* 
# 1.生成根证书：生成CA私钥（.key）–>生成CA证书请求（.csr）–>自签名得到根证书（.crt）（CA给自已颁发的证书）
# 1.1 生成根证书私钥
	openssl genrsa -out certs/ca.key 2048 
# 1.2 生成CA证书请求
	openssl req -new -key certs/ca.key -out certs/ca.csr -subj "/C=cn/OU=myorg/O=mytest/CN=localhost"
# 1.3 自签名得到根证书
	openssl x509 -req -days 3650 -in certs/ca.csr -signkey certs/ca.key -out certs/ca.crt

# 2.生成SAN的服务端证书 生成服务端私钥（serve.key）–>生成服务端证书请求（server.csr）–>CA对服务端请求文件签名，生成服务端证书（server.pem）
# 2.1 生成服务端证书私钥
	openssl genrsa -out certs/server.key 2048
# 2.2 根据私钥server.key生成证书请求文件server.csr
	openssl req -new -nodes -key certs/server.key -out certs/server.csr -subj "/C=cn/OU=myorg/O=mytest/CN=localhost" -config ./openssl.cnf -extensions v3_req
# 2.3 请求CA对证书请求文件签名，生成最终证书文件
	openssl x509 -req -days 365 -in certs/server.csr -out certs/server.pem -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -extfile ./openssl.cnf -extensions v3_req


.PHONY: proto genSANCert