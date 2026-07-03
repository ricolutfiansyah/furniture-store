package service

import (
	"context"
	"furniture-api/internal/domain"
)

type ProductRepository interface {
	GetActive(ctx context.Context, limit, offset int) ([]domain.Product, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Product, error)
	GetVariantsByProductID(ctx context.Context, productID int) ([]domain.ProductVariant, error)
	GetImagesByProductID(ctx context.Context, productID int) ([]domain.ProductImage, error)
	GetCategoryByID(ctx context.Context, categoryID int) (*domain.Category, error)
}

type ProductVariantRepository interface {
	GetByID(ctx context.Context, id int) (*domain.ProductVariant, error)
}

type ProductService struct {
	productRepo ProductRepository
}

func NewProductService(productRepo ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	return s.productRepo.GetActive(ctx, pageSize, offset)
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	product, err := s.productRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}

	variants, err := s.productRepo.GetVariantsByProductID(ctx, product.ID)
	if err == nil {
		product.Variants = variants
	}

	images, err := s.productRepo.GetImagesByProductID(ctx, product.ID)
	if err == nil {
		product.Images = images
	}

	categories, err := s.productRepo.GetCategoryByID(ctx, product.CategoryID)
	if err == nil {
		product.Category = categories
	}

	return product, nil
}
