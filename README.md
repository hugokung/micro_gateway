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
cd go_gateway
go mod tidy
```
- 创建数据库并导入
```shell
mysql -h localhost -u root -p -e "CREATE DATABASE go_gateway DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;"
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

### 💻API文档
待完善......