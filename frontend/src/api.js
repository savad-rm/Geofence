import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const geofenceAPI = {
  create: (data) => api.post('/geofences', data),
  getAll: (category) => api.get('/geofences', { params: category ? { category } : {} }),
};

export const vehicleAPI = {
  register: (data) => api.post('/vehicles', data),
  getAll: () => api.get('/vehicles'),
  updateLocation: (data) => api.post('/vehicles/location', data),
  getLocation: (vehicleID) => api.get(`/vehicles/location/${vehicleID}`),
};

export const alertAPI = {
  configure: (data) => api.post('/alerts/configure', data),
  getAll: (params) => api.get('/alerts', { params }),
};

export const violationAPI = {
  getHistory: (params) => api.get('/violations/history', { params }),
};

export default api;
