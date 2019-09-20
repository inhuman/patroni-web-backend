package db

import (
	"github.com/jinzhu/gorm"

	"github.com/pkg/errors"
)

type Memo struct {
	gorm.Model
	Author   string `json:"author" gorm:"not null"`
	Text     string `json:"text"`
	Name     string `json:"name" gorm:"not null"`
	ObjectId string `json:"object_id" gorm:"not null"`
}

func (m *Memo) Create() error {
	return Stor.Db().Create(m).Error
}

func (m *Memo) Save() error {
	return Stor.Db().Save(m).Error
}

func (m *Memo) Delete() error {
	return Stor.Db().Unscoped().Delete(m).Error
}

func (m *Memo) Check() error {

	if m.Author == "" {
		return errors.New("Author can not be empty")
	}

	if m.Name == "" {
		return errors.New("Name can not be empty")
	}

	if m.ObjectId == "" {
		return errors.New("ObjectId can not be empty")
	}

	return nil
}

func GetMemosById(objectId string) []*Memo {

	var memos []*Memo

	Stor.Db().Where("object_id = ?", objectId).Find(&memos)

	return memos
}
