package domain

import "time"

// Product represents the core product business entity
type Product struct {
	ID           string
	Name         string
	Price        int
	ImageURL     string
	League       string
	KitType      string
	Year         int
	IsTopSelling bool
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Sizes        []ProductSize 
}