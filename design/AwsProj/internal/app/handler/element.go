package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"AwsProj/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ElementView struct {
	ID            int
	Image         string
	Title         string
	Concentration string // строка для отображения
	PH            string // строка для отображения
}

type ElementResponse struct {
	ID            int    `json:"id"`
	Image         string `json:"image"`
	Title         string `json:"title"`
	Concentration string `json:"concentration"`
	PH            string `json:"ph"`
}

func (h *Handler) GetAllElements(ctx *gin.Context) {
	var elements []ds.Elements
	var err error

	search := ctx.Query("query")
	if search == "" {
		elements, err = h.Repository.GetAllElements()
	} else {
		elements, err = h.Repository.SearchElementByName(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		logrus.Error(err)
		return
	}

	// Преобразуем в JSON-ответ
	//сериализация данных из бд в джсон
	var response []ElementResponse
	for _, elem := range elements {
		response = append(response, ElementResponse{
			ID:            elem.ID,
			Image:         elem.Img,
			Title:         elem.Name,
			Concentration: formatConcentration(elem.Concentration),
			PH:            formatPH(elem.Ph),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"meta": gin.H{
			"total":    len(response),
			"query":    search,
			"has_more": false, // можно добавить пагинацию позже
		},
	})
}

// Вспомогательные функции для форматирования
func formatConcentration(conc float32) string {
	return fmt.Sprintf("%.2fM", conc)
}

func formatPH(ph float32) string {
	return fmt.Sprintf("%.1f", ph)
}

// func (h *Handler) GetElementById(ctx *gin.Context) {
// 	strId := ctx.Param("id")
// 	id, err := strconv.Atoi(strId)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		logrus.Error(err)
// 		return
// 	}

// 	element, err := h.Repository.GetElementByID(id)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		logrus.Error(err)
// 		return
// 	}

// 	ctx.HTML(http.StatusOK, "element.html", element)
// }

func (h *Handler) GetElementById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid ID format",
			"message": "ID must be a number",
		})
		logrus.Error("Invalid ID format:", err)
		return
	}

	element, err := h.Repository.GetElementByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Element not found",
			"message": err.Error(),
		})
		logrus.Error("Element not found:", err)
		return
	}

	// Сериализация в JSON-ответ
	response := ElementResponse{
		ID:            element.ID,
		Image:         element.Img,
		Title:         element.Name,
		Concentration: formatConcentration(element.Concentration),
		PH:            formatPH(element.Ph),
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

func (h *Handler) DeleteElement(ctx *gin.Context) {
	// считываем значение из формы, которую мы добавим в наш шаблон
	strId := ctx.PostForm("element_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	// Вызов функции добавления чата в заявку
	err = h.Repository.DeleteElement(uint(id))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return
	}

	// после вызова сразу произойдет обновление страницы
	ctx.Redirect(http.StatusFound, "/chemistry")
}

// deepseek
func (h *Handler) AddToMixing(ctx *gin.Context) {
	// 1. Получаем ID элемента из JSON тела запроса
	var request struct {
		ElementID int     `json:"element_id" binding:"required"`
		Volume    float32 `json:"volume,omitempty"` // опционально
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// 2. Получаем ID пользователя (пока хардкод, потом через аутентификацию)
	userID := uint(1)

	// 3. Объём по умолчанию
	volume := request.Volume
	if volume == 0 {
		volume = 100.0 // значение по умолчанию
	}

	// 4. Добавляем элемент в корзину
	err := h.Repository.AddElementToCart(userID, uint(request.ElementID), volume)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to add element to mixing",
			"message": err.Error(),
		})
		return
	}

	// 5. Возвращаем JSON ответ вместо редиректа
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Element added to mixing successfully",
		"data": gin.H{
			"element_id": request.ElementID,
			"volume":     volume,
			"user_id":    userID,
		},
	})
}

func (h *Handler) GetMixingPage(ctx *gin.Context) {
	// Захардкоженный пользователь
	userID := uint(1)

	// Получаем корзину и элементы
	cart, cartItems, err := h.Repository.GetUserCart(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get mixing data",
			"message": err.Error(),
		})
		return
	}

	// Если корзины нет или она пустая
	if cart == nil || len(cartItems) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []gin.H{},
			"meta": gin.H{
				"total_items": 0,
				"cart_id":     0, // корзины нет, поэтому ID = 0
				"user_id":     userID,
				"message":     "Mixing cart is empty",
			},
		})
		return
	}

	// Преобразуем элементы корзины в JSON-формат
	var elements []gin.H
	for _, item := range cartItems {
		elements = append(elements, gin.H{
			"id":            item.Element.ID,
			"title":         item.Element.Name,
			"image":         item.Element.Img,
			"ph":            item.Element.Ph,
			"concentration": item.Element.Concentration,
			"volume":        item.Volume, // Добавляем объем из корзины
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    elements,
		"meta": gin.H{
			"total_items": len(elements),
			"cart_id":     cart.ID, // теперь cart не nil, можно безопасно использовать
			"user_id":     userID,
		},
	})
}

// func (h *Handler) CreateMixing(ctx *gin.Context) {
// 	userID := uint(1)

// 	// Получаем объем добавленной воды из формы
// 	addedWaterStr := ctx.PostForm("added_water")
// 	addedWater, err := strconv.ParseFloat(addedWaterStr, 64)
// 	if err != nil {
// 		addedWater = 100.0
// 	}

// 	// Вызываем метод с двумя аргументами и получаем два значения
// 	calculatedPH, err := h.Repository.CompleteCartAndCreateNew(userID, addedWater)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Показываем результат
// 	ctx.HTML(http.StatusOK, "calculatepage.html", gin.H{
// 		"data":         []ds.Elements{}, // пустая корзина
// 		"calculatedPH": calculatedPH,
// 		"added_water":  addedWater,
// 		"showResult":   true,
// 		"message":      "Заявка успешно сформирована!",
// 	})
// }

func (h *Handler) CreateMixing(ctx *gin.Context) {
	// 1. Получаем данные из JSON тела запроса
	var request struct {
		AddedWater float64 `json:"added_water" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		// Если JSON невалидный, используем значение по умолчанию
		request.AddedWater = 100.0
	}

	userID := uint(1)

	// 2. Вызываем метод репозитория
	calculatedPH, err := h.Repository.CompleteCartAndCreateNew(userID, request.AddedWater)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create mixing",
			"message": err.Error(),
		})
		return
	}

	// 3. Возвращаем JSON результат вместо HTML
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mixing successfully created",
		"data": gin.H{
			"calculated_ph": calculatedPH,
			"added_water":   request.AddedWater,
			"user_id":       userID,
		},
	})
}

func (h *Handler) RemoveFromMixing(ctx *gin.Context) {
	// Получаем ID элемента из формы
	elementIDStr := ctx.PostForm("element_id")
	elementID, err := strconv.Atoi(elementIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid element ID"})
		return
	}

	// Захардкоженный пользователь
	userID := uint(1)

	// Удаляем элемент из корзины
	err = h.Repository.RemoveFromCart(userID, uint(elementID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Перенаправляем обратно на страницу корзины
	ctx.Redirect(http.StatusFound, "/mixingpage")
}
