package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func phash(pass string) string {
	h := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(h[:])
}

// AuthMiddleware возвращает middleware с учётом конфигурации
func AuthMiddleware(cfg Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// если пароль не задан — аутентификация отключена
			if cfg.Password == "" {
				next(w, r)
				return
			}

			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !validateJWT(cookie.Value, cfg.Password) {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			next(w, r)
		}
	}
}

func makeJWT(pass string) (string, error) {
	claims := jwt.MapClaims{
		"phash": phash(pass),
		"exp":   time.Now().Add(8 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(pass))
}

func validateJWT(tokenString, pass string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return []byte(pass), nil
	})
	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	p, _ := claims["phash"].(string)
	return p == phash(pass)
}
