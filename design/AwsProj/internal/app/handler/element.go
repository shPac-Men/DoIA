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

func (h *Handler) GetAllElements(ctx *gin.Context) {
	var elements []ds.Elements
	var err error

	search := ctx.Query("query") // получаем "query" из URL
	if search == "" {
		elements, err = h.Repository.GetAllElements()
	} else {
		elements, err = h.Repository.SearchElementByName(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		logrus.Error(err)
		return
	}

	var viewElements []ElementView
	for _, elem := range elements {
		viewElements = append(viewElements, ElementView{
			ID:            elem.ID,
			Image:         elem.Img,
			Title:         elem.Name,
			Concentration: formatConcentration(elem.Concentration),
			PH:            formatPH(elem.Ph),
		})
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"data":       viewElements,
		"cart_count": h.Repository.GetCartCount(),
		"query":      search, // ← ИСПРАВИТЬ: должно быть "query" (как в шаблоне)
	})
}

// Вспомогательные функции для форматирования
func formatConcentration(conc float32) string {
	return fmt.Sprintf("%.2fM", conc)
}

func formatPH(ph float32) string {
	return fmt.Sprintf("%.1f", ph)
}

func (h *Handler) GetElementById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	element, err := h.Repository.GetElementByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "element.html", element)
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
	// Получаем ID элемента из формы
	userID := uint(1)
	elementIDStr := ctx.PostForm("element_id")
	elementID, err := strconv.Atoi(elementIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid element ID"})
		return
	}

	// Получаем пользователя (предположим, что у вас есть аутентификация)
	//userID, exists := ctx.Get("userID")
	// if !exists {
	// 	ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
	// 	return
	// }

	// Объём по умолчанию (можно сделать настраиваемым)
	volume := float32(100.0) // 100 мл по умолчанию

	// Добавляем элемент в корзину
	err = h.Repository.AddElementToCart(uint(userID), uint(elementID), volume)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Перенаправляем обратно на страницу химии или показываем сообщение об успехе
	ctx.Redirect(http.StatusFound, "/chemistry")
}

func (h *Handler) GetMixingPage(ctx *gin.Context) {
	// Захардкоженный пользователь
	userID := uint(1)

	// Получаем корзину и элементы
	cart, cartItems, err := h.Repository.GetUserCart(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Если корзины нет или она пустая
	if cart == nil || len(cartItems) == 0 {
		ctx.HTML(http.StatusOK, "calculatepage.html", gin.H{
			"data": []ds.Elements{},
		})
		return
	}

	// Преобразуем элементы корзины в формат для шаблона
	var elements []gin.H
	for _, item := range cartItems {
		elements = append(elements, gin.H{
			"ID":            item.Element.ID,
			"Title":         item.Element.Name,
			"Image":         item.Element.Img,
			"PH":            item.Element.Ph,
			"Concentration": item.Element.Concentration,
			"Volume":        item.Volume, // Добавляем объем из корзины
		})
	}

	ctx.HTML(http.StatusOK, "calculatepage.html", gin.H{
		"data": elements,
	})
}

func (h *Handler) CreateMixing(ctx *gin.Context) {
	userID := uint(1)

	// Получаем объем добавленной воды из формы
	addedWaterStr := ctx.PostForm("added_water")
	addedWater, err := strconv.ParseFloat(addedWaterStr, 64)
	if err != nil {
		addedWater = 100.0
	}

	// Вызываем метод с двумя аргументами и получаем два значения
	calculatedPH, err := h.Repository.CompleteCartAndCreateNew(userID, addedWater)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Показываем результат
	ctx.HTML(http.StatusOK, "calculatepage.html", gin.H{
		"data":         []ds.Elements{}, // пустая корзина
		"calculatedPH": calculatedPH,
		"added_water":  addedWater,
		"showResult":   true,
		"message":      "Заявка успешно сформирована!",
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
