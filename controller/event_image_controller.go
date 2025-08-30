package controller

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"ticketing/service"
)

type EventImageController struct {
	service service.EventImageService
}

func NewEventImageController(s service.EventImageService) *EventImageController {
	return &EventImageController{service: s}
}

// Upload image (admin only)
func (ctl *EventImageController) UploadImage(c *gin.Context) {
	eventID := c.Param("eventId")
	var pid uint
	if _, err := fmt.Sscan(eventID, &pid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid event ID",
			"detail":  err.Error(),
		})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Missing image",
			"detail":  err.Error(),
		})
		return
	}

	isPrimary := c.DefaultPostForm("is_primary", "false") == "true"

	tempPath := filepath.Join("/tmp", file.Filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to save image, try again later",
			"detail":  err.Error(),
		})
		return
	}

	img, err := ctl.service.Upload(c, pid, tempPath, isPrimary)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to upload image",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Image successfully uploaded",
		"detail":  img,
	})
}

// Delete image (admin only)
func (ctl *EventImageController) DeleteImage(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscan(c.Param("imageId"), &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid image ID",
			"detail":  err.Error(),
		})
		return
	}

	if err := ctl.service.Delete(c, id); err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to delete image, try again later"

		if err.Error() == "image not found" {
			status = http.StatusNotFound
			msg = "Image not found"
		}

		c.JSON(status, gin.H{
			"message": msg,
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image successfully deleted",
	})
}

// Recover image (admin only)
func (ctl *EventImageController) RecoverImage(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscan(c.Param("imageId"), &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid image ID",
			"detail":  err.Error(),
		})
		return
	}

	if err := ctl.service.Recover(c, id); err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to recover image, try again later"

		if err.Error() == "image not found" {
			status = http.StatusNotFound
			msg = "Image not found"
		}

		c.JSON(status, gin.H{
			"message": msg,
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image successfully recovered",
	})
}
