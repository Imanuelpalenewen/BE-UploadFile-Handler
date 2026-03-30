package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func CheckAPI(route http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		validKey := os.Getenv("API_KEY")
		apiKey := r.Header.Get("X-API-Key")

		if apiKey == "" || apiKey != validKey {
			fmt.Println("Unauthorized API access attempt")
			res := struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    401,
				Message: "Unauthorized Access",
			}

			jsonData, err := json.Marshal(res)
			if err != nil {
				fmt.Println(err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)
			return
		}

		route.ServeHTTP(w, r)
	}
}
