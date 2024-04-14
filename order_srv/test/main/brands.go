package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"mxshop_srvs/order_srv/proto"
)

var brandClient proto.OrderClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("10.111.196.244:50054", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	brandClient = proto.NewOrderClient(conn)
}

func TestCreateCartItem() {
	_, err := brandClient.CreateCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  12,
		Nums:    1,
		GoodsId: 421,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("添加成功")

}

func TestCartItem() {
	rsp, err := brandClient.CartItem(context.Background(), &proto.UserInfo{})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(rsp.Data)

}

func TestUpdateCartItem() {
	_, err := brandClient.UpdateCartItem(context.Background(), &proto.CartItemRequest{
		Id:      2,
		Nums:    12,
		Checked: true,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("修改成功")

}

func TestDeleteCartItem() {
	_, err := brandClient.DeleteCartItem(context.Background(), &proto.CartItemRequest{
		Id: 1,
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("删除购物车成功")

}

func TestCreateOrder() {
	_, err := brandClient.CreateOrder(context.Background(), &proto.OrderRequest{
		UserId:  12,
		Address: "湖南省",
		Name:    "付伟",
		Mobile:  "12345678900",
		Post:    "无",
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("生成订单成功")

}

//	func TestOrderList(goodID int32, num int32) {
//		_, err := brandClient.SetInv(context.Background(), &proto.GoodsInvInfo{
//			GoodsId: goodID,
//			Num:     num,
//		})
//		if err != nil {
//			fmt.Println(err)
//		}
//		fmt.Println("设置库存成功")
//
// }
//
//	func TestOrderDetail(goodID int32, num int32) {
//		_, err := brandClient.SetInv(context.Background(), &proto.GoodsInvInfo{
//			GoodsId: goodID,
//			Num:     num,
//		})
//		if err != nil {
//			fmt.Println(err)
//		}
//		fmt.Println("设置库存成功")
//
// }
//
//	func TestUpdateOrderStatus(goodID int32, num int32) {
//		_, err := brandClient.SetInv(context.Background(), &proto.GoodsInvInfo{
//			GoodsId: goodID,
//			Num:     num,
//		})
//		if err != nil {
//			fmt.Println(err)
//		}
//		fmt.Println("设置库存成功")
//
// }
func main() {
	Init()
	TestCreateCartItem()
	//TestCartItem()
	//TestUpdateCartItem()
	//TestDeleteCartItem()
	//TestCreateOrder()
}
