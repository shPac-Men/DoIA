package handler

// ========== AUTH DTOs ==========
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ExpiresIn   int64  `json:"expires_in"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type PingResponse struct {
	Auth    bool   `json:"auth"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type RegisterRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
}

type RegisterResponse struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

type UserProfileResponse struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Email    string `json:"email"`
	UserUUID string `json:"user_uuid"`
}

type UpdateUserRequest struct {
	Email *string `json:"email,omitempty" binding:"omitempty,email"`
	Login *string `json:"login,omitempty" binding:"omitempty,min=3"`
}

// ========== ELEMENT DTOs ==========
type ElementView struct {
	ID            int
	Image         string
	Title         string
	Concentration string
	PH            string
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

type UploadImageResponse struct {
	ID      int    `json:"id"`
	Image   string `json:"image"`
	Message string `json:"message"`
}

// ========== MIXING DTOs ==========
type CartIconResponse struct {
	DraftOrderID int `json:"draft_order_id"`
	ItemsCount   int `json:"items_count"`
}

type AddToMixingRequest struct {
	ElementID int `json:"element_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

type RemoveFromMixingRequest struct {
	ElementID int `json:"element_id" binding:"required"`
}

type CreateMixingRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type CreateMixingResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// ========== COMMON DTOs ==========
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginationRequest struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=20" binding:"min=1,max=100"`
}

type PaginationResponse struct {
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
	Data       interface{} `json:"data"`
}
