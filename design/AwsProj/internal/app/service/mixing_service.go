package service

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
