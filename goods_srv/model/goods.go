package model

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        int32          `gorm:"primarykey;type:int" json:"id"`
	CreatedAt time.Time      `gorm:"column:add_time" json:"-"`
	UpdatedAt time.Time      `gorm:"column:update_time" json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"`
	IsDeleted bool           `json:"-"`
}

type GormList []string

func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)

}

// 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		if page <= 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// 分类表 加上级分类
type Category struct {
	BaseModel
	Name             string      `gorm:"type:varchar(20);not null" json:"name"`
	ParentCategoryID int32       `json:"parent"`
	ParentCategory   *Category   `json:"-"`
	SubCategory      []*Category `gorm:"foreignKey:ParentCategoryID;reference:ID" json:"sub_category"`
	Level            int32       `gorm:"type:int;not null;default:1" json:"level"`
	IsTab            bool        `gorm:"default:false;not null comment '是否展示'" json:"is_tab"`
}

// 品牌表
type Brands struct {
	BaseModel
	Name string `gorm:"type:varchar(20);not null"`
	Logo string `gorm:"type:varchar(200);not null;default:''"`
}

type GoodsCategoryBrand struct {
	BaseModel
	CategoryId int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Category   Category
	BrandsId   int32 `gorm:"type:int;index:idx_category_brand,unique"`
	Brands     Brands
}

// 重载表名
func (GoodsCategoryBrand) TableName() string {
	return "goodscategorybrand"
}

// 轮播图
type Banner struct {
	BaseModel
	Image string `gorm:"type:varchar(200);not null"`
	Url   string `gorm:"type:varchar(200);not null"`
	Index int32  `gorm:"type:int;default:1;not null comment '访问次数'"`
}

// 商品表
type Goods struct {
	BaseModel
	CategoryId int32 `gorm:"type:int;not null"`
	Category   Category
	BrandsId   int32 `gorm:"type:int;not null"`
	Brands     Brands

	OnSale   bool `gorm:"default:false;not null comment '是否上架'"`
	ShipFree bool `gorm:"default:false;not null comment '是否免运费'"`
	IsNew    bool `gorm:"default:false;not null comment '是否是新品'"`
	IsHot    bool `gorm:"default:false;not null comment '是否热销'"`

	Name            string   `gorm:"type:varchar(50);not null"`
	GoodsSn         string   `gorm:"type:varchar(50);not null comment '商家自己仓库的商品编号'"`
	ClickNum        int32    `gorm:"type:int;default:0;not null comment '点击数'"`
	SoldNum         int32    `gorm:"type:int;default:0;not null comment '已售商品数'"`
	FavNum          int32    `gorm:"type:int;default:0;not null comment '收藏数'"`
	MarketPrice     float32  `gorm:"not null comment '商品价格'"`
	ShopPrice       float32  `gorm:"not null comment '市场价格'"`
	GoodsBrief      string   `gorm:"type:varchar(100);not null comment '简介'"`
	Images          GormList `gorm:"type:varchar(1000);not null comment '展示图片'"`
	DescImages      GormList `gorm:"type:varchar(1000);not null comment '简介简介'"`
	GoodsFrontImage string   `gorm:"type:varchar(200);not null comment '封面图'"`
}
