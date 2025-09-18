package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

// type Order struct { // вот наша новая структура
// 	ID    int    // поля структур, которые передаются в шаблон
// 	Title string // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
// }

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
			Image:         "/static/img/NaOH.jpg",
			Concentration: "1M",
			PH:            "14",
		},
		{
			ID:            2,
			Title:         "HCl",
			Image:         "/static/img/Hcl.jpg",
			Concentration: "0.5M",
			PH:            "1",
		},
		{
			ID:            3,
			Title:         "H2SO4",
			Image:         "/static/img/64352489.jpg",
			Concentration: "2M",
			PH:            "0",
		},
		{
			ID:            4,
			Title:         "NaCl",
			Image:         "/static/img/nacl.jpg",
			Concentration: "0.15M",
			PH:            "7",
		},
		{
			ID:            5,
			Title:         "NH3",
			Image:         "/static/img/nh3.jpg",
			Concentration: "1M",
			PH:            "11",
		},
	}

	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
	if len(orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return orders, nil
}

func (r *Repository) GetOrder(id int) (Order, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
	orders, err := r.GetOrders()
	if err != nil {
		return Order{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, order := range orders {
		if order.ID == id {
			return order, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return Order{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
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
