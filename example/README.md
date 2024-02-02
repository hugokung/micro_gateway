## Http代理服务模拟
- 进入`http_real_server`目录，执行。
```go
go run real_server.go
```
- 在网址中输入127.0.0.1:8081/`你设置的httpRule`

## Tcp代理服务模拟
- 进入`tcp_real_server`目录，执行。
```go
go run real_server.go
```
- 新建窗口执行
```shell
telnet 127.0.0.1 你设置的代理端口
```

## GRPC代理服务模拟
- 进入`grpc_real_server`目录。
- 进入`server`目录，启动服务端。
```go
go run main.go
```
- 进入`client`目录，在代码中设置好你的代理端口，然后启动客户端。
```go
go run main.go
```

### GRPC测试环境配置

`https://github.com/grpc-ecosystem/grpc-gateway`
- 开启 go mod `export GO111MODULE=on`
- 开启代理 go mod `export GOPROXY=https://goproxy.io`
- 执行安装命令

``` shell
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go install  github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go install github.com/golang/protobuf/protoc-gen-go
```

### 构建grpc-gateway 测试服务端

- 编写 `echo-gateway.proto`
- 运行IDL生成命令
```
protoc -I/usr/local/include -I. -I/root/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis --go_out=plugins=grpc:proto echo-gateway.proto
```
- 运行gateway生成命令
```
protoc -I/usr/local/include -I. -I/root/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis --grpc-gateway_out=logtostderr=true:proto echo-gateway.proto
```