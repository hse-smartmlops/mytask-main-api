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
		subscriptionGroup.GET("/all/:page/:pagesize", GetAllSubscriptions)
		subscriptionGroup.GET("/:id", GetSubscriptionById)
		subscriptionGroup.GET("/user/:id/:page/:pagesize", GetSubscriptionsByUserId)
		subscriptionGroup.GET("/sub-object/:id/:type/:page/:pagesize", GetSubscriptionBySubObject)
		subscriptionGroup.POST("", CreateSubscription)
		subscriptionGroup.DELETE("/:id", DeleteSubscription)
	}
}

// GetAllSubscriptions godoc
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
func GetAllSubscriptions(c echo.Context) error{
	if err := Authorize(c); err != nil { return err }

	pageReq := c.Param("page")
	pageSizeReq := c.Param("pagesize")

	page, err := strconv.Atoi(pageReq)
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(pageSizeReq)
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге размера страницы"})
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Subscription{}).
		Where("deleted = FALSE").
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count subscriptions): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте подписок"})
	}

	var subs []models.Subscription
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Subscription{}).
		Where("deleted = FALSE").
		Limit(pageSize).Offset(offset).
		Find(&subs).Error; err != nil {
		log.Printf("DB error (find subscriptions): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении подписок"})
	}

	out := response.SubscriptionListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}
	for _, s := range subs {
		var subscriptionId string
		if s.SubscriptionId != nil {
			subscriptionId = s.SubscriptionId.String()
		}
		out.Subscriptions = append(out.Subscriptions, response.SubscriptionResponse{
			ID:             s.ID.String(),
			UserId:         s.UserID.String(),
			SubscriptionId: subscriptionId,
			TypeId:         utils.GetInt8(s.TypeID),
			CreatedAt:      utils.GetTime(s.CreatedAt),
		})
	}

	return c.JSON(http.StatusOK, out)
}

// GetSubscriptionsByUserId godoc
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
func GetSubscriptionsByUserId(c echo.Context) error{
	if err := Authorize(c); err != nil { return err }

	userUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор подписки"})
	}
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге размера страницы"})
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Subscription{}).
		Where("deleted = FALSE AND user_id = ?", userUUID).
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count subscriptions by user): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте подписок"})
	}

	var subs []models.Subscription
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Subscription{}).
		Where("deleted = FALSE AND user_id = ?", userUUID).
		Limit(pageSize).Offset(offset).
		Find(&subs).Error; err != nil {
		log.Printf("DB error (find subscriptions by user): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении подписок"})
	}

	out := response.SubscriptionListByUserIdResponse{
		UserId:     c.Param("id"),
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}
	for _, s := range subs {
		var subscriptionId string
		if s.SubscriptionId != nil {
			subscriptionId = s.SubscriptionId.String()
		}
		out.Subscriptions = append(out.Subscriptions, response.SubscriptionResponse{
			ID:             s.ID.String(),
			UserId:         s.UserID.String(),
			SubscriptionId: subscriptionId,
			TypeId:         utils.GetInt8(s.TypeID),
			CreatedAt:      utils.GetTime(s.CreatedAt),
		})
	}

	return c.JSON(http.StatusOK, out)
}

// GetSubscriptionBySubObject godoc
// @Summary Получение подписок по объекту подписки
// @Description Получение списка подписок для конкретного объекта подписки и типа с пагинацией (deleted = false)
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Param id path string true "UUID объекта подписки"
// @Param type path int true "Тип объекта подписки"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров или данных запроса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении подписок"
// @Success 200 {object} response.SubscriptionListBySubObjectResponse "Список подписок успешно получен"
// @Router /subscription/sub-object/{id}/{type}/{page}/{pagesize} [get]
func GetSubscriptionBySubObject(c echo.Context) error{
	if err := Authorize(c); err != nil { return err }

	subObjUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор объекта"})
	}
	typeId, err := strconv.Atoi(c.Param("type"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге типа объекта"})
	}
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге размера страницы"})
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Subscription{}).
		Where("deleted = FALSE AND subscription_id = ? AND type_id = ?", subObjUUID, typeId).
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count subscriptions by sub-object): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте подписок"})
	}

	var subs []models.Subscription
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Subscription{}).
		Where("deleted = FALSE AND subscription_id = ? AND type_id = ?", subObjUUID, typeId).
		Limit(pageSize).Offset(offset).
		Find(&subs).Error; err != nil {
		log.Printf("DB error (find subscriptions by sub-object): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении подписок"})
	}

	out := response.SubscriptionListBySubObjectResponse{
		SubscriptionId: c.Param("id"),
		TypeId:         int8(typeId),
		Page:           page,
		PageSize:       pageSize,
		TotalCount:     totalCount,
	}
	for _, s := range subs {
		var subscriptionId string
		if s.SubscriptionId != nil {
			subscriptionId = s.SubscriptionId.String()
		}
		out.Subscriptions = append(out.Subscriptions, response.SubscriptionResponse{
			ID:             s.ID.String(),
			UserId:         s.UserID.String(),
			SubscriptionId: subscriptionId,
			TypeId:         utils.GetInt8(s.TypeID),
			CreatedAt:      utils.GetTime(s.CreatedAt),
		})
	}

	return c.JSON(http.StatusOK, out)
}

// GetSubscriptionById godoc
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
func GetSubscriptionById(c echo.Context) error{
	if err := Authorize(c); err != nil { return err }

	subId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор подписки"})
	}

	var subscription models.Subscription
	result := DBConn.Session(&gorm.Session{}).
		Model(&models.Subscription{}).
		Where("deleted = FALSE AND id = ?", subId).
		First(&subscription)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound){
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Подписка не найдена"})
		}
		log.Printf("DB error (find subscription by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении подписки"})
	}

	var subscriptionId string
	if subscription.SubscriptionId != nil {
		subscriptionId = subscription.SubscriptionId.String()
	}
	return c.JSON(http.StatusOK, response.SubscriptionResponse{
		ID:             subscription.ID.String(),
		UserId:         subscription.UserID.String(),
		SubscriptionId: subscriptionId,
		TypeId:         utils.GetInt8(subscription.TypeID),
		CreatedAt:      utils.GetTime(subscription.CreatedAt),
	})
}

// CreateSubscription godoc
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
func CreateSubscription(c echo.Context) error{
	if err := Authorize(c); err != nil { return err }

	var req request.SubscriptionCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	newUUID := uuid.New()
	now := time.Now()
	del := false

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор пользователя"})
	}
	subId, err := uuid.Parse(req.SubscriptionId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор подписки"})
	}
	typeId := *req.TypeId

	switch typeId {
	case 0:
		var task models.Task
		res := DBConn.Session(&gorm.Session{}).
			Model(&models.Task{}).
			Where("id = ? AND deleted = FALSE", subId).
			First(&task)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
			}
			log.Printf("DB error (check task): %v", res.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке задачи"})
		}
	case 1:
		var problem models.Problem
		res := DBConn.Session(&gorm.Session{}).
			Model(&models.Problem{}).
			Where("id = ? AND deleted = FALSE", subId).
			First(&problem)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Проблема не найдена"})
			}
			log.Printf("DB error (check problem): %v", res.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке проблемы"})
		}
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный тип объекта"})
	}

	sub := models.Subscription{
		ID:             newUUID,
		UserID:         userId,
		SubscriptionId: &subId,
		TypeID:         req.TypeId,
		CreatedAt:      &now,
		Deleted:        &del,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		return tx.Session(&gorm.Session{}).
			Model(&models.Subscription{}).
			Create(&sub).Error
	}); txErr != nil {
		log.Printf("DB transaction error (create subscription): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании подписки"})
	}

	return c.JSON(http.StatusCreated, response.SubscriptionUniversalResponse{
		ID:      sub.ID.String(),
		Message: "Подписка создана",
	})
}

// DeleteSubscription godoc
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
func DeleteSubscription(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	subUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор подписки"})
	}

	updateData := map[string]interface{}{"deleted": true}
	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.Subscription{}).
			Where("id = ?", subUUID).
			Updates(updateData)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete subscription): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении подписки"})
	}

	return c.JSON(http.StatusOK, response.SubscriptionUniversalResponse{
		ID:      subUUID.String(),
		Message: "Подписка удалена",
	})
}
