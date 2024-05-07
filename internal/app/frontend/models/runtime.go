package models

import (
	"time"

	"github.com/lib/pq"
)

type Runtime struct {
	Name               string         `gorm:"primary_key; not null"`
	Version            string         `gorm:"primary_key; not null"`
	BinaryPath         string         `gorm:"not null"`
	BinaryArgs         pq.StringArray `gorm:"type:text[]; not null"`
	DisplayName        string         `gorm:"not null"`
	DefaultHandler     string         `gorm:"not null"`
	DefaultMemoryLimit int32          `gorm:"not null"`
	DefaultVCpuCores   int32          `gorm:"not null"`
	KernelName         string         `gorm:"not null"`
	KernelVersion      string         `gorm:"not null"`
	KernelArchitecture string         `gorm:"primary_key; not null"`
	Kernel             Kernel         `gorm:"foreignkey:KernelName,KernelVersion,KernelArchitecture;References:Name,Version,Architecture"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
