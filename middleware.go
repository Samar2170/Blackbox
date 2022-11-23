package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt"
)

func checkAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Token not found"))
			return
		}
		tokenStringTrunc := strings.Split(tokenString, " ")[1]
		log.Println(tokenString)
		token, err := jwt.ParseWithClaims(tokenStringTrunc, &jwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized" + err.Error()))
			return
		}
		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
		claims, ok := token.Claims.(*jwtCustomClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
		log.Println(claims)
		r.Header.Set("userId", strconv.Itoa(int(claims.Id)))
		next.ServeHTTP(w, r)

	})
}
