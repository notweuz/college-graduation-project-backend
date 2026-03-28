package model

type Image struct {
	ID   uint64 `gorm:"primaryKey;autoIncrement"`
	Path string `gorm:"size:255;not null"`
}

func NewImage(id uint64, path string) *Image {
	return &Image{
		ID:   id,
		Path: path,
	}
}
