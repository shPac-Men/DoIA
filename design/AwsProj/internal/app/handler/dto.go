package handler

// ========== AUTH DTOs ==========
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string   `json:"token"`
	ExpiresIn int64    `json:"expires_in"`
	TokenType string   `json:"token_type"`
	User      UserInfo `json:"user"`
}

type UserInfo struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	Role        string `json:"role"`
	IsModerator bool   `json:"is_moderator"`
}

type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=25"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterResponse struct {
	User    UserInfo `json:"user"`
	Message string   `json:"message"`
}

type UserProfileResponse struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	Role        string `json:"role"`
	IsModerator bool   `json:"is_moderator"`
}

type UpdateUserRequest struct {
	Login    *string `json:"login,omitempty" binding:"omitempty,min=3,max=25"`
	Password *string `json:"password,omitempty" binding:"omitempty,min=6"`
}

type UpdateUserRoleRequest struct {
	IsModerator bool `json:"is_moderator"`
}

// ========== ELEMENT DTOs ==========
type ElementResponse struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Ph            float32 `json:"ph"`
	Concentration float32 `json:"concentration"`
	Image         string  `json:"image"`
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

type UploadImageResponse struct {
	ID      int    `json:"id"`
	Image   string `json:"image"`
	Message string `json:"message"`
}

// ========== MIXING DTOs ==========
type AddToMixingRequest struct {
	ElementID int     `json:"element_id" binding:"required"`
	Volume    float32 `json:"volume" binding:"required,min=0"`
}

type RemoveFromMixingRequest struct {
	ElementID int `json:"element_id" binding:"required"`
}

type SubmitMixedRequest struct {
	AddedWater float64 `json:"added_water"`
}

type CartIconResponse struct {
	DraftOrderID int `json:"draft_order_id"`
	ItemsCount   int `json:"items_count"`
}

type MixingResponse struct {
	Items      []MixingItemResponse `json:"items"`
	TotalItems int                  `json:"total_items"`
	CartID     uint                 `json:"cart_id"`
	UserID     uint                 `json:"user_id"`
}

type MixingItemResponse struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Image         string  `json:"image"`
	Ph            float32 `json:"ph"`
	Concentration float32 `json:"concentration"`
	Volume        float32 `json:"volume"`
}

// ========== COMMON DTOs ==========
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type PingResponse struct {
	Status  bool   `json:"status"`
	Auth    bool   `json:"auth"`
	UserID  uint   `json:"user_id,omitempty"`
	Role    string `json:"role,omitempty"`
	Message string `json:"message"`
}

type MixedListResponse struct {
	Items []MixedListItem `json:"items"`
	Total int             `json:"total"`
}

type MixedListItem struct {
	ID             uint    `json:"id"`
	Status         string  `json:"status"`
	DateCreate     string  `json:"date_create"`
	DateUpdate     string  `json:"date_update"`
	DateFinish     string  `json:"date_finish,omitempty"`
	CreatorLogin   string  `json:"creator_login"`
	ModeratorLogin string  `json:"moderator_login,omitempty"`
	Ph             float32 `json:"ph"`
	Concentration  float32 `json:"concentration"`
	TotalVolume    float64 `json:"total_volume"`
	AddedWater     float64 `json:"added_water"`
	ItemsCount     int     `json:"items_count"`      // <--- Добавляем
	ProcessedCount int     `json:"processed_count"` // Количество записей с ph > 0
}

type MixedDetailResponse struct {
	ID             uint              `json:"id"`
	Status         string            `json:"status"`
	DateCreate     string            `json:"date_create"`
	DateUpdate     string            `json:"date_update"`
	DateFinish     string            `json:"date_finish,omitempty"`
	CreatorLogin   string            `json:"creator_login"`
	ModeratorLogin string            `json:"moderator_login,omitempty"`
	Ph             float32           `json:"ph"`
	Concentration  float32           `json:"concentration"`
	TotalVolume    float64           `json:"total_volume"`
	AddedWater     float64           `json:"added_water"`
	ItemsCount     int               `json:"items_count"` // <--- Добавляем с JSON тегом
	Items          []MixedDetailItem `json:"items"`
}

type MixedDetailItem struct {
	ElementID     int     `json:"element_id"`
	Title         string  `json:"title"`
	Image         string  `json:"image"`
	Ph            float32 `json:"ph"`
	Concentration float32 `json:"concentration"`
	Volume        float32 `json:"volume"`
	Comment       string  `json:"comment,omitempty"`
}

type UpdateMixedRequest struct {
	Status        *string  `json:"status,omitempty"`
	Concentration *float32 `json:"concentration,omitempty"`
	Ph            *float32 `json:"ph,omitempty"`
	TotalVolume   *float64 `json:"total_volume,omitempty"`
	AddedWater    *float64 `json:"added_water,omitempty"`
}

type CreateMixingRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type DeleteFromMixedRequest struct {
	ElementID  uint `json:"element_id" binding:"required"`
	HardDelete bool `json:"hard_delete"`
}

type ProcessingResultDTO struct {
	// Используем указатель, чтобы отличать переданный 0 от отсутствия значения,
	// или просто float64, если Python всегда шлет число.
	Result float64 `json:"result"`
}
