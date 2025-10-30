// package repository

// import (
// 	"fmt"
// 	"strings"
// )

// type Repository struct {
// }

// func NewRepository() (*Repository, error) {
// 	return &Repository{}, nil
// }

// type Order struct {
// 	ID            int
// 	Title         string
// 	Image         string
// 	Concentration string
// 	PH            string
// }

// func (r *Repository) GetOrders() ([]Order, error) {
// 	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
// 	orders := []Order{
// 		{
// 			ID:            1,
// 			Title:         "NaOH",
// 			Image:         "http://localhost:9000/staticimages/NaOH.jpg",
// 			Concentration: "1M",
// 			PH:            "14",
// 		},
// 		{
// 			ID:            2,
// 			Title:         "HCl",
// 			Image:         "http://localhost:9000/staticimages/Hcl.jpg",
// 			Concentration: "0.5M",
// 			PH:            "1",
// 		},
// 		{
// 			ID:            3,
// 			Title:         "H2SO4",
// 			Image:         "http://localhost:9000/staticimages/64352489.jpg",
// 			Concentration: "2M",
// 			PH:            "0",
// 		},
// 		{
// 			ID:            4,
// 			Title:         "NaCl",
// 			Image:         "http://localhost:9000/staticimages/nacl.jpg",
// 			Concentration: "0.15M",
// 			PH:            "7",
// 		},
// 		{
// 			ID:            5,
// 			Title:         "NH3",
// 			Image:         "http://localhost:9000/staticimages/nh3.jpg",
// 			Concentration: "1M",
// 			PH:            "11",
// 		},
// 	}

// 	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
// 	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
// 	if len(orders) == 0 {
// 		return nil, fmt.Errorf("массив пустой")
// 	}

// 	return orders, nil
// }

// // GetCartItems - отдельный метод для получения элементов корзины (только 2 элемента)
// func (r *Repository) GetCartItems() ([]Order, error) {
// 	// Хардкодим только 2 элемента для корзины
// 	cartItems := []Order{
// 		{
// 			ID:            1,
// 			Title:         "NaOH",
// 			Image:         "http://localhost:9000/staticimages/NaOH.jpg",
// 			Concentration: "1M",
// 			PH:            "14",
// 		},
// 		{
// 			ID:            2,
// 			Title:         "HCl",
// 			Image:         "http://localhost:9000/staticimages/Hcl.jpg",
// 			Concentration: "0.5M",
// 			PH:            "1",
// 		},
// 	}

// 	if len(cartItems) == 0 {
// 		return nil, fmt.Errorf("корзина пустая")
// 	}

// 	return cartItems, nil
// }

// func (r *Repository) GetOrder(id int) (Order, error) {
// 	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
// 	orders, err := r.GetOrders()
// 	if err != nil {
// 		return Order{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
// 	}

// 	for _, order := range orders {
// 		if order.ID == id {
// 			return order, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
// 		}
// 	}
// 	return Order{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
// }

// func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
// 	orders, err := r.GetOrders()
// 	if err != nil {
// 		return []Order{}, err
// 	}

// 	var result []Order
// 	for _, order := range orders {
// 		if strings.Contains(strings.ToLower(order.Title), strings.ToLower(title)) {
// 			result = append(result, order)
// 		}
// 	}

// 	return result, nil
// }

package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
	// Добавляем словарь для корзины: ключ - ID корзины, значение - массив ID элементов
	carts map[int][]int
}

func NewRepository() (*Repository, error) {
	return &Repository{
		carts: map[int][]int{
			1: {1, 2}, // Корзина с ID 1 содержит элементы с ID 1 и 2
		},
	}, nil
}

type Order struct {
	ID            int
	Title         string
	Image         string
	Concentration string
	PH            string
}

func (r *Repository) GetOrders() ([]Order, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	orders := []Order{
		{
			ID:            1,
			Title:         "NaOH",
			Image:         "http://localhost:9000/staticimages/NaOH.jpg",
			Concentration: "1M",
			PH:            "14",
		},
		{
			ID:            2,
			Title:         "HCl",
			Image:         "http://localhost:9000/staticimages/Hcl.jpg",
			Concentration: "0.5M",
			PH:            "1",
		},
		{
			ID:            3,
			Title:         "H2SO4",
			Image:         "http://localhost:9000/staticimages/64352489.jpg",
			Concentration: "2M",
			PH:            "0",
		},
		{
			ID:            4,
			Title:         "NaCl",
			Image:         "http://localhost:9000/staticimages/nacl.jpg",
			Concentration: "0.15M",
			PH:            "7",
		},
		{
			ID:            5,
			Title:         "NH3",
			Image:         "http://localhost:9000/staticimages/nh3.jpg",
			Concentration: "1M",
			PH:            "11",
		},
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return orders, nil
}

// GetCartItems - отдельный метод для получения элементов корзины (только 2 элемента)
func (r *Repository) GetCartItems() ([]Order, error) {
	// Получаем ID элементов из корзины
	cartID := 1 // используем корзину с ID 1
	elementIDs, exists := r.carts[cartID]
	if !exists || len(elementIDs) == 0 {
		return nil, fmt.Errorf("корзина пустая")
	}

	// Получаем все элементы
	allOrders, err := r.GetOrders()
	if err != nil {
		return nil, err
	}

	// Фильтруем только те элементы, которые есть в корзине
	var cartItems []Order
	for _, order := range allOrders {
		for _, elementID := range elementIDs {
			if order.ID == elementID {
				cartItems = append(cartItems, order)
				break
			}
		}
	}

	if len(cartItems) == 0 {
		return nil, fmt.Errorf("корзина пустая")
	}

	return cartItems, nil
}

func (r *Repository) GetOrder(id int) (Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return Order{}, err
	}

	for _, order := range orders {
		if order.ID == id {
			return order, nil
		}
	}
	return Order{}, fmt.Errorf("заказ не найден")
}

func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return []Order{}, err
	}

	var result []Order
	for _, order := range orders {
		if strings.Contains(strings.ToLower(order.Title), strings.ToLower(title)) {
			result = append(result, order)
		}
	}

	return result, nil
}
