package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"emplacc-api/internal/utils"
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
		subscriptionGroup.GET("/all/:page/:pagesize", getAllSubscriptions)
		subscriptionGroup.GET("/:id", getSubscriptionById)
		subscriptionGroup.GET("/user/:id/:page/:pagesize", getSubscriptionsByUserId)
		subscriptionGroup.GET("/sub-object/:id/:type/:page/:pagesize", getSubscriptionBySubObject)
		subscriptionGroup.POST("", createSubscription)
		subscriptionGroup.DELETE("/:id", deleteSubscription)
	}
}

// getAllSubscriptions godoc
// @Summary Получение всех подписок
// @Description Получение списка всех подписок с пагинацией (deleted = false)
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении подписок"
// @Success 200 {object} response.SubscriptionListResponse "Список подписок успешно получен"
// @Router /subscription/all/{page}/{pagesize} [get]
func getAllSubscriptions(c echo.Context) error{
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
	if err := dbConn.Session(&gorm.Session{}).Model(models.Subscription{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&subs).Error; err != nil{
			log.Printf("DB error (find subscription):%v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении подписок из базы данных", 
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
		subsList.Subscriptions = append(subsList.Subscriptions, response.SubscriptionResponse{
			ID: subscription.ID.String(),
			UserId: userId,
			SubscriptionId: subscriptionId,
			TypeId: utils.GetInt8(subscription.TypeID),
			CreatedAt: utils.GetTime(subscription.CreatedAt),	
		})
	}

	return c.JSON(http.StatusOK, subsList)
}

// getSubscriptionsByUserId godoc
// @Summary Получение подписок по ID пользователя
// @Description Получение списка подписок для конкретного пользователя с пагинацией (deleted = false)
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param id path string true "user_id"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров или данных запроса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении подписок"
// @Success 200 {object} response.SubscriptionListByUserIdResponse "Список подписок успешно получен"
// @Router /subscription/user/{id}/{page}/{pagesize} [get]
func getSubscriptionsByUserId(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор подписки",
		})
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

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.Subscription{}).Where("deleted = ? and user_id = ?", false, userId).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count projects): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете проектов",
		})
	}

	var subs []models.Subscription
	if err := dbConn.Session(&gorm.Session{}).Model(models.Subscription{}).Where("deleted = ? and user_id = ?", false, userId).
		Limit(pageSize).
		Offset(offset).
		Find(&subs).Error; err != nil{
			log.Printf("DB error (find subscription by id): %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении подписок из базы данныхэ", 
			})
	}

	subsList := response.SubscriptionListByUserIdResponse{
		Page: page,
		PageSize: pageSize,
		TotalCount: totalCount,
		UserId: id,
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
		subsList.Subscriptions = append(subsList.Subscriptions, response.SubscriptionResponse{
			ID: subscription.ID.String(),
			UserId: userId,
			SubscriptionId: subscriptionId,
			TypeId: utils.GetInt8(subscription.TypeID),
			CreatedAt: utils.GetTime(subscription.CreatedAt),	
		})
	}

	return c.JSON(http.StatusOK, subsList)
}

// getSubscriptionBySubObject godoc
// @Summary Получение подписок по объекту подписки
// @Description Получение списка подписок для конкретного объекта подписки и типа с пагинацией (deleted = false)
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Param subId path string true "UUID объекта подписки"
// @Param typeId path int true "Тип объекта подписки"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров или данных запроса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении подписок"
// @Success 200 {object} response.SubscriptionListBySubObjectResponse "Список подписок успешно получен"
// @Router /subscription/sub-object/{subId}/{typeId}/{page}/{pagesize} [get]
func getSubscriptionBySubObject(c echo.Context) error{
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

	id := c.Param("id")
	log.Print(id)
	log.Print(c.Request().URL.String())
	subId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор объекта",
		})
	}

	typeIdReq := c.Param("type")
	typeId, err := strconv.Atoi(typeIdReq)
	if err != nil{
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге типа объекта",
		})
	}

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.Subscription{}).Where("deleted = ? and subscription_id = ? and type_id = ?", false, subId, typeId).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count projects): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете проектов",
		})
	}

	var subs []models.Subscription
	if err := dbConn.Session(&gorm.Session{}).Model(models.Subscription{}).Where("deleted = ? and subscription_id = ? and type_id = ?", false, subId, typeId).
		Limit(pageSize).
		Offset(offset).
		Find(&subs).Error; err != nil{
			log.Printf("DB error (find subscription by id): %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении подписок из базы данных", 
			})
	}

	subsList := response.SubscriptionListBySubObjectResponse{
		Page: page,
		PageSize: pageSize,
		TotalCount: totalCount,
		TypeId: int8(typeId),
		SubscriptionId: id,
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
		subsList.Subscriptions = append(subsList.Subscriptions, response.SubscriptionResponse{
			ID: subscription.ID.String(),
			UserId: userId,
			SubscriptionId: subscriptionId,
			TypeId: utils.GetInt8(subscription.TypeID),
			CreatedAt: utils.GetTime(subscription.CreatedAt),	
		})
	}

	return c.JSON(http.StatusOK, subsList)
}

// getSubscriptionById godoc
// @Summary Получение подписки по ID
// @Description Получение детальной информации о подписке по её ID (deleted = false)
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Некорректный ID подписки"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Подписка не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении подписки"
// @Success 200 {object} response.SubscriptionResponse "Подписка успешно получена"
// @Router /subscription/{id} [get]
func getSubscriptionById(c echo.Context) error{
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
	result := dbConn.Session(&gorm.Session{}).Model(models.Subscription{}).Where("deleted = ? and id = ?", false, subId).First(&subscription)
	if result.Error != nil{
		if errors.Is(result.Error, gorm.ErrRecordNotFound){
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Подписка не найдена",
			})
		}
		log.Printf("DB error (find subscription by id): %v", result.Error)
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
	subResponse := response.SubscriptionResponse{
		ID: subscription.ID.String(),
		UserId: userId,
		SubscriptionId: subscriptionId,
		TypeId: utils.GetInt8(subscription.TypeID),
		CreatedAt: utils.GetTime(subscription.CreatedAt),	
	}
	return c.JSON(http.StatusOK, subResponse)
}

// createSubscription godoc
// @Summary Создание подписки
// @Description Создание новой подписки с указанными данными
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param body body request.SubscriptionCreateRequest true "Данные для создания подписки"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге данных запроса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании подписки"
// @Success 201 {object} response.SubscriptionUniversalResponse "Подписка успешно создана"
// @Router /subscription [post]
func createSubscription(c echo.Context) error{
	if err := authorize(c); err != nil{
		return err
	}
	var req request.SubscriptionCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	now := time.Now()

	del := false

	userId, err := uuid.Parse(req.UserId)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор пользователя",
		})
	}

	subId, err := uuid.Parse(req.SubscriptionId)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор подписки",
		})
	}

	typeId := *req.TypeId

	switch typeId{
	case 0:
		var task models.Task
		result := dbConn.Session(&gorm.Session{}).Model(models.Task{}).First(&task, "id = ? AND deleted = ?", subId, false)
		if result.Error != nil {
			log.Printf("DB error %v", result.Error)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Задача не найдена",
				})
			}
			log.Printf("DB error (find project by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении задачи из базы данных",
			})
		}
	case 1:
		var  problem models.Problem
		result := dbConn.Session(&gorm.Session{}).Model(models.Problem{}).First(&problem, "id = ? AND deleted = ?", subId, false)
		if result.Error != nil{
			log.Printf("DB error %v", result.Error)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Проблема не найдена",
				})
			}
			log.Printf("DB error (find problem by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении проблемы из базы данных",
			})
		}
	default:
		log.Print("incorect type of object")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный тип объекта",
		})
	}
 
	sub := models.Subscription{
		ID: &newUUID,
		UserID: &userId,
		SubscriptionId: &subId,
		TypeID: req.TypeId,
		CreatedAt: &now,
		Deleted: &del,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).Model(models.Subscription{}).Create(&sub); res.Error != nil{
			log.Printf("DB error(create subscription)")
		}
		return nil
	}); txErr != nil{
		log.Printf("DB transaction error (create project): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании проекта"})
	}

	createResponse := response.SubscriptionUniversalResponse{
		ID:       sub.ID.String(),
		Message: "Подписка создана",
	}

	return c.JSON(http.StatusCreated, createResponse)
}

// deleteSubscription godoc
// @Summary Удаление подписки
// @Description Логическое удаление подписки по ID (поле deleted = true)
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.SubscriptionUniversalResponse "Подписка успешно удалена"
// @Failure 404 {object} map[string]string "Подписка не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении подписки"
// @Router /subscription/{id} [delete]
func deleteSubscription(c echo.Context) error {
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

	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.Subscription{}).Where("id = ?", subId).Updates(updateData)
		if res.Error != nil{
			log.Printf("DB error (delete subscription): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0{
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		return nil
	}); txErr != nil{
		if he, ok := txErr.(*echo.HTTPError); ok{
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete project): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении подписки"})
	}

	delResponse :=  response.SubscriptionUniversalResponse{
		ID: id,
		Message: "Подписка удалена",
	}

	return c.JSON(http.StatusOK, delResponse)
}