package middleware

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SimpleAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		addr, err := netip.ParseAddr(c.ClientIP())
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if addr.IsLoopback() {
			log.Printf("SimpleAuth: skipping auth for loopback address %v", addr)
			return
		}

		log.Printf("SimpleAuth: path=%s, method=%s", c.Request.URL.Path, c.Request.Method)

		authHeader := c.GetHeader("Authorization")
		log.Printf("SimpleAuth: auth header = %q", authHeader)

		if authHeader == "" {
			log.Printf("SimpleAuth: no auth header")
			c.Header("WWW-Authenticate", `Basic realm="ecms"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Basic ") {
			log.Printf("SimpleAuth: not basic auth")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(authHeader[6:])
		if err != nil {
			log.Printf("SimpleAuth: decode error: %v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			log.Printf("SimpleAuth: invalid parts: %v", parts)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		username := parts[0]
		password := parts[1]
		log.Printf("SimpleAuth: username=%s, password=%s", username, password)

		if !validatePassword(username, password) {
			log.Printf("SimpleAuth: password validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "401000000",
				"message": "Invalid username or password",
			})
			c.Abort()
			return
		}

		log.Printf("SimpleAuth: auth success")
		c.Set("username", username)
	}
}

func validatePassword(username, password string) bool {
	now := time.Now().Unix()
	epochMin := now - now%60

	log.Printf("validatePassword: username=%s, password=%s, now=%d, epochMin=%d", username, password, now, epochMin)

	for _, offset := range []int64{0, -1, 1} {
		epoch := epochMin + 60*offset
		want := generateToken(username, epoch)
		log.Printf("validatePassword: checking epoch=%d, want=%s, match=%v", epoch, want, password == want)
		if password == want {
			return true
		}
	}
	return false
}

func generateToken(username string, epoch int64) string {
	seed := fmt.Sprintf("%s:%d", username, epoch)
	hash := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%x", hash)
}
