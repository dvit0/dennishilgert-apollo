package models

import "time"

type HttpTrigger struct {
	UrlId        string   `gorm:"primary_key"`
	FunctionUuid string   `gorm:"not null"`
	Function     Function `gorm:"foreignkey:FunctionUuid;References:Uuid"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
