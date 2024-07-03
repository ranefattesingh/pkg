package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	signingMethod jwt.SigningMethod
	secretKey     string
}

func NewJWT(sm jwt.SigningMethod, sk string) *JWT {
	return &JWT{
		signingMethod: sm,
		secretKey:     sk,
	}
}

// Generate new JWT Token.
func (j *JWT) GenerateToken(c jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(j.signingMethod, c)

	// Register the JWT string
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *JWT) Parse(ts string) (jwt.MapClaims, bool, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(ts, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, false, err
	}

	return claims, token.Valid, nil
}
