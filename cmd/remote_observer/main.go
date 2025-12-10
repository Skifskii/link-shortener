package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Payload struct {
	Message string `json:"message"`
	Value   int    `json:"value"`
}

func main() {
	http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Логируем сырое тело
		log.Println("Raw body:", string(body))

		// Парсим JSON (необязательно, но полезно)
		var data Payload
		if err := json.Unmarshal(body, &data); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Parsed JSON: %+v\n", data)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Received OK")
	})

	log.Println("Server started on :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
