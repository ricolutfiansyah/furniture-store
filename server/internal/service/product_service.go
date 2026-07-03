package service

import (
	"context"
	"errors"
	"fmt"
	"furniture-api/internal/domain"
	"furniture-api/internal/repository"
	"log"
)

type ProductRepository interface {
	GetActive(ctx context.Context, limit, offset int) ([]domain.Product, error)
	CountActive(ctx context.Context) (int, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Product, error)
	GetByID(ctx context.Context, id int) (*domain.Product, error)
	GetVariantsByProductID(ctx context.Context, productID int) ([]domain.ProductVariant, error)
	GetImagesByProductID(ctx context.Context, productID int) ([]domain.ProductImage, error)
	GetCategoryByID(ctx context.Context, categoryID int) (*domain.Category, error)
	GetVariantByID(ctx context.Context, id int) (*domain.ProductVariant, error)
}

type ProductService struct {
	productRepo ProductRepository
}

func NewProductService(productRepo ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) GetAll(ctx context.Context, page, pageSize int) (*ProductListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	products, err := s.productRepo.GetActive(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("get active products: %w", err)
	}

	total, err := s.productRepo.CountActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("count active products: %w", err)
	}

	return &ProductListResult{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	product, err := s.productRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by slug: %w", err)
	}

	variants, err := s.productRepo.GetVariantsByProductID(ctx, product.ID)
	if err != nil {
		log.Printf("get variants for product %d: %v", product.ID, err)
	} else {
		product.Variants = variants
	}

	images, err := s.productRepo.GetImagesByProductID(ctx, product.ID)
	if err != nil {
		log.Printf("get images for product %d: %v", product.ID, err)
	} else {
		product.Images = images
	}

	category, err := s.productRepo.GetCategoryByID(ctx, product.CategoryID)
	if err != nil {
		log.Printf("get category %d for product %d: %v", product.CategoryID, product.ID, err)
	} else {
		product.Category = category
	}

	return product, nil
}
