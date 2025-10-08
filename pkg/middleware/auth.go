package middleware

import (
	"strings"

	"crm-platform/pkg/config"
	"crm-platform/pkg/errors"
	
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT UTILS

// Extract Bearer token from Authorization header
func extractTokenFromHeader(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

// Validate JWT token signature and extract claims
func validateJWT(token string) (map[string]interface{}, error) {
	parsedToken, err := jwt.Parse(token, func (token *jwt.Token)(interface{}, error) {
		// Ensure HMAC signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrJWT("invalid signing method") 
		}

		// Get shared secret from environment via config package
		secret := config.GetJWTSecret()
		if secret == "" {
			return nil, errors.ErrJWT("JWT secret not configured in environment")
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check token validity
	if !parsedToken.Valid {
		return nil, errors.ErrJWT("token is invalid")
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims) 
	if !ok {
		return nil, errors.ErrJWT("could not parse claims")
	}
	
	return claims, nil
}

// MIDDLEWARE

// Handle user authentication and set context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Since no auth server yet, use headers in development mode
		if config.IsDevelopmentMode() {
			// Get tenant from header (allows dynamic tenant in dev/test)
			tenantID := c.GetHeader("X-Tenant-ID")
			if tenantID == "" {
				tenantID = "01HK153X003BMPJNJB6JHKXK8T" // fallback
			}
			
			userID := c.GetHeader("X-User-ID")
			if userID == "" {
				userID = "dev-user"
			}
			
			c.Set("user_id", userID)
			c.Set("user_permissions", []string{"deals:read", "deals:write"})
			c.Set("tenant_id", tenantID)
			c.Next()
			return
		}

		// Get JWT from header
		token := extractTokenFromHeader(c)
		if token == "" {
			c.JSON(401, gin.H{"error": errors.ErrAuth("authorization header required").Error()})
			c.Abort()
			return
		}

		// Validate and sign token, receive claims
		claims, err := validateJWT(token)
		if err != nil {
			c.JSON(401, gin.H{"error": errors.ErrAuth("token validation failed").Error()})
			c.Abort()
			return
		}
		
		// Set claims in context for handlers to use
		c.Set("user_id", claims["user_id"])
		c.Set("user_permissions", claims["user_permissions"])
		c.Set("tenant_id", claims["tenant_id"])
		c.Next()
	}
}

// CONTEXT EXTRACTION

// Get authenticated user ID from request context
func ExtractUserId(c *gin.Context) string {
	userId, exists := c.Get("user_id")

	if !exists {
		return ""
	}

	if id, ok := userId.(string); ok {
		return id
	}

	return ""
}

// Get user permissions from request context
func ExtractPermissions(c *gin.Context) []string {
	perms, exists := c.Get("user_permissions")

	if !exists {
		return []string{}
	}

	if permissions, ok := perms.([]string); ok {
		return permissions
	}
	
	return []string{}
}

// Check if user has specific permission
func HasPermission(c *gin.Context, permission string) bool {
	permissions := ExtractPermissions(c)

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}

	return false
}