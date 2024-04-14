package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"mxshop_srvs/inventory_srv/proto"
	"sync"
)

var brandClient proto.InventoryClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("192.168.1.21:50053", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	brandClient = proto.NewInventoryClient(conn)
}

func TestSetInv(goodID int32, num int32) {
	_, err := brandClient.SetInv(context.Background(), &proto.GoodsInvInfo{
		GoodsId: goodID,
		Num:     num,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("设置库存成功")

}
func TestInvDetail(goodID int32) {
	r, err := brandClient.InvDetail(context.Background(), &proto.GoodsInvInfo{
		GoodsId: goodID,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("goodId : %d stock : %d", r.GoodsId, r.Num)

}

func TestSell(wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := brandClient.Sell(context.Background(), &proto.SellInfo{
		GoodsIndo: []*proto.GoodsInvInfo{
			{
				GoodsId: 421,
				Num:     1,
			},
		},
	})
	if err != nil {
		fmt.Println("出库失败")
	}
	fmt.Println("出库成功")

}

func TestReBack(wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := brandClient.ReBack(context.Background(), &proto.SellInfo{
		GoodsIndo: []*proto.GoodsInvInfo{
			{
				GoodsId: 421,
				Num:     1,
			},
		},
	})
	if err != nil {
		fmt.Println("回库失败")
	}
	fmt.Println("回库成功")

}

func main() {
	Init()
	var rw sync.WaitGroup
	rw.Add(20)
	for i := 0; i < 20; i++ {
		go TestReBack(&rw)

	}
	rw.Wait()
	//TestInvDetail(421)
	//TestSell()
	//TestReBack()
	conn.Close()
}
