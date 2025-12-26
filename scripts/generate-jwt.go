package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in JWT token
type JWTClaims struct {
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run generate-jwt.go <secret> <user_id> <tenant_id>")
		fmt.Println("Example: go run generate-jwt.go my-secret user-123 tenant-abc")
		os.Exit(1)
	}

	secret := os.Args[1]
	userID := os.Args[2]
	tenantID := os.Args[3]

	// Create claims
	claims := JWTClaims{
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated JWT Token:")
	fmt.Println(tokenString)
	fmt.Println()
	fmt.Println("Token Details:")
	fmt.Printf("  User ID: %s\n", userID)
	fmt.Printf("  Tenant ID: %s\n", tenantID)
	fmt.Printf("  Expires: %s\n", claims.ExpiresAt.Time.Format(time.RFC3339))
	fmt.Println()
	fmt.Println("Example Usage:")
	fmt.Printf("  curl -H \"Authorization: Bearer %s\" http://localhost:8000/api/v1/employees/list\n", tokenString)
}

