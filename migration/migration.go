package migration

import (
	"log"

	"vestra-ecommerce/src/model"
	database "vestra-ecommerce/utils/databases"
)

func Migrate() {
	if err := database.PgSQLDB.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.ProductSize{},
		&model.Cart{},
		&model.CartItem{},
        &model.Wishlist{},
        &model.Order{},
        &model.OrderItem{},
	); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	log.Println("✅ Database migrated successfully")
}
