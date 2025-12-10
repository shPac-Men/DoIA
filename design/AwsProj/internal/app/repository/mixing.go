package repository

import (
	"AwsProj/internal/app/ds"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

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

// GetMixedList возвращает список заявок с фильтрацией
func (r *Repository) GetMixedList(filters map[string]interface{}) ([]map[string]interface{}, error) {
	var mixedList []ds.Mixed

	query := r.db.Model(&ds.Mixed{}).
		Preload("Creator").
		Preload("Moderator")

	// --- ФИЛЬТРЫ ---
	if creatorID, ok := filters["creator_id"]; ok {
		query = query.Where("creator_id = ?", creatorID)
	}
	if dateFrom, ok := filters["date_from"]; ok {
		query = query.Where("date_create >= ?", dateFrom)
	}
	if dateTo, ok := filters["date_to"]; ok {
		query = query.Where("date_create <= ?", dateTo)
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	// ----------------

	query = query.Order("date_create DESC")

	err := query.Find(&mixedList).Error
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(mixedList))
	for i, mixed := range mixedList {
		// Логика для ModeratorID типа uint
		modLogin := ""
		if mixed.ModeratorID != 0 { // Проверяем на 0, а не на nil
			modLogin = mixed.Moderator.Login
		}

		result[i] = map[string]interface{}{
			"id":              mixed.ID,
			"status":          mixed.Status,
			"date_create":     mixed.DateCreate,
			"date_update":     mixed.DateUpdate,
			"date_finish":     mixed.DateFinish,
			"creator_login":   mixed.Creator.Login,
			"moderator_login": modLogin,
			"ph":              mixed.Ph,
			"concentration":   mixed.Concentartion,
			"total_volume":    mixed.TotalVolume,
			"added_water":     mixed.AddedWater,
		}
	}

	return result, nil
}

// Вспомогательная функция для получения логина модератора
func getModeratorLogin(moderator ds.Users) string {
	if moderator.ID == 0 {
		return ""
	}
	return moderator.Login
}

// GetMixedByID возвращает заявку по ID с ее элементами
func (r *Repository) GetMixedByID(mixedID uint) (*ds.Mixed, []ds.ElemMix, error) {
	var mixed ds.Mixed

	// Ищем заявку с предзагрузкой создателя и модератора
	err := r.db.Preload("Creator").Preload("Moderator").
		Where("id = ? AND status != ?", mixedID, "deleted").
		First(&mixed).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("заявка не найдена")
		}
		return nil, nil, err
	}

	// Получаем элементы заявки (только не удаленные)
	var elemMixes []ds.ElemMix
	err = r.db.Preload("Element").
		Where("mixed_id = ? AND is_delete = ?", mixedID, false).
		Find(&elemMixes).Error
	if err != nil {
		return nil, nil, err
	}

	return &mixed, elemMixes, nil
}

// UpdateMixed обновляет поля заявки
func (r *Repository) UpdateMixed(mixedID uint, updates map[string]interface{}) error {
	// Сначала проверяем существование заявки
	var mixed ds.Mixed
	err := r.db.Where("id = ? AND status != ?", mixedID, "deleted").First(&mixed).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("заявка не найдена")
		}
		return err
	}

	// Добавляем дату обновления
	updates["date_update"] = time.Now()

	// Обновляем заявку
	err = r.db.Model(&ds.Mixed{}).
		Where("id = ?", mixedID).
		Updates(updates).Error
	if err != nil {
		return err
	}

	return nil
}

// GetMixedByIDForUpdate возвращает заявку для обновления (без лишних прелоадов)
func (r *Repository) GetMixedByIDForUpdate(mixedID uint) (*ds.Mixed, error) {
	var mixed ds.Mixed
	err := r.db.Where("id = ? AND status != ?", mixedID, "deleted").First(&mixed).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("заявка не найдена")
		}
		return nil, err
	}
	return &mixed, nil
}

// ValidateMixedForCompletion проверяет обязательные поля для формирования заявки
func (r *Repository) ValidateMixedForCompletion(mixedID uint) error {
	var mixed ds.Mixed
	err := r.db.Preload("Creator").Preload("Moderator").
		Where("id = ? AND status = ?", mixedID, "draft").
		First(&mixed).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("черновик заявки не найден")
		}
		return err
	}

	// Проверяем, что в заявке есть элементы
	var itemsCount int64
	err = r.db.Model(&ds.ElemMix{}).
		Where("mixed_id = ? AND is_delete = ?", mixedID, false).
		Count(&itemsCount).Error
	if err != nil {
		return err
	}

	if itemsCount == 0 {
		return fmt.Errorf("заявка не может быть пустой")
	}

	return nil
}

// CompleteMixed формирует заявку (меняет статус и проставляет дату формирования)
func (r *Repository) CompleteMixed(mixedID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var mixed ds.Mixed
		err := tx.Where("id = ? AND status = ?", mixedID, "draft").First(&mixed).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("черновик заявки не найден")
			}
			return err
		}

		// Обновляем заявку - формируем ее
		updates := map[string]interface{}{
			"status":      "completed",
			"date_update": time.Now(),
		}

		err = tx.Model(&ds.Mixed{}).
			Where("id = ?", mixedID).
			Updates(updates).Error
		if err != nil {
			return err
		}

		return nil
	})
}

// GetMixedItems возвращает элементы заявки для проверки
func (r *Repository) GetMixedItems(mixedID uint) ([]ds.ElemMix, error) {
	var items []ds.ElemMix
	err := r.db.Preload("Element").
		Where("mixed_id = ? AND is_delete = ?", mixedID, false).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) CompleteMixedWithData(mixedID uint, ph, volume, concentration float64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var mixed ds.Mixed

		// Проверяем существование (статус уже не важен, если модератор решил пересчитать/завершить повторно,
		// но обычно проверяют "draft" или "pending")
		err := tx.Where("id = ?", mixedID).First(&mixed).Error
		if err != nil {
			return err
		}

		// Данные для обновления
		updates := map[string]interface{}{
			"status":        "completed",
			"date_update":   time.Now(),
			"date_finish":   time.Now(),    // <--- Ставим дату завершения
			"ph":            ph,            // <--- Записываем расчеты
			"total_volume":  volume,        // <--- Записываем расчеты
			"concentartion": concentration, // <--- Записываем расчеты (сохраняем орфографию БД)
		}

		// Выполняем обновление
		err = tx.Model(&ds.Mixed{}).
			Where("id = ?", mixedID).
			Updates(updates).Error

		return err
	})
}

func (r *Repository) GetMixedByIDBasic(id uint) (*ds.Mixed, error) {
	mixed := &ds.Mixed{}
	err := r.db.First(mixed, id).Error
	if err != nil {
		return nil, err
	}
	return mixed, nil
}

func (r *Repository) DeleteMixed(mixedID uint) error {
	// Сначала проверяем существование заявки
	var mixed ds.Mixed
	err := r.db.Where("id = ? AND status != ?", mixedID, "deleted").First(&mixed).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("заявка не найдена")
		}
		return err
	}

	// Выполняем soft delete - меняем статус на deleted и проставляем дату обновления
	updates := map[string]interface{}{
		"status":      "deleted",
		"date_update": time.Now(),
	}

	err = r.db.Model(&ds.Mixed{}).
		Where("id = ?", mixedID).
		Updates(updates).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) HardDeleteMixed(mixedID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Удаляем связанные элементы корзины
		err := tx.Where("mixed_id = ?", mixedID).Delete(&ds.ElemMix{}).Error
		if err != nil {
			return err
		}

		// Удаляем саму заявку
		err = tx.Where("id = ?", mixedID).Delete(&ds.Mixed{}).Error
		if err != nil {
			return err
		}

		return nil
	})
}

// DeleteFromMixed удаляет элемент из заявки по mixed_id и element_id
func (r *Repository) DeleteFromMixed(mixedID uint, elementID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Проверяем существование заявки
		var mixed ds.Mixed
		err := tx.Where("id = ? AND status = ?", mixedID, "draft").First(&mixed).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("черновик заявки не найден")
			}
			return err
		}

		// Проверяем существование элемента в заявке
		var elemMix ds.ElemMix
		err = tx.Where("mixed_id = ? AND element_id = ? AND is_delete = ?", mixedID, elementID, false).
			First(&elemMix).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("элемент не найден в заявке")
			}
			return err
		}

		// Выполняем soft delete элемента
		result := tx.Model(&ds.ElemMix{}).
			Where("mixed_id = ? AND element_id = ?", mixedID, elementID).
			Update("is_delete", true)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("элемент не был удален")
		}

		// Обновляем дату изменения заявки
		mixed.DateUpdate = time.Now()
		if err := tx.Save(&mixed).Error; err != nil {
			return err
		}

		return nil
	})
}

// HardDeleteFromMixed полностью удаляет элемент из заявки
func (r *Repository) HardDeleteFromMixed(mixedID uint, elementID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Проверяем существование заявки
		var mixed ds.Mixed
		err := tx.Where("id = ? AND status = ?", mixedID, "draft").First(&mixed).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("черновик заявки не найден")
			}
			return err
		}

		// Полностью удаляем элемент
		result := tx.Where("mixed_id = ? AND element_id = ?", mixedID, elementID).
			Delete(&ds.ElemMix{})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("элемент не найден в заявке")
		}

		// Обновляем дату изменения заявки
		mixed.DateUpdate = time.Now()
		if err := tx.Save(&mixed).Error; err != nil {
			return err
		}

		return nil
	})
}
