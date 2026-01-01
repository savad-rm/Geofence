package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Geofence struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Coordinates [][2]float64  `json:"coordinates"`
	Category    string        `json:"category"`
	Status      string        `json:"status"`
	CreatedAt   string        `json:"created_at"`
}

type Vehicle struct {
	ID            string `json:"id"`
	VehicleNumber string `json:"vehicle_number"`
	DriverName    string `json:"driver_name"`
	VehicleType   string `json:"vehicle_type"`
	Phone         string `json:"phone"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp string  `json:"timestamp"`
}

type CurrentGeofence struct {
	GeofenceID   string `json:"geofence_id"`
	GeofenceName string `json:"geofence_name"`
	Category     string `json:"category,omitempty"`
	Status       string `json:"status,omitempty"`
}

type AlertConfig struct {
	AlertID    string `json:"alert_id"`
	GeofenceID string `json:"geofence_id"`
	VehicleID  string `json:"vehicle_id"`
	EventType  string `json:"event_type"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

type Violation struct {
	ID            string  `json:"id"`
	VehicleID     string  `json:"vehicle_id"`
	VehicleNumber string  `json:"vehicle_number"`
	GeofenceID    string  `json:"geofence_id"`
	GeofenceName  string  `json:"geofence_name"`
	EventType     string  `json:"event_type"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Timestamp     string  `json:"timestamp"`
}

func createGeofence(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	var req struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Coordinates [][2]float64 `json:"coordinates"`
		Category    string      `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Coordinates) < 4 {
		http.Error(w, "Minimum 4 points required (3 unique + 1 closing point)", http.StatusBadRequest)
		return
	}

	if req.Coordinates[0] != req.Coordinates[len(req.Coordinates)-1] {
		http.Error(w, "First and last coordinates must be identical", http.StatusBadRequest)
		return
	}

	coordJSON, _ := json.Marshal(req.Coordinates)
	id := "geo_" + uuid.New().String()

	_, err := db.Exec(
		`INSERT INTO geofences (id, name, description, coordinates, category, status)
		VALUES ($1, $2, $3, $4, $5, 'active')`,
		id, req.Name, req.Description, string(coordJSON), req.Category,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":     id,
		"name":   req.Name,
		"status": "active",
	}, startTime)
}

func getGeofences(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	category := r.URL.Query().Get("category")

	query := "SELECT id, name, description, coordinates, category, status, created_at FROM geofences"
	var args []interface{}

	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var geofences []Geofence
	for rows.Next() {
		var g Geofence
		var coordStr string
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &coordStr, &g.Category, &g.Status, &g.CreatedAt); err != nil {
			log.Fatal(err)
		}

		json.Unmarshal([]byte(coordStr), &g.Coordinates)
		geofences = append(geofences, g)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"geofences": geofences,
	}, startTime)
}

func registerVehicle(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	var req struct {
		VehicleNumber string `json:"vehicle_number"`
		DriverName    string `json:"driver_name"`
		VehicleType   string `json:"vehicle_type"`
		Phone         string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := "veh_" + uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO vehicles (id, vehicle_number, driver_name, vehicle_type, phone, status)
		VALUES ($1, $2, $3, $4, $5, 'active')`,
		id, req.VehicleNumber, req.DriverName, req.VehicleType, req.Phone,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             id,
		"vehicle_number": req.VehicleNumber,
		"status":         "active",
	}, startTime)
}

func getVehicles(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	rows, err := db.Query(
		`SELECT id, vehicle_number, driver_name, vehicle_type, phone, status, created_at
		FROM vehicles ORDER BY created_at DESC`,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var vehicles []Vehicle
	for rows.Next() {
		var v Vehicle
		if err := rows.Scan(&v.ID, &v.VehicleNumber, &v.DriverName, &v.VehicleType, &v.Phone, &v.Status, &v.CreatedAt); err != nil {
			log.Fatal(err)
		}
		vehicles = append(vehicles, v)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"vehicles": vehicles,
	}, startTime)
}

func updateVehicleLocation(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	var req struct {
		VehicleID string  `json:"vehicle_id"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Timestamp string  `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	locID := "loc_" + uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO locations (id, vehicle_id, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4, $5)`,
		locID, req.VehicleID, req.Latitude, req.Longitude, req.Timestamp,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	currentGeofences := checkGeofences(req.VehicleID, req.Latitude, req.Longitude)
	checkAndTriggerAlerts(req.VehicleID, req.Latitude, req.Longitude, req.Timestamp)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"vehicle_id":       req.VehicleID,
		"location_updated": true,
		"current_geofences": currentGeofences,
	}, startTime)
}

func getVehicleLocation(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	vehicleID := chi.URLParam(r, "vehicleID")

	var veh Vehicle
	err := db.QueryRow(
		`SELECT id, vehicle_number, driver_name, vehicle_type, phone, status, created_at
		FROM vehicles WHERE id = $1`,
		vehicleID,
	).Scan(&veh.ID, &veh.VehicleNumber, &veh.DriverName, &veh.VehicleType, &veh.Phone, &veh.Status, &veh.CreatedAt)

	if err != nil {
		http.Error(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	var loc Location
	err = db.QueryRow(
		`SELECT latitude, longitude, timestamp FROM locations
		WHERE vehicle_id = $1 ORDER BY timestamp DESC LIMIT 1`,
		vehicleID,
	).Scan(&loc.Latitude, &loc.Longitude, &loc.Timestamp)

	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"vehicle_id":       vehicleID,
			"vehicle_number":   veh.VehicleNumber,
			"current_location": nil,
			"current_geofences": []interface{}{},
		}, startTime)
		return
	}

	currentGeofences := checkGeofences(vehicleID, loc.Latitude, loc.Longitude)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"vehicle_id":       vehicleID,
		"vehicle_number":   veh.VehicleNumber,
		"current_location": loc,
		"current_geofences": currentGeofences,
	}, startTime)
}

func configureAlert(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	var req struct {
		GeofenceID string `json:"geofence_id"`
		VehicleID  string `json:"vehicle_id"`
		EventType  string `json:"event_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	alertID := "alert_" + uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO alert_configs (id, geofence_id, vehicle_id, event_type, status)
		VALUES ($1, $2, $3, $4, 'active')`,
		alertID, req.GeofenceID, req.VehicleID, req.EventType,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"alert_id":   alertID,
		"geofence_id": req.GeofenceID,
		"vehicle_id": req.VehicleID,
		"event_type": req.EventType,
		"status":     "active",
	}, startTime)
}

func getAlerts(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	geofenceID := r.URL.Query().Get("geofence_id")
	vehicleID := r.URL.Query().Get("vehicle_id")

	query := `SELECT ac.id, ac.geofence_id, g.name, ac.vehicle_id, v.vehicle_number, ac.event_type, ac.status, ac.created_at
	FROM alert_configs ac
	JOIN geofences g ON ac.geofence_id = g.id
	LEFT JOIN vehicles v ON ac.vehicle_id = v.id WHERE 1=1`
	var args []interface{}
	argCount := 1

	if geofenceID != "" {
		query += fmt.Sprintf(" AND ac.geofence_id = $%d", argCount)
		args = append(args, geofenceID)
		argCount++
	}
	if vehicleID != "" {
		query += fmt.Sprintf(" AND ac.vehicle_id = $%d", argCount)
		args = append(args, vehicleID)
		argCount++
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var alerts []map[string]interface{}
	for rows.Next() {
		var alertID, geofID, geoName, vehID, vehNum, eventType, status, createdAt string
		if err := rows.Scan(&alertID, &geofID, &geoName, &vehID, &vehNum, &eventType, &status, &createdAt); err != nil {
			log.Fatal(err)
		}

		alert := map[string]interface{}{
			"alert_id":      alertID,
			"geofence_id":   geofID,
			"geofence_name": geoName,
			"event_type":    eventType,
			"status":        status,
			"created_at":    createdAt,
		}
		if vehID != "" {
			alert["vehicle_id"] = vehID
			alert["vehicle_number"] = vehNum
		}
		alerts = append(alerts, alert)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
	}, startTime)
}

func getViolationsHistory(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	vehicleID := r.URL.Query().Get("vehicle_id")
	geofenceID := r.URL.Query().Get("geofence_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	query := `SELECT v.id, v.vehicle_id, veh.vehicle_number, v.geofence_id, g.name, v.event_type, v.latitude, v.longitude, v.timestamp
	FROM violations v
	JOIN vehicles veh ON v.vehicle_id = veh.id
	JOIN geofences g ON v.geofence_id = g.id WHERE 1=1`
	var args []interface{}
	argCount := 1

	if vehicleID != "" {
		query += fmt.Sprintf(" AND v.vehicle_id = $%d", argCount)
		args = append(args, vehicleID)
		argCount++
	}
	if geofenceID != "" {
		query += fmt.Sprintf(" AND v.geofence_id = $%d", argCount)
		args = append(args, geofenceID)
		argCount++
	}
	if startDate != "" {
		query += fmt.Sprintf(" AND v.timestamp >= $%d", argCount)
		args = append(args, startDate)
		argCount++
	}
	if endDate != "" {
		query += fmt.Sprintf(" AND v.timestamp <= $%d", argCount)
		args = append(args, endDate)
		argCount++
	}

	countQuery := strings.Replace(query, "SELECT v.id, v.vehicle_id, veh.vehicle_number, v.geofence_id, g.name, v.event_type, v.latitude, v.longitude, v.timestamp", "SELECT COUNT(*)", 1)
	var totalCount int
	db.QueryRow(countQuery, args...).Scan(&totalCount)

	query += fmt.Sprintf(" ORDER BY v.timestamp DESC LIMIT $%d", argCount)
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var violations []Violation
	for rows.Next() {
		var v Violation
		if err := rows.Scan(&v.ID, &v.VehicleID, &v.VehicleNumber, &v.GeofenceID, &v.GeofenceName, &v.EventType, &v.Latitude, &v.Longitude, &v.Timestamp); err != nil {
			log.Fatal(err)
		}
		violations = append(violations, v)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"violations":  violations,
		"total_count": totalCount,
	}, startTime)
}
