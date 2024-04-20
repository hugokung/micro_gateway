package main

import (
	"context"
	"go.etcd.io/etcd/client/v3"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建 Gin 实例
	router := gin.Default()

	// 定义路由
	router.GET("/test_http_service/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	// 启动 HTTP 服务
	go func() {
		if err := router.Run(":8771"); err != nil {
			log.Fatal("Error starting HTTP server:", err)
		}
	}()

	// 上报 IP 地址到 etcd
	reportIPToEtcd()

	// 保持程序运行
	select {}
}

// 获取本机IP地址
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// 上报 IP 地址到 etcd
func reportIPToEtcd() {
	// 连接 etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal("Error connecting to etcd:", err)
	}
	defer cli.Close()

	// 获取本机IP地址
	ipAddr := getLocalIP()
	if ipAddr == "" {
		log.Fatal("Failed to get local IP address")
	}

	// 注册服务到 etcd
	key := "/test_http_service"
	value := ipAddr + ":8771"
	ctx := context.Background()
	//leaseResp, err := cli.Grant(ctx, 10) // 设置 TTL 为 10 秒
	//if err != nil {
	//log.Fatal("Error granting lease:", err)
	//}
	_, err = cli.Put(ctx, key, value)
	if err != nil {
		log.Fatal("Error putting value to etcd:", err)
	}

	log.Printf("Successfully registered service to etcd: key=%s, value=%s\n", key, value)
}
