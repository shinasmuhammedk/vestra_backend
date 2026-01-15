package repo

import (
	"gorm.io/gorm"
)

type IPgSQLRepository interface {
	Insert(req interface{}) error
	FindById(obj interface{}, id interface{}) error
	Update(obj interface{}, id interface{}, update interface{}) error
	UpdateByFields(obj interface{}, id interface{}, fields map[string]interface{}) error
	Delete(obj interface{}, id interface{}) error
	HardDelete(obj interface{}) error
	FindAll(obj interface{}) error
	FindAllWhere(obj interface{}, query interface{}, args ...interface{}) error
	FindOneWhere(out interface{}, query string, args ...interface{}) error
	InsertAndReturnID(req interface{}) (uint, error)
	FindDistinct(obj interface{}, field string, query interface{}, args ...interface{}) error
	Raw(sql string, values ...interface{}) *gorm.DB
	Save(req interface{}) error
}
