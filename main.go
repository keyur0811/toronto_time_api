package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "path/filepath"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/gorilla/mux"
)

// Global DB variable
var db *sql.DB

// Initialize Database Connection
func init() {
    var err error
    dsn := "api_user:password@tcp(127.0.0.1:3306)/toronto_time_db"
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }

    if err = db.Ping(); err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    fmt.Println("Connected to the database.")
}

// Struct for JSON Response
type TimeResponse struct {
    CurrentTime string `json:"current_time"`
}

// /current-time Endpoint
func getCurrentTime(w http.ResponseWriter, r *http.Request) {
    loc, err := time.LoadLocation("America/Toronto") // Load Toronto timezone
    if err != nil {
        http.Error(w, "Timezone error", http.StatusInternalServerError)
        log.Printf("Error: %v", err)
        return
    }

    torontoTime := time.Now().In(loc) // Convert to Toronto time
    log.Printf("Current Toronto Time: %v", torontoTime)

    // Log the time in the database
    _, err = db.Exec("INSERT INTO time_log (timestamp) VALUES (?)", torontoTime)
    if err != nil {
        log.Printf("Error inserting time into database: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // Respond with JSON
    response := TimeResponse{CurrentTime: torontoTime.Format("2006-01-02 15:04:05")}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// /logged-times Endpoint
func getLoggedTimes(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, timestamp FROM time_log")
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        log.Printf("Error querying database: %v", err)
        return
    }
    defer rows.Close()

    var logs []TimeResponse
    for rows.Next() {
        var id int
        var timestamp string // Use string to handle raw MySQL DATETIME
        if err := rows.Scan(&id, &timestamp); err != nil {
            log.Printf("Error scanning row: %v", err)
            continue
        }

        // Convert timestamp string to Go's time.Time
        parsedTime, err := time.Parse("2006-01-02 15:04:05", timestamp)
        if err != nil {
            log.Printf("Error parsing timestamp: %v", err)
            continue
        }

        logs = append(logs, TimeResponse{CurrentTime: parsedTime.Format("2006-01-02 15:04:05")})
    }

    // Handle any row iteration errors
    if err := rows.Err(); err != nil {
        log.Printf("Row iteration error: %v", err)
        http.Error(w, "Error retrieving logs", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(logs)
}

// Serve static files (HTML, CSS, JS)
func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
    // Serve the HTML file
    if r.URL.Path == "/" || r.URL.Path == "/index.html" {
        http.ServeFile(w, r, "index.html")
        return
    }

    // Serve static assets like CSS, JS files
    staticDir := "./static"
    http.ServeFile(w, r, filepath.Join(staticDir, r.URL.Path))
}

func main() {
    r := mux.NewRouter()

    // Route to serve the HTML file
    r.HandleFunc("/", serveStaticFiles).Methods("GET")
    r.HandleFunc("/current-time", getCurrentTime).Methods("GET")
    r.HandleFunc("/logged-times", getLoggedTimes).Methods("GET")

    fmt.Println("Server running on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", r))
}
