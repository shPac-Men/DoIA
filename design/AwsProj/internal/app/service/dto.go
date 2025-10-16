package service

type MixingResponse struct {
	Items      []MixingItem `json:"data"`
	TotalItems int          `json:"total_items"`
	CartID     uint         `json:"cart_id"`
	UserID     uint         `json:"user_id"`
}

type MixingItem struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Image         string  `json:"image"`
	PH            float32 `json:"ph"`
	Concentration float32 `json:"concentration"`
	Volume        float32 `json:"volume"`
}

// DTO для элементов
type ElementsResponse struct {
	Items   []ElementItem `json:"data"`
	Total   int           `json:"total"`
	Query   string        `json:"query"`
	HasMore bool          `json:"has_more"`
}

type ElementItem struct {
	ID            int    `json:"id"`
	Image         string `json:"image"`
	Title         string `json:"title"`
	Concentration string `json:"concentration"`
	PH            string `json:"ph"`
}

// post
type CreateElementRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	Ph            float32 `json:"ph" binding:"required"`
	Concentration float32 `json:"concentration" binding:"required"`
}

type CreateElementResponse struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Ph            float32 `json:"ph"`
	Concentration float32 `json:"concentration"`
	Message       string  `json:"message"`
}

// put
type UpdateElementRequest struct {
	Name          *string  `json:"name,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Ph            *float32 `json:"ph,omitempty"`
	Concentration *float32 `json:"concentration,omitempty"`
}

type ElementResponse struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Ph            float32 `json:"ph"`
	Concentration float32 `json:"concentration"`
	Image         string  `json:"image"`
	Message       string  `json:"message"`
}
