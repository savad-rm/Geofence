package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	fmt.Println("Database URL:", dsn)
	if dsn == "" {
		dsn = "postgres://postgres:admin@localhost:5432/geofencing?sslmode=disable"
	}

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	log.Println("Connected to database")
	initDB()
}

func initDB() {
	schema := `
	CREATE TABLE IF NOT EXISTS geofences (
		id VARCHAR(50) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		coordinates TEXT NOT NULL,
		category VARCHAR(50) NOT NULL,
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS vehicles (
		id VARCHAR(50) PRIMARY KEY,
		vehicle_number VARCHAR(50) UNIQUE NOT NULL,
		driver_name VARCHAR(255) NOT NULL,
		vehicle_type VARCHAR(50) NOT NULL,
		phone VARCHAR(20) NOT NULL,
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS locations (
		id VARCHAR(50) PRIMARY KEY,
		vehicle_id VARCHAR(50) NOT NULL,
		latitude DECIMAL(10, 8) NOT NULL,
		longitude DECIMAL(11, 8) NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
	);

	CREATE TABLE IF NOT EXISTS alert_configs (
		id VARCHAR(50) PRIMARY KEY,
		geofence_id VARCHAR(50) NOT NULL,
		vehicle_id VARCHAR(50),
		event_type VARCHAR(20) NOT NULL,
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (geofence_id) REFERENCES geofences(id),
		FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
	);

	CREATE TABLE IF NOT EXISTS violations (
		id VARCHAR(50) PRIMARY KEY,
		vehicle_id VARCHAR(50) NOT NULL,
		geofence_id VARCHAR(50) NOT NULL,
		event_type VARCHAR(20) NOT NULL,
		latitude DECIMAL(10, 8) NOT NULL,
		longitude DECIMAL(11, 8) NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (vehicle_id) REFERENCES vehicles(id),
		FOREIGN KEY (geofence_id) REFERENCES geofences(id)
	);

	CREATE TABLE IF NOT EXISTS alert_history (
		id VARCHAR(50) PRIMARY KEY,
		geofence_id VARCHAR(50) NOT NULL,
		vehicle_id VARCHAR(50) NOT NULL,
		event_type VARCHAR(20) NOT NULL,
		latitude DECIMAL(10, 8) NOT NULL,
		longitude DECIMAL(11, 8) NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (geofence_id) REFERENCES geofences(id),
		FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
	);

	CREATE TABLE IF NOT EXISTS vehicle_geofence_state (
	vehicle_id  VARCHAR(50) NOT NULL,
	geofence_id VARCHAR(50) NOT NULL,
	status      VARCHAR(10) NOT NULL, -- inside | outside
	updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (vehicle_id, geofence_id)
	);

	CREATE INDEX IF NOT EXISTS idx_vehicle_id ON locations(vehicle_id);
	CREATE INDEX IF NOT EXISTS idx_geofence_id ON violations(geofence_id);
	CREATE INDEX IF NOT EXISTS idx_vehicle_id_violations ON violations(vehicle_id);
	`

	_, err := db.Exec(schema)
	if err != nil {
		log.Fatal("Failed to create schema:", err)
	}
	log.Println("Database schema initialized")
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}, startTime time.Time) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	elapsed := time.Since(startTime).Nanoseconds()
	response := map[string]interface{}{
		"time_ns": fmt.Sprintf("%d", elapsed),
	}

	if dataMap, ok := data.(map[string]interface{}); ok {
		for k, v := range dataMap {
			response[k] = v
		}
	} else {
		response["data"] = data
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(enableCORS)

	r.Post("/geofences", createGeofence)
	r.Get("/geofences", getGeofences)
	r.Post("/vehicles", registerVehicle)
	r.Get("/vehicles", getVehicles)
	r.Post("/vehicles/location", updateVehicleLocation)
	r.Get("/vehicles/location/{vehicleID}", getVehicleLocation)
	r.Post("/alerts/configure", configureAlert)
	r.Get("/alerts", getAlerts)
	r.Get("/violations/history", getViolationsHistory)

	hub := NewAlertHub()
	go hub.Run()
	r.HandleFunc("/ws/alerts", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
