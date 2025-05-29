import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { CssBaseline, ThemeProvider, createTheme } from '@mui/material';
import { Dashboard } from './pages/Dashboard';
import { TrafficMonitor } from './pages/TrafficMonitor';
import { ArtifactExplorer } from './pages/ArtifactExplorer';
import { ThreatAnalysis } from './pages/ThreatAnalysis';
import { LoginPage } from './pages/LoginPage';
import { Navbar } from './components/Navbar';
import { AuthProvider } from './context/AuthContext';

const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#2196f3',
    },
    secondary: {
      main: '#f50057',
    },
    background: {
      default: '#121212',
      paper: '#1e1e1e',
    },
  },
  typography: {
    fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
  },
});

function App() {
  return (
    <Router>
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AuthProvider>
          <Navbar />
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/traffic" element={<TrafficMonitor />} />
            <Route path="/artifacts" element={<ArtifactExplorer />} />
            <Route path="/threats" element={<ThreatAnalysis />} />
            <Route path="/login" element={<LoginPage />} />
          </Routes>
      </AuthProvider>
    </ThemeProvider>
    </Router>
  );
}

export default App;