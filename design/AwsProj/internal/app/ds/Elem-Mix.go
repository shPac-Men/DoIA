package ds

// Связь элементов с заявкой (m-m)
type ElemMix struct {
	ID        uint `gorm:"primaryKey"`
	MixedID   uint `gorm:"not null;uniqueIndex:idx_mixed_elements"`
	ElementID uint `gorm:"not null;uniqueIndex:idx_mixed_elements"`

	// поля для пользовательского ввода
	Volume  float32 `gorm:"not null"` // объём в мл
	Comment string  `gorm:"type:varchar(100)"`

	IsDelete bool `gorm:"type:boolean;default:false"` // ← ДОБАВИТЬ ЭТО ПОЛЕ
	// связи
	Mixed   Mixed    `gorm:"foreignKey:MixedID"`
	Element Elements `gorm:"foreignKey:ElementID"`
}
