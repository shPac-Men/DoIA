package handler

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

type CreateElementRequest struct {
	Name          string  `json:"name" binding:"required,min=1,max=25"`
	Description   string  `json:"description" binding:"max=100"`
	Ph            float32 `json:"ph" binding:"required,min=0,max=14"`
	Concentration float32 `json:"concentration" binding:"required,min=0"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type UpdateElementRequest struct {
	Name          *string  `json:"name,omitempty" binding:"omitempty,min=1,max=25"`
	Description   *string  `json:"description,omitempty" binding:"omitempty,max=100"`
	Ph            *float32 `json:"ph,omitempty" binding:"omitempty,min=0,max=14"`
	Concentration *float32 `json:"concentration,omitempty" binding:"omitempty,min=0"`
}

type UpdateElementResponse struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Ph            float32 `json:"ph"`
	Concentration float32 `json:"concentration"`
	Image         string  `json:"image"`
	Message       string  `json:"message"`
}
