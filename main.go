package main

import (
		"encoding/json"
		"log"
		"math"
		"net/http"
		"regexp"
		"strconv"
		"strings"
		"time"

		"github.com/google/uuid"
		"github.com/gorilla/mux"
)

type Receipt struct {
		Retailer     string `json:"retailer"`
		PurchaseDate string `json:"purchaseDate"`
		PurchaseTime string `json:"purchaseTime"`
		Items        []Item `json:"items"`
		Total        string `json:"total"`
}

type Item struct {
		ShortDescription string `json:"shortDescription"`
		Price            string `json:"price"`
}

type ProcessResponse struct {
		ID string `json:"id"`
}

type PointsResponse struct {
		Points int `json:"points"`
}

var receipts = make(map[string]Receipt)
var points = make(map[string]int)

func main() {
		r := mux.NewRouter()

		r.HandleFunc("/receipts/process", processReceipt).Methods("POST")
		r.HandleFunc("/receipts/{id}/points", getPoints).Methods("GET")

		log.Fatal(http.ListenAndServe(":8080", r))
}

func processReceipt(w http.ResponseWriter, r *http.Request) {
		var receipt Receipt
		err := json.NewDecoder(r.Body).Decode(&receipt)
		if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
		}

		id := uuid.New().String()
		receipts[id] = receipt
		points[id] = calculatePoints(receipt)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ProcessResponse{ID: id})
}

func getPoints(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		point, exists := points[id]
		if !exists {
				http.Error(w, "Receipt not found", http.StatusNotFound)
				return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PointsResponse{Points: point})
}

func calculatePoints(receipt Receipt) int {
		points := 0

		// Rule 1: One point for every alphanumeric character in the retailer name
		points += len(regexp.MustCompile(`[a-zA-Z0-9]`).FindAllString(receipt.Retailer, -1))

		// Rule 2: 50 points if the total is a round dollar amount with no cents
		total, _ := strconv.ParseFloat(receipt.Total, 64)
		if total == math.Floor(total) {
				points += 50
		}

		// Rule 3: 25 points if the total is a multiple of 0.25
		if math.Mod(total, 0.25) == 0 {
				points += 25
		}

		// Rule 4: 5 points for every two items on the receipt
		points += (len(receipt.Items) / 2) * 5

		// Rule 5: Points based on item description length
		for _, item := range receipt.Items {
				descLen := len(strings.TrimSpace(item.ShortDescription))
				if descLen%3 == 0 {
						price, _ := strconv.ParseFloat(item.Price, 64)
						points += int(math.Ceil(price * 0.2))
				}
		}

		// Rule 6: 6 points if the day in the purchase date is odd
		date, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
		if date.Day()%2 != 0 {
				points += 6
		}

		// Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm
		t, _ := time.Parse("15:04", receipt.PurchaseTime)
		if t.Hour() == 14 || (t.Hour() == 15 && t.Minute() < 60) {
				points += 10
		}

		return points
}
