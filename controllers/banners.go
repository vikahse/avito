package controllers

import (
	"avito/auth"
	"avito/database"
	"avito/models"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserBanner(c *gin.Context) {
	var payload UserBannerPayload
	var banner models.Banner

	config, err := godotenv.Read(".env")

	if err != nil {
		log.Fatal("Error reading .env file")
	}

	token := c.GetHeader("token")
	if token == "" {
		c.JSON(401, gin.H{
			"Code":  401,
			"Error": "Пользователь не авторизован",
		})
		c.Abort()
		return
	}

	err = c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(400, gin.H{
			"Code":  400,
			"Error": "Некорректные данные",
		})
		c.Abort()
		return
	}

	jwtWrapper := auth.JwtWrapper{
		SecretKey:         config["SECRET_KEY"],
		Issuer:            "BannerService",
		ExpirationMinutes: 20,
		ExpirationHours:   12,
	}

	claims, err := jwtWrapper.ValidateToken(token)
	if err != nil {
		fmt.Print(err)
		c.JSON(403, gin.H{
			"Code":  403,
			"Error": "Пользователь не имеет доступа",
		})
		c.Abort()
		return
	}

	if !claims.IsAdmin { // если не админ, то смотрим только активные банеры по tag_id и feature_id
		if payload.UseLastVersion {
			result := database.GlobalDB.Joins("JOIN banner_tags ON banners.id = banner_tags.banner_id").
				Joins("JOIN tags ON banner_tags.tag_id = tags.id").
				Where("tags.value = ? AND feature_id = ? AND is_active = ?", payload.TagId, payload.FeatureId, true).
				Order("updated_at desc").
				First(&banner)
			if result.Error != nil {
				if result.Error == gorm.ErrRecordNotFound {
					c.JSON(404, gin.H{
						"Code":  404,
						"Error": "Баннер не найден",
					})
					c.Abort()
					return
				} else {
					c.JSON(500, gin.H{
						"Code":  500,
						"Error": "Внутренняя ошибка сервера",
					})
					c.Abort()
					return
				}
			}
		} else {
			err := database.CacheClient.HGetAll(c, bannerCacheKey(payload.TagId, payload.FeatureId)).Scan(banner) ///
			if err != nil {                                                                                       // не получилось достать из кеша
				result := database.GlobalDB.Joins("JOIN banner_tags ON banners.id = banner_tags.banner_id").
					Joins("JOIN tags ON banner_tags.tag_id = tags.id").
					Where("tags.value = ? AND feature_id = ? AND is_active = ?", payload.TagId, payload.FeatureId, true).
					Order("updated_at desc").
					First(&banner)
				if result.Error != nil {
					if result.Error == gorm.ErrRecordNotFound {
						c.JSON(404, gin.H{
							"Code":  404,
							"Error": "Баннер не найден",
						})
						c.Abort()
						return
					} else {
						c.JSON(500, gin.H{
							"Code":  500,
							"Error": "Внутренняя ошибка сервера",
						})
						c.Abort()
						return
					}
				}
				database.CacheClient.HSet(c, bannerCacheKey(payload.TagId, payload.FeatureId), banner)
			}
		}
	} else { // если админ, то имеет доступ и к выключеным банерам
		if payload.UseLastVersion {
			result := database.GlobalDB.Joins("JOIN banner_tags ON banners.id = banner_tags.banner_id").
				Joins("JOIN tags ON banner_tags.tag_id = tags.id").
				Where("tags.value = ? AND feature_id = ?", payload.TagId, payload.FeatureId).
				Order("updated_at desc").
				First(&banner)
			if result.Error != nil {
				if result.Error == gorm.ErrRecordNotFound {
					c.JSON(404, gin.H{
						"Code":  404,
						"Error": "Баннер не найден",
					})
					c.Abort()
					return
				} else {
					c.JSON(500, gin.H{
						"Code":  500,
						"Error": "Внутренняя ошибка сервера",
					})
					c.Abort()
					return
				}
			}
		} else {
			b, err := database.CacheClient.Get(c, bannerCacheKey(payload.TagId, payload.FeatureId)).Bytes()
			if err != nil {
				result := database.GlobalDB.Joins("JOIN banner_tags ON banners.id = banner_tags.banner_id").
					Joins("JOIN tags ON banner_tags.tag_id = tags.id").
					Where("tags.value = ? AND feature_id = ?", payload.TagId, payload.FeatureId).
					Order("updated_at desc").
					First(&banner)
				if result.Error != nil {
					if result.Error == gorm.ErrRecordNotFound {
						c.JSON(404, gin.H{
							"Code":  404,
							"Error": "Баннер не найден",
						})
						c.Abort()
						return
					} else {
						c.JSON(500, gin.H{
							"Code":  500,
							"Error": "Внутренняя ошибка сервера",
						})
						c.Abort()
						return
					}
				}

				bannerBytes, err := json.Marshal(banner)
				if err != nil {
					c.JSON(500, gin.H{
						"Code":  500,
						"Error": "Internal error",
					})
					c.Abort()
					return
				}

				err = database.CacheClient.Set(c, bannerCacheKey(payload.TagId, payload.FeatureId), bannerBytes, time.Minute*5).Err()
				if err != nil {
					c.JSON(500, gin.H{
						"Code":  500,
						"Error": "Не удалось записать данные в кеш",
					})
					c.Abort()
					return
				}
			} else {
				err = json.Unmarshal(b, &banner)
				if err != nil {
					c.JSON(500, gin.H{
						"Code":  500,
						"Error": "Internal error",
					})
					c.Abort()
					return
				}
			}
		}
	}

	c.JSON(200, gin.H{
		"Code": 200,
		"JSON-отображение баннера": banner.Content,
	})
}

func CreateBanner(c *gin.Context) {
	var payload CreateBannerPayload
	var banner models.Banner

	config, err := godotenv.Read(".env")

	if err != nil {
		log.Fatal("Error reading .env file")
	}

	token := c.GetHeader("token")
	if token == "" {
		c.JSON(401, gin.H{
			"Code":  401,
			"Error": "Пользователь не авторизован",
		})
		c.Abort()
		return
	}

	err = c.ShouldBindJSON(&payload)
	if err != nil {
		fmt.Print(err)
		c.JSON(400, gin.H{
			"Code":  400,
			"Error": "Некорректные данные",
		})
		c.Abort()
		return
	}

	jwtWrapper := auth.JwtWrapper{
		SecretKey:         config["SECRET_KEY"],
		Issuer:            "BannerService",
		ExpirationMinutes: 20,
		ExpirationHours:   12,
	}

	claims, err := jwtWrapper.ValidateToken(token)
	if err != nil {
		c.JSON(401, gin.H{
			"Code":  401,
			"Error": "Пользователь не авторизован",
		})
		c.Abort()
		return
	}

	if !claims.IsAdmin {
		c.JSON(403, gin.H{
			"Code":  403,
			"Error": "Пользователь не имеет доступа",
		})
		c.Abort()
		return
	}

	var tags []models.Tag

	for _, tag := range payload.TagIDs {
		newTag := models.Tag{
			Value: tag,
		}

		tags = append(tags, newTag)
	}

	banner = models.Banner{
		TagIds:    tags,
		FeatureId: payload.FeatureID,
		Content:   payload.Content,
		IsActive:  payload.IsActive,
	}

	result := database.GlobalDB.Create(&banner)
	if result.Error != nil {
		c.JSON(500, gin.H{
			"Code":  500,
			"Error": "Внутренняя ошибка сервера",
		})
		c.Abort()
		return
	}

	c.JSON(201, gin.H{
		"Code":      201,
		"banner_id": banner.ID,
	})
}

func GetBanners(c *gin.Context) {
	var payload GetBannersPayload
	var banners []models.Banner

	config, err := godotenv.Read(".env")

	if err != nil {
		log.Fatal("Error reading .env file")
	}

	token := c.GetHeader("token")
	if token == "" {
		c.JSON(401, gin.H{
			"Code":  401,
			"Error": "Пользователь не авторизован",
		})
		c.Abort()
		return
	}

	err = c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(400, gin.H{
			"Code":  400,
			"Error": "Неверный формат данных",
		})
		c.Abort()
		return
	}

	jwtWrapper := auth.JwtWrapper{
		SecretKey:         config["SECRET_KEY"],
		Issuer:            "BannerService",
		ExpirationMinutes: 20,
		ExpirationHours:   12,
	}

	_, err = jwtWrapper.ValidateToken(token)
	if err != nil {
		c.JSON(401, gin.H{
			"Code":  401,
			"Error": "Пользователь не авторизован",
		})
		c.Abort()
		return
	}

	query := database.GlobalDB.Joins("JOIN banner_tags ON banners.id = banner_tags.banner_id").
		Joins("JOIN tags ON banner_tags.tag_id = tags.id")

	if payload.TagId != nil {
		query = query.Where("tags.value = ?", *payload.TagId)
	}

	if payload.FeatureID != nil {
		query = query.Where("feature_id = ?", *payload.FeatureID)
	}

	if payload.Offset != nil {
		query = query.Offset(*payload.Offset)
	} else {
		query = query.Offset(0)
	}

	if payload.Limit != nil {
		query = query.Limit(*payload.Limit)
	} else {
		query = query.Limit(10)
	}

	result := query.Order("updated_at desc").Find(&banners)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{
				"Code":  404,
				"Error": "Баннеры не найдены",
			})
			c.Abort()
			return
		} else {
			c.JSON(500, gin.H{
				"Code":  500,
				"Error": "Внутренняя ошибка сервера",
			})
			c.Abort()
			return
		}
	}

	var tagsIDs []int
	var tagsValues []int
	var result_arr []map[string]interface{}
	for _, banner := range banners {
		item := make(map[string]interface{})
		item["banner_id"] = banner.ID

		database.GlobalDB.Table("banner_tags").Select("tag_id").Where("banner_id IN (?)", banner.ID).Pluck("tag_id", &tagsIDs)
		database.GlobalDB.Table("tags").Select("value").Where("id IN (?)", tagsIDs).Pluck("value", &tagsValues)

		item["tag_ids"] = tagsValues
		item["feature_id"] = banner.FeatureId
		item["content"] = banner.Content
		item["is_active"] = banner.IsActive
		item["created_at"] = banner.CreatedAt
		item["updated_at"] = banner.UpdatedAt
		result_arr = append(result_arr, item)

		tagsIDs = tagsIDs[:0]
		tagsValues = tagsValues[:0]
	}

	c.JSON(200, gin.H{
		"Code":    200,
		"Content": result_arr,
	})

}

func UpdateBannerById(c *gin.Context) {
	var payload UpdateBannerPayload
	var banner models.Banner

	config, err := godotenv.Read(".env")

	if err != nil {
		log.Fatal("Error reading .env file")
	}

	id := c.Param("id")

	token := c.GetHeader("token")
	if token == "" {
		c.JSON(401, gin.H{
			"Code":  401,
			"Error": "Пользователь не авторизован",
		})
		c.Abort()
		return
	}

	err = c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(400, gin.H{
			"Code":  400,
			"Error": "Некорректные данные",
		})
		c.Abort()
		return
	}

	jwtWrapper := auth.JwtWrapper{
		SecretKey:         config["SECRET_KEY"],
		Issuer:            "BannerService",
		ExpirationMinutes: 20,
		ExpirationHours:   12,
	}

	claims, err := jwtWrapper.ValidateToken(token)
	if err != nil {
		c.JSON(401, gin.H{
			"Code":  401,
			"Error": "Пользователь не авторизован",
		})
		c.Abort()
		return
	}

	if !claims.IsAdmin {
		c.JSON(403, gin.H{
			"Code":  403,
			"Error": "Пользователь не имеет доступа",
		})
		c.Abort()
		return
	}

	id_number, _ := strconv.Atoi(id)

	result := database.GlobalDB.First(&banner, "id = ?", id_number)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{
				"Code":  404,
				"Error": "Баннер не найден",
			})
			c.Abort()
			return
		} else {
			c.JSON(500, gin.H{
				"Code":  500,
				"Error": "Внутренняя ошибка сервера",
			})
			c.Abort()
			return
		}
	}

	var existingTags []models.Tag
	database.GlobalDB.Model(&banner).Association("TagIds").Find(&existingTags)
	for _, existingTag := range existingTags {
		database.GlobalDB.Model(&banner).Association("TagIds").Delete(existingTag)
	}

	var tags []models.Tag

	if len(payload.TagIDs) != 0 {

		for _, tag := range payload.TagIDs {
			newTag := models.Tag{
				Value: tag,
			}

			tags = append(tags, newTag)
		}
	}

	if payload.FeatureID != nil {
		banner.FeatureId = *payload.FeatureID
	}

	if payload.Content != nil {
		banner.Content = payload.Content
	}

	if payload.IsActive != nil {
		banner.IsActive = *payload.IsActive
	}

	database.GlobalDB.Save(&banner).Association("TagIds").Append(tags)

	c.JSON(200, gin.H{
		"Code": 200,
	})

}

func DeleteBannerById(c *gin.Context) {
	var banner models.Banner

	config, err := godotenv.Read(".env")

	if err != nil {
		log.Fatal("Error reading .env file")
	}

	id := c.Param("id")

	token := c.GetHeader("token")
	if token == "" {
		c.JSON(401, gin.H{
			"Code":  401,
			"Error": "Пользователь не авторизован",
		})
		c.Abort()
		return
	}

	jwtWrapper := auth.JwtWrapper{
		SecretKey:         config["SECRET_KEY"],
		Issuer:            "BannerService",
		ExpirationMinutes: 20,
		ExpirationHours:   12,
	}

	claims, err := jwtWrapper.ValidateToken(token)
	if err != nil {
		c.JSON(401, gin.H{
			"Code":  401,
			"Error": "Пользователь не авторизован",
		})
		c.Abort()
		return
	}

	if !claims.IsAdmin {
		c.JSON(403, gin.H{
			"Code":  403,
			"Error": "Пользователь не имеет доступа",
		})
		c.Abort()
		return
	}

	id_number, _ := strconv.Atoi(id)
	result := database.GlobalDB.Where("id = ?", id_number).First(&banner)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{
				"Code":  404,
				"Error": "Баннер для тэга не найден",
			})
			c.Abort()
			return
		} else {
			c.JSON(500, gin.H{
				"Code":  500,
				"Error": "Внутренняя ошибка сервера",
			})
			c.Abort()
			return
		}
	}

	//удаляем связи в join таблице
	var existingTags []models.Tag
	database.GlobalDB.Model(&banner).Association("TagIds").Find(&existingTags)
	for _, existingTag := range existingTags {
		database.GlobalDB.Model(&banner).Association("TagIds").Delete(existingTag)
	}

	database.GlobalDB.Where("id = ?", id_number).Delete(&banner)

	c.JSON(200, gin.H{
		"Code": 200,
	})

}

func bannerCacheKey(tagID, featureID int) string {
	return fmt.Sprintf("%d-%d", tagID, featureID)
}
