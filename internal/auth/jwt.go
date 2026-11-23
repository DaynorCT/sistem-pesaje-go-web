package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	cookieAccess  = "access_token"
	cookieRefresh = "refresh_token"
)

type Claims struct {
	UserID int64  `json:"userId"`
	Email  string `json:"email"`
	Rol    string `json:"rol"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID int64 `json:"userId"`
	jwt.RegisteredClaims
}

func CreateAccessToken(secret string, userID int64, email, rol string) (string, error) {
	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		Email:  email,
		Rol:    rol,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	})
	return t.SignedString([]byte(secret))
}

func CreateRefreshToken(secret string, userID int64) (string, error) {
	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	})
	return t.SignedString([]byte(secret))
}

func VerifyAccess(secret, token string) (*Claims, error) {
	p, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("algoritmo inválido")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := p.Claims.(*Claims)
	if !ok || !p.Valid {
		return nil, errors.New("token inválido")
	}
	return c, nil
}

func VerifyRefresh(secret, token string) (int64, error) {
	p, err := jwt.ParseWithClaims(token, &RefreshClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	c, ok := p.Claims.(*RefreshClaims)
	if !ok || !p.Valid {
		return 0, errors.New("refresh inválido")
	}
	return c.UserID, nil
}

func UserIDFromRequest(r *http.Request, secret string) (int64, bool) {
	tok := CookieValue(r, cookieAccess)
	if tok == "" {
		return 0, false
	}
	c, err := VerifyAccess(secret, tok)
	if err != nil {
		return 0, false
	}
	return c.UserID, true
}

func CookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil || c.Value == "" {
		// Cabecera Cookie sin http.Cookie (cliente escritorio)
		h := r.Header.Get("Cookie")
		if h == "" {
			return ""
		}
		prefix := name + "="
		for _, part := range strings.Split(h, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, prefix) {
				return strings.TrimPrefix(part, prefix)
			}
		}
		return ""
	}
	return c.Value
}

func SetAuthCookies(w http.ResponseWriter, access, refresh string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieAccess,
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   15 * 60,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     cookieRefresh,
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60,
	})
}
