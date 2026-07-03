package service

import "furniture-api/internal/domain"

type ProductListResult struct {
	Products []domain.Product
	Total    int
	Page     int
	PageSize int
}
