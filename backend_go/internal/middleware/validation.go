package middleware

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateJSON validates the request body against the provided struct
func ValidateJSON(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request body is empty
		if c.Request.ContentLength == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []ValidationError{
					{
						Field:   "request",
						Message: "Request body cannot be empty",
					},
				},
			})
			c.Abort()
			return
		}

		// Read the raw body
		rawBody, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []ValidationError{
					{
						Field:   "request",
						Message: "Failed to read request body",
					},
				},
			})
			c.Abort()
			return
		}

		// Check if body is empty JSON object or whitespace
		trimmedBody := strings.TrimSpace(string(rawBody))
		if trimmedBody == "{}" || trimmedBody == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []ValidationError{
					{
						Field:   "request",
						Message: "Request body cannot be empty",
					},
				},
			})
			c.Abort()
			return
		}

		// Reset the request body for ShouldBindJSON
		c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

		// Create a new instance of the model
		newModel := reflect.New(reflect.TypeOf(model).Elem()).Interface()

		if err := c.ShouldBindJSON(newModel); err != nil {
			var errors []ValidationError
			if ve, ok := err.(validator.ValidationErrors); ok {
				for _, e := range ve {
					errors = append(errors, ValidationError{
						Field:   e.Field(),
						Message: getErrorMsg(e),
					})
				}
			} else {
				errors = append(errors, ValidationError{
					Field:   "request",
					Message: "Invalid JSON format",
				})
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errors,
			})
			c.Abort()
			return
		}

		// Check if any required fields are missing
		val := reflect.ValueOf(newModel).Elem()
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			if field.Tag.Get("binding") == "required" && val.Field(i).IsZero() {
				c.JSON(http.StatusBadRequest, gin.H{
					"errors": []ValidationError{
						{
							Field:   field.Name,
							Message: "This field is required",
						},
					},
				})
				c.Abort()
				return
			}
		}

		c.Set("validated", newModel)
		c.Next()
	}
}

// ValidateQueryParams validates query parameters
func ValidateQueryParams(params interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindQuery(params); err != nil {
			var errors []ValidationError
			if ve, ok := err.(validator.ValidationErrors); ok {
				for _, e := range ve {
					errors = append(errors, ValidationError{
						Field:   e.Field(),
						Message: getErrorMsg(e),
					})
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errors,
			})
			c.Abort()
			return
		}
		c.Set("validated_params", params)
		c.Next()
	}
}

// getErrorMsg returns a more user-friendly error message
func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value must be greater than " + fe.Param()
	case "max":
		return "Value must be less than " + fe.Param()
	default:
		return "Invalid value"
	}
}

// Common validation structs
type PaginationParams struct {
	Page         int `form:"page" binding:"required,min=1"`
	ItemsPerPage int `form:"items_per_page" binding:"required,min=1,max=100"`
}

type WordReviewRequest struct {
	Correct *bool `json:"correct" binding:"required"`
}

type StudySessionRequest struct {
	GroupID         int `json:"group_id" binding:"required"`
	StudyActivityID int `json:"study_activity_id" binding:"required"`
}

type StudyActivityRequest struct {
	GroupID int    `json:"group_id" binding:"required"`
	Name    string `json:"name" binding:"required"`
}

// WordRequest represents word creation request
type WordRequest struct {
	Tamil   string                 `json:"tamil" binding:"required"`
	Romaji  string                 `json:"romaji" binding:"required"`
	English string                 `json:"english" binding:"required"`
	Parts   map[string]interface{} `json:"parts" binding:"required"`
}

// GroupRequest represents group creation request
type GroupRequest struct {
	Name string `json:"name" binding:"required"`
}
