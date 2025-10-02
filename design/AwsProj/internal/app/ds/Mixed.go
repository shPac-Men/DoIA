package ds

import (
	"database/sql"
	"time"
)

//то что получиться - заявка

type Mixed struct {
	ID            uint      `gorm:"primaryKey"`
	Status        string    `gorm:"type:varchar(15);not null"`
	DateCreate    time.Time `gorm:"not null"`
	DateUpdate    time.Time
	Concentartion float32
	Ph            float32
	TotalVolume   float64      `gorm:"type:decimal(10,2)"`
	AddedWater    float64      `gorm:"type:decimal(10,2)"`
	DateFinish    sql.NullTime `gorm:"default:null"`
	CreatorID     uint         `gorm:"not null"`
	ModeratorID   uint

	Creator   Users `gorm:"foreignKey:CreatorID"`
	Moderator Users `gorm:"foreignKey:ModeratorID"`
}
