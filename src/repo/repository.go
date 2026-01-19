package repo

import (
	"errors"
	"reflect"
	database "vestra-ecommerce/utils/databases"

	"gorm.io/gorm"
)

type PgSQLRepository struct{}

var IPgSQLRepo IPgSQLRepository

func PgSQLInit() {
	IPgSQLRepo = &PgSQLRepository{}
}

// Getter to inject into services
func GetPgSQLRepository() IPgSQLRepository {
	if IPgSQLRepo == nil {
		panic("PGSQLRepo not initialized! Call PgSQLInit first.")
	}
	return IPgSQLRepo
}


// Insert data
func (r *PgSQLRepository) Insert(req interface{}) error {
	if err := database.PgSQLDB.Debug().Create(req).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) Save(req interface{}) error {
	if err := database.PgSQLDB.Debug().Save(req).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) InsertAndReturnID(req interface{}) (uint, error) {
	if err := database.PgSQLDB.Create(req).Error; err != nil {
		return 0, err
	}

	value := reflect.ValueOf(req).Elem()
	idField := value.FieldByName("ID")
	if !idField.IsValid() {
		return 0, errors.New("ID field not found")
	}

	return uint(idField.Uint()), nil
}

func (r *PgSQLRepository) FindById(obj interface{}, id interface{}) error {
	if err := database.PgSQLDB.Debug().Where("id = ?", id).First(obj).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) FindAll(obj interface{}) error {
	if err := database.PgSQLDB.Debug().Find(obj).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) FindOneWhere(out interface{}, query string, args ...interface{}) error {
	return database.PgSQLDB.Debug().Where(query, args...).First(out).Error
}

func (r *PgSQLRepository) FindAllWhere(obj interface{}, query interface{}, args ...interface{}) error {
	if err := database.PgSQLDB.Debug().Where(query, args...).Find(obj).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) Update(obj interface{}, id interface{}, update interface{}) error {
	if err := database.PgSQLDB.Debug().Where("id = ?", id).First(obj).Updates(update).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) UpdateByFields(obj interface{}, id interface{}, fields map[string]interface{}) error {
	if err := database.PgSQLDB.Debug().Model(obj).Where("id = ?", id).Updates(fields).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) Delete(obj interface{}, id interface{}) error {
	if err := database.PgSQLDB.Debug().Where("id = ?", id).Delete(obj).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) HardDelete(obj interface{}) error {
	if err := database.PgSQLDB.Unscoped().Delete(obj).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) FindDistinct(obj interface{}, field string, query interface{}, args ...interface{}) error {
	if err := database.PgSQLDB.Debug().Model(obj).Distinct(field).Where(query, args...).Find(obj).Error; err != nil {
		return err
	}
	return nil
}

func (r *PgSQLRepository) Raw(query string, args ...interface{}) *gorm.DB {
	return database.PgSQLDB.Raw(query, args...)
}


func (r *PgSQLRepository) Exec(sql string, values ...interface{}) *gorm.DB {
	return database.PgSQLDB.Exec(sql, values...)
}


func (r *PgSQLRepository) FindByIdWithPreload(obj interface{}, id interface{}, preloads ...string) error {
	db := database.PgSQLDB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	return db.Where("id = ?", id).First(obj).Error
}

func (r *PgSQLRepository) FindWhereWithPreload(obj interface{}, query string, args []interface{}, preloads ...string) error {
	db := database.PgSQLDB
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	return db.Where(query, args...).Find(obj).Error
}
