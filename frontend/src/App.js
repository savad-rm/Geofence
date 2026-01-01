import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import './App.css';
import Geofences from './pages/Geofences';
import Vehicles from './pages/Vehicles';
import Locations from './pages/Locations';
import Alerts from './pages/Alerts';
import Violations from './pages/Violations';
import AlertNotifications from './components/AlertNotifications';

function App() {
  const [alerts, setAlerts] = useState([]);

  useEffect(() => {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsURL = `${wsProtocol}//${window.location.host.split(':')[0]}:8080/ws/alerts`;

    try {
      const ws = new WebSocket(wsURL);

      ws.onopen = () => {
        console.log('WebSocket connected');
      };

      ws.onmessage = (event) => {
        const alert = JSON.parse(event.data);
        setAlerts((prev) => [alert, ...prev.slice(0, 9)]);
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      ws.onclose = () => {
        console.log('WebSocket disconnected');
      };

      return () => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.close();
        }
      };
    } catch (error) {
      console.error('WebSocket connection error:', error);
    }
  }, []);

  return (
    <Router>
      <div className="App">
        <nav className="navbar">
          <div className="nav-container">
            <h1 className="nav-logo">Geofencing System</h1>
            <ul className="nav-menu">
              <li><Link to="/">Geofences</Link></li>
              <li><Link to="/vehicles">Vehicles</Link></li>
              <li><Link to="/locations">Locations</Link></li>
              <li><Link to="/alerts">Alerts</Link></li>
              <li><Link to="/violations">Violations</Link></li>
            </ul>
          </div>
        </nav>

        <AlertNotifications alerts={alerts} />

        <main className="main-content">
          <Routes>
            <Route path="/" element={<Geofences />} />
            <Route path="/vehicles" element={<Vehicles />} />
            <Route path="/locations" element={<Locations />} />
            <Route path="/alerts" element={<Alerts />} />
            <Route path="/violations" element={<Violations />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
