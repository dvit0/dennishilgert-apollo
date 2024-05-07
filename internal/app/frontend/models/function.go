package models

import (
	"time"

	frontendpb "github.com/dennishilgert/apollo/internal/pkg/proto/frontend/v1"
)

type Function struct {
	Uuid                string                    `gorm:"primary_key; not null"`
	Name                string                    `gorm:"not null"`
	Version             string                    `gorm:"not null"`
	Handler             string                    `gorm:"not null"`
	MemoryLimit         int32                     `gorm:"not null"`
	VCpuCores           int32                     `gorm:"not null"`
	Status              frontendpb.FunctionStatus `gorm:"not null"`
	IdleTtl             int32                     `gorm:"not null"`
	LogLevel            string                    `gorm:"not null"`
	RuntimeName         string                    `gorm:"not null"`
	RuntimeVersion      string                    `gorm:"not null"`
	RuntimeArchitecture string                    `gorm:"not null"`
	Runtime             Runtime                   `gorm:"foreignkey:RuntimeName,RuntimeVersion,RuntimeArchitecture;References:Name,Version,KernelArchitecture"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
