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

// func (r *Repository) AddElementToCart(userID, elementID uint, volume float32) error {
// 	return r.db.Transaction(func(tx *gorm.DB) error {
// 		// 1. Ищем активную корзину (черновик) для пользователя
// 		var cart ds.Mixed
// 		err := tx.Where("creator_id = ? AND status = ?", userID, "draft").First(&cart).Error

// 		// Если корзина не найдена, создаём новую
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			cart = ds.Mixed{
// 				Status:        "draft",
// 				DateCreate:    time.Now(),
// 				DateUpdate:    time.Now(),
// 				CreatorID:     userID,
// 				ModeratorID:   userID,
// 				Concentartion: 0,
// 				Ph:            0,
// 			}
// 			if err := tx.Create(&cart).Error; err != nil {
// 				return err
// 			}
// 		} else if err != nil {
// 			return err
// 		}

// 		// 2. Проверяем, есть ли уже этот элемент в корзине (включая удаленные)
// 		var existingElem ds.ElemMix
// 		err = tx.Unscoped().Where("mixed_id = ? AND element_id = ?", cart.ID, elementID).First(&existingElem).Error

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			// Элемент ещё не в корзине - добавляем новый
// 			elemMix := ds.ElemMix{
// 				MixedID:   cart.ID,
// 				ElementID: elementID,
// 				Volume:    volume,
// 				Comment:   "",
// 				IsDelete:  false, // явно указываем false
// 			}
// 			if err := tx.Create(&elemMix).Error; err != nil {
// 				return err
// 			}
// 		} else if err != nil {
// 			return err
// 		} else {
// 			// Элемент уже существует в корзине (возможно удаленный)
// 			if existingElem.IsDelete {
// 				// Восстанавливаем удаленный элемент
// 				existingElem.IsDelete = false
// 				existingElem.Volume = volume // обновляем объем на новый
// 				existingElem.Comment = ""    // сбрасываем комментарий
// 			} else {
// 				// Элемент уже активен в корзине - можно обновить объем или оставить как есть
// 				// existingElem.Volume += volume // если хотим суммировать объемы
// 				// existingElem.Volume = volume  // если хотим заменить объем
// 			}

// 			if err := tx.Save(&existingElem).Error; err != nil {
// 				return err
// 			}
// 		}

// 		// 3. Обновляем данные mixed (дату обновления)
// 		cart.DateUpdate = time.Now()
// 		if err := tx.Save(&cart).Error; err != nil {
// 			return err
// 		}

// 		return nil
// 	})
// }

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
		newCart := ds.Mixed{
			Status:        "draft",
			DateCreate:    time.Now(),
			DateUpdate:    time.Now(),
			CreatorID:     userID,
			ModeratorID:   userID,
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

// // GetElementImagePath возвращает текущий путь к изображению элемента
// func (r *Repository) GetElementImagePath(elementID int) (string, error) {
// 	var element ds.Elements
// 	err := r.db.Select("img").Where("id = ? AND is_delete = ?", elementID, false).First(&element).Error
// 	if err != nil {
// 		return "", err
// 	}
// 	return element.Img, nil
// }
