package controller

import (
	"net/http"
	"strconv"
	"time"

	"ticketing/entity"
	"ticketing/service"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	service service.CategoryService
}

func NewCategoryController(s service.CategoryService) *CategoryController {
	return &CategoryController{service: s}
}

// Struct request for endpoint create category and update
type createCategoryReq struct {
	Name string `json:"name" binding:"required"`
}

type updateCategoryReq struct {
	Name *string `json:"name,omitempty"`
}

// GET categories
func (ctl *CategoryController) GetCategories(c *gin.Context) {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(offsetStr)

	categories, err := ctl.service.GetCategories(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch categories, try again later",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// Create category (admin only)
func (ctl *CategoryController) CreateCategory(c *gin.Context) {
	var req createCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}

	category := &entity.EventCategories{
		Name: req.Name,
	}
	if err := ctl.service.CreateCategory(c.Request.Context(), category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create category, try again later",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category successfully created",
		"data":    category,
	})
}

// Update category (admin only)
func (ctl *CategoryController) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid category ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	var req updateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}

	existing, err := ctl.service.GetCategoryByIDIncludeDeleted(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Category not found",
			"detail":  err.Error(),
		})
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	existing.UpdatedAt = time.Now()

	if err := ctl.service.UpdateCategory(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update category, try again later",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category successfully updated",
		"data":    existing,
	})
}

// Delete category (admin only)
func (ctl *CategoryController) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid category ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	if err := ctl.service.DeleteCategory(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to delete category, try again later"

		if err.Error() == "category not found" {
			status = http.StatusNotFound
			msg = "Category not found"
		}

		c.JSON(status, gin.H{
			"message": msg,
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Category successfully deleted",
	})
}

// Recover category (admin only)
func (ctl *CategoryController) RecoverCategory(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid category ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	if err := ctl.service.RecoverCategory(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to recover category, try again later"

		if err.Error() == "category not found" {
			status = http.StatusNotFound
			msg = "Category not found"
		}

		c.JSON(status, gin.H{
			"message": msg,
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Category successfully recovered",
	})
}
