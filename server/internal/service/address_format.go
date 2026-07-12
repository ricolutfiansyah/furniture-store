package service

import (
	"furniture-api/internal/domain"
	"strings"
)

func formatShippingAddress(addr *domain.UserAddress) string {
	var sb strings.Builder

	sb.WriteString(addr.RecipientName)
	sb.WriteString(" (")
	sb.WriteString(addr.Phone)
	sb.WriteString(")\n")
	sb.WriteString(addr.AddressLine)
	sb.WriteString(", ")
	sb.WriteString(addr.District)
	sb.WriteString(", ")
	sb.WriteString(addr.City)
	sb.WriteString(", ")
	sb.WriteString(addr.Province)
	sb.WriteString(" ")
	sb.WriteString(addr.PostalCode)

	return sb.String()
}
