package service

import (
	"AwsProj/internal/app/ds"
	"errors"
	"fmt"
	"time"
)

func (s *MixingService) GetUserMixing(userID uint) (*MixingResponse, error) {
	cart, cartItems, err := s.repo.GetUserCart(userID)
	if err != nil {
		return nil, err
	}

	// Бизнес-логика здесь!
	var items []MixingItem
	for _, item := range cartItems {
		items = append(items, MixingItem{
			ID:            item.Element.ID,
			Title:         item.Element.Name,
			Image:         item.Element.Img,
			PH:            item.Element.Ph,
			Concentration: item.Element.Concentration,
			Volume:        item.Volume,
		})
	}

	return &MixingResponse{
		Items:      items,
		TotalItems: len(items),
		CartID:     cart.ID,
		UserID:     userID,
	}, nil
}

func (s *MixingService) AddElementToMixing(userID uint, req *AddToMixingRequest) (*AddToMixingResponse, error) {
	// Валидация бизнес-правил
	if req.ElementID <= 0 {
		return nil, errors.New("ID элемента должен быть положительным числом")
	}

	if req.Volume < 0 {
		return nil, errors.New("объем не может быть отрицательным")
	}

	// Устанавливаем объем по умолчанию
	volume := req.Volume
	if volume == 0 {
		volume = 100.0
	}

	// Проверяем существование элемента
	_, err := s.repo.GetElementByID(req.ElementID)
	if err != nil {
		return nil, errors.New("элемент не найден")
	}

	// Добавляем элемент в корзину
	err = s.repo.AddElementToCart(userID, uint(req.ElementID), volume)
	if err != nil {
		return nil, fmt.Errorf("ошибка добавления в корзину: %v", err)
	}

	return &AddToMixingResponse{
		ElementID: req.ElementID,
		Volume:    volume,
		UserID:    userID,
		Message:   "Элемент успешно добавлен в корзину",
	}, nil
}

// GetCartIcon возвращает информацию для иконки корзины
func (s *MixingService) GetCartIcon() (*CartIconResponse, error) {
	// Используем хардкод как в вашем существующем коде
	cartID, itemsCount, err := s.repo.GetCartIconInfo(1) // userID = 1
	if err != nil {
		return nil, err
	}

	return &CartIconResponse{
		DraftOrderID: cartID,
		ItemsCount:   itemsCount,
	}, nil
}

// GetMixedList возвращает список заявок
func (s *MixingService) GetMixedList(filters MixedListRequest) ([]MixedListItem, error) {
	// Подготавливаем фильтры для репозитория
	repoFilters := make(map[string]interface{})

	if filters.Status != "" {
		repoFilters["status"] = filters.Status
	}

	// Парсим даты из строкового формата
	if filters.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", filters.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %v", err)
		}
		repoFilters["date_from"] = dateFrom
	}

	if filters.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", filters.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %v", err)
		}
		// Добавляем время конца дня для корректного фильтра
		dateTo = dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		repoFilters["date_to"] = dateTo
	}

	// Получаем данные из репозитория
	mixedData, err := s.repo.GetMixedList(repoFilters)
	if err != nil {
		return nil, err
	}

	// Преобразуем в структурированный ответ
	result := make([]MixedListItem, len(mixedData))
	for i, item := range mixedData {
		result[i] = MixedListItem{
			ID:             item["id"].(uint),
			Status:         item["status"].(string),
			DateCreate:     item["date_create"].(time.Time),
			DateUpdate:     item["date_update"].(time.Time),
			CreatorLogin:   item["creator_login"].(string),
			ModeratorLogin: item["moderator_login"].(string),
			Ph:             item["ph"].(float32),
			Concentration:  item["concentration"].(float32),
			TotalVolume:    item["total_volume"].(float64),
			AddedWater:     item["added_water"].(float64),
		}

		// Обрабатываем nullable date_finish
		if dateFinish, ok := item["date_finish"].(time.Time); ok && !dateFinish.IsZero() {
			result[i].DateFinish = dateFinish
		}
	}

	return result, nil
}

func (s *MixingService) GetMixedByID(mixedID uint) (*MixedDetailResponse, error) {
	mixed, elemMixes, err := s.repo.GetMixedByID(mixedID)
	if err != nil {
		return nil, err
	}

	// Формируем список элементов заявки
	items := make([]MixedDetailItem, len(elemMixes))
	for i, elemMix := range elemMixes {
		items[i] = MixedDetailItem{
			ElementID:     int(elemMix.Element.ID),
			Title:         elemMix.Element.Name,
			Image:         elemMix.Element.Img,
			PH:            elemMix.Element.Ph,
			Concentration: elemMix.Element.Concentration,
			Volume:        elemMix.Volume,
			Comment:       elemMix.Comment,
		}
	}

	response := &MixedDetailResponse{
		ID:             mixed.ID,
		Status:         mixed.Status,
		DateCreate:     mixed.DateCreate,
		DateUpdate:     mixed.DateUpdate,
		CreatorLogin:   mixed.Creator.Login,
		ModeratorLogin: getModeratorLogin(mixed.Moderator),
		Ph:             mixed.Ph,
		Concentration:  mixed.Concentartion,
		TotalVolume:    mixed.TotalVolume,
		AddedWater:     mixed.AddedWater,
		Items:          items,
	}

	// Обрабатываем nullable date_finish
	if mixed.DateFinish.Valid {
		response.DateFinish = &mixed.DateFinish.Time
	}

	return response, nil
}

// Вспомогательная функция для получения логина модератора
func getModeratorLogin(moderator ds.Users) string {
	if moderator.ID == 0 {
		return ""
	}
	return moderator.Login
}

func (s *MixingService) UpdateMixed(mixedID uint, req *UpdateMixedRequest) error {
	// Валидация бизнес-правил
	if req.Status != "" {
		allowedStatuses := []string{"draft", "completed", "pending", "rejected"}
		validStatus := false
		for _, status := range allowedStatuses {
			if req.Status == status {
				validStatus = true
				break
			}
		}
		if !validStatus {
			return fmt.Errorf("недопустимый статус: %s", req.Status)
		}
	}

	if req.Ph < 0 || req.Ph > 14 {
		return fmt.Errorf("pH должен быть в диапазоне от 0 до 14")
	}

	if req.Concentration < 0 {
		return fmt.Errorf("концентрация не может быть отрицательной")
	}

	if req.TotalVolume < 0 {
		return fmt.Errorf("общий объем не может быть отрицательным")
	}

	if req.AddedWater < 0 {
		return fmt.Errorf("добавленная вода не может быть отрицательной")
	}

	// Подготавливаем обновления
	updates := make(map[string]interface{})

	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Concentration != 0 {
		updates["concentartion"] = req.Concentration
	}
	if req.Ph != 0 {
		updates["ph"] = req.Ph
	}
	if req.TotalVolume != 0 {
		updates["total_volume"] = req.TotalVolume
	}
	if req.AddedWater != 0 {
		updates["added_water"] = req.AddedWater
	}

	// Если нет полей для обновления
	if len(updates) == 0 {
		return fmt.Errorf("нет полей для обновления")
	}

	// Выполняем обновление
	err := s.repo.UpdateMixed(mixedID, updates)
	if err != nil {
		return err
	}

	return nil
}

// CompleteMixedSimple упрощенная версия формирования заявки
// CompleteMixed формирует заявку создателем
func (s *MixingService) CompleteMixed(mixedID uint, req *CompleteMixedRequest) (*CompleteMixedResponse, error) {
	// 1. Проверяем обязательные поля и валидность заявки
	err := s.repo.ValidateMixedForCompletion(mixedID)
	if err != nil {
		return nil, err
	}

	// 2. Дополнительная бизнес-логика проверки
	items, err := s.repo.GetMixedItems(mixedID)
	if err != nil {
		return nil, err
	}

	// Проверяем что все элементы имеют корректные объемы
	for _, item := range items {
		if item.Volume <= 0 {
			return nil, fmt.Errorf("объем элемента '%s' должен быть положительным", item.Element.Name)
		}
	}

	// 3. Формируем заявку
	err = s.repo.CompleteMixed(mixedID)
	if err != nil {
		return nil, err
	}

	// 4. Получаем обновленную заявку для ответа
	updatedMixed, err := s.repo.GetMixedByIDBasic(mixedID)
	if err != nil {
		return nil, err
	}

	return &CompleteMixedResponse{
		MixedID:    mixedID,
		Status:     updatedMixed.Status,
		DateUpdate: updatedMixed.DateUpdate,
		Message:    "Заявка успешно сформирована",
	}, nil
}

func (s *MixingService) DeleteMixed(mixedID uint, req *DeleteMixedRequest) (*DeleteMixedResponse, error) {
	var err error
	var message string

	if req != nil && req.HardDelete {
		// Полное удаление
		err = s.repo.HardDeleteMixed(mixedID)
		message = "Заявка полностью удалена"
	} else {
		// Soft delete
		err = s.repo.DeleteMixed(mixedID)
		message = "Заявка перемещена в архив"
	}

	if err != nil {
		return nil, err
	}

	return &DeleteMixedResponse{
		MixedID:   mixedID,
		Status:    "deleted",
		DeletedAt: time.Now(),
		Message:   message,
	}, nil
}

// DeleteFromMixed удаляет элемент из заявки
func (s *MixingService) DeleteFromMixed(mixedID uint, req *DeleteFromMixedRequest) (*DeleteFromMixedResponse, error) {
	// Валидация
	if req.ElementID == 0 {
		return nil, fmt.Errorf("element_id обязателен")
	}

	var err error
	var message string

	if req.HardDelete {
		// Полное удаление
		err = s.repo.HardDeleteFromMixed(mixedID, req.ElementID)
		message = "Элемент полностью удален из заявки"
	} else {
		// Soft delete
		err = s.repo.DeleteFromMixed(mixedID, req.ElementID)
		message = "Элемент удален из заявки"
	}

	if err != nil {
		return nil, err
	}

	return &DeleteFromMixedResponse{
		MixedID:   mixedID,
		ElementID: req.ElementID,
		DeletedAt: time.Now(),
		Message:   message,
	}, nil
}
