package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Employee struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Birthday   time.Time `json:"birthday"`
	Subscribed bool      `json:"subscribed"`
}

var employees []Employee
var nextID = 1

func registerEmployee(w http.ResponseWriter, r *http.Request) {
	var emp Employee
	if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	emp.ID = nextID
	nextID++
	employees = append(employees, emp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(emp)
}

func subscribe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	for i, emp := range employees {
		if emp.ID == id {
			employees[i].Subscribed = true
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Employee not found", http.StatusNotFound)
}

func unsubscribe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	for i, emp := range employees {
		if emp.ID == id {
			employees[i].Subscribed = false
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "Employee not found", http.StatusNotFound)
}

func sendBirthdayNotifications() {
	for _, emp := range employees {
		if emp.Subscribed && emp.Birthday.Month() == time.Now().Month() && emp.Birthday.Day() == time.Now().Day() {
			// Отправить уведомление (тут можно интегрировать отправку уведомлений по email или другим каналам)
			fmt.Printf("Сегодня день рождения у %s!\n", emp.Name)
		}
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Bearer valid-token" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()
	r.Use(authMiddleware)
	r.HandleFunc("/register", registerEmployee).Methods("POST")
	r.HandleFunc("/subscribe/{id}", subscribe).Methods("POST")
	r.HandleFunc("/unsubscribe/{id}", unsubscribe).Methods("POST")

	go func() {
		for {
			time.Sleep(24 * time.Hour)
			sendBirthdayNotifications()
		}
	}()

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", handlers.CORS(handlers.AllowedOrigins([]string{"*"}))(r)); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
