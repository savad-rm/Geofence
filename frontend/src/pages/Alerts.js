import React, { useState, useEffect, useCallback } from 'react';
import { alertAPI, geofenceAPI, vehicleAPI } from '../api';

function Alerts() {
  const [alerts, setAlerts] = useState([]);
  const [geofences, setGeofences] = useState([]);
  const [vehicles, setVehicles] = useState([]);
  const [formData, setFormData] = useState({
    geofence_id: '',
    vehicle_id: '',
    event_type: 'entry',
  });
  const [filter, setFilter] = useState({
    geofence_id: '',
    vehicle_id: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchData = useCallback(async () => {
    try {
      const [geoRes, vehRes] = await Promise.all([
        geofenceAPI.getAll(),
        vehicleAPI.getAll(),
      ]);
      setGeofences(geoRes.data.geofences || []);
      setVehicles(vehRes.data.vehicles || []);
    } catch (err) {
      console.error('Failed to fetch data', err);
    }
  }, []);

  const fetchAlerts = useCallback(async () => {
    try {
      setLoading(true);
      const response = await alertAPI.getAll(filter);
      setAlerts(response.data.alerts || []);
    } catch (err) {
      setError('Failed to fetch alerts');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    fetchAlerts();
  }, [fetchAlerts]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleFilterChange = (e) => {
    const { name, value } = e.target;
    setFilter((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await alertAPI.configure(formData);
      setFormData({
        geofence_id: '',
        vehicle_id: '',
        event_type: 'entry',
      });
      fetchAlerts();
      setError('');
    } catch (err) {
      setError('Failed to configure alert');
      console.error(err);
    }
  };

  return (
    <div className="page">
      <h1 className="page-title">Alert Configuration</h1>

      <form onSubmit={handleSubmit}>
        <div className="form-row">
          <div className="form-group">
            <label>Geofence *</label>
            <select
              name="geofence_id"
              value={formData.geofence_id}
              onChange={handleInputChange}
              required
            >
              <option value="">Select a geofence</option>
              {geofences.map((geo) => (
                <option key={geo.id} value={geo.id}>
                  {geo.name} ({geo.category})
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>Vehicle (Optional)</label>
            <select
              name="vehicle_id"
              value={formData.vehicle_id}
              onChange={handleInputChange}
            >
              <option value="">All Vehicles</option>
              {vehicles.map((veh) => (
                <option key={veh.id} value={veh.id}>
                  {veh.vehicle_number} - {veh.driver_name}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label>Event Type</label>
            <select name="event_type" value={formData.event_type} onChange={handleInputChange}>
              <option value="entry">Entry</option>
              <option value="exit">Exit</option>
              <option value="both">Both</option>
            </select>
          </div>
        </div>

        {error && <div className="error">{error}</div>}

        <button type="submit" className="btn btn-primary">
          Configure Alert
        </button>
      </form>

      <div className="list-container">
        <h2 className="list-title">Configured Alerts</h2>

        <div className="filter-group">
          <select
            name="geofence_id"
            value={filter.geofence_id}
            onChange={handleFilterChange}
          >
            <option value="">All Geofences</option>
            {geofences.map((geo) => (
              <option key={geo.id} value={geo.id}>
                {geo.name}
              </option>
            ))}
          </select>

          <select
            name="vehicle_id"
            value={filter.vehicle_id}
            onChange={handleFilterChange}
          >
            <option value="">All Vehicles</option>
            {vehicles.map((veh) => (
              <option key={veh.id} value={veh.id}>
                {veh.vehicle_number}
              </option>
            ))}
          </select>
        </div>

        {loading ? (
          <div className="loading">Loading alerts...</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Geofence</th>
                <th>Vehicle</th>
                <th>Event Type</th>
                <th>Status</th>
                <th>Created At</th>
              </tr>
            </thead>
            <tbody>
              {alerts.map((alert) => (
                <tr key={alert.alert_id}>
                  <td>{alert.geofence_name}</td>
                  <td>{alert.vehicle_number || 'All Vehicles'}</td>
                  <td>{alert.event_type}</td>
                  <td>
                    <span className="badge badge-success">{alert.status}</span>
                  </td>
                  <td>{new Date(alert.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export default Alerts;
