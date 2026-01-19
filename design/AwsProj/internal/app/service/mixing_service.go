package service

import (
	"AwsProj/internal/app/ds"
	"AwsProj/internal/app/repository" // <--- Обязательно добавь этот импорт!
	"bytes"
	"database/sql" // <--- Обязательно добавь этот импорт!
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"
)

const GoToPythonSecret = "super-secret-key-8b"

type MixingService struct {
	repo            *repository.Repository
	asyncServiceURL string
}

func NewMixingService(repo *repository.Repository, asyncServiceURL string) *MixingService {
	return &MixingService{
		repo:            repo,
		asyncServiceURL: asyncServiceURL,
	}
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

	if filters.Status != "" {
		repoFilters["status"] = filters.Status
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

	if filters.Status != "" {
		repoFilters["status"] = filters.Status
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
		// Безопасное извлечение count
		itemsCount := 0
		if val, ok := item["items_count"]; ok {
			if count, ok := val.(int); ok {
				itemsCount = count
			}
		}

		processedCount := 0
		if val, ok := item["processed_count"]; ok {
			if count, ok := val.(int); ok {
				processedCount = count
			}
		}

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
			ItemsCount:     itemsCount,     // <--- Добавили
			ProcessedCount: processedCount, // Количество записей с ph > 0
		}

		if dateFinish, ok := item["date_finish"].(sql.NullTime); ok && dateFinish.Valid {
			result[i].DateFinish = dateFinish.Time
		} else if dateFinish, ok := item["date_finish"].(time.Time); ok && !dateFinish.IsZero() {
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

		Items:      items,
		ItemsCount: len(items), // <--- ВОТ ЗДЕСЬ СЧИТАЕМ КОЛИЧЕСТВО
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

	// --- РАСЧЕТ ОБЪЕМА И КОНЦЕНТРАЦИИ (pH будет рассчитан асинхронно) ---
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
		totalVolume += itemVol

		// Для расчета концентрации
		concVal := float64(item.Element.Concentration)
		mass := concVal * itemVol
		totalMass += mass
	}

	// Итоговая концентрация
	var finalConcentration float64
	if totalVolume > 0 {
		finalConcentration = totalMass / totalVolume
	}

	// Округляем до 2 знаков
	finalConcentration = math.Round(finalConcentration*100) / 100
	// pH остается 0, будет рассчитан асинхронно
	finalPh := 0.0
	// --- КОНЕЦ РАСЧЕТА ---

	// 3. Вызываем репозиторий с новыми данными (pH = 0, будет обновлен асинхронно)
	err = s.repo.CompleteMixedWithData(mixedID, finalPh, totalVolume, finalConcentration)
	if err != nil {
		return nil, err
	}

	// 4. Асинхронно вызываем Python-сервис для расчета pH (только модератор может вызвать CompleteMixed)
	go s.callAsyncService(mixedID)

	// 5. Получаем обновленные данные для ответа
	updatedMixed, err := s.repo.GetMixedByIDBasic(mixedID)
	if err != nil {
		return nil, err
	}

	return &CompleteMixedResponse{
		MixedID:    mixedID,
		Status:     updatedMixed.Status,
		DateUpdate: updatedMixed.DateUpdate,
		Message:    fmt.Sprintf("Заявка завершена. pH будет рассчитан асинхронно. V: %.2f", totalVolume),
	}, nil
}

// SubmitMixedForProcessing отправляет заявку на обработку (для пользователя)
// Устанавливает статус "pending", pH остается 0 (будет рассчитан модератором)
func (s *MixingService) SubmitMixedForProcessing(mixedID uint, addedWater float64) (*CompleteMixedResponse, error) {
	// 1. Получаем элементы заказа для расчета
	items, err := s.repo.GetMixedItems(mixedID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("заказ пуст, невозможно отправить на обработку")
	}

	// 2. Валидация объемов
	for _, item := range items {
		if item.Volume <= 0 {
			return nil, fmt.Errorf("объем элемента '%s' должен быть положительным", item.Element.Name)
		}
	}

	// --- РАСЧЕТ ОБЪЕМА И КОНЦЕНТРАЦИИ (pH будет рассчитан асинхронно) ---
	var totalVolume float64
	var totalMass float64

	// Учитываем добавленную воду
	totalVolume = addedWater

	for _, item := range items {
		itemVol := float64(item.Volume)
		totalVolume += itemVol

		concVal := float64(item.Element.Concentration)
		mass := concVal * itemVol
		totalMass += mass
	}

	// Итоговая концентрация
	var finalConcentration float64
	if totalVolume > 0 {
		finalConcentration = totalMass / totalVolume
	}
	finalConcentration = math.Round(finalConcentration*100) / 100
	// pH будет рассчитан асинхронно, поэтому передаем 0.0
	// --- КОНЕЦ РАСЧЕТА ---

	// 3. Устанавливаем статус "pending" (в работе)
	// НЕ вызываем Python-сервис - это делает только модератор при подтверждении (кнопка "Готов")
	err = s.repo.SubmitMixedForProcessing(mixedID, addedWater, totalVolume, finalConcentration)
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
		Message:    fmt.Sprintf("Заявка отправлена на обработку. pH будет рассчитан модератором. V: %.2f", totalVolume),
	}, nil
}

// callAsyncService вызывает Python-сервис для асинхронного расчета pH
func (s *MixingService) callAsyncService(mixedID uint) {

	if s.asyncServiceURL == "" {
		return // Если URL не настроен, пропускаем вызов
	}

	// Получаем элементы заявки для передачи в Python
	items, err := s.repo.GetMixedItems(mixedID)
	if err != nil {
		fmt.Printf("Error getting mixed items for async service: %v\n", err)
		return
	}

	// Получаем информацию о добавленной воде
	currentMixed, err := s.repo.GetMixedByIDBasic(mixedID)
	if err != nil {
		fmt.Printf("Error getting mixed info for async service: %v\n", err)
		return
	}

	// Формируем данные элементов для расчета pH
	type ElementData struct {
		Ph     float32 `json:"ph"`
		Volume float32 `json:"volume"`
	}

	elementsData := make([]ElementData, len(items))
	for i, item := range items {
		elementsData[i] = ElementData{
			Ph:     item.Element.Ph,
			Volume: item.Volume,
		}
	}

	url := fmt.Sprintf("%s/process", s.asyncServiceURL)
	payload := map[string]interface{}{
		"mixed_id":    mixedID,
		"added_water": currentMixed.AddedWater,
		"elements":    elementsData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Secret-Key", GoToPythonSecret)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error calling async service: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("Successfully called async service for mixed %d\n", mixedID)
	} else {
		fmt.Printf("Async service returned status %d for mixed %d\n", resp.StatusCode, mixedID)
	}
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

// НОВЫЙ МЕТОД DeleteFromMixed (Исправленный)
func (s *MixingService) DeleteFromMixed(mixedID uint, userID uint, userRole string, req *DeleteFromMixedRequest) (*DeleteMixedResponse, error) {
	if req.ElementID == 0 {
		return nil, errors.New("element_id обязателен")
	}

	// 1. Получаем черновик.
	// ВАЖНО: Ваш репозиторий возвращает 3 значения: (mixed, items, err)
	// Нам нужен только mixed для проверки владельца.
	mixed, _, err := s.repo.GetMixedByID(mixedID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения черновика: %v", err)
	}

	// 2. Проверка прав
	if userRole != "admin" && mixed.CreatorID != userID {
		return nil, errors.New("forbidden: у вас нет прав на изменение этого черновика")
	}

	// 3. Удаление
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

	// 4. Возврат ответа (БЕЗ поля ElementID, так как его нет в вашей структуре)
	return &DeleteMixedResponse{
		MixedID:   mixedID,
		Status:    "updated", // Или mixed.Status
		DeletedAt: time.Now(),
		Message:   message,
	}, nil
}

// service.go

// Интерфейс для Service (если вы используете интерфейсы)
// type MixedsService interface {
//     UpdatePH(ctx context.Context, id int, ph float64) error
// }

func (s *MixingService) UpdateProcessingResult(mixedID uint, result float64) error {
	// Можно добавить бизнес-логику, если нужно (например, валидацию значения pH)
	return s.repo.UpdatePH(mixedID, result)
}
