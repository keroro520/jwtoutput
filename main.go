package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

func JWTAuthToken(jwtSecret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": &jwt.NumericDate{Time: time.Now()},
	})
	signedToken, err := token.SignedString(jwtSecret[:])
	if err != nil {
		return "", err
	}

	// Include the signed JWT as a Bearer token in the Authorization header
	return fmt.Sprintf("Bearer %s", signedToken), nil
}

func main() {
	jwtSecretFlag := flag.String("jwt-secret", "", "JWT secret")
	flag.Parse()

	jwtSecret := common.FromHex(*jwtSecretFlag)
	jwtToken, _ := JWTAuthToken(jwtSecret)

    fmt.Printf("Authorization: %s", jwtToken)
}
