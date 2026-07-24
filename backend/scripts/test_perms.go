package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := []byte("e83e0b7d5b30893982f9e73988fc3031fa3cee5e5fa7f5fa3c226d8534c7c9b0")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  "USR-ADMIN-001",
		"username": "admin",
		"role_id":  "ROLE-001",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/users/USR-AUDITEE-001/permissions", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(body))
}
