package repository

import (
	"AwsProj/internal/app/ds"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (r *Repository) GetAllElements() ([]ds.Elements, error) {
	var elements []ds.Elements
	err := r.db.Where("is_delete = false").Find(&elements).Error // find заполняет слайс
	if err != nil {
		return nil, err //сгнатура ребует два значения, поэтоу nil, err
	}
	return elements, nil
}

func (r *Repository) GetElementByID(id int) (*ds.Elements, error) {
	// 	type Elements struct {
	// 	ID            int    `gorm:"primaryKey"`
	// 	IsDelete      bool   `gorm:"type:boolean not null;default:false"`
	// 	Img           string `gorm:"type:varchar(100)"`
	// 	Name          string `gorm:"type:varchar(25);not null"`
	// 	Description   string `gorm:"type:varchar(100)"`
	// 	Ph            float32
	// 	Concentration float32
	// }

	query := "SELECT id, img, name, description, ph, Concentration FROM elements WHERE id = $1 and is_delete = false" //$1 плейсхолдер от инъекций и для Производительности - БД может кэшировать план запроса

	row := r.db.Raw(query, id).Row() //r.db.Raw(query, id).Row() - это выполнение "сырого" SQL запроса и получение одной строки результата.

	elements := &ds.Elements{}

	err := row.Scan(
		&elements.ID,
		&elements.Imgage,
		&elements.Name,
		&elements.Description,
		&elements.Ph,
		&elements.Concentration,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Возвращаем nil, если записи нет
		}
		return nil, err
	}

	return elements, nil
}

func (r *Repository) SearchElementByName(name string) ([]ds.Elements, error) {
	var elements []ds.Elements
	err := r.db.Where("name LIKE ? and is_delete = ?", "%"+name+"%", false).Find(&elements).Error // добавили условие
	if err != nil {
		return nil, err
	}
	return elements, nil
}

// понять!!
func (r *Repository) GetCartCount() int64 {
	var mixedID uint
	var count int64
	creatorID := 1
	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

	err := r.db.Model(&ds.Mixed{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("id").First(&mixedID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.ElemMix{}).Where("mixed_id = ?", mixedID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return count
}

func (r *Repository) DeleteElement(ElementID uint) error {
	err := r.db.Model(&ds.Elements{}).Where("id = ?", ElementID).UpdateColumn("is_delete", true).Error
	fmt.Println(ElementID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении с id %d: %w", ElementID, err)
	}

	return nil
}
