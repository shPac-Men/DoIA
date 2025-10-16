package service

import "time"

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

// удаление
type DeleteElementResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

type AddToMixingRequest struct {
	ElementID int     `json:"element_id" binding:"required"`
	Volume    float32 `json:"volume,omitempty"`
}

// AddToMixingResponse DTO для ответа
type AddToMixingResponse struct {
	ElementID int     `json:"element_id"`
	Volume    float32 `json:"volume"`
	UserID    uint    `json:"user_id"`
	Message   string  `json:"message"`
}

type UploadImageResponse struct {
	ID      int    `json:"id"`
	Image   string `json:"image"`
	Message string `json:"message"`
}

type CartIconResponse struct {
	DraftOrderID int `json:"draft_order_id"`
	ItemsCount   int `json:"items_count"`
}

type MixedListRequest struct {
	Status   string `form:"status"`
	DateFrom string `form:"date_from"` // Изменяем на string
	DateTo   string `form:"date_to"`   // Изменяем на string
}

// MixedListItem - элемент списка заявок
type MixedListItem struct {
	ID             uint      `json:"id"`
	Status         string    `json:"status"`
	DateCreate     time.Time `json:"date_create"`
	DateUpdate     time.Time `json:"date_update"`
	DateFinish     time.Time `json:"date_finish,omitempty"`
	CreatorLogin   string    `json:"creator_login"`
	ModeratorLogin string    `json:"moderator_login,omitempty"`
	Ph             float32   `json:"ph"`
	Concentration  float32   `json:"concentration"`
	TotalVolume    float64   `json:"total_volume"`
	AddedWater     float64   `json:"added_water"`
}

// MixedDetailResponse - ответ с деталями заявки
type MixedDetailResponse struct {
	ID             uint              `json:"id"`
	Status         string            `json:"status"`
	DateCreate     time.Time         `json:"date_create"`
	DateUpdate     time.Time         `json:"date_update"`
	DateFinish     *time.Time        `json:"date_finish,omitempty"`
	CreatorLogin   string            `json:"creator_login"`
	ModeratorLogin string            `json:"moderator_login,omitempty"`
	Ph             float32           `json:"ph"`
	Concentration  float32           `json:"concentration"`
	TotalVolume    float64           `json:"total_volume"`
	AddedWater     float64           `json:"added_water"`
	Items          []MixedDetailItem `json:"items"`
}

// MixedDetailItem - элемент заявки
type MixedDetailItem struct {
	ElementID     int     `json:"element_id"`
	Title         string  `json:"title"`
	Image         string  `json:"image"`
	PH            float32 `json:"ph"`
	Concentration float32 `json:"concentration"`
	Volume        float32 `json:"volume"`
	Comment       string  `json:"comment,omitempty"`
}

// UpdateMixedRequest - запрос на обновление заявки
type UpdateMixedRequest struct {
	Status        string  `json:"status,omitempty"`
	Concentration float32 `json:"concentration,omitempty"`
	Ph            float32 `json:"ph,omitempty"`
	TotalVolume   float64 `json:"total_volume,omitempty"`
	AddedWater    float64 `json:"added_water,omitempty"`
}

// CompleteMixedRequest - запрос на формирование заявки
type CompleteMixedRequest struct {
	// Можно добавить дополнительные поля если нужны
}

// CompleteMixedResponse - ответ после формирования заявки
type CompleteMixedResponse struct {
	MixedID    uint      `json:"mixed_id"`
	Status     string    `json:"status"`
	DateUpdate time.Time `json:"date_update"`
	Message    string    `json:"message"`
}

// DeleteMixedRequest - запрос на удаление заявки
type DeleteMixedRequest struct {
	HardDelete bool `json:"hard_delete,omitempty"` // true - полное удаление, false - soft delete
}

// DeleteMixedResponse - ответ после удаления заявки
type DeleteMixedResponse struct {
	MixedID   uint      `json:"mixed_id"`
	Status    string    `json:"status"`
	DeletedAt time.Time `json:"deleted_at"`
	Message   string    `json:"message"`
}

// DeleteFromMixedRequest - запрос на удаление элемента из заявки
type DeleteFromMixedRequest struct {
	ElementID  uint `json:"element_id" binding:"required"`
	HardDelete bool `json:"hard_delete,omitempty"`
}

// DeleteFromMixedResponse - ответ после удаления элемента
type DeleteFromMixedResponse struct {
	MixedID   uint      `json:"mixed_id"`
	ElementID uint      `json:"element_id"`
	DeletedAt time.Time `json:"deleted_at"`
	Message   string    `json:"message"`
}

type RegisterRequest struct {
	Login       string `json:"login" binding:"required,min=3,max=25"`
	Password    string `json:"password" binding:"required,min=6"`
	IsModerator bool   `json:"is_moderator,omitempty"`
}

// RegisterResponse - ответ после регистрации
type RegisterResponse struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
	Message     string `json:"message"`
}

type UserProfileResponse struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
}

type UpdateUserRequest struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

// LoginRequest - запрос на аутентификацию
type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse - ответ после аутентификации
type LoginResponse struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	IsModerator bool   `json:"is_moderator"`
	Message     string `json:"message"`
}

// LogoutRequest - запрос на деавторизацию (может быть пустым)
type LogoutRequest struct {
	// Можно добавить поля если нужны (например, token для blacklist)
}

// LogoutResponse - ответ после деавторизации
type LogoutResponse struct {
	Message string `json:"message"`
}
