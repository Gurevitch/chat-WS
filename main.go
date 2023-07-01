package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"database/sql"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow connections from localhost:3000
			return r.Header.Get("Origin") == "http://localhost:3000"
		},
	}
	clients = make(map[*websocket.Conn]bool)
)

var db *sql.DB

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func main() {
	err := connectToDatabase()
	if err != nil {
		log.Println("Failed to connect to the database,the error recived is :", err)
		return
	}
	e := echo.New()

	// Enable CORS middleware
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/ws", handleWebSocket)
	e.POST("/login", handleLogin)

	e.Static("/", "public")

	e.Logger.Fatal(e.Start(":8000"))
}

func handleLogin(c echo.Context) error {
	// Parse the request body into a User struct
	var user User
	if err := c.Bind(&user); err != nil {
		return c.String(http.StatusBadRequest, "Invalid request payload")
	}

	// Check if the username and password match a record in the database
	stmt := "SELECT COUNT(*) FROM users WHERE username = ? AND password = ?"
	var count int
	err := db.QueryRow(stmt, user.Username, user.Password).Scan(&count)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	// Check the count to determine if the login is successful
	if count == 1 {
		response := LoginResponse{
			Success: true,
			Message: "Login successful",
		}
		return c.JSON(http.StatusOK, response)
	} else {
		// Insert new user to the db
		createUser(user)
		response := LoginResponse{
			Success: true,
			Message: "Login successful",
		}
		return c.JSON(http.StatusOK, response)
	}
}

func createUser(user User) error {

	// Prepare the SQL statement for inserting the user
	stmt, err := db.Prepare("INSERT INTO users (username, password) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the SQL statement with the user data
	_, err = stmt.Exec(user.Username, user.Password)
	if err != nil {
		return err
	}

	return nil
}

func connectToDatabase() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	dbPath := filepath.Join(currentDir, "Chat.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Create the database file
		_, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	createTableSQL := "CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY AUTOINCREMENT,user VARCHAR(255),content TEXT,created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	createTableSQL = "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT,username VARCHAR(255),password VARCHAR(255))"
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}
func handleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer ws.Close()

	clients[ws] = true

	for {
		// Read message from the browser
		_, msgBytes, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, ws)
			break
		}

		// Broadcast the received message to all clients
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msgBytes)
			if err != nil {
				log.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
	}

	return nil
}
