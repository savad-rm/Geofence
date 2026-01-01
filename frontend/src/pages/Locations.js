import React, { useState, useEffect } from 'react';
import { vehicleAPI } from '../api';

function Locations() {
  const [vehicles, setVehicles] = useState([]);
  const [formData, setFormData] = useState({
    vehicle_id: '',
    latitude: '',
    longitude: '',
  });
  const [locationData, setLocationData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [selectedVehicle, setSelectedVehicle] = useState('');

  useEffect(() => {
    fetchVehicles();
  }, []);

  const fetchVehicles = async () => {
    try {
      const response = await vehicleAPI.getAll();
      setVehicles(response.data.vehicles || []);
    } catch (err) {
      console.error('Failed to fetch vehicles', err);
    }
  };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      setLoading(true);
      await vehicleAPI.updateLocation({
        ...formData,
        latitude: parseFloat(formData.latitude),
        longitude: parseFloat(formData.longitude),
        timestamp: new Date().toISOString(),
      });
      setFormData({
        vehicle_id: '',
        latitude: '',
        longitude: '',
      });
      setLocationData(null);
      setError('');
    } catch (err) {
      setError('Failed to update location');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleGetLocation = async (e) => {
    e.preventDefault();
    if (!selectedVehicle) {
      setError('Please select a vehicle');
      return;
    }
    try {
      setLoading(true);
      const response = await vehicleAPI.getLocation(selectedVehicle);
      setLocationData(response.data);
      setError('');
    } catch (err) {
      setError('Failed to fetch location');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page">
      <h1 className="page-title">Vehicle Location Management</h1>

      <div style={{ marginBottom: '2rem' }}>
        <h2 style={{ marginBottom: '1rem', fontSize: '1.3rem' }}>Update Vehicle Location</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Vehicle</label>
            <select
              name="vehicle_id"
              value={formData.vehicle_id}
              onChange={handleInputChange}
              required
            >
              <option value="">Select a vehicle</option>
              {vehicles.map((vehicle) => (
                <option key={vehicle.id} value={vehicle.id}>
                  {vehicle.vehicle_number} - {vehicle.driver_name}
                </option>
              ))}
            </select>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>Latitude</label>
              <input
                type="number"
                name="latitude"
                step="0.0001"
                value={formData.latitude}
                onChange={handleInputChange}
                placeholder="e.g., 37.7749"
                required
              />
            </div>
            <div className="form-group">
              <label>Longitude</label>
              <input
                type="number"
                name="longitude"
                step="0.0001"
                value={formData.longitude}
                onChange={handleInputChange}
                placeholder="e.g., -122.4194"
                required
              />
            </div>
          </div>

          {error && <div className="error">{error}</div>}

          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Updating...' : 'Update Location'}
          </button>
        </form>
      </div>

      <div style={{ borderTop: '1px solid #ddd', paddingTop: '2rem' }}>
        <h2 style={{ marginBottom: '1rem', fontSize: '1.3rem' }}>Get Vehicle Location</h2>
        <form onSubmit={handleGetLocation}>
          <div className="form-group">
            <label>Vehicle</label>
            <select
              value={selectedVehicle}
              onChange={(e) => setSelectedVehicle(e.target.value)}
              required
            >
              <option value="">Select a vehicle</option>
              {vehicles.map((vehicle) => (
                <option key={vehicle.id} value={vehicle.id}>
                  {vehicle.vehicle_number} - {vehicle.driver_name}
                </option>
              ))}
            </select>
          </div>

          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Fetching...' : 'Get Location'}
          </button>
        </form>

        {locationData && (
          <div style={{ marginTop: '2rem', padding: '1rem', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
            <h3>Location Information</h3>
            <p>
              <strong>Vehicle Number:</strong> {locationData.vehicle_number}
            </p>
            {locationData.current_location ? (
              <>
                <p>
                  <strong>Latitude:</strong> {locationData.current_location.latitude}
                </p>
                <p>
                  <strong>Longitude:</strong> {locationData.current_location.longitude}
                </p>
                <p>
                  <strong>Timestamp:</strong> {new Date(locationData.current_location.timestamp).toLocaleString()}
                </p>
              </>
            ) : (
              <p>No location data available</p>
            )}

            <h4 style={{ marginTop: '1rem' }}>Current Geofences</h4>
            {locationData.current_geofences && locationData.current_geofences.length > 0 ? (
              <table className="table">
                <thead>
                  <tr>
                    <th>Geofence Name</th>
                    <th>Category</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {locationData.current_geofences.map((geo) => (
                    <tr key={geo.geofence_id}>
                      <td>{geo.geofence_name}</td>
                      <td>{geo.category}</td>
                      <td>
                        <span className="badge badge-info">{geo.status}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <p>Vehicle is not inside any geofences</p>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default Locations;
