package models

import "time"

type Kernel struct {
	Name         string `gorm:"primary_key; not null"`
	Version      string `gorm:"primary_key; not null"`
	Architecture string `gorm:"primary_key; not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
