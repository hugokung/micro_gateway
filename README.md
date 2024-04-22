![Github Repo Stars](https://img.shields.io/github/stars/hugokung/micro_gateway?style=plastic
)
![License](https://img.shields.io/github/license/hugokung/micro_gateway?style=plastic&color=green
)
![Issue](https://img.shields.io/github/issues-search/hugokung/micro_gateway?query=is%3Aopen%20label%3Aenhancement&style=plastic&color=red
)
![Verison](https://img.shields.io/github/v/tag/hugokung/micro_gateway?sort=semver&style=plastic&label=version&color=yellow
)
![Build](https://img.shields.io/github/actions/workflow/status/hugokung/micro_gateway/release.yml?style=plastic
)
![Commit](https://img.shields.io/github/commits-since/hugokung/micro_gateway/latest?style=plastic&color=pink
)
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
### 功能
![功能脑图](./assets/功能脑图.png)
### 技术栈
#### 后端
- Golang
- Gin
- Gorm
- Redis
- MySql
- Swagger
- Docker
- K8s
#### 前端
- Vue.js
- Vue-element-admin

### 特性
#### 代理协议
- [x] Http/Https
- [x] Tcp
- [x] Grpc
- [ ] WebSocket

#### 代理功能
- [x] 流量统计
- [x] 流量限制
- [x] 熔断
- [x] 黑白名单
- [x] 错误重试(Http/Https)

#### 服务发现
- [x] 静态配置
- [x] ETCD
- [x] Zookeeper
- [ ] Nacos  

#### 插件
- [ ] 用户自定义插件

#### 灰度发布
- [x] 按权重分流

#### 性能监测
- [x] pprof
- [ ] Prometheus

#### 部署方式
- [x] 单机部署
- [x] Docker
- [x] K8s
- [ ] DockerCompose

### 环境依赖
- Golang版本要求Golang1.12+
- 下载类库依赖
```shell
export GO111MODULE=on && export GOPROXY=https://goproxy.cn
cd micro_gateway
go mod tidy
```
- 创建数据库并导入
```shell
mysql -h localhost -u root -p -e "CREATE DATABASE micro_gateway DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;"
mysql -h localhost -u root -p micro_gateway < gateway.sql --default-character-set=utf8
```
### 前端部署
#### 控制面板前端与后端服务分开部署时，前端项目需要如下设置：  
  - 在`vue.config.js`文件中设置`publicPath`为`/`
  - 在`.env.production`文件中设置`VUE_APP_BASE_API`为自己需要的url前缀，本项目设置为`/prod-api`。
  - 编译。
  ```sh
  npm run build:prod
  ```
  - 通过nginx代理实现与后端接口服务的同域访问。
  ```sh
   server {
        listen       8884;
        server_name  localhost;
        root /dashboard编译生成的结果的路径;
        index  index.html index.htm index.php;

        location / {
            try_files $uri $uri/ /index.html?$args;
        }

        location /prod-api/ {
            proxy_pass http://127.0.0.1:8880/; #后端服务接口
        }
  }
  ```
  - 访问`http://你的ip:8884`即可。
#### 控制面板前端与后端项目合并部署   
  - 在`vue.config.js`文件中设置`publicPath`为`/dist`
  - 在`.env.production`文件中设置`VUE_APP_BASE_API`为空。
  - 在后端项目的`router`包的`route.go`文件中增加代码
  ```go
  router.Static("/dist", "./dist")
  ``` 
  - 编译后放入到后端项目的根目录下。
  - 访问`http://后端IP:后端port/dist`
  
### 后端部署
#### 直接编译源码运行
```shell
make build_dev
sh run.sh
```
#### 使用Docker部署  
- 部署网关管理服务
```shell
docker build -f dockerfile-dashboard -t gateway-dashboard .
docker run --name dashboard --net host -e TZ=Asia/Shanghai -d gateway-dashboard:latest
```
- 部署代理服务
```shell
docker build -f dockerfile-server -t gateway-server .
docker run --name server --net host -e TZ=Asia/Shanghai -d gateway-server:latest
```
- 需要再额外自己部署Redis和Mysql服务器。

#### 使用K8s部署
```shell
kubectl apply -f k8s_gateway_mysql.yaml
kubectl apply -f k8s_gateway_redis.yaml
kubectl apply -f k8s_dashboard.yaml
kubectl apply -f k8s_server.yaml
```

### 测试  
- `example`目录为模拟下游服务节点的代码。

### 代理规则
- `HTTP/HTTPS`代理：通过`HttpRule.Rule`字段以前缀匹配的形式实现不同下游服务的转发
- `TCP`代理：通过`TcpRule.Port`字段实现不同tcp服务的转发
- `GRPC`代理：通过`GrpcRule.Port`字段实现不同GRPC服务的转发


### 💻API文档
生成接口文档：`swag init`  
然后启动服务器：`go run main.go`，浏览地址: `http://127.0.0.1:8880/swagger/index.html`
