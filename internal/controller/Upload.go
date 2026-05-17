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
	g.POST("/avatar/:user_id", ctrl.UploadAvatar)
}

// UploadImage godoc
// @Summary Загрузить изображение
// @Description Загружает файл в S3/RustFS и возвращает публичный URL
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Файл изображения"
// @Security BearerAuth
// @Success 200 {object} map[string]string "url"
// @Router /upload/image [post]
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

	url, err := u.storageService.UploadFile(file, fileHeader, "images")
	if err != nil {
		log.Printf("upload image error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка загрузки файла"})
	}

	return c.JSON(http.StatusOK, map[string]string{"url": url})
}

// UploadAvatar godoc
// @Summary Загрузить аватар пользователя
// @Description Загружает аватар в S3/RustFS и обновляет URL в профиле
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param user_id path string true "UUID пользователя"
// @Param file formData file true "Файл аватара"
// @Security BearerAuth
// @Success 200 {object} map[string]string "url"
// @Router /upload/avatar/{user_id} [post]
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

	avatarURL, err := u.storageService.UploadFile(file, fileHeader, "avatars")
	if err != nil {
		log.Printf("upload avatar error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ошибка загрузки аватара"})
	}

	// Сохраняем URL в профиле пользователя
	if err := u.userService.UpdateAvatarURL(userID, avatarURL); err != nil {
		log.Printf("update avatar url error: %v", err)
		// Не фатально — URL всё равно возвращаем
	}

	return c.JSON(http.StatusOK, map[string]string{"url": avatarURL})
}
