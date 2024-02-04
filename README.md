<div align="center">

<h3 align="center">Micro Gateway</h3>
  <p align="center">
    🧱一个高性能微服务网关
    <br />
  </p>
</div>

### Micro Gateway 管理后台Demo
![demo1](./assets/dashboard.png)
![demo2](./assets/service_list.png)
![demo3](./assets/app_list.png)
### ✨功能
![功能脑图](./assets/功能脑图.png)
### 🔧技术栈
#### 后端
- Golang
- Gin
- Gorm
- Redis
- MySql
- Swagger
- Docker
#### 前端
- Vue.js
- Vue-element-admin

### 🚀快速开始
- Golang版本要求Golang1.12+
- 下载类库依赖
```shell
export GO111MODULE=on && export GOPROXY=https://goproxy.cn
cd mirco_gateway
go mod tidy
```
- 创建数据库并导入
```shell
mysql -h localhost -u root -p -e "CREATE DATABASE mirco_gateway DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;"
mysql -h localhost -u root -p mirco_gateway < gateway.sql --default-character-set=utf8
```
- 脚本快速编译部署
```shell
sh onekeyupdate.sh
```
- 使用Docker部署  
部署网关管理服务
```shell
docker run --name dashboard --net host -e TZ=Asia/Shanghai -d dockerfile-dashboard:latest
```
部署代理服务
```shell
docker run --name gateway_server --net host -e TZ=Asia/Shanghai -d dockerfile-server:latest
```
测试  
- `example`目录为模拟下游服务节点的代码。

代理方式
- Http/Https代理：通过`HttpRule.Rule`字段以前缀匹配的形式实现不同下游服务的转发
- TCP代理：通过`TcpRule.Port`字段实现不同tcp服务的转发
- GRPC代理：通过`GrpcRule.Port`字段实现不同GRPC服务的转发

### TODO
- 指标监控

### 💻API文档
生成接口文档：swag init  
然后启动服务器：go run main.go，浏览地址: http://127.0.0.1:8880/swagger/index.html