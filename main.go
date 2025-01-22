package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type KeyValue struct {
	ID        int       `json:"id"`
	Key       string    `json:"key" binding:"required"`
	Value     string    `json:"value" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var db *sql.DB

func main() {
	os.MkdirAll("data", 0755)
	var err error
	db, err = sql.Open("sqlite3", "data/kvstore.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS key_values (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_key ON key_values(key)")
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_value ON key_values(value)")

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	router.GET("/", indexHandler)
	router.GET("/api/entries", listEntries)
	router.POST("/api/entries", createEntry)
	router.GET("/api/entries/:id", getEntry)
	router.PUT("/api/entries/:id", updateEntry)
	router.DELETE("/api/entries/:id", deleteEntry)
	router.POST("/api/entries/generate-dummy", generateDummyData)
	router.POST("/api/entries/truncate", truncateDB)

	log.Fatal(router.Run(":8080"))
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func listEntries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	search := c.Query("search")
	sortColumn := c.DefaultQuery("sort", "updated_at")
	sortOrder := c.DefaultQuery("order", "desc")

	validColumns := map[string]bool{"key": true, "value": true, "created_at": true, "updated_at": true}
	validOrders := map[string]bool{"asc": true, "desc": true}

	if !validColumns[sortColumn] || !validOrders[sortOrder] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort parameters"})
		return
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, key, value, created_at, updated_at 
		FROM key_values 
		WHERE key LIKE ? OR value LIKE ?
		ORDER BY %s %s
		LIMIT ? OFFSET ?`, sortColumn, sortOrder)

	searchTerm := "%" + search + "%"
	rows, err := db.Query(query, searchTerm, searchTerm, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var entries []KeyValue
	for rows.Next() {
		var entry KeyValue
		err := rows.Scan(&entry.ID, &entry.Key, &entry.Value, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		entries = append(entries, entry)
	}

	if entries == nil {
		entries = make([]KeyValue, 0)
	}

	c.JSON(http.StatusOK, entries)
}

func createEntry(c *gin.Context) {
	var entry KeyValue
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("INSERT OR IGNORE INTO key_values (key, value) VALUES (?, ?)", entry.Key, entry.Value)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "Key already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	entry.ID = int(id)
	c.JSON(http.StatusCreated, entry)
}

func getEntry(c *gin.Context) {
	id := c.Param("id")
	var entry KeyValue

	err := db.QueryRow("SELECT id, key, value, created_at, updated_at FROM key_values WHERE id = ?", id).Scan(
		&entry.ID, &entry.Key, &entry.Value, &entry.CreatedAt, &entry.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func updateEntry(c *gin.Context) {
	id := c.Param("id")
	var entry KeyValue

	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM key_values WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	_, err = db.Exec("UPDATE key_values SET key = ?, value = ? WHERE id = ?", entry.Key, entry.Value, id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "Key already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func deleteEntry(c *gin.Context) {
	id := c.Param("id")

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM key_values WHERE id = ?)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	_, err = db.Exec("DELETE FROM key_values WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func generateDummyData(c *gin.Context) {
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO key_values (key, value) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d-%s", i, randomString(5))
		value := fmt.Sprintf("value-%d-%s", i, randomString(10))
		_, err := stmt.Exec(key, value)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func truncateDB(c *gin.Context) {
	_, err := db.Exec("DELETE FROM key_values")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = db.Exec("VACUUM")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
