package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"mxshop_srvs/goods_srv/proto"
)

var brandClient proto.GoodsClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("192.168.1.21:50052", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	brandClient = proto.NewGoodsClient(conn)
}
func TestGetBrandList() {
	rsp, err := brandClient.BrandList(context.Background(), &proto.BrandFilterRequest{
		Pages:       1,
		PagePerNums: 10,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	for _, brand := range rsp.Data {
		fmt.Println(brand.Name)
	}
}

func TestCreateBrand() {
	rsp, err := brandClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: "454圣",
		Logo: "",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Id)
}

func main() {
	Init()
	TestGetBrandList()
	//TestCreateBrand()
	conn.Close()
}
