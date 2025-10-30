package handler

import (
	"AwsProj/internal/app/service"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AddToMixing(ctx *gin.Context) {
	var req service.AddToMixingRequest

	// Валидация входных данных
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": err.Error(),
		})
		return
	}

	// Получаем ID пользователя (пока хардкод)
	userID := uint(1)

	// Вызов сервиса
	result, err := h.MixingService.AddElementToMixing(userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "не найден") ||
			strings.Contains(err.Error(), "положительным") ||
			strings.Contains(err.Error(), "отрицательным") {
			statusCode = http.StatusBadRequest
		}

		ctx.JSON(statusCode, gin.H{
			"error":   "Failed to add to mixing",
			"message": err.Error(),
		})
		return
	}

	// Успешный ответ
	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) GetCartIcon(ctx *gin.Context) {
	response, err := h.MixingService.GetCartIcon()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get cart icon",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"cart_id":     response.DraftOrderID,
		"total_items": response.ItemsCount,
	})
}

func (h *Handler) GetMixedList(ctx *gin.Context) {
	var filters service.MixedListRequest

	// Парсим query параметры
	if err := ctx.ShouldBindQuery(&filters); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"message": err.Error(),
		})
		return
	}

	// Получаем список заявок
	mixedList, err := h.MixingService.GetMixedList(filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get mixed list",
			"message": err.Error(),
		})
		return
	}

	// Возвращаем данные с total, но без success и data обертки
	ctx.JSON(http.StatusOK, gin.H{
		"items": mixedList,
		"total": len(mixedList),
	})
}

func (h *Handler) GetMixedByID(ctx *gin.Context) {
	// Получаем ID из параметров пути
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid mixed ID",
			"message": "ID must be a positive integer",
		})
		return
	}

	// Получаем детальную информацию о заявке
	mixedDetail, err := h.MixingService.GetMixedByID(uint(mixedID))
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Mixed not found",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get mixed details",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, mixedDetail)
}

func (h *Handler) UpdateMixed(ctx *gin.Context) {
	// Получаем ID из параметров пути
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid mixed ID",
			"message": "ID must be a positive integer",
		})
		return
	}

	// Парсим тело запроса
	var updateReq service.UpdateMixedRequest
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Выполняем обновление
	err = h.MixingService.UpdateMixed(uint(mixedID), &updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Mixed not found",
				"message": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "недопустимый") ||
			strings.Contains(err.Error(), "не может быть") ||
			strings.Contains(err.Error(), "диапазон") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update mixed",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "Заявка успешно обновлена",
		"mixed_id": mixedID,
	})
}

func (h *Handler) CompleteMixed(ctx *gin.Context) {
	// Получаем ID из параметров пути
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid mixed ID",
			"message": "ID must be a positive integer",
		})
		return
	}

	// Пустой запрос, так как все данные берутся из существующей заявки
	var completeReq service.CompleteMixedRequest
	if err := ctx.ShouldBindJSON(&completeReq); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Выполняем формирование заявки
	response, err := h.MixingService.CompleteMixed(uint(mixedID), &completeReq)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Mixed not found",
				"message": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "не может быть пустой") ||
			strings.Contains(err.Error(), "должен быть положительным") {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":   "Validation failed",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to complete mixed",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteMixed(ctx *gin.Context) {
	// Получаем ID из параметров пути
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid mixed ID",
			"message": "ID must be a positive integer",
		})
		return
	}

	// Парсим тело запроса (опционально)
	var deleteReq service.DeleteMixedRequest
	if err := ctx.ShouldBindJSON(&deleteReq); err != nil && err != io.EOF {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Выполняем удаление
	response, err := h.MixingService.DeleteMixed(uint(mixedID), &deleteReq)
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Mixed not found",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete mixed",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteFromMixed(ctx *gin.Context) {
	// Получаем ID заявки из параметров пути
	idStr := ctx.Param("id")
	mixedID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid mixed ID",
			"message": "ID must be a positive integer",
		})
		return
	}

	// Парсим тело запроса
	var deleteReq service.DeleteFromMixedRequest
	if err := ctx.ShouldBindJSON(&deleteReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Выполняем удаление элемента
	response, err := h.MixingService.DeleteFromMixed(uint(mixedID), &deleteReq)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete from mixed",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
