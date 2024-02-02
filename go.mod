module github.com/hugokung/micro_gateway

go 1.14

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/boj/redistore v0.0.0-20180917114910-cd5dcc76aeff // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/e421083458/grpc-proxy v0.2.0
	github.com/garyburd/redigo v1.6.0
	github.com/gin-gonic/contrib v0.0.0-20190526021735-7fb7810ed2a0
	github.com/gin-gonic/gin v1.7.7
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/spec v0.20.8 // indirect
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang/protobuf v1.5.3
	github.com/gorilla/sessions v1.1.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/mwitkow/grpc-proxy v0.0.0-20230212185441-f345521cb9c9 // indirect
	github.com/pkg/errors v0.8.1
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/spf13/viper v1.4.0
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.2.0
	github.com/swaggo/swag v1.8.10
	golang.org/x/time v0.3.0
	golang.org/x/tools v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20210401141331-865547bb08e2
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.28.1
	gopkg.in/go-playground/validator.v9 v9.29.0
	gorm.io/driver/mysql v1.2.1
	gorm.io/gorm v1.22.4
)

replace github.com/gin-contrib/sse v0.1.0 => github.com/e421083458/sse v0.1.1
