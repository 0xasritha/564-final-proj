package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 32‑byte AES key (in prod, load from ENV or KMS)
var aesKey = []byte("0123456789abcdef0123456789abcdef")

type tokenPayload struct {
	ID  uint32 `json:"id"`
	Exp int64  `json:"exp"`
}

type Implant struct {
	ID uint32 `json:"id"`
}

// createToken encrypts id+expiry into a compact, base64 URL‑safe string
func createToken(id uint32) (string, error) {
	// prepare plaintext JSON
	pl := tokenPayload{ID: id, Exp: time.Now().Add(24 * time.Hour).Unix()}
	plain, err := json.Marshal(pl)
	if err != nil {
		return "", err
	}

	// AES‑GCM setup
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// allocate nonce + ciphertext
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nil, nonce, plain, nil)

	// prefix nonce to ciphertext, then base64‑URL encode
	token := append(nonce, ct...)
	return base64.RawURLEncoding.EncodeToString(token), nil
}

// verifyToken reverses createToken: decrypts, checks expiry, returns the ID
func verifyToken(tok string) (uint32, error) {
	data, err := base64.RawURLEncoding.DecodeString(tok)
	if err != nil {
		return 0, err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return 0, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return 0, errors.New("malformed token")
	}
	nonce, ct := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return 0, err
	}

	var pl tokenPayload
	if err := json.Unmarshal(plain, &pl); err != nil {
		return 0, err
	}
	if time.Now().Unix() > pl.Exp {
		return 0, errors.New("token expired")
	}
	return pl.ID, nil
}

// authMiddleware pulls Bearer token, calls verifyToken, aborts on error
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		const prefix = "Bearer "
		h := c.GetHeader("Authorization")
		if len(h) <= len(prefix) || h[:len(prefix)] != prefix {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed auth header"})
			return
		}
		id, err := verifyToken(h[len(prefix):])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// store ID if you need it later: c.Set("implantID", id)
		c.Next()
	}
}

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"hello": "world"})
	})

	// issue AES‑GCM token
	r.POST("/login", func(c *gin.Context) {
		var inp Implant
		if err := c.ShouldBindJSON(&inp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if inp.ID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid implant ID"})
			return
		}
		tok, err := createToken(inp.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tok})
	})

	// protected routes
	api := r.Group("/api", authMiddleware())
	api.GET("/beacon", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tasks": []string{"do-stuff", "report-status"}})
	})

	if err := r.RunTLS(":443", "./tls.crt", "./tls.key"); err != nil {
		panic(err)
	}
}
