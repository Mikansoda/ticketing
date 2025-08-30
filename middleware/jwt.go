package middleware

import (
	"net/http"
	"strings"
	"time"

	"ticketing/service"
	"github.com/gin-gonic/gin"
)

func Auth(requiredRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(requiredRoles))
	for _, r := range requiredRoles {
		roleSet[r] = struct{}{}
	}
    // Ambil header Authorization
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		// Pastikan formatnya Bearer <token>
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing bearer token",
			})
			return
		}
		// buang prefix "Bearer "
		token := strings.TrimPrefix(auth, "Bearer ")

		// Cek token udah blacklist atau belum.
		if exp, ok := service.AccessBlacklistLookup(token); ok {
			if time.Now().Before(exp) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Token blacklisted",
				})
				return
			}
		}
        // Parse/bongkar isi token → ambil claim (UserID, Role, dsb).
		claims, err := service.ParseAccessForMiddleware(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}
		// Taruh claim ke gin.Context biar bisa dipakai controller.
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// role check
		if len(roleSet) > 0 {
			if _, ok := roleSet[claims.Role]; !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "Forbidden",
				})
				return
			}
		}
		// lanjut ke request kalo passed
		c.Next()
	}
}

