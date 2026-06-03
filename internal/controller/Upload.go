package controller

import (
	"emplacc-api/internal/service"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UploadController struct {
	storageService service.StorageService
	userService    service.UserService
}

func NewUploadController(storageService service.StorageService, userService service.UserService) *UploadController {
	return &UploadController{storageService: storageService, userService: userService}
}

func RegisterUploadRoutes(e *echo.Echo, storageService service.StorageService, userService service.UserService) {
	ctrl := NewUploadController(storageService, userService)
	g := e.Group("/upload")
	g.POST("/image", ctrl.UploadImage)
	g.POST("/file", ctrl.UploadFile)
	g.POST("/avatar/:user_id", ctrl.UploadAvatar)
	g.GET("/refresh", ctrl.RefreshURL)
}

func (u *UploadController) UploadImage(c echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "файл не найден в запросе"})
	}
	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "не удалось открыть файл"})
	}
	defer file.Close()

	presignedURL, objectPath, err := u.storageService.UploadFile(file, fileHeader, "images")
	if err != nil {
		log.Printf("upload image error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка загрузки файла"})
	}
	return c.JSON(http.StatusOK, map[string]string{"url": presignedURL, "path": objectPath})
}

func (u *UploadController) UploadFile(c echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "файл не найден в запросе"})
	}
	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "не удалось открыть файл"})
	}
	defer file.Close()

	presignedURL, objectPath, err := u.storageService.UploadFile(file, fileHeader, "files")
	if err != nil {
		log.Printf("upload file error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка загрузки файла"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"url":          presignedURL,
		"path":         objectPath,
		"name":         fileHeader.Filename,
		"size":         fileHeader.Size,
		"content_type": fileHeader.Header.Get("Content-Type"),
	})
}

func (u *UploadController) UploadAvatar(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный user_id"})
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "файл не найден в запросе"})
	}
	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "не удалось открыть файл"})
	}
	defer file.Close()

	presignedURL, objectPath, err := u.storageService.UploadFile(file, fileHeader, "avatars")
	if err != nil {
		log.Printf("upload avatar error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка загрузки аватара"})
	}
	// Храним objectPath (путь), а не presigned URL — URL истекает через 7 дней,
	// путь — вечный. FreshAvatarURL генерирует свежий URL при каждом запросе.
	if err := u.userService.UpdateAvatarURL(userID, objectPath); err != nil {
		log.Printf("update avatar url error: %v", err)
	}
	return c.JSON(http.StatusOK, map[string]string{"url": presignedURL})
}

// RefreshURL godoc
// @Summary Обновить presigned URL для файла
// @Description Генерирует новый presigned URL для объекта по его пути
// @Tags Upload
// @Produce json
// @Param path query string true "Путь к объекту (например: images/uuid.jpg)"
// @Security BearerAuth
// @Success 200 {object} map[string]string "url"
// @Router /upload/refresh [get]
func (u *UploadController) RefreshURL(c echo.Context) error {
	objectPath := c.QueryParam("path")
	if objectPath == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "path обязателен"})
	}
	newURL, err := u.storageService.RefreshURL(objectPath)
	if err != nil {
		log.Printf("refresh url error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка обновления ссылки"})
	}
	return c.JSON(http.StatusOK, map[string]string{"url": newURL})
}
