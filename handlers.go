package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handlers struct {
	receipts map[string]Receipt
	points   map[string]int
	mutex    sync.RWMutex
}

func NewHandlers() *Handlers {
	return &Handlers{
		receipts: make(map[string]Receipt),
		points:   make(map[string]int),
	}
}

func (h *Handlers) ProcessReceipt(w http.ResponseWriter, r *http.Request) {

	var receipt Receipt

	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	// Validate receipt
	if err := validateReceipt(receipt); err != nil {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest)
		return
	}

	// Generate ID
	id := uuid.New().String()

	// Calculate points
	points := calculatePoints(receipt)

	// Store receipt and points
	h.mutex.Lock()
	h.receipts[id] = receipt
	h.points[id] = points
	h.mutex.Unlock()

	// Set status code to 200
	w.WriteHeader(http.StatusOK)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProcessResponse{ID: id})
}

func (h *Handlers) GetPoints(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	h.mutex.RLock()
	points, exists := h.points[id]
	h.mutex.RUnlock()

	if !exists {
		http.Error(w, "No receipt found for that ID.", http.StatusNotFound)
		return
	}

	// Set status code to 200
	w.WriteHeader(http.StatusOK)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PointsResponse{Points: points})
}

func validateReceipt(receipt Receipt) error {

	if !regexp.MustCompile(`^[\w\s\-&]+$`).MatchString(receipt.Retailer) {
		return fmt.Errorf("invalid retailer")
	}

	if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
		return fmt.Errorf("invalid date")
	}

	if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
		return fmt.Errorf("invalid time")
	}

	if !regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(receipt.Total) {
		return fmt.Errorf("invalid total")
	}

	if len(receipt.Items) < 1 {
		return fmt.Errorf("receipt must have at least one item")
	}

	for _, item := range receipt.Items {
		if !regexp.MustCompile(`^[\w\s\-]+$`).MatchString(item.ShortDescription) {
			return fmt.Errorf("invalid item description")
		}
		if !regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(item.Price) {
			return fmt.Errorf("invalid item price")
		}
	}

	return nil
}

func calculatePoints(receipt Receipt) int {

	total := 0

	// 1: One point for every alphanumeric character in the retailer name
	for _, char := range receipt.Retailer {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			total++
		}
	}

	// 2: 50 points if the total is a round dollar amount
	if strings.HasSuffix(receipt.Total, ".00") {
		total += 50
	}

	// 3: 25 points if the total is a multiple of 0.25
	if totalFloat, err := strconv.ParseFloat(receipt.Total, 64); err == nil {
		if math.Mod(totalFloat*100, 25) == 0 {
			total += 25
		}
	}

	// 4: 5 points for every two items
	total += (len(receipt.Items) / 2) * 5

	// 5: Points for items with description length multiple of 3
	for _, item := range receipt.Items {
		trimmedLen := len(strings.TrimSpace(item.ShortDescription))
		if trimmedLen%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			total += int(math.Ceil(price * 0.2))
		}
	}

	// 6: 6 points if the day in the purchase date is odd
	if purchaseDate, err := time.Parse("2006-01-02", receipt.PurchaseDate); err == nil {
		if purchaseDate.Day()%2 == 1 {
			total += 6
		}
	}

	// 7: 10 points if purchase time is between 2:00pm and 4:00pm
	if purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime); err == nil {
		startTime, _ := time.Parse("15:04", "14:00")
		endTime, _ := time.Parse("15:04", "16:00")
		if purchaseTime.After(startTime) && purchaseTime.Before(endTime) {
			total += 10
		}
	}

	return total
}
