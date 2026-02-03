package repo

import (
	"errors"
	"reflect"
	database "vestra-ecommerce/utils/databases"
	"vestra-ecommerce/utils/logging"

	"gorm.io/gorm"
)

type PgSQLRepository struct{}

var IPgSQLRepo IPgSQLRepository

// PgSQLInit initializes the PostgreSQL repository
func PgSQLInit() {
	logging.Debug.Println("Initializing PostgreSQL Repository")
	IPgSQLRepo = &PgSQLRepository{}
	logging.Debug.Println("PostgreSQL Repository initialized successfully")
}

// GetPgSQLRepository returns the repository instance
func GetPgSQLRepository() IPgSQLRepository {
	if IPgSQLRepo == nil {
		logging.Error.Fatalln("PgSQLRepository not initialized! Call PgSQLInit first.")
	}
	return IPgSQLRepo
}

// Insert creates a new record in the database
func (r *PgSQLRepository) Insert(req interface{}) error {
	logging.Debug.Printf("Inserting record: %T", req)
	if err := database.PgSQLDB.Debug().Create(req).Error; err != nil {
		logging.Error.Printf("Insert failed for %T: %v", req, err)
		return err
	}
	logging.Debug.Printf("Record inserted successfully: %T", req)
	return nil
}

// Save updates or creates a record
func (r *PgSQLRepository) Save(req interface{}) error {
	logging.Debug.Printf("Saving record: %T", req)
	if err := database.PgSQLDB.Debug().Save(req).Error; err != nil {
		logging.Error.Printf("Save failed for %T: %v", req, err)
		return err
	}
	logging.Debug.Printf("Record saved successfully: %T", req)
	return nil
}

// InsertAndReturnID inserts record and returns its ID
func (r *PgSQLRepository) InsertAndReturnID(req interface{}) (uint, error) {
	logging.Debug.Printf("Inserting and returning ID for: %T", req)
	if err := database.PgSQLDB.Create(req).Error; err != nil {
		logging.Error.Printf("InsertAndReturnID failed: %v", err)
		return 0, err
	}

	value := reflect.ValueOf(req).Elem()
	idField := value.FieldByName("ID")
	if !idField.IsValid() {
		logging.Error.Println("ID field not found in struct")
		return 0, errors.New("ID field not found")
	}

	id := uint(idField.Uint())
	logging.Debug.Printf("Inserted record ID: %d", id)
	return id, nil
}

// FindById finds a record by ID
func (r *PgSQLRepository) FindById(obj interface{}, id interface{}) error {
	logging.Debug.Printf("Finding by ID: %T with id: %v", obj, id)
	if err := database.PgSQLDB.Debug().Where("id = ?", id).First(obj).Error; err != nil {
		logging.Error.Printf("FindById failed for id %v: %v", id, err)
		return err
	}
	logging.Debug.Printf("Record found by ID: %v", id)
	return nil
}

// FindAll retrieves all records
func (r *PgSQLRepository) FindAll(obj interface{}) error {
	logging.Debug.Printf("Finding all records: %T", obj)
	if err := database.PgSQLDB.Debug().Find(obj).Error; err != nil {
		logging.Error.Printf("FindAll failed for %T: %v", obj, err)
		return err
	}
	logging.Debug.Printf("Found all records: %T", obj)
	return nil
}

// FindOneWhere finds single record with conditions
func (r *PgSQLRepository) FindOneWhere(out interface{}, query string, args ...interface{}) error {
	logging.Debug.Printf("FindOneWhere: %s, args: %v", query, args)
	err := database.PgSQLDB.Debug().Where(query, args...).First(out).Error
	if err != nil {
		logging.Debug.Printf("FindOneWhere not found: %s, args: %v", query, args)
	}
	return err
}

// FindAllWhere finds all records with conditions
func (r *PgSQLRepository) FindAllWhere(obj interface{}, query interface{}, args ...interface{}) error {
	logging.Debug.Printf("FindAllWhere: %v, args: %v", query, args)
	if err := database.PgSQLDB.Debug().Where(query, args...).Find(obj).Error; err != nil {
		logging.Error.Printf("FindAllWhere failed: %v, args: %v - %v", query, args, err)
		return err
	}
	logging.Debug.Printf("FindAllWhere completed: %T", obj)
	return nil
}

// Update updates a record by ID
func (r *PgSQLRepository) Update(obj interface{}, id interface{}, update interface{}) error {
	logging.Debug.Printf("Updating record: %T, id: %v", obj, id)
	if err := database.PgSQLDB.Debug().Where("id = ?", id).First(obj).Updates(update).Error; err != nil {
		logging.Error.Printf("Update failed for id %v: %v", id, err)
		return err
	}
	logging.Debug.Printf("Record updated: id %v", id)
	return nil
}

// UpdateByFields updates specific fields of a record
func (r *PgSQLRepository) UpdateByFields(obj interface{}, id interface{}, fields map[string]interface{}) error {
	logging.Debug.Printf("UpdateByFields: %T, id: %v, fields: %v", obj, id, fields)
	if err := database.PgSQLDB.Debug().Model(obj).Where("id = ?", id).Updates(fields).Error; err != nil {
		logging.Error.Printf("UpdateByFields failed for id %v: %v", id, err)
		return err
	}
	logging.Debug.Printf("Fields updated for id %v", id)
	return nil
}

// Delete soft deletes a record
func (r *PgSQLRepository) Delete(obj interface{}, id interface{}) error {
	logging.Debug.Printf("Deleting record: %T, id: %v", obj, id)
	if err := database.PgSQLDB.Debug().Where("id = ?", id).Delete(obj).Error; err != nil {
		logging.Error.Printf("Delete failed for id %v: %v", id, err)
		return err
	}
	logging.Debug.Printf("Record deleted: id %v", id)
	return nil
}

// HardDelete permanently removes a record
func (r *PgSQLRepository) HardDelete(obj interface{}) error {
	logging.Debug.Printf("Hard deleting record: %T", obj)
	if err := database.PgSQLDB.Unscoped().Delete(obj).Error; err != nil {
		logging.Error.Printf("HardDelete failed: %v", err)
		return err
	}
	logging.Debug.Printf("Record hard deleted: %T", obj)
	return nil
}

// FindDistinct finds distinct values of a field
func (r *PgSQLRepository) FindDistinct(obj interface{}, field string, query interface{}, args ...interface{}) error {
	logging.Debug.Printf("FindDistinct: field=%s, query=%v", field, query)
	if err := database.PgSQLDB.Debug().Model(obj).Distinct(field).Where(query, args...).Find(obj).Error; err != nil {
		logging.Error.Printf("FindDistinct failed: field=%s - %v", field, err)
		return err
	}
	return nil
}

// Raw executes a raw SQL query
func (r *PgSQLRepository) Raw(query string, args ...interface{}) *gorm.DB {
	logging.Debug.Printf("Executing raw query: %s, args: %v", query, args)
	return database.PgSQLDB.Raw(query, args...)
}

// Exec executes a raw SQL statement
func (r *PgSQLRepository) Exec(sql string, values ...interface{}) *gorm.DB {
	logging.Debug.Printf("Executing SQL: %s, values: %v", sql, values)
	return database.PgSQLDB.Exec(sql, values...)
}

// FindByIdWithPreload finds record by ID with preloaded relations
func (r *PgSQLRepository) FindByIdWithPreload(obj interface{}, id interface{}, preloads ...string) error {
	logging.Debug.Printf("FindByIdWithPreload: id=%v, preloads=%v", id, preloads)
	db := database.PgSQLDB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	err := db.Where("id = ?", id).First(obj).Error
	if err != nil {
		logging.Error.Printf("FindByIdWithPreload failed for id %v: %v", id, err)
	}
	return err
}

// FindWhereWithPreload finds records with conditions and preloads
func (r *PgSQLRepository) FindWhereWithPreload(obj interface{}, query string, args []interface{}, preloads ...string) error {
	logging.Debug.Printf("FindWhereWithPreload: query=%s, preloads=%v", query, preloads)
	db := database.PgSQLDB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Where(query, args...).Find(obj).Error; err != nil {
		logging.Error.Printf("FindWhereWithPreload failed: %v", err)
		return err
	}
	return nil
}

// FindAllWithPreload finds all records with preloaded relations
func (r *PgSQLRepository) FindAllWithPreload(obj interface{}, preloads ...string) error {
	logging.Debug.Printf("FindAllWithPreload: %T, preloads: %v", obj, preloads)
	db := database.PgSQLDB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Find(obj).Error; err != nil {
		logging.Error.Printf("FindAllWithPreload failed: %v", err)
		return err
	}
	return nil
}

// Begin starts a database transaction
func (r *PgSQLRepository) Begin() *gorm.DB {
	logging.Debug.Println("Beginning database transaction")
	return database.PgSQLDB.Begin()
}

// Commit commits a transaction
func (r *PgSQLRepository) Commit(tx *gorm.DB) error {
	logging.Debug.Println("Committing transaction")
	if err := tx.Commit().Error; err != nil {
		logging.Error.Printf("Commit failed: %v", err)
		return err
	}
	logging.Debug.Println("Transaction committed")
	return nil
}

// Rollback rolls back a transaction
func (r *PgSQLRepository) Rollback(tx *gorm.DB) error {
	logging.Debug.Println("Rolling back transaction")
	if err := tx.Rollback().Error; err != nil {
		logging.Error.Printf("Rollback failed: %v", err)
		return err
	}
	logging.Debug.Println("Transaction rolled back")
	return nil
}

func (r *PgSQLRepository) DB() *gorm.DB {
	return database.PgSQLDB
}
