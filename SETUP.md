# Geofencing and Real-time Alert System - Setup Guide

## 📋 Overview

This guide provides instructions for setting up and running the Geofencing and Real-time Alert System locally and with Docker.

## Prerequisites

### Local Development
- **Go** 1.21 or higher
- **Node.js** 18+ and npm
- **PostgreSQL** 12+ with PostGIS extension
- **Git**

### Docker Setup
- **Docker** 20.10+
- **Docker Compose** 2.0+

## 🚀 Quick Start with Docker Compose

The fastest way to get the entire system running is using Docker Compose:

```bash
docker-compose up --build
```

This will start:
- PostgreSQL database (port 5432)
- Go backend API (port 8080)
- React frontend (port 3000)

Access the application at: **http://localhost:3000**

## 📝 Local Development Setup

### 1. Database Setup

Install PostgreSQL with PostGIS extension:

```bash
# On macOS with Homebrew
brew install postgresql postgis

# On Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib postgis postgresql-13-postgis-3

# On Windows - Download from https://www.postgresql.org/download/windows/
```

Create the database:

```bash
createdb geofencing
```

Enable PostGIS extension:

```bash
psql -d geofencing -c "CREATE EXTENSION IF NOT EXISTS postgis;"
```

### 2. Backend Setup

Navigate to the backend directory:

```bash
cd backend
```

Download Go dependencies:

```bash
go mod download
```

Set up environment variables (create a `.env` file or set them in your shell):

```bash
export DATABASE_URL="postgres://user:password@localhost:5432/geofencing?sslmode=disable"
export PORT=8080
```

Run the backend:

```bash
go run main.go handlers.go geofence.go websocket.go
```

The backend API will be available at: **http://localhost:8080**

### 3. Frontend Setup

Navigate to the frontend directory:

```bash
cd frontend
```

Install dependencies:

```bash
npm install
```

Start the development server:

```bash
npm start
```

The frontend will open automatically at: **http://localhost:3000**

## 🔌 API Testing Guide

### Base URL
```
http://localhost:8080
```

### 1. Create a Geofence

```bash
curl -X POST http://localhost:8080/geofences \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Downtown Delivery Zone",
    "description": "Main delivery area",
    "coordinates": [
      [37.7749, -122.4194],
      [37.7849, -122.4194],
      [37.7849, -122.4094],
      [37.7749, -122.4094],
      [37.7749, -122.4194]
    ],
    "category": "delivery_zone"
  }'
```

### 2. Get All Geofences

```bash
curl http://localhost:8080/geofences
```

Filter by category:
```bash
curl "http://localhost:8080/geofences?category=delivery_zone"
```

### 3. Register a Vehicle

```bash
curl -X POST http://localhost:8080/vehicles \
  -H "Content-Type: application/json" \
  -d '{
    "vehicle_number": "KA-01-AB-1234",
    "driver_name": "John Doe",
    "vehicle_type": "truck",
    "phone": "+1234567890"
  }'
```

### 4. Get All Vehicles

```bash
curl http://localhost:8080/vehicles
```

### 5. Update Vehicle Location

```bash
curl -X POST http://localhost:8080/vehicles/location \
  -H "Content-Type: application/json" \
  -d '{
    "vehicle_id": "veh_<uuid>",
    "latitude": 37.7849,
    "longitude": -122.4194,
    "timestamp": "2025-01-15T10:35:00Z"
  }'
```

### 6. Get Vehicle Location

```bash
curl http://localhost:8080/vehicles/location/veh_<uuid>
```

### 7. Configure an Alert

```bash
curl -X POST http://localhost:8080/alerts/configure \
  -H "Content-Type: application/json" \
  -d '{
    "geofence_id": "geo_<uuid>",
    "vehicle_id": "veh_<uuid>",
    "event_type": "entry"
  }'
```

### 8. Get All Alerts

```bash
curl http://localhost:8080/alerts
```

Filter by geofence:
```bash
curl "http://localhost:8080/alerts?geofence_id=geo_<uuid>"
```

### 9. Get Violation History

```bash
curl "http://localhost:8080/violations/history?limit=50"
```

Filter by vehicle and date range:
```bash
curl "http://localhost:8080/violations/history?vehicle_id=veh_<uuid>&start_date=2025-01-01T00:00:00Z&end_date=2025-01-31T23:59:59Z&limit=100"
```

### 10. WebSocket Connection

Connect to the real-time alerts WebSocket:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/alerts');

ws.onopen = () => {
  console.log('Connected to alerts stream');
};

ws.onmessage = (event) => {
  const alert = JSON.parse(event.data);
  console.log('Alert received:', alert);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

## 🎨 Frontend Usage Guide

### Main Features

1. **Geofence Management**
   - Create new geofences by specifying coordinates
   - View all geofences with filtering by category
   - Edit geofence coordinates on the map

2. **Vehicle Management**
   - Register new vehicles with driver and contact information
   - View all registered vehicles
   - Check vehicle status

3. **Location Updates**
   - Update vehicle locations with latitude/longitude
   - View current geofence status for each vehicle
   - Get real-time location information

4. **Alert Configuration**
   - Create alert rules for geofence entry/exit events
   - Configure alerts for specific vehicles or all vehicles
   - View all configured alerts

5. **Violation History**
   - View all geofence violations (entries and exits)
   - Filter by vehicle, geofence, or date range
   - Download or export violation data

6. **Real-time Notifications**
   - Receive instant notifications when vehicles enter/exit geofences
   - Toast notifications display alert details
   - Notifications persist in the alerts feed

## 🗂️ Project Structure

```
geofencing-system/
├── backend/
│   ├── main.go           # Main application entry point
│   ├── handlers.go       # API request handlers
│   ├── geofence.go       # Geofencing logic (point-in-polygon)
│   ├── websocket.go      # WebSocket implementation
│   ├── go.mod            # Go dependencies
│   └── Dockerfile        # Backend Docker image
├── frontend/
│   ├── public/
│   │   └── index.html    # HTML entry point
│   ├── src/
│   │   ├── App.js        # Main React component
│   │   ├── App.css       # Styling
│   │   ├── index.js      # React DOM render
│   │   ├── api.js        # API client utilities
│   │   ├── components/
│   │   │   └── AlertNotifications.js  # Alert UI component
│   │   └── pages/
│   │       ├── Geofences.js    # Geofence management page
│   │       ├── Vehicles.js     # Vehicle management page
│   │       ├── Locations.js    # Location update page
│   │       ├── Alerts.js       # Alert configuration page
│   │       └── Violations.js   # Violation history page
│   ├── package.json      # NPM dependencies
│   └── Dockerfile        # Frontend Docker image
├── docker-compose.yml    # Docker Compose orchestration
├── README.md             # Project overview
└── SETUP.md              # This file
```

## 🚢 Deployment

### Backend Deployment

1. Build and push Docker image to Docker Hub:

```bash
docker build -t yourusername/geofencing-backend:1.0 ./backend
docker push yourusername/geofencing-backend:1.0
```

2. Deploy to your hosting platform:
   - AWS ECS
   - Google Cloud Run
   - DigitalOcean App Platform
   - Any Kubernetes cluster

### Frontend Deployment to Netlify

1. Build the frontend:

```bash
cd frontend
npm run build
```

2. Connect to Netlify:
   - Push your code to GitHub
   - Connect your GitHub repo to Netlify
   - Netlify will automatically build and deploy on every push

3. Configure environment variables in Netlify:
   - Set `REACT_APP_API_URL` to your backend URL

## 📊 Database Schema

The system uses the following PostgreSQL tables:

- **geofences**: Stores geofence polygons and metadata
- **vehicles**: Stores vehicle registration information
- **locations**: Stores historical location updates
- **alert_configs**: Stores configured alert rules
- **violations**: Stores geofence entry/exit events
- **alert_history**: Stores alert event history

All tables are automatically created on first run.

## 🐛 Troubleshooting

### Backend Connection Issues
- Ensure PostgreSQL is running and accessible
- Check DATABASE_URL environment variable
- Verify port 8080 is not in use

### Frontend Can't Connect to Backend
- Verify backend is running on http://localhost:8080
- Check REACT_APP_API_URL environment variable
- Ensure CORS is enabled (already configured in backend)

### Docker Issues
- Run `docker-compose down -v` to clean up volumes
- Rebuild images with `docker-compose up --build`
- Check logs with `docker-compose logs -f`

## 📝 API Response Format

All API responses include a `time_ns` field indicating execution time in nanoseconds:

```json
{
  "data": {...},
  "time_ns": "1234567"
}
```

## 🔐 Security Notes

- In production, set strong database passwords
- Use environment variables for sensitive data
- Implement API authentication (JWT/API keys)
- Use HTTPS/WSS for production deployments
- Add rate limiting to prevent abuse
- Validate all input parameters

## 📞 Support

For issues or questions:
1. Check the README.md for overview
2. Review API documentation in this SETUP.md
3. Check backend logs: `docker-compose logs backend`
4. Check frontend console for errors

---

**Last Updated**: January 2025
