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

func (s *GoodsServer) BannerList(ctx context.Context, req *emptypb.Empty) (*proto.BannerListResponse, error) {

	bannerListResponse := proto.BannerListResponse{}
	var total int64
	global.DB.Model(&model.Brands{}).Count(&total)
	bannerListResponse.Total = int32(total)
	var banners []model.Banner
	global.DB.Find(&banners)
	var bannerResponses []*proto.BannerResponse
	for _, banner := range banners {
		bannerResponse := proto.BannerResponse{
			Id:    banner.ID,
			Index: banner.Index,
			Image: banner.Image,
			Url:   banner.Url,
		}
		bannerResponses = append(bannerResponses, &bannerResponse)
	}
	bannerListResponse.Data = bannerResponses
	return &bannerListResponse, nil

}
func (s *GoodsServer) CreateBanner(ctx context.Context, req *proto.BannerRequest) (*proto.BannerResponse, error) {
	if result := global.DB.Where("image=?", req.Image).First(&model.Banner{}); result.RowsAffected == 1 {
		return nil, status.Errorf(codes.InvalidArgument, "轮播图已存在")
	}
	banner := model.Banner{}
	banner.Url = req.Url
	banner.Index = req.Index
	banner.Image = banner.Image
	global.DB.Save(&banner)
	return &proto.BannerResponse{Id: banner.ID}, nil
}
func (s *GoodsServer) DeleteBanner(ctx context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Banner{}, req.Id).First(&model.Banner{}); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "轮播图不存在")
	}
	return &emptypb.Empty{}, nil
}
func (s *GoodsServer) UpdateBanner(ctx context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	if result := global.DB.Where("id=?", req.Id).First(&model.Banner{}); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "轮播图不存在")
	}
	banner := model.Banner{}
	if req.Url != "" {
		banner.Url = req.Url
	}
	if req.Image != "" {
		banner.Image = req.Image
	}
	if req.Index != 0 {
		banner.Index = req.Index
	}
	global.DB.Save(&banner)
	return &emptypb.Empty{}, nil
}
