package models

import (
	"grocery/shared"
)

type (
	Product struct {
		Code  string  `json:"code"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
)

func (p *Product) String() string {
	return shared.String(p)
}
