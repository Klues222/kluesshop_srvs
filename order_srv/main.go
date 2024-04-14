package main

import (
	"flag"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"mxshop_srvs/order_srv/global"
	"mxshop_srvs/order_srv/handler"
	"mxshop_srvs/order_srv/utils"
	"mxshop_srvs/order_srv/utils/register/consul"
	"net"
	"os"
	"os/signal"
	"syscall"

	"mxshop_srvs/order_srv/initialize"
	"mxshop_srvs/order_srv/proto"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50054, "ip地址")
	//初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	initialize.InitSrvs()
	flag.Parse()
	zap.S().Info("ip: ", *IP)
	//
	debug := os.Getenv("DEBUG")
	if debug == "false" {
		*Port, _ = utils.GetFreePort()
	}

	zap.S().Info("port: ", *Port)
	server := grpc.NewServer()
	proto.RegisterOrderServer(server, &handler.OrderServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("fail to listen:" + err.Error())
	}
	//注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	//启动服务
	go func() {
		err = server.Serve(lis)
		if err != nil {
			panic("fail to start grpc:" + err.Error())
		}
	}()
	//服务注册
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	err = register_client.Register(global.ServerConfig.Host, *Port, global.ServerConfig.Name, global.ServerConfig.Tags, serviceId)
	if err != nil {
		zap.S().Panic("服务注册失败:", err.Error())
	}
	zap.S().Debugf("启动服务器, 端口： %d", *Port)

	//接受终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Info("注销失败:", err.Error())
	} else {
		zap.S().Info("注销成功:")
	}

}
