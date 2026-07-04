package service

type AddToCartRequest struct {
	VariantID int `json:"variant_id"`
	Quantity  int `json:"quantity"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity"`
}
