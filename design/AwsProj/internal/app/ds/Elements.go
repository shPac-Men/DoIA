package ds

type Elements struct {
	ID            int    `gorm:"primaryKey"`
	IsDelete      bool   `gorm:"type:boolean not null;default:false"`
	Img           string `gorm:"type:varchar(100)"`
	Name          string `gorm:"type:varchar(25);not null"`
	Description   string `gorm:"type:varchar(100)"`
	Ph            float32
	Concentration float32
}
