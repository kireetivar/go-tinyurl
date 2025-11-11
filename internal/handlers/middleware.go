package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (s *Server) ShortenMiddleware(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}
	tokenString := parts[1]

	_ ,err := s.ValidateToken(tokenString)
	if err!=nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	s.ShortenHandler(w, r)
}

// validateToken is a helper that parses and validates a JWT string.
func (s *Server) ValidateToken(tokenString string) (*jwt.Token, error) {
	// 1. Define the Keyfunc. This function is called by the parser
	//    to get the secret key for validation.
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// 2. IMPORTANT: Check the token's signing method.
		//    Ensure it's the one you used (HS256).
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// 3. If the method is correct, return your server's secret key.
		return []byte(s.jwtSecret), nil
	}

	// 4. Parse the token.
	//    This will check the signature, expiration, and "issued at" time.
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, err
	}

	// 5. Final check that the token is valid.
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}