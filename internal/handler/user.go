package handler

import (
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"college-graduation-project-backend/internal/service"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userService  service.UserService
	imageService service.ImageService
}

func NewUserHandler(userService service.UserService, imageService service.ImageService) *UserHandler {
	return &UserHandler{userService: userService, imageService: imageService}
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Возвращает профиль текущего авторизованного пользователя.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.UserShort
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/users/me [get]
func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	userId, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	user, err := h.userService.FindByID(userId)
	if err != nil {
		return err
	}
	userShort := response.NewUserShort(user.ID, user.FullName, user.Email)

	return c.Status(fiber.StatusOK).JSON(userShort)
}

// GetPublicProfile godoc
// @Summary Get user public profile
// @Description Возвращает краткую публичную информацию о пользователе: id, full_name и avatar.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID пользователя"
// @Success 200 {object} response.UserPublicShort
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/users/{id} [get]
func (h *UserHandler) GetPublicProfile(c fiber.Ctx) error {
	id := fiber.Params[uint64](c, "id")

	user, err := h.userService.FindByID(id)
	if err != nil {
		return err
	}

	avatarPath, err := h.imageService.GetUserAvatar(id)
	if err != nil {
		return err
	}

	userPublic := response.NewUserPublicShort(user.ID, user.FullName, avatarPath)
	return c.Status(fiber.StatusOK).JSON(userPublic)
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Обновляет профиль текущего авторизованного пользователя.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body request.UpdateProfile true "Поля профиля для обновления"
// @Success 200 {object} response.UserShort
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/users/me [patch]
func (h *UserHandler) UpdateProfile(c fiber.Ctx) error {
	var updateProfile request.UpdateProfile
	if err := c.Bind().Body(&updateProfile); err != nil {
		return err
	}
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	user, err := h.userService.UpdateProfile(userID, updateProfile)
	if err != nil {
		return err
	}
	userShort := response.NewUserShort(user.ID, user.FullName, user.Email)
	return c.Status(fiber.StatusOK).JSON(userShort)
}

// GetRole godoc
// @Summary Get current user role
// @Description Возвращает роль текущего авторизованного пользователя.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Role
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/users/me/role [get]
func (h *UserHandler) GetRole(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}
	user, err := h.userService.FindByID(userID)
	if err != nil {
		return err
	}
	userRole := response.Role{Role: user.Role.String()}
	return c.Status(fiber.StatusOK).JSON(userRole)
}

// UploadAvatar godoc
// @Summary Upload current user avatar
// @Description Загружает/обновляет аватар текущего пользователя (jpg/jpeg/png/gif/webp, до 10MB).
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param image formData file true "Файл изображения"
// @Success 201 {object} UploadImageResponse
// @Failure 400 {object} BadRequestErrorResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/users/me/avatar [put]
func (h *UserHandler) UploadAvatar(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("image")
	if err != nil {
		return errs.BadRequest("Invalid file", "No file uploaded")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		return errs.BadRequest("Invalid file type", "Only jpg, jpeg, png, gif, webp files are allowed")
	}

	if file.Size > 10*1024*1024 {
		return errs.BadRequest("File too large", "Maximum file size is 10MB")
	}

	uploadsDir := "uploads/images"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return errs.InternalServerError("Cannot create upload directory", err.Error())
	}

	filename := fmt.Sprintf("user_%d_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadsDir, filename)
	if err := c.SaveFile(file, filePath); err != nil {
		return errs.InternalServerError("Cannot save file", err.Error())
	}

	imagePath := fmt.Sprintf("/api/images/%s", filename)
	if err := h.imageService.SetUserAvatar(userID, imagePath); err != nil {
		_ = os.Remove(filePath)
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"image_path": imagePath,
	})
}

// GetAvatar godoc
// @Summary Get current user avatar
// @Description Возвращает путь к аватару текущего пользователя.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UploadImageResponse
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/users/me/avatar [get]
func (h *UserHandler) GetAvatar(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	avatarPath, err := h.imageService.GetUserAvatar(userID)
	if err != nil {
		return err
	}

	if avatarPath == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"image_path": nil})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"image_path": *avatarPath,
	})
}

// DeleteAvatar godoc
// @Summary Delete current user avatar
// @Description Удаляет аватар текущего пользователя.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 204 {string} string "No Content"
// @Failure 401 {object} UnauthorizedErrorResponse
// @Failure 404 {object} NotFoundErrorResponse
// @Failure 500 {object} InternalServerErrorResponse
// @Router /api/users/me/avatar [delete]
func (h *UserHandler) DeleteAvatar(c fiber.Ctx) error {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		return err
	}

	avatarPath, err := h.imageService.GetUserAvatar(userID)
	if err != nil {
		return err
	}

	if err := h.imageService.DeleteUserAvatar(userID); err != nil {
		return err
	}

	if avatarPath != nil {
		if filename, extractErr := imageFilenameFromPath(*avatarPath); extractErr == nil && filename != "" {
			_ = os.Remove(filepath.Join("uploads/images", filename))
		}
	}

	return c.Status(fiber.StatusNoContent).JSON(nil)
}

func imageFilenameFromPath(imagePath string) (string, error) {
	parsed, err := url.Parse(imagePath)
	if err != nil {
		return "", err
	}
	return filepath.Base(parsed.Path), nil
}
