import React, { useEffect, useState } from 'react';

function AlertNotifications({ alerts }) {
  const [displayedAlerts, setDisplayedAlerts] = useState([]);

  useEffect(() => {
    if (alerts.length > 0) {
      const newAlert = alerts[0];
      setDisplayedAlerts((prev) => [...prev, newAlert]);

      const timer = setTimeout(() => {
        setDisplayedAlerts((prev) => prev.filter((a) => a.event_id !== newAlert.event_id));
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [alerts]);

  return (
    <div className="alerts-container">
      {displayedAlerts.map((alert) => (
        <div key={alert.event_id} className={`alert-notification ${alert.event_type}`}>
          <strong>{alert.event_type.toUpperCase()}</strong>
          <p>
            <strong>Vehicle:</strong> {alert.vehicle.vehicle_number} ({alert.vehicle.driver_name})
          </p>
          <p>
            <strong>Geofence:</strong> {alert.geofence.geofence_name} ({alert.geofence.category})
          </p>
          <p>
            <strong>Location:</strong> {alert.location.latitude.toFixed(4)}, {alert.location.longitude.toFixed(4)}
          </p>
          <p className="timestamp">{new Date(alert.timestamp).toLocaleString()}</p>
        </div>
      ))}
    </div>
  );
}

export default AlertNotifications;
