package repository

import (
	"AwsProj/internal/app/ds"
	"context"
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

// добавление Post
func (r *Repository) CreateElement(element *ds.Elements) error {
	return r.db.Create(element).Error
}

// CheckElementExists проверяет существование элемента по имени
func (r *Repository) CheckElementExists(name string) (bool, error) {
	var count int64
	err := r.db.Model(&ds.Elements{}).
		Where("name = ? AND is_delete = ?", name, false).
		Count(&count).Error
	return count > 0, err
}

// PUT
func (r *Repository) UpdateElement(id int, updates map[string]interface{}) error {
	return r.db.Model(&ds.Elements{}).
		Where("id = ? AND is_delete = ?", id, false).
		Updates(updates).Error
}

// понять!!
func (r *Repository) GetCartCount() int64 {
	var mixedID uint
	var count int64
	creatorID := 1
	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT

	err := r.db.Model(&ds.Mixed{}).Where("creator_id = ? AND status = ?", creatorID, "draft").Select("id").First(&mixedID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.ElemMix{}).Where("mixed_id = ?", mixedID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return count
}

func (r *Repository) DeleteElement(id int) (string, error) {
	// Сначала получаем элемент чтобы узнать путь к изображению
	var element ds.Elements
	err := r.db.Where("id = ? AND is_delete = ?", id, false).First(&element).Error
	if err != nil {
		return "", err
	}

	// Выполняем soft delete
	err = r.db.Model(&ds.Elements{}).
		Where("id = ?", id).
		Update("is_delete", true).Error
	if err != nil {
		return "", err
	}

	// Возвращаем путь к изображению для удаления из MinIO
	return element.Img, nil
}

func (r *Repository) GetElementImagePath(id int) (string, error) {
	var element ds.Elements
	err := r.db.Select("img").Where("id = ? AND is_delete = ?", id, false).First(&element).Error
	if err != nil {
		return "", err
	}
	return element.Img, nil
}

func (r *Repository) GetUserCart(userID uint) (*ds.Mixed, []ds.ElemMix, error) {
	// Ищем корзину пользователя
	var cart ds.Mixed
	err := r.db.Where("creator_id = ? AND status = ?", userID, "draft").First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil // Корзина не существует
		}
		return nil, nil, err
	}

	// Получаем элементы корзины (только не удаленные)
	var cartItems []ds.ElemMix
	err = r.db.Preload("Element").
		Where("mixed_id = ? AND is_delete = ?", cart.ID, false).
		Find(&cartItems).Error
	if err != nil {
		return nil, nil, err
	}

	return &cart, cartItems, nil
}

func (r *Repository) CompleteCartAndCreateNew(userID uint, addedWater float64) (float32, error) {
	var calculatedPH float32

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Находим активную корзину (draft)
		var cart ds.Mixed
		err := tx.Where("creator_id = ? AND status = ?", userID, "draft").First(&cart).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("корзина не найдена")
			}
			return err
		}

		// 2. Рассчитываем pH (заглушка)
		calculatedPH = 6.5 // временная заглушка

		// 3. Обновляем корзину
		cart.Status = "completed"
		cart.Ph = calculatedPH
		cart.DateUpdate = time.Now()
		cart.DateFinish = sql.NullTime{Time: time.Now(), Valid: true}

		if err := tx.Save(&cart).Error; err != nil {
			return err
		}

		// 4. Создаем новую корзину (draft)
		// ModeratorID должен быть NULL (не 0 и не userID) для черновика, чтобы не нарушать foreign key constraint
		newCart := ds.Mixed{
			Status:        "draft",
			DateCreate:    time.Now(),
			DateUpdate:    time.Now(),
			CreatorID:     userID,
			ModeratorID:   nil, // NULL для черновика - модератор назначится при подтверждении
			TotalVolume:   0,
			Concentartion: 0,
			Ph:            0,
		}

		if err := tx.Create(&newCart).Error; err != nil {
			return err
		}

		return nil
	})

	return calculatedPH, err
}

// Упрощенный расчет pH (заглушка)
func (r *Repository) CalculatePH(mixedID uint) float32 {
	// Здесь может быть сложная логика расчета pH на основе элементов
	// Пока просто возвращаем случайное значение или фиксированное

	// Пример: получаем элементы корзины и рассчитываем средний pH
	var elemMixes []ds.ElemMix
	err := r.db.Preload("Element").Where("mixed_id = ?", mixedID).Find(&elemMixes).Error
	if err != nil || len(elemMixes) == 0 {
		return 7.0 // нейтральный pH по умолчанию
	}

	// Простой расчет: среднее значение pH элементов
	var totalPH float32
	for _, elem := range elemMixes {
		totalPH += elem.Element.Ph
	}

	return totalPH / float32(len(elemMixes))
}

func (r *Repository) RemoveFromCart(userID, elementID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Находим активную корзину пользователя
		var cart ds.Mixed
		err := tx.Where("creator_id = ? AND status = ?", userID, "draft").First(&cart).Error
		if err != nil {
			return err
		}

		// 2. Помечаем элемент как удаленный (soft delete)
		result := tx.Model(&ds.ElemMix{}).
			Where("mixed_id = ? AND element_id = ?", cart.ID, elementID).
			Update("is_delete", true)

		if result.Error != nil {
			return result.Error
		}

		// 3. Если элемент не был найден
		if result.RowsAffected == 0 {
			return fmt.Errorf("элемент не найден в корзине")
		}

		// 4. Обновляем дату изменения корзины
		cart.DateUpdate = time.Now()
		if err := tx.Save(&cart).Error; err != nil {
			return err
		}

		return nil
	})
}

// UpdateElementImage обновляет путь к изображению элемента
func (r *Repository) UpdateElementImage(elementID int, imagePath string) error {
	return r.db.Model(&ds.Elements{}).
		Where("id = ? AND is_delete = ?", elementID, false).
		Update("img", imagePath).Error
}

// GetOrCreateUserDraftOrder возвращает или создает черновик корзины пользователя
func (r *Repository) GetOrCreateUserDraftOrder(ctx context.Context, userID int) (*ds.Mixed, error) {
	var cart ds.Mixed

	// Используем существующую логику из GetUserCart
	err := r.db.WithContext(ctx).
		Where("creator_id = ? AND status = ?", userID, "draft").
		First(&cart).Error

	if err != nil {
		// Если черновик не найден, создаем новый
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ModeratorID должен быть NULL (не 0) для черновика, чтобы не нарушать foreign key constraint
			newCart := &ds.Mixed{
				CreatorID:     uint(userID),
				ModeratorID:   nil, // NULL для черновика - модератор назначится при подтверждении
				Status:        "draft",
				DateCreate:    time.Now(),
				DateUpdate:    time.Now(),
				Concentartion: 0,
				Ph:            0,
				TotalVolume:   0,
				AddedWater:    0,
			}
			if err := r.db.WithContext(ctx).Create(newCart).Error; err != nil {
				return nil, err
			}
			return newCart, nil
		}
		return nil, err
	}

	return &cart, nil
}

// CountCartItems возвращает количество элементов в корзине
func (r *Repository) CountCartItems(ctx context.Context, cartID int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ds.ElemMix{}).
		Where("mixed_id = ? AND is_delete = ?", cartID, false).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetCartIconInfo возвращает информацию для иконки корзины
func (r *Repository) GetCartIconInfo(userID int) (int, int, error) {
	// ✅ ИСПОЛЬЗУЙТЕ ПЕРЕДАННЫЙ userID
	cart, err := r.GetOrCreateUserDraftOrder(context.Background(), userID)
	if err != nil {
		return 0, 0, err
	}

	itemsCount, err := r.CountCartItems(context.Background(), int(cart.ID))
	if err != nil {
		return 0, 0, err
	}

	return int(cart.ID), itemsCount, nil
}
