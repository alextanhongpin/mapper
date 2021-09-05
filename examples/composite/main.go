package main

import "errors"

//go:generate go run github.com/alextanhongpin/mapper/cmd/mapper -type ProductMapper
type ProductMapper interface {
	ProductToProductSummary(Products) (*ProductSummary, error)
}

type Products struct {
	Items []int64
}

type ProductSummary struct {
	Items      bool  `map:",IsValidStatus"`
	TotalCount int64 `map:"Items,CountItems"`
}

func CountItems(items []int64) int64 {
	var result int64
	for _, item := range items {
		result += item
	}
	return result
}

func IsValidStatus(items []int64) (bool, error) {
	return false, errors.New("not implemented")
}
