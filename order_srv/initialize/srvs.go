package initialize

import (
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"mxshop_srvs/order_srv/global"
	"mxshop_srvs/order_srv/proto"
)

func InitSrvs() {
	//初始化第三方微服务连接
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host,
			global.ServerConfig.ConsulInfo.Port, global.ServerConfig.GoodSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatal("[InitSrvConn] 连接 【商品服务失败】")
	}
	inventoryConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host,
			global.ServerConfig.ConsulInfo.Port, global.ServerConfig.InventorySrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatal("[InitSrvConn] 连接 【库存服务失败】")
	}

	goodsSrvClient := proto.NewGoodsClient(goodsConn)
	inventorySrvCLient := proto.NewInventoryClient(inventoryConn)
	global.GoodsSrvClient = goodsSrvClient
	global.InventorySrvClient = inventorySrvCLient
}
