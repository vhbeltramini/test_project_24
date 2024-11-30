package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Value struct {
	ID        int       `json:"id"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	dataStore  = make(map[int]Value)
	currentSum float64
	idCounter  = 0
	mu         sync.Mutex
)

func main() {
	r := gin.Default()

	// Route to add a value
	r.POST("/sum/add", addValue)

	// Route to get the sum
	r.GET("/sum", getSum)

	r.GET("/sum/history", getSumHistory)

	// Route to delete a value
	r.DELETE("/sum/delete/:id", deleteValue)

	r.Run(":8080") // Start the server on port 8080
}

// Handler to add a value
func addValue(c *gin.Context) {
	var newValue struct {
		Value float64 `json:"value"`
	}

	if err := c.ShouldBindJSON(&newValue); err != nil || newValue.Value == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A numeric value is required."})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	idCounter++
	timestamp := time.Now()
	dataStore[idCounter] = Value{
		ID:        idCounter,
		Value:     newValue.Value,
		Timestamp: timestamp,
	}
	currentSum += newValue.Value

	c.JSON(http.StatusCreated, gin.H{"id": idCounter, "timestamp": timestamp})
}

// Handler to get the current sum
func getSum(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"sum": currentSum})
}

// Handler to get the current sum
func getSumHistory(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	var values []Value
	for _, v := range dataStore {
		values = append(values, v)
	}

	// Convert the slice to JSON
	jsonData, err := json.Marshal(values)
	if err != nil {
		errWraper, _ := fmt.Println("Error converting to JSON:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errWraper})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": string(jsonData)})
}

// Handler to delete a value by ID
func deleteValue(c *gin.Context) {
	id := c.Param("id")

	mu.Lock()
	defer mu.Unlock()

	// Convert id to integer
	var idInt int
	if _, err := fmt.Sscanf(id, "%d", &idInt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID."})
		return
	}

	if value, exists := dataStore[idInt]; exists {
		currentSum -= value.Value
		delete(dataStore, idInt)
		c.JSON(http.StatusOK, gin.H{"message": "Value removed successfully."})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID not found."})
	}
}
