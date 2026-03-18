package model

type HallImage struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement"`
	HallID uint64 `gorm:"not null;index"`
	Path   string `gorm:"size:255;not null"`
}

func NewHallImage(id, hallID uint64, path string) *HallImage {
	return &HallImage{
		ID:     id,
		HallID: hallID,
		Path:   path,
	}
}
