package service

import "time"

// ========== AUTH ==========
type RegisterRequest struct {
	Login    string
	Password string
}

type RegisterResponse struct {
	ID          uint
	Login       string
	IsModerator bool
	Message     string
}

type LoginRequest struct {
	Login    string
	Password string
}

type LoginResponse struct {
	Token       string
	ID          uint
	Login       string
	Role        string
	IsModerator bool
	Message     string
}

type UserProfileResponse struct {
	ID          uint
	Login       string
	Role        string
	IsModerator bool
}

type UpdateUserRequest struct {
	Login    string
	Password string
}

type LogoutResponse struct {
	Message string
}

// ========== ELEMENTS ==========
type ElementsResponse struct {
	Items   []ElementItem
	Total   int
	Query   string
	HasMore bool
}

type ElementItem struct {
	ID            int
	Image         string
	Title         string
	Concentration string
	PH            string
}

type CreateElementRequest struct {
	Name          string
	Description   string
	Ph            float32
	Concentration float32
}

type CreateElementResponse struct {
	ID            int
	Name          string
	Description   string
	Ph            float32
	Concentration float32
	Message       string
}

type UpdateElementRequest struct {
	Name          *string
	Description   *string
	Ph            *float32
	Concentration *float32
}

type ElementResponse struct {
	ID            int
	Name          string
	Description   string
	Ph            float32
	Concentration float32
	Image         string
	Message       string
}

type DeleteElementResponse struct {
	ID      int
	Message string
}

type UploadImageResponse struct {
	ID      int
	Image   string
	Message string
}

// ========== MIXING (КОРЗИНА) ==========
type MixingResponse struct {
	Items      []MixingItem
	TotalItems int
	CartID     uint
	UserID     uint
}

type MixingItem struct {
	ID            int
	Title         string
	Image         string
	PH            float32
	Concentration float32
	Volume        float32
}

type AddToMixingRequest struct {
	ElementID int
	Volume    float32
}

type AddToMixingResponse struct {
	ElementID int
	Volume    float32
	UserID    uint
	Message   string
}

type CartIconResponse struct {
	DraftOrderID int
	ItemsCount   int
}

// ========== MIXED (ЗАКАЗЫ) ==========
type MixedListRequest struct {
	DateFrom string
	DateTo   string
}

type MixedListItem struct {
	ID             uint
	Status         string
	DateCreate     time.Time
	DateUpdate     time.Time
	DateFinish     time.Time
	CreatorLogin   string
	ModeratorLogin string
	Ph             float32
	Concentration  float32
	TotalVolume    float64
	AddedWater     float64
	ItemsCount     int // <--- Добавляе
}

type MixedDetailResponse struct {
	ID             uint
	Status         string
	DateCreate     time.Time
	DateUpdate     time.Time
	DateFinish     *time.Time
	CreatorLogin   string
	ModeratorLogin string
	Ph             float32
	Concentration  float32
	TotalVolume    float64
	AddedWater     float64
	ItemsCount     int // <--- Добавляем поле сюда
	Items          []MixedDetailItem
}

type MixedDetailItem struct {
	ElementID     int
	Title         string
	Image         string
	PH            float32
	Concentration float32
	Volume        float32
	Comment       string
}

type UpdateMixedRequest struct {
	Status        string
	Concentration float32
	Ph            float32
	TotalVolume   float64
	AddedWater    float64
}

type CompleteMixedRequest struct {
	// Дополнительные поля при необходимости
}

type CompleteMixedResponse struct {
	MixedID    uint
	Status     string
	DateUpdate time.Time
	Message    string
}

type DeleteMixedRequest struct {
	HardDelete bool
}

type DeleteMixedResponse struct {
	MixedID   uint
	Status    string
	DeletedAt time.Time
	Message   string
}

type DeleteFromMixedRequest struct {
	ElementID  uint
	HardDelete bool
}

type DeleteFromMixedResponse struct {
	MixedID   uint
	ElementID uint
	DeletedAt time.Time
	Message   string
}
