package repository

import (
	"AwsProj/internal/app/ds"
	"errors"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) AddElementToCart(userID, elementID uint, volume float32) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Ищем активную корзину (черновик) для пользователя
		var cart ds.Mixed
		err := tx.Where("creator_id = ? AND status = ?", userID, "draft").First(&cart).Error

		// Если корзина не найдена, создаём новую
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cart = ds.Mixed{
				Status:        "draft",
				DateCreate:    time.Now(),
				DateUpdate:    time.Now(),
				CreatorID:     userID,
				ModeratorID:   userID,
				Concentartion: 0,
				Ph:            0,
			}
			if err := tx.Create(&cart).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// 2. Проверяем, есть ли уже этот элемент в корзине (включая удаленные)
		var existingElem ds.ElemMix
		err = tx.Unscoped().Where("mixed_id = ? AND element_id = ?", cart.ID, elementID).First(&existingElem).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Элемент ещё не в корзине - добавляем новый
			elemMix := ds.ElemMix{
				MixedID:   cart.ID,
				ElementID: elementID,
				Volume:    volume,
				Comment:   "",
				IsDelete:  false,
			}
			if err := tx.Create(&elemMix).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			// Элемент уже существует в корзине
			if existingElem.IsDelete {
				// Восстанавливаем удаленный элемент
				existingElem.IsDelete = false
				existingElem.Volume = volume
				existingElem.Comment = ""
			}
			// Если элемент уже активен - оставляем как есть

			if err := tx.Save(&existingElem).Error; err != nil {
				return err
			}
		}

		// 3. Обновляем дату обновления корзины
		cart.DateUpdate = time.Now()
		if err := tx.Save(&cart).Error; err != nil {
			return err
		}

		return nil
	})
}
