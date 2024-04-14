package handler

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"mxshop_srvs/inventory_srv/global"
	"mxshop_srvs/inventory_srv/model"
	"mxshop_srvs/inventory_srv/proto"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

func (*InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	//设置库存
	var inv model.Inventory
	global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv)
	if inv.Goods == 0 {
		inv.Goods = req.GoodsId
	}
	inv.Stocks = req.Num

	global.DB.Save(&inv)
	return &emptypb.Empty{}, nil
}

func (*InventoryServer) InvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	//获取商品库存详情
	var inv model.Inventory
	if result := global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "库存信息不存在")
	}

	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

func (*InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存
	//事务
	//并发 可能出现超卖
	//悲观锁 tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv);
	//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
	//	return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
	//}
	//乐观锁
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsIndo {
		var inv model.Inventory
		for {
			if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
				tx.Rollback()
				return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
			}
			if inv.Stocks < goodInfo.Num {
				return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
			}
			//扣减
			if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version = ?", goodInfo.GoodsId, inv.Version).Updates(model.Inventory{
				Stocks:  inv.Stocks - goodInfo.Num,
				Version: inv.Version + 1,
			}); result.RowsAffected == 0 {
				zap.S().Info("库存扣减失败")
			} else {
				break
			}
		}
		//tx.Save(&inv)
	}
	tx.Commit() //需要手动提交操作
	//sellLock.Unlock()
	return &emptypb.Empty{}, nil
}

//func (*InventoryServer) ReBack(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
//	//库存归还 : 1 订单超时归还 2 订单创建失败 3 手动归还
//	client := goredislib.NewClient(&goredislib.Options{
//		Addr: "127.0.0.1:6379",
//	})
//	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
//	rs := redsync.New(pool)
//
//	tx := global.DB.Begin()
//	for _, goodInfo := range req.GoodsIndo {
//		var inv model.Inventory
//
//		mutex := rs.NewMutex("goods_%d", goodInfo.GoodsId)
//		if err := mutex.Lock(); err != nil {
//			return nil, status.Errorf(codes.Internal, "内部错误")
//		}
//
//		fmt.Println("上锁")
//		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
//			tx.Rollback()
//			return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
//		}
//		inv.Stocks += goodInfo.Num
//		tx.Save(&inv)
//		fmt.Println("执行业务")
//		if ok, err := mutex.Unlock(); !ok || err != nil {
//			return nil, status.Errorf(codes.Internal, "内部错误")
//		}
//		fmt.Println("关锁")
//		//扣减
//
//	}
//	tx.Commit() //需要手动提交操作
//	return &emptypb.Empty{}, nil
//}

func (*InventoryServer) ReBack(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//库存归还 : 1 订单超时归还 2 订单创建失败 3 手动归还
	//client := goredislib.NewClient(&goredislib.Options{
	//	Addr: "127.0.0.1:6379",
	//})
	//pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	//rs := redsync.New(pool)
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsIndo {
		var inv model.Inventory
		for {
			if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
				tx.Rollback()
				return nil, status.Errorf(codes.InvalidArgument, "库存信息不存在")
			}
			inv.Stocks += goodInfo.Num
			if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version = ?", goodInfo.GoodsId, inv.Version).Updates(model.Inventory{
				Stocks:  inv.Stocks,
				Version: inv.Version + 1,
			}); result.RowsAffected == 0 {
				zap.S().Info("库存扣减失败")
			} else {
				break
			}
		}
	}
	tx.Commit() //需要手动提交操作
	return &emptypb.Empty{}, nil
}
