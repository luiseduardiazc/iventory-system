package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth valida una API Key simple en el header X-API-Key
func APIKeyAuth(validAPIKeys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "missing X-API-Key header",
			})
			c.Abort()
			return
		}

		// Verificar si la API key es válida
		storeName, valid := validAPIKeys[apiKey]
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "invalid API key",
			})
			c.Abort()
			return
		}

		// Guardar información en el contexto
		c.Set("api_key", apiKey)
		c.Set("store_name", storeName)

		c.Next()
	}
}

// OptionalAPIKeyAuth valida la API key si está presente, pero no la requiere
func OptionalAPIKeyAuth(validAPIKeys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		if apiKey != "" {
			if storeName, valid := validAPIKeys[apiKey]; valid {
				c.Set("api_key", apiKey)
				c.Set("store_name", storeName)
			}
		}

		c.Next()
	}
}
