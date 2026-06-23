package service

import (
	"context"
	"furniture-api/internal/domain"
)

type ProductRepository interface {
	GetAll(ctx context.Context, limit, offset int) ([]domain.Product, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Product, error)
	GetVariantsByProductID(ctx context.Context, productID int) ([]domain.ProductVariant, error)
	GetImagesByProductID(ctx context.Context, productID int) ([]domain.ProductImage, error)
	GetCategoryByID(ctx context.Context, categoryID int) (*domain.Category, error)
}

type ProductService struct {
	productRepo ProductRepository
}

func NewProductService(productRepo ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, error) {
	if page > 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (pageSize - 1) * pageSize

	return s.productRepo.GetAll(ctx, pageSize, offset)
}
