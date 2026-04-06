package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS returns a gin middleware that adds permissive CORS headers for vitacoin.network.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := false
		for _, o := range []string{
			"https://vitacoin.network",
			"https://www.vitacoin.network",
			"https://app.vitacoin.network",
			"http://localhost:3000",
			"http://localhost:3001",
		} {
			if origin == o {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			// Non-browser / server-to-server
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
