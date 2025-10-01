package ds

type ElemMix struct {
	ID uint `gorm:"primaryKey"`
	// здесь создаем Unique key, указывая общий uniqueIndex
	PonID      uint `gorm:"not null;uniqueIndex:idx_mixed_elements"`
	ElementsID uint `gorm:"not null;uniqueIndex:idx_mixed_elements"`

	Sound bool `gorm:"default:true"`

	Mixed    Mixed    `gorm:"foreignKey:MixedID"`
	Elements Elements `gorm:"foreignKey:ElementID"`
}

//много польз могут создавать много заявок

//черновик = корзина

//ид кислоты ид корзины(заявки)

//у пользователя 1 черновик. после формирования черновик превращается в заявку и удаляется.
