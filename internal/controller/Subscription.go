package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterSubscriptionRoutes(e *echo.Echo){
	subscriptionGroup := e.Group("/subscription")
	subscriptionGroup.Use(KeycloakAuthMiddleware)
	{
		subscriptionGroup.GET("/all/:page/:pagesize", GetAllAttendances)
		subscriptionGroup.GET("/:id", GetSubscriptionById)
	}
}

func GetAllSubscriptions(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	pageReq := c.Param("page")
	pageSizeReq := c.Param("pagesize")
	// Значения по умолчанию
	page, err := strconv.Atoi(pageReq)
	if err != nil{
		log.Printf("failed to parse page: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге страницы",
		})
	}
	if page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeReq)
	if err != nil{
		log.Printf("failed to parse pagesize: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге номера страницы",
		})
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.Subscription{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count projects): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете проектов",
		})
	}

	var subs []models.Subscription
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&subs).Error; err != nil{
			log.Printf("DB error (find subscription)")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении подписок из базы данныхэ", 
			})
	}

	subsList := response.SubscriptionListResponse{
		Page: page,
		PageSize: pageSize,
		TotalCount: totalCount,
	}

	for _, subscription := range subs{
		var userId string
		if subscription.UserID != nil{
			userId = subscription.UserID.String()
		}
		var subscriptionId string
		if subscription.SubscriptionId != nil{
			subscriptionId = subscription.SubscriptionId.String()
		}
		var typeId int8
		if subscription.TypeID != nil{
			typeId = *subscription.TypeID
		}
		var createdAt time.Time
		if subscription.CreatedAt != nil{
			createdAt = *subscription.CreatedAt
		}	
		subsList.Subscriptions = append(subsList.Subscriptions, response.SubscriptionResponse{
			ID: subscription.ID.String(),
			UserId: userId,
			SubscriptionId: subscriptionId,
			TypeId: typeId,
			CreatedAt: createdAt,	
		})
	}

	return c.JSON(http.StatusOK, subsList)
}

func GetSubscriptionsByUserId(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	pageReq := c.Param("page")
	pageSizeReq := c.Param("pagesize")
	// Значения по умолчанию
	page, err := strconv.Atoi(pageReq)
	if err != nil{
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге страницы",
		})
	}
	if page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeReq)
	if err != nil{
		log.Printf("failed to parse pagesize: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге номера страницы",
		})
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var req request.SubscriptionListByUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	userId, err := uuid.Parse(req.UserID)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор пользователя",
		})
	}

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.Subscription{}).Where("deleted = ? and user_id = ?", false, userId).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count projects): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете проектов",
		})
	}

	var subs []models.Subscription
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ? and user_id = ?", false, userId).
		Limit(pageSize).
		Offset(offset).
		Find(&subs).Error; err != nil{
			log.Printf("DB error (find subscription)")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении подписок из базы данныхэ", 
			})
	}

	subsList := response.SubscriptionListByUserIdResponse{
		Page: page,
		PageSize: pageSize,
		TotalCount: totalCount,
		UserId: req.UserID,
	}

	for _, subscription := range subs{
		var userId string
		if subscription.UserID != nil{
			userId = subscription.UserID.String()
		}
		var subscriptionId string
		if subscription.SubscriptionId != nil{
			subscriptionId = subscription.SubscriptionId.String()
		}
		var typeId int8
		if subscription.TypeID != nil{
			typeId = *subscription.TypeID
		}
		var createdAt time.Time
		if subscription.CreatedAt != nil{
			createdAt = *subscription.CreatedAt
		}	
		subsList.Subscriptions = append(subsList.Subscriptions, response.SubscriptionResponse{
			ID: subscription.ID.String(),
			UserId: userId,
			SubscriptionId: subscriptionId,
			TypeId: typeId,
			CreatedAt: createdAt,	
		})
	}

	return c.JSON(http.StatusOK, subsList)
}

func GetSubscriptionBySubObject(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	pageReq := c.Param("page")
	pageSizeReq := c.Param("pagesize")
	// Значения по умолчанию
	page, err := strconv.Atoi(pageReq)
	if err != nil{
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге страницы",
		})
	}
	if page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeReq)
	if err != nil{
		log.Printf("failed to parse pagesize: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге номера страницы",
		})
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var req request.SubscriptionListBySubObjectRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	subscriptionId, err := uuid.Parse(req.SubscriptionId)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор объекта подписки",
		})
	}

	typeId := req.TypeId

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.Subscription{}).Where("deleted = ? and subscription_id = ? and type_id = ?", false, subscriptionId, typeId).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count projects): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете проектов",
		})
	}

	var subs []models.Subscription
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ? and subscription_id = ? and type_id = ?", false, subscriptionId, typeId).
		Limit(pageSize).
		Offset(offset).
		Find(&subs).Error; err != nil{
			log.Printf("DB error (find subscription)")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении подписок из базы данныхэ", 
			})
	}

	subsList := response.SubscriptionListBySubObjectResponse{
		Page: page,
		PageSize: pageSize,
		TotalCount: totalCount,
		TypeId: typeId,
		SubscriptionId: req.SubscriptionId,
	}

	for _, subscription := range subs{
		var userId string
		if subscription.UserID != nil{
			userId = subscription.UserID.String()
		}
		var subscriptionId string
		if subscription.SubscriptionId != nil{
			subscriptionId = subscription.SubscriptionId.String()
		}
		var typeId int8
		if subscription.TypeID != nil{
			typeId = *subscription.TypeID
		}
		var createdAt time.Time
		if subscription.CreatedAt != nil{
			createdAt = *subscription.CreatedAt
		}	
		subsList.Subscriptions = append(subsList.Subscriptions, response.SubscriptionResponse{
			ID: subscription.ID.String(),
			UserId: userId,
			SubscriptionId: subscriptionId,
			TypeId: typeId,
			CreatedAt: createdAt,	
		})
	}

	return c.JSON(http.StatusOK, subsList)
}

func GetSubscriptionById(c echo.Context) error{
	if err := authorize(c); err != nil{
		return err
	}
	id := c.Param("id")
	subId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор подписки",
		})
	}

	var subscription models.Subscription
	result := dbConn.Select(&gorm.Session{}).Where("deleted = ? and id = ?", false, subId).First(&subscription)
	if result.Error != nil{
		if errors.Is(result.Error, gorm.ErrRecordNotFound){
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Подписка не найдена",
			})
		}
		log.Printf("DB error (find subscription by id)")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении подписки из бд",
		})
	}

	var userId string
	if subscription.UserID != nil{
		userId = subscription.UserID.String()
	}
	var subscriptionId string
	if subscription.SubscriptionId != nil{
		subscriptionId = subscription.SubscriptionId.String()
	}
	var typeId int8
	if subscription.TypeID != nil{
		typeId = *subscription.TypeID
	}
	var createdAt time.Time
	if subscription.CreatedAt != nil{
		createdAt = *subscription.CreatedAt
	}

	subResponse := response.SubscriptionResponse{
		ID: subscription.ID.String(),
		UserId: userId,
		SubscriptionId: subscriptionId,
		TypeId: typeId,
		CreatedAt: createdAt,
	}
	return c.JSON(http.StatusOK, subResponse)
}