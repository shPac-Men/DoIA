package service

import (
	"AwsProj/internal/app/ds"
	"AwsProj/internal/app/repository"
	"errors"
	"fmt"
	"math"
	"time"
)

type MixingService struct {
	repo *repository.Repository
}

func NewMixingService(repo *repository.Repository) *MixingService {
	return &MixingService{repo: repo}
}

// GetUserMixing получает корзину пользователя
func (s *MixingService) GetUserMixing(userID uint) (*MixingResponse, error) {
	cart, cartItems, err := s.repo.GetUserCart(userID)
	if err != nil {
		return nil, err
	}

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

// AddElementToMixing добавляет элемент в корзину
func (s *MixingService) AddElementToMixing(userID uint, req *AddToMixingRequest) (*AddToMixingResponse, error) {
	if req.ElementID <= 0 {
		return nil, errors.New("ID элемента должен быть положительным числом")
	}

	if req.Volume < 0 {
		return nil, errors.New("объем не может быть отрицательным")
	}

	volume := req.Volume
	if volume == 0 {
		volume = 100.0
	}

	_, err := s.repo.GetElementByID(req.ElementID)
	if err != nil {
		return nil, errors.New("элемент не найден")
	}

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
func (s *MixingService) GetCartIcon(userID uint) (*CartIconResponse, error) {
	cartID, itemsCount, err := s.repo.GetCartIconInfo(int(userID))
	if err != nil {
		return nil, err
	}

	return &CartIconResponse{
		DraftOrderID: cartID,
		ItemsCount:   itemsCount,
	}, nil
}

// GetMixedList возвращает список всех заявок
func (s *MixingService) GetMixedList(filters MixedListRequest) ([]MixedListItem, error) {
	repoFilters := make(map[string]interface{})

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
		dateTo = dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		repoFilters["date_to"] = dateTo
	}

	mixedData, err := s.repo.GetMixedList(repoFilters)
	if err != nil {
		return nil, err
	}

	return s.convertToMixedListItems(mixedData), nil
}

// GetMixedListByUser возвращает заявки конкретного пользователя
func (s *MixingService) GetMixedListByUser(userID uint, filters MixedListRequest) ([]MixedListItem, error) {
	repoFilters := make(map[string]interface{})
	repoFilters["creator_id"] = userID

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
		dateTo = dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		repoFilters["date_to"] = dateTo
	}

	mixedData, err := s.repo.GetMixedList(repoFilters)
	if err != nil {
		return nil, err
	}

	return s.convertToMixedListItems(mixedData), nil
}

// convertToMixedListItems конвертирует данные из репозитория
func (s *MixingService) convertToMixedListItems(mixedData []map[string]interface{}) []MixedListItem {
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

		if dateFinish, ok := item["date_finish"].(time.Time); ok && !dateFinish.IsZero() {
			result[i].DateFinish = dateFinish
		}
	}
	return result
}

// GetMixedByID получает детали заявки
func (s *MixingService) GetMixedByID(mixedID uint) (*MixedDetailResponse, error) {
	mixed, elemMixes, err := s.repo.GetMixedByID(mixedID)
	if err != nil {
		return nil, err
	}

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

	if mixed.DateFinish.Valid {
		response.DateFinish = &mixed.DateFinish.Time
	}

	return response, nil
}

func getModeratorLogin(moderator ds.Users) string {
	if moderator.ID == 0 {
		return ""
	}
	return moderator.Login
}

// UpdateMixed обновляет заявку
func (s *MixingService) UpdateMixed(mixedID uint, req *UpdateMixedRequest) error {
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

	if req.Ph != 0 && (req.Ph < 0 || req.Ph > 14) {
		return fmt.Errorf("pH должен быть в диапазоне от 0 до 14")
	}

	if req.Concentration != 0 && req.Concentration < 0 {
		return fmt.Errorf("концентрация не может быть отрицательной")
	}

	if req.TotalVolume != 0 && req.TotalVolume < 0 {
		return fmt.Errorf("общий объем не может быть отрицательным")
	}

	if req.AddedWater != 0 && req.AddedWater < 0 {
		return fmt.Errorf("добавленная вода не может быть отрицательной")
	}

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

	if len(updates) == 0 {
		return fmt.Errorf("нет полей для обновления")
	}

	return s.repo.UpdateMixed(mixedID, updates)
}

// CompleteMixed формирует заявку
func (s *MixingService) CompleteMixed(mixedID uint, req *CompleteMixedRequest) (*CompleteMixedResponse, error) {
	// 1. Получаем элементы заказа для расчета
	items, err := s.repo.GetMixedItems(mixedID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("заказ пуст, невозможно завершить")
	}

	// 2. Валидация объемов
	for _, item := range items {
		if item.Volume <= 0 {
			return nil, fmt.Errorf("объем элемента '%s' должен быть положительным", item.Element.Name)
		}
	}

	// --- НАЧАЛО РАСЧЕТА ---
	var totalProtons float64
	var totalVolume float64
	var totalMass float64

	// Получаем текущее значение добавленной воды (оно уже есть в заказе)
	currentMixed, err := s.repo.GetMixedByIDBasic(mixedID)
	if err != nil {
		return nil, err
	}

	// Учитываем добавленную воду в общем объеме
	totalVolume = currentMixed.AddedWater

	for _, item := range items {
		// Приводим item.Volume к float64
		itemVol := float64(item.Volume)

		// Исправляем: totalVolume += float64(item.Volume)
		totalVolume += itemVol

		// Для расчета pH
		// item.Element.Ph скорее всего float32, приводим к float64 для math.Pow
		phVal := float64(item.Element.Ph)
		hConcentration := math.Pow(10, -phVal)

		molesH := hConcentration * itemVol
		totalProtons += molesH

		// Для расчета концентрации
		// item.Element.Concentration скорее всего float32
		concVal := float64(item.Element.Concentration)
		mass := concVal * itemVol
		totalMass += mass
	}

	// Итоговый pH
	var finalPh float64
	if totalVolume > 0 {
		finalHConcentration := totalProtons / totalVolume
		if finalHConcentration > 0 {
			finalPh = -math.Log10(finalHConcentration)
		} else {
			finalPh = 7.0 // Нейтральная среда, если нет протонов
		}
	} else {
		finalPh = 7.0
	}

	// Итоговая концентрация
	var finalConcentration float64
	if totalVolume > 0 {
		finalConcentration = totalMass / totalVolume
	}

	// Округляем до 2 знаков
	finalPh = math.Round(finalPh*100) / 100
	finalConcentration = math.Round(finalConcentration*100) / 100
	// --- КОНЕЦ РАСЧЕТА ---

	// 3. Вызываем репозиторий с новыми данными
	err = s.repo.CompleteMixedWithData(mixedID, finalPh, totalVolume, finalConcentration) // <-- Заменил finalVolume на totalVolume
	if err != nil {
		return nil, err
	}

	// 4. Получаем обновленные данные для ответа
	updatedMixed, err := s.repo.GetMixedByIDBasic(mixedID)
	if err != nil {
		return nil, err
	}

	return &CompleteMixedResponse{
		MixedID:    mixedID,
		Status:     updatedMixed.Status,
		DateUpdate: updatedMixed.DateUpdate,
		Message:    fmt.Sprintf("Заявка завершена. pH: %.2f, V: %.2f", finalPh, totalVolume),
	}, nil
}

// DeleteMixed удаляет заявку
func (s *MixingService) DeleteMixed(mixedID uint, req *DeleteMixedRequest) (*DeleteMixedResponse, error) {
	var err error
	var message string

	if req != nil && req.HardDelete {
		err = s.repo.HardDeleteMixed(mixedID)
		message = "Заявка полностью удалена"
	} else {
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
	if req.ElementID == 0 {
		return nil, fmt.Errorf("element_id обязателен")
	}

	var err error
	var message string

	if req.HardDelete {
		err = s.repo.HardDeleteFromMixed(mixedID, req.ElementID)
		message = "Элемент полностью удален из заявки"
	} else {
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
