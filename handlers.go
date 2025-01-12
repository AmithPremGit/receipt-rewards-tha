package main

import (
	"net/http"
	"sync"
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

// Add the function signatures - TODO: implement
func (h *Handlers) ProcessReceipt(w http.ResponseWriter, r *http.Request) {}
func (h *Handlers) GetPoints(w http.ResponseWriter, r *http.Request)      {}
func calculatePoints(receipt Receipt) int                                 {}
func validateReceipt(receipt Receipt) error                               {}
