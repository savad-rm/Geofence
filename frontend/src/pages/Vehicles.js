import React, { useState, useEffect } from 'react';
import { vehicleAPI } from '../api';

function Vehicles() {
  const [vehicles, setVehicles] = useState([]);
  const [formData, setFormData] = useState({
    vehicle_number: '',
    driver_name: '',
    vehicle_type: 'truck',
    phone: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchVehicles();
  }, []);

  const fetchVehicles = async () => {
    try {
      setLoading(true);
      const response = await vehicleAPI.getAll();
      setVehicles(response.data.vehicles || []);
    } catch (err) {
      setError('Failed to fetch vehicles');
      console.error(err);
    } finally {
      setLoading(false);
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
      await vehicleAPI.register(formData);
      setFormData({
        vehicle_number: '',
        driver_name: '',
        vehicle_type: 'truck',
        phone: '',
      });
      fetchVehicles();
      setError('');
    } catch (err) {
      setError('Failed to register vehicle');
      console.error(err);
    }
  };

  return (
    <div className="page">
      <h1 className="page-title">Vehicle Management</h1>

      <form onSubmit={handleSubmit}>
        <div className="form-row">
          <div className="form-group">
            <label>Vehicle Number</label>
            <input
              type="text"
              name="vehicle_number"
              value={formData.vehicle_number}
              onChange={handleInputChange}
              placeholder="e.g., KA-01-AB-1234"
              required
            />
          </div>
          <div className="form-group">
            <label>Driver Name</label>
            <input
              type="text"
              name="driver_name"
              value={formData.driver_name}
              onChange={handleInputChange}
              required
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label>Vehicle Type</label>
            <select name="vehicle_type" value={formData.vehicle_type} onChange={handleInputChange}>
              <option value="truck">Truck</option>
              <option value="car">Car</option>
              <option value="van">Van</option>
              <option value="motorcycle">Motorcycle</option>
            </select>
          </div>
          <div className="form-group">
            <label>Phone</label>
            <input
              type="tel"
              name="phone"
              value={formData.phone}
              onChange={handleInputChange}
              placeholder="+1234567890"
              required
            />
          </div>
        </div>

        {error && <div className="error">{error}</div>}

        <button type="submit" className="btn btn-primary">
          Register Vehicle
        </button>
      </form>

      <div className="list-container">
        <h2 className="list-title">All Vehicles</h2>

        {loading ? (
          <div className="loading">Loading vehicles...</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Vehicle Number</th>
                <th>Driver Name</th>
                <th>Type</th>
                <th>Phone</th>
                <th>Status</th>
                <th>Created At</th>
              </tr>
            </thead>
            <tbody>
              {vehicles.map((vehicle) => (
                <tr key={vehicle.id}>
                  <td>{vehicle.vehicle_number}</td>
                  <td>{vehicle.driver_name}</td>
                  <td>{vehicle.vehicle_type}</td>
                  <td>{vehicle.phone}</td>
                  <td>
                    <span className="badge badge-success">{vehicle.status}</span>
                  </td>
                  <td>{new Date(vehicle.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export default Vehicles;
