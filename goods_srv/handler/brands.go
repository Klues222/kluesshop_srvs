package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"mxshop_srvs/goods_srv/global"
	"mxshop_srvs/goods_srv/model"
	"mxshop_srvs/goods_srv/proto"
)

func (g *GoodsServer) BrandList(ctx context.Context, req *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	brandListResponse := proto.BrandListResponse{}
	var brands []model.Brands
	global.DB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&brands)
	var total int64
	global.DB.Model(&model.Brands{}).Count(&total)
	brandListResponse.Total = int32(total)
	var brandResponses []*proto.BrandInfoResponse
	for _, brands := range brands {
		brandResponse := proto.BrandInfoResponse{
			Id:   brands.ID,
			Name: brands.Name,
			Logo: brands.Logo,
		}
		brandResponses = append(brandResponses, &brandResponse)
	}
	brandListResponse.Data = brandResponses
	return &brandListResponse, nil
}

func (g *GoodsServer) CreateBrand(ctx context.Context, req *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	//新建品牌
	//先检测品牌是否存在
	if result := global.DB.Where("name=?", req.Name).First(&model.Brands{}); result.RowsAffected == 1 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌已存在")

	}

	brand := &model.Brands{}
	brand.Name = req.Name
	brand.Logo = req.Logo
	global.DB.Save(brand)
	return &proto.BrandInfoResponse{Id: brand.ID}, nil
}
func (g *GoodsServer) DeleteBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Brands{}, req.Id).First(&model.Brands{}); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}
	return &emptypb.Empty{}, nil

}

func (g *GoodsServer) UpdateBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	brands := model.Brands{}
	if result := global.DB.Where("id=?", req.Id).First(&brands); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌不存在")

	}
	if req.Name != "" {
		brands.Name = req.Name
	}
	if req.Logo != "" {
		brands.Logo = req.Logo
	}
	global.DB.Save(&brands)
	return &emptypb.Empty{}, nil
}
