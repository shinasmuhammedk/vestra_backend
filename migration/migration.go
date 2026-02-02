package migration

import (
	"vestra-ecommerce/src/model"
	database "vestra-ecommerce/utils/databases"
	"vestra-ecommerce/utils/logging"
)

// Migrate runs database migrations for all models
func Migrate() {
	logging.Debug.Println("Starting database migration")

	models := []interface{}{
		&model.User{},
		&model.Product{},
		&model.ProductSize{},
		&model.Cart{},
		&model.CartItem{},
		&model.Wishlist{},
		&model.Order{},
		&model.OrderItem{},
		&model.UserAddress{},
		&model.Payment{},
	}

	logging.Debug.Printf("Migrating %d models", len(models))

	for i, model := range models {
		logging.Debug.Printf("Migrating model %d: %T", i+1, model)
	}

	if err := database.PgSQLDB.AutoMigrate(models...); err != nil {
		logging.Error.Fatalf("Database migration failed: %v", err)
	}

	logging.Debug.Println("Database migration completed successfully")
}