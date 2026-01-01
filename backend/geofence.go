package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math"
)

func checkGeofences(vehicleID string, lat float64, lon float64) []CurrentGeofence {
	var currentGeofences []CurrentGeofence

	rows, err := db.Query(
		`SELECT id, name, category, coordinates FROM geofences WHERE status = 'active'`,
	)
	if err != nil {
		log.Println("Error querying geofences:", err)
		return currentGeofences
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, category, coordStr string
		if err := rows.Scan(&id, &name, &category, &coordStr); err != nil {
			log.Println("Error scanning geofence:", err)
			continue
		}

		var coordinates [][2]float64
		if err := json.Unmarshal([]byte(coordStr), &coordinates); err != nil {
			log.Println("Error unmarshaling coordinates:", err)
			continue
		}

		if isPointInPolygon(lat, lon, coordinates) {
			currentGeofences = append(currentGeofences, CurrentGeofence{
				GeofenceID:   id,
				GeofenceName: name,
				Category:     category,
				Status:       "inside",
			})
		}
	}

	return currentGeofences
}

func isPointInPolygon(lat float64, lon float64, polygon [][2]float64) bool {
	n := len(polygon)
	inside := false

	p1lat, p1lon := polygon[0][0], polygon[0][1]
	for i := 1; i <= n; i++ {
		p2lat, p2lon := polygon[i%n][0], polygon[i%n][1]
		if lon > math.Min(p1lon, p2lon) {
			if lon <= math.Max(p1lon, p2lon) {
				if lat <= math.Max(p1lat, p2lat) {
					if p1lon != p2lon {
						xinters := (lon-p1lon)*(p2lat-p1lat)/(p2lon-p1lon) + p1lat
						if p1lat == p2lat || lat <= xinters {
							inside = !inside
						}
					}
				}
			}
		}
		p1lat, p1lon = p2lat, p2lon
	}

	return inside
}

func getPreviousState(vehicleID, geofenceID string) string {
	var state string
	err := db.QueryRow(
		`SELECT status FROM vehicle_geofence_state
		 WHERE vehicle_id = $1 AND geofence_id = $2`,
		vehicleID, geofenceID,
	).Scan(&state)

	if err == sql.ErrNoRows {
		return "outside"
	}
	return state
}

func updateGeofenceState(vehicleID, geofenceID, state string) {
	db.Exec(
		`INSERT INTO vehicle_geofence_state (vehicle_id, geofence_id, status)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (vehicle_id, geofence_id)
		 DO UPDATE SET status = $3, updated_at = NOW()`,
		vehicleID, geofenceID, state,
	)
}

func checkAndTriggerAlerts(vehicleID string, lat, lon float64, timestamp string) {
	current := checkGeofences(vehicleID, lat, lon)

	currentMap := make(map[string]bool)
	for _, g := range current {
		currentMap[g.GeofenceID] = true
	}

	rows, err := db.Query(
		`SELECT id, geofence_id, event_type
		 FROM alert_configs
		 WHERE vehicle_id = $1 AND status = 'active'`,
		vehicleID,
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var alertID, geofenceID, eventType string
		rows.Scan(&alertID, &geofenceID, &eventType)

		prevState := getPreviousState(vehicleID, geofenceID)
		currState := "outside"
		if currentMap[geofenceID] {
			currState = "inside"
		}

		if prevState == "outside" && currState == "inside" && eventType == "entry" {
			recordViolation(vehicleID, geofenceID, "entry", lat, lon, timestamp)
			triggerAlert(vehicleID, geofenceID, "entry", lat, lon, timestamp)
		}

		if prevState == "inside" && currState == "outside" && eventType == "exit" {
			recordViolation(vehicleID, geofenceID, "exit", lat, lon, timestamp)
			triggerAlert(vehicleID, geofenceID, "exit", lat, lon, timestamp)
		}

		updateGeofenceState(vehicleID, geofenceID, currState)
	}
}

func recordViolation(vehicleID string, geofenceID string, eventType string, lat float64, lon float64, timestamp string) {
	var vehNum, geoName string
	db.QueryRow(`SELECT vehicle_number FROM vehicles WHERE id = $1`, vehicleID).Scan(&vehNum)
	db.QueryRow(`SELECT name FROM geofences WHERE id = $1`, geofenceID).Scan(&geoName)

	violID := "viol_" + randomID()
	_, err := db.Exec(
		`INSERT INTO violations (id, vehicle_id, geofence_id, event_type, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		violID, vehicleID, geofenceID, eventType, lat, lon, timestamp,
	)

	if err != nil {
		log.Println("Error recording violation:", err)
	}
}

func triggerAlert(vehicleID string, geofenceID string, eventType string, lat float64, lon float64, timestamp string) {
	var configs []map[string]interface{}

	rows, err := db.Query(
		`SELECT id FROM alert_configs
		WHERE geofence_id = $1 AND status = 'active'
		AND (vehicle_id = $2 OR vehicle_id IS NULL)
		AND (event_type = $3 OR event_type = 'both')`,
		geofenceID, vehicleID, eventType,
	)
	if err != nil {
		log.Println("Error querying alert configs:", err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		var veh Vehicle
		var geo Geofence

		db.QueryRow(`SELECT id, vehicle_number, driver_name, phone FROM vehicles WHERE id = $1`, vehicleID).Scan(&veh.ID, &veh.VehicleNumber, &veh.DriverName, &veh.Phone)
		db.QueryRow(`SELECT id, name, category FROM geofences WHERE id = $1`, geofenceID).Scan(&geo.ID, &geo.Name, &geo.Category)

		alert := map[string]interface{}{
			"event_id":   "evt_" + randomID(),
			"event_type": eventType,
			"timestamp":  timestamp,
			"vehicle": map[string]string{
				"vehicle_id":     veh.ID,
				"vehicle_number": veh.VehicleNumber,
				"driver_name":    veh.DriverName,
			},
			"geofence": map[string]string{
				"geofence_id":   geo.ID,
				"geofence_name": geo.Name,
				"category":      geo.Category,
			},
			"location": map[string]float64{
				"latitude":  lat,
				"longitude": lon,
			},
		}

		alertHistID := "ah_" + randomID()
		_, err := db.Exec(
			`INSERT INTO alert_history (id, geofence_id, vehicle_id, event_type, latitude, longitude, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			alertHistID, geofenceID, vehicleID, eventType, lat, lon, timestamp,
		)
		if err != nil {
			log.Println("Error recording alert history:", err)
		}

		configs = append(configs, alert)

		hub.broadcast <- alert
	}
}

func randomID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = "0123456789abcdef"[i%16]
	}
	return string(b)
}
