import React, { useState, useEffect, useCallback } from 'react';
import { violationAPI, vehicleAPI, geofenceAPI } from '../api';

function Violations() {
  const [violations, setViolations] = useState([]);
  const [vehicles, setVehicles] = useState([]);
  const [geofences, setGeofences] = useState([]);
  const [filter, setFilter] = useState({
    vehicle_id: '',
    geofence_id: '',
    start_date: '',
    end_date: '',
    limit: '50',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchData = useCallback(async () => {
    try {
      const [vehRes, geoRes] = await Promise.all([
        vehicleAPI.getAll(),
        geofenceAPI.getAll(),
      ]);
      setVehicles(vehRes.data.vehicles || []);
      setGeofences(geoRes.data.geofences || []);
    } catch (err) {
      console.error('Failed to fetch data', err);
    }
  }, []);

  const fetchViolations = useCallback(async () => {
    try {
      setLoading(true);
      const params = {};
      if (filter.vehicle_id) params.vehicle_id = filter.vehicle_id;
      if (filter.geofence_id) params.geofence_id = filter.geofence_id;
      if (filter.start_date) params.start_date = filter.start_date;
      if (filter.end_date) params.end_date = filter.end_date;
      if (filter.limit) params.limit = filter.limit;

      const response = await violationAPI.getHistory(params);
      setViolations(response.data.violations || []);
    } catch (err) {
      setError('Failed to fetch violations');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    fetchViolations();
  }, [fetchViolations]);

  const handleFilterChange = (e) => {
    const { name, value } = e.target;
    setFilter((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleReset = () => {
    setFilter({
      vehicle_id: '',
      geofence_id: '',
      start_date: '',
      end_date: '',
      limit: '50',
    });
  };

  return (
    <div className="page">
      <h1 className="page-title">Violation History</h1>

      <div className="filter-group" style={{ flexWrap: 'wrap' }}>
        <select
          name="vehicle_id"
          value={filter.vehicle_id}
          onChange={handleFilterChange}
          style={{ flex: '1 1 200px' }}
        >
          <option value="">All Vehicles</option>
          {vehicles.map((veh) => (
            <option key={veh.id} value={veh.id}>
              {veh.vehicle_number}
            </option>
          ))}
        </select>

        <select
          name="geofence_id"
          value={filter.geofence_id}
          onChange={handleFilterChange}
          style={{ flex: '1 1 200px' }}
        >
          <option value="">All Geofences</option>
          {geofences.map((geo) => (
            <option key={geo.id} value={geo.id}>
              {geo.name}
            </option>
          ))}
        </select>

        <input
          type="datetime-local"
          name="start_date"
          value={filter.start_date}
          onChange={(e) =>
            setFilter((prev) => ({
              ...prev,
              start_date: e.target.value,
            }))
          }
          style={{ flex: '1 1 200px' }}
        />

        <input
          type="datetime-local"
          name="end_date"
          value={filter.end_date}
          onChange={(e) =>
            setFilter((prev) => ({
              ...prev,
              end_date: e.target.value,
            }))
          }
          style={{ flex: '1 1 200px' }}
        />

        <select
          name="limit"
          value={filter.limit}
          onChange={handleFilterChange}
          style={{ flex: '1 1 100px' }}
        >
          <option value="10">10</option>
          <option value="50">50</option>
          <option value="100">100</option>
          <option value="500">500</option>
        </select>

        <button
          onClick={handleReset}
          className="btn btn-secondary"
          style={{ flex: '1 1 100px' }}
        >
          Reset
        </button>
      </div>

      {error && <div className="error">{error}</div>}

      {loading ? (
        <div className="loading">Loading violations...</div>
      ) : (
        <table className="table" style={{ marginTop: '2rem' }}>
          <thead>
            <tr>
              <th>Vehicle Number</th>
              <th>Geofence Name</th>
              <th>Event Type</th>
              <th>Latitude</th>
              <th>Longitude</th>
              <th>Timestamp</th>
            </tr>
          </thead>
          <tbody>
            {violations.length > 0 ? (
              violations.map((violation) => (
                <tr key={violation.id}>
                  <td>{violation.vehicle_number}</td>
                  <td>{violation.geofence_name}</td>
                  <td>
                    <span
                      className={`badge ${violation.event_type === 'entry'
                          ? 'badge-warning'
                          : 'badge-info'
                        }`}
                    >
                      {violation.event_type}
                    </span>
                  </td>
                  <td>{violation.latitude.toFixed(6)}</td>
                  <td>{violation.longitude.toFixed(6)}</td>
                  <td>{new Date(violation.timestamp).toLocaleString()}</td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan="6" style={{ textAlign: 'center' }}>
                  No violations found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      )}
    </div>
  );
}

export default Violations;
