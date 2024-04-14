package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"mxshop_srvs/order_srv/global"
	"mxshop_srvs/order_srv/model"
	"mxshop_srvs/order_srv/proto"
)

type OrderServer struct {
	proto.UnimplementedOrderServer
}

func (*OrderServer) CartItem(ctx context.Context, req *proto.UserInfo) (*proto.CartItemListResponse, error) {
	//获取用户购物车列表
	var shopCarts []model.ShoppingCart

	result := global.DB.Where(&model.ShoppingCart{User: req.Id}).Find(&shopCarts)
	if result.Error != nil {
		return nil, result.Error
	}
	rsp := proto.CartItemListResponse{
		Total: int32(result.RowsAffected),
	}
	for _, shopCart := range shopCarts {
		rsp.Data = append(rsp.Data, &proto.ShopCartInfoResponse{
			Id:      shopCart.ID,
			UserId:  shopCart.User,
			GoodsId: shopCart.Goods,
			Nums:    shopCart.Nums,
			Checked: shopCart.Checked,
		})
	}
	return &rsp, nil
}

func (*OrderServer) CreateCartItem(ctx context.Context, req *proto.CartItemRequest) (*proto.ShopCartInfoResponse, error) {
	//新建商品到购物车
	var shopCart model.ShoppingCart

	result := global.DB.Where(&model.ShoppingCart{Goods: req.GoodsId, User: req.UserId}).Find(&shopCart)
	if result.RowsAffected == 1 {
		shopCart.Nums += req.Nums
	} else {
		shopCart.User = req.UserId
		shopCart.Goods = req.GoodsId
		shopCart.Nums = req.Nums
		shopCart.Checked = false
	}
	global.DB.Save(&shopCart)
	return &proto.ShopCartInfoResponse{
		Id: shopCart.ID,
	}, nil
}

func (*OrderServer) UpdateCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	//更新购物车信息
	var shopCart model.ShoppingCart

	result := global.DB.Where("goods=? and user=?", req.GoodsId, req.UserId).First(&shopCart)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "记录不存在")
	}
	if req.Nums > 0 {
		shopCart.Nums = req.Nums
	}
	shopCart.Checked = req.Checked
	global.DB.Save(&shopCart)
	return &emptypb.Empty{}, nil
}

func (*OrderServer) DeleteCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	if result := global.DB.Where("goods=? and user=?", req.GoodsId, req.UserId).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}
	return &emptypb.Empty{}, nil
}

func (*OrderServer) CreateOrder(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	//1 获取购物车中的商品
	var shopCarts []model.ShoppingCart
	var goodIds []int32
	tx := global.DB.Begin()
	goodsNumsMap := make(map[int32]int32)
	result := global.DB.Where(&model.ShoppingCart{
		User:    req.UserId,
		Checked: true,
	}).Find(&shopCarts)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "未选中结算的商品")
	}
	for _, shopcart := range shopCarts {
		goodIds = append(goodIds, shopcart.Goods)
		goodsNumsMap[shopcart.Goods] = shopcart.Nums
	}
	//跨服务调用
	goods, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{Id: goodIds})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "批量查询商品信息失败 %s", err.Error())
	}
	var orderAmount float32
	var orderGoods []*model.OrderGoods
	var goodsInv []*proto.GoodsInvInfo
	for _, good := range goods.Data {
		orderAmount += good.ShopPrice * float32(goodsNumsMap[good.Id])
		orderGoods = append(orderGoods, &model.OrderGoods{
			Goods:      good.Id,
			GoodsName:  good.Name,
			GoodsImage: good.GoodsFrontImage,
			GoodsPrice: good.ShopPrice,
			Nums:       goodsNumsMap[good.Id],
		})
		goodsInv = append(goodsInv, &proto.GoodsInvInfo{
			GoodsId: good.Id,
			Num:     goodsNumsMap[good.Id],
		})
	}
	//扣减库存
	_, err = global.InventorySrvClient.Sell(context.Background(), &proto.SellInfo{GoodsIndo: goodsInv})
	if err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "扣减库存失败")
	}
	//生成订单表
	order := model.OrderInfo{
		User:         req.UserId,
		OrderSn:      GenerateOrderSn(req.UserId),
		OrderMount:   orderAmount,
		Address:      req.Address,
		SignerName:   req.Name,
		SignerMobile: req.Mobile,
		Post:         req.Post,
	}
	if result = tx.Save(&order); result.RowsAffected == 0 {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "创建订单失败")
	}
	for _, orderGood := range orderGoods {
		orderGood.Order = order.ID
	}
	if result = tx.CreateInBatches(orderGoods, 100); result.RowsAffected == 0 {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "创建订单失败")
	}
	if result = tx.Where(&model.ShoppingCart{
		User:    req.UserId,
		Checked: true,
	}).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "创建订单失败")
	}
	tx.Commit()

	return &proto.OrderInfoResponse{Id: order.ID, OrderSn: order.OrderSn, Total: order.OrderMount}, nil
}

func (*OrderServer) OrderList(ctx context.Context, req *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	var orders []model.OrderInfo
	var rsp proto.OrderListResponse
	var total int64
	global.DB.Where(&model.OrderInfo{User: req.UserId}).Count(&total)
	rsp.Total = int32(total)
	global.DB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Where(&model.OrderInfo{User: req.UserId}).Find(&orders)
	for _, order := range orders {
		rsp.Data = append(rsp.Data, &proto.OrderInfoResponse{
			Id:      order.ID,
			UserId:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SignerMobile,
			AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &rsp, nil
}

func (*OrderServer) OrderDetail(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	var order model.OrderInfo
	var rsp proto.OrderInfoDetailResponse
	var orderGoods []model.OrderGoods

	result := global.DB.Where(&model.OrderInfo{BaseModel: model.BaseModel{ID: req.Id}, User: req.UserId}).First(&order)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	orderInfo := proto.OrderInfoResponse{}
	orderInfo.Id = order.ID
	orderInfo.UserId = order.User
	orderInfo.OrderSn = order.OrderSn
	orderInfo.PayType = order.PayType
	orderInfo.Status = order.Status
	orderInfo.Post = order.Post
	orderInfo.Total = order.OrderMount
	orderInfo.Name = order.SignerName
	orderInfo.Mobile = order.SignerName
	orderInfo.Address = order.Address
	rsp.OrderInfo = &orderInfo

	global.DB.Where(&model.OrderGoods{Order: order.ID}).Find(&orderGoods)
	for _, orderGood := range orderGoods {
		rsp.Goods = append(rsp.Goods, &proto.OrderItemResponse{
			GoodsId:    orderGood.Goods,
			GoodsName:  orderGood.GoodsName,
			GoodsPrice: orderGood.GoodsPrice,
			Nums:       orderGood.Nums,
		})
	}

	return &rsp, nil
}

func (*OrderServer) UpdateOrderStatus(ctx context.Context, req *proto.OrderStatus) (*emptypb.Empty, error) {
	result := global.DB.Model(&model.OrderInfo{}).Where("order_sn = ?", req.OrderSn).Update("status", req.Status)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	return &emptypb.Empty{}, nil
}
