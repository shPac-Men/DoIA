package service

import (
	"errors"
	"fmt"
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
