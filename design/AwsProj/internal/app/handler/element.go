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

// func (h *Handler) GetAllElements(ctx *gin.Context) {
// 	var elements []ds.Elements
// 	var err error

// 	search := ctx.Query("search")
// 	if search == "" {
// 		elements, err = h.Repository.GetAllElements()
// 	} else {
// 		elements, err = h.Repository.SearchElementByName(search)
// 	}

// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		logrus.Error(err)
// 		return
// 	}

// 	ctx.HTML(http.StatusOK, "index.html", gin.H{
// 		"data":       elements,
// 		"cart_count": h.Repository.GetCartCount(),
// 		"search":     search,
// 	})
// }

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
