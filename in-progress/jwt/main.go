package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

// shared HMAC secret
var secretKey = []byte("supersecretkey")

type Implant struct {
	ID uint32 `json:"id"`
}

// createToken signs a JWT with the implant ID and 24h expiry
func createToken(id uint32) (string, error) {
	claims := jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// authMiddleware extracts & validates the Bearer token
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if len(auth) < len(prefix) || auth[:len(prefix)] != prefix {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or malformed"})
			return
		}

		tknStr := auth[len(prefix):]
		tkn, err := jwt.Parse(tknStr, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return secretKey, nil
		})
		if err != nil || !tkn.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// you can pull claims out here if you need them:
		// claims := tkn.Claims.(jwt.MapClaims)
		c.Next()
	}
}

func main() {
	r := gin.Default()

	// Health‑check
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"hello": "world"})
	})

	// Login issues a JWT
	r.POST("/login", func(c *gin.Context) {
		var inp Implant
		if err := c.ShouldBindJSON(&inp); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if inp.ID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid implant ID"})
			return
		}

		token, err := createToken(inp.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	// group protected routes under /api/*
	api := r.Group("/api", authMiddleware())
	{
		api.GET("/beacon", func(c *gin.Context) {
			// In a real C2 you'd fetch pending tasks here.
			c.JSON(http.StatusOK, gin.H{"tasks": []string{"do-stuff", "report-status"}})
		})
	}

	// serve TLS
	if err := r.RunTLS(":443", "./tls.crt", "./tls.key"); err != nil {
		panic(err)
	}
}
