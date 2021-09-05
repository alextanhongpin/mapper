package main

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type ProductMapper
type ProductMapper interface {
	ProductToProductSummary(Products) *ProductSummary
}

type Products struct {
	Items []int64
}

type ProductSummary struct {
	TotalCount int64 `map:"Items,CountItems"`
}

func CountItems(items []int64) int64 {
	var result int64
	for _, item := range items {
		result += item
	}
	return result
}
