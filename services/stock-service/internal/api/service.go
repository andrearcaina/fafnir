package api

import "fafnir/stock-service/internal/db"

type Service struct {
	db *db.Database
}

func NewStockService(database *db.Database) *Service {
	return &Service{
		db: database,
	}
}
