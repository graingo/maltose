package internal

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/gorm"
)

func LoadTracing(db *gorm.DB) error {
	// Load tracing
	return db.Use(otelgorm.NewPlugin())
}
