package model

type HallImage struct {
	HallID  uint64 `gorm:"primaryKey"`
	ImageID uint64 `gorm:"primaryKey"`

	Hall  Hall  `gorm:"foreignKey:HallID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Image Image `gorm:"foreignKey:ImageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func NewHallImage(hallID, imageID uint64) *HallImage {
	return &HallImage{
		HallID:  hallID,
		ImageID: imageID,
	}
}

type UserImage struct {
	UserID  uint64 `gorm:"primaryKey"`
	ImageID uint64 `gorm:"not null"`

	User  User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Image Image `gorm:"foreignKey:ImageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func NewUserImage(userID, imageID uint64) *UserImage {
	return &UserImage{
		UserID:  userID,
		ImageID: imageID,
	}
}
