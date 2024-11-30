package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Sum struct {
	ID        int       `json:"id"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	dataStore  = make(map[int]Sum)
	currentSum float64
	idCounter  = 0
	mu         sync.Mutex
)

func main() {
	r := gin.Default()

	r.POST("/sum/add", addSum)
	r.GET("/sum", getSum)
	r.GET("/sum/history", getSumHistory)
	r.DELETE("/sum/delete/:id", deleteSum)
	r.Run(":8080")
}

func addSum(c *gin.Context) {
	var newValue struct {
		Value float64 `json:"value"`
	}

	if err := c.ShouldBindJSON(&newValue); err != nil || newValue.Value == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A numeric value is required."})
		return
	}

	mu.Lock()

	idCounter++
	timestamp := time.Now()
	dataStore[idCounter] = Sum{
		ID:        idCounter,
		Value:     newValue.Value,
		Timestamp: timestamp,
	}
	currentSum += newValue.Value

	mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{"id": idCounter, "timestamp": timestamp})
}

func getSum(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"sum": currentSum})
}

func getSumHistory(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	var values []Sum
	for _, v := range dataStore {
		values = append(values, v)
	}

	c.JSON(http.StatusOK, gin.H{"history": values})
}

func deleteSum(c *gin.Context) {
	id := c.Param("id")

	mu.Lock()
	defer mu.Unlock()

	var idInt int
	if _, err := fmt.Sscanf(id, "%d", &idInt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID."})
		return
	}

	if value, exists := dataStore[idInt]; exists {
		currentSum -= value.Value
		delete(dataStore, idInt)
		c.JSON(http.StatusOK, gin.H{"message": "Sum removed successfully."})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID not found."})
	}
}
