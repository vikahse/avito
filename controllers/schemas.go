package controllers

type LoginPayload struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserBannerPayload struct {
	TagId          int  `json:"tag_id" binding:"required"`
	FeatureId      int  `json:"feature_id" binding:"required"`
	UseLastVersion bool `json:"use_last_version"`
}

type CreateBannerPayload struct {
	TagIDs    []int                  `json:"tag_ids" binding:"required"`
	FeatureID int                    `json:"feature_id" binding:"required"`
	Content   map[string]interface{} `json:"content" binding:"required" gorm:"serializer:json"`
	IsActive  bool                   `json:"is_active"`
}

type GetBannersPayload struct {
	FeatureID *int `json:"feature_id"`
	TagId     *int `json:"tag_id"`
	Limit     *int `json:"limit"`
	Offset    *int `json:"offset"`
}

type UpdateBannerPayload struct {
	TagIDs    []int                  `json:"tag_ids"`
	FeatureID *int                    `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  *bool                   `json:"is_active"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshtoken"`
}