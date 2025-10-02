package repository

import (
	"AwsProj/internal/app/ds"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

	query := "SELECT id, img, name, description, ph, Concentration FROM elements WHERE id = $1 and is_delete = false" //$1 плейсхолдер от инъекций и для Производительности - БД может кэшировать план запроса

	row := r.db.Raw(query, id).Row() //r.db.Raw(query, id).Row() - это выполнение "сырого" SQL запроса и получение одной строки результата.

	elements := &ds.Elements{}

	err := row.Scan(
		&elements.ID,
		&elements.Img,
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
	// ILIKE - case-insensitive LIKE (только в PostgreSQL)
	err := r.db.Where("name ILIKE ? and is_delete = ?", "%"+name+"%", false).Find(&elements).Error
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

func (r *Repository) AddElementToCart(userID, elementID uint, volume float32) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Ищем активную корзину (черновик) для пользователя
		var cart ds.Mixed
		err := tx.Where("creator_id = ? AND status = ?", userID, "draft").First(&cart).Error

		// Если корзина не найдена, создаём новую
		if errors.Is(err, gorm.ErrRecordNotFound) {
			moderatorID := uint(1) // ← ХАРДКОД moderator_id
			cart = ds.Mixed{
				Status:      "draft",
				DateCreate:  time.Now(),
				DateUpdate:  time.Now(),
				CreatorID:   userID,
				ModeratorID: moderatorID, // ← используем хардкод
				Ph:          0,
			}
			if err := tx.Create(&cart).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// 2. Проверяем, не добавлен ли уже этот элемент в корзину
		var existingElem ds.ElemMix
		err = tx.Where("mixed_id = ? AND element_id = ?", cart.ID, elementID).First(&existingElem).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Элемент ещё не в корзине - добавляем
			elemMix := ds.ElemMix{
				MixedID:   cart.ID,
				ElementID: elementID,
				Volume:    volume,
				Comment:   "",
			}
			if err := tx.Create(&elemMix).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// 3. Обновляем данные mixed (дату обновления)
		cart.DateUpdate = time.Now()
		if err := tx.Save(&cart).Error; err != nil {
			return err
		}

		return nil
	})
}
