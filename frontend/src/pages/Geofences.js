import React, { useState, useEffect, useCallback } from 'react';
import { geofenceAPI } from '../api';

function Geofences() {
  const [geofences, setGeofences] = useState([]);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: 'delivery_zone',
    coordinates: [[37.7749, -122.4194], [37.7849, -122.4194], [37.7849, -122.4094], [37.7749, -122.4094], [37.7749, -122.4194]],
  });
  const [filter, setFilter] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchGeofences = useCallback(async () => {
    try {
      setLoading(true);
      const response = await geofenceAPI.getAll(filter || undefined);
      setGeofences(response.data.geofences || []);
    } catch (err) {
      setError('Failed to fetch geofences');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [filter]);

  useEffect(() => {
    fetchGeofences();
  }, [fetchGeofences]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleCoordinateChange = (index, field, value) => {
    const newCoords = [...formData.coordinates];
    newCoords[index][field === 'lat' ? 0 : 1] = parseFloat(value);
    setFormData((prev) => ({
      ...prev,
      coordinates: newCoords,
    }));
  };

  const addCoordinate = () => {
    setFormData((prev) => ({
      ...prev,
      coordinates: [...prev.coordinates.slice(0, -1), [0, 0], prev.coordinates[prev.coordinates.length - 1]],
    }));
  };

  const removeCoordinate = (index) => {
    if (formData.coordinates.length > 4) {
      const newCoords = formData.coordinates.filter((_, i) => i !== index);
      setFormData((prev) => ({
        ...prev,
        coordinates: newCoords,
      }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await geofenceAPI.create(formData);
      setFormData({
        name: '',
        description: '',
        category: 'delivery_zone',
        coordinates: [[37.7749, -122.4194], [37.7849, -122.4194], [37.7849, -122.4094], [37.7749, -122.4094], [37.7749, -122.4194]],
      });
      fetchGeofences();
      setError('');
    } catch (err) {
      setError('Failed to create geofence');
      console.error(err);
    }
  };

  return (
    <div className="page">
      <h1 className="page-title">Geofence Management</h1>

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label>Name</label>
          <input
            type="text"
            name="name"
            value={formData.name}
            onChange={handleInputChange}
            required
          />
        </div>

        <div className="form-group">
          <label>Description</label>
          <textarea
            name="description"
            value={formData.description}
            onChange={handleInputChange}
            rows="3"
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label>Category</label>
            <select name="category" value={formData.category} onChange={handleInputChange}>
              <option value="delivery_zone">Delivery Zone</option>
              <option value="restricted_zone">Restricted Zone</option>
              <option value="toll_zone">Toll Zone</option>
              <option value="customer_area">Customer Area</option>
            </select>
          </div>
        </div>

        <div className="coordinates-input">
          <label>Coordinates (Latitude, Longitude)</label>
          {formData.coordinates.map((coord, index) => (
            <div key={index} className="coordinate-point">
              <input
                type="number"
                step="0.0001"
                value={coord[0]}
                onChange={(e) => handleCoordinateChange(index, 'lat', e.target.value)}
                placeholder="Latitude"
              />
              <input
                type="number"
                step="0.0001"
                value={coord[1]}
                onChange={(e) => handleCoordinateChange(index, 'lon', e.target.value)}
                placeholder="Longitude"
              />
              {formData.coordinates.length > 4 && index !== 0 && index !== formData.coordinates.length - 1 && (
                <button type="button" onClick={() => removeCoordinate(index)}>
                  Remove
                </button>
              )}
            </div>
          ))}
          <button type="button" onClick={addCoordinate} className="btn btn-secondary" style={{ marginTop: '0.5rem' }}>
            Add Coordinate
          </button>
        </div>

        {error && <div className="error">{error}</div>}

        <button type="submit" className="btn btn-primary" style={{ marginTop: '1rem' }}>
          Create Geofence
        </button>
      </form>

      <div className="list-container">
        <h2 className="list-title">All Geofences</h2>

        <div className="filter-group">
          <select value={filter} onChange={(e) => setFilter(e.target.value)}>
            <option value="">All Categories</option>
            <option value="delivery_zone">Delivery Zone</option>
            <option value="restricted_zone">Restricted Zone</option>
            <option value="toll_zone">Toll Zone</option>
            <option value="customer_area">Customer Area</option>
          </select>
        </div>

        {loading ? (
          <div className="loading">Loading geofences...</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Category</th>
                <th>Status</th>
                <th>Created At</th>
              </tr>
            </thead>
            <tbody>
              {geofences.map((geo) => (
                <tr key={geo.id}>
                  <td>{geo.name}</td>
                  <td>{geo.description || '-'}</td>
                  <td>{geo.category}</td>
                  <td>
                    <span className="badge badge-success">{geo.status}</span>
                  </td>
                  <td>{new Date(geo.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export default Geofences;
