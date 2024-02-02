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