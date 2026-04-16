package middleware

import (
	"net/http"
	"strings"
	"time"

	"certhub-backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint64, email, role string) (string, error) {
	expireAt := time.Now().Add(time.Duration(config.C.JWT.ExpireHours) * time.Hour)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.JWT.Secret))
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "未授权",
				"data":    nil,
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Token 格式错误",
				"data":    nil,
			})
			return
		}

		tokenStr := parts[1]
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.C.JWT.Secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Token 无效或已过期",
				"data":    nil,
			})
			return
		}

		if claims, ok := token.Claims.(*Claims); ok {
			c.Set("userID", claims.UserID)
			c.Set("userEmail", claims.Email)
			c.Set("userRole", claims.Role)
		}

		c.Next()
	}
}

// AdminOnly ensures that only admin users can access the route.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleAny, exists := c.Get("userRole")
		if !exists || roleAny.(string) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "权限不足",
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}


