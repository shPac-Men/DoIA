package service

import (
	"AwsProj/internal/app/ds"
	"AwsProj/internal/app/repository"
	"errors"
	"fmt"
)

type ElementService struct {
	repo *repository.Repository
}

func NewElementService(repo *repository.Repository) *ElementService {
	return &ElementService{repo: repo}
}

func (s *ElementService) GetAllElements(search string) (*ElementsResponse, error) {
	var elements []ds.Elements
	var err error

	if search == "" {
		elements, err = s.repo.GetAllElements()
	} else {
		elements, err = s.repo.SearchElementByName(search)
	}

	if err != nil {
		return nil, err
	}

	var items []ElementItem
	for _, elem := range elements {
		items = append(items, ElementItem{
			ID:            elem.ID,
			Image:         elem.Img,
			Title:         elem.Name,
			Concentration: s.formatConcentration(elem.Concentration),
			PH:            s.formatPH(elem.Ph),
		})
	}

	return &ElementsResponse{
		Items:   items,
		Total:   len(items),
		Query:   search,
		HasMore: false,
	}, nil
}

func (s *ElementService) CreateElement(req *CreateElementRequest) (*CreateElementResponse, error) {
	// Валидация бизнес-правил
	if s == nil {
		return nil, errors.New("ElementService is nil")
	}
	if s.repo == nil {
		return nil, errors.New("repository is nil in ElementService")
	}

	if req.Name == "" {
		return nil, errors.New("название элемента обязательно")
	}

	if req.Ph < 0 || req.Ph > 14 {
		return nil, errors.New("pH должен быть в диапазоне 0-14")
	}

	if req.Concentration < 0 {
		return nil, errors.New("концентрация не может быть отрицательной")
	}

	// Проверка на уникальность названия
	exists, err := s.repo.CheckElementExists(req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("элемент с таким названием уже существует")
	}

	// Создаем элемент
	element := &ds.Elements{
		Name:          req.Name,
		Description:   req.Description,
		Ph:            req.Ph,
		Concentration: req.Concentration,
		Img:           "", // изображение будет добавлено отдельно
		IsDelete:      false,
	}

	// Сохраняем в БД
	err = s.repo.CreateElement(element)
	if err != nil {
		return nil, err
	}

	// Возвращаем ответ
	return &CreateElementResponse{
		ID:            element.ID,
		Name:          element.Name,
		Description:   element.Description,
		Ph:            element.Ph,
		Concentration: element.Concentration,
		Message:       "Элемент успешно создан",
	}, nil
}

func (s *ElementService) formatConcentration(conc float32) string {
	return fmt.Sprintf("%.2f%%", conc)
}

func (s *ElementService) formatPH(ph float32) string {
	return fmt.Sprintf("%.1f", ph)
}

func (s *ElementService) UpdateElement(elementID int, req *UpdateElementRequest) (*ElementResponse, error) {
	// Проверяем существование элемента
	existingElement, err := s.repo.GetElementByID(elementID) // ← теперь int
	if err != nil {
		return nil, errors.New("элемент не найден")
	}

	// Валидация бизнес-правил
	if req.Name != nil {
		if *req.Name == "" {
			return nil, errors.New("название элемента не может быть пустым")
		}
		// Проверка уникальности названия (кроме текущего элемента)
		exists, err := s.repo.CheckElementExists(*req.Name)
		if err != nil {
			return nil, err
		}
		if exists && *req.Name != existingElement.Name {
			return nil, errors.New("элемент с таким названием уже существует")
		}
	}

	if req.Ph != nil && (*req.Ph < 0 || *req.Ph > 14) {
		return nil, errors.New("pH должен быть в диапазоне 0-14")
	}

	if req.Concentration != nil && *req.Concentration < 0 {
		return nil, errors.New("концентрация не может быть отрицательной")
	}

	// Подготавливаем обновления
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Ph != nil {
		updates["ph"] = *req.Ph
	}
	if req.Concentration != nil {
		updates["concentration"] = *req.Concentration
	}

	// Если нет полей для обновления
	if len(updates) == 0 {
		return nil, errors.New("нет данных для обновления")
	}

	// Обновляем элемент в БД
	err = s.repo.UpdateElement(elementID, updates)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления элемента: %v", err)
	}

	// Получаем обновленный элемент
	updatedElement, err := s.repo.GetElementByID(elementID)
	if err != nil {
		return nil, err
	}

	// Возвращаем ответ
	return &ElementResponse{
		ID:            updatedElement.ID,
		Name:          updatedElement.Name,
		Description:   updatedElement.Description,
		Ph:            updatedElement.Ph,
		Concentration: updatedElement.Concentration,
		Image:         updatedElement.Img,
		Message:       "Элемент успешно обновлен",
	}, nil
}
