import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { CssBaseline, ThemeProvider, createTheme } from '@mui/material';
import { Dashboard } from './pages/Dashboard';
import { TrafficMonitor } from './pages/TrafficMonitor';
import { ArtifactExplorer } from './pages/ArtifactExplorer';
import { AttackPatterns } from './pages/AttackPatterns';
import { APTTimeline } from './pages/APTTimeline';
import { AIModels } from './pages/AIModels';
import { CounterAttack } from './pages/CounterAttack';
import { Navbar } from './components/Navbar';
import { Test2FA } from './components/Test2FA';
import { AuthProvider } from './context/AuthContext';
import { Register } from './pages/Register';
import { Login } from './pages/Login';
import { PrivateRoute } from './components/PrivateRoute';
import { FileScanner } from './pages/FileScanner';
import { Box } from '@mui/material';
import { TwoFASettings } from './components/TwoFASettings';
import { Profile } from './pages/Profile';

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
    h4: {
      fontWeight: 600,
    },
  },
});

function App() {
  return (
    <Router>
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AuthProvider>
          <Navbar />
          <Box sx={{ p: 3 }}>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
              <Route path="/test-2fa" element={<Test2FA />} />
              <Route path="/profile" element={<PrivateRoute><Profile /></PrivateRoute>} />
              <Route path="/" element={<PrivateRoute><Dashboard /></PrivateRoute>} />
              <Route path="/traffic" element={<PrivateRoute><TrafficMonitor /></PrivateRoute>} />
              <Route path="/artifacts" element={<PrivateRoute><ArtifactExplorer /></PrivateRoute>} />
              <Route path="/attack-patterns" element={<PrivateRoute><AttackPatterns /></PrivateRoute>} />
              <Route path="/apt-timeline" element={<PrivateRoute><APTTimeline /></PrivateRoute>} />
              <Route path="/ai-models" element={<PrivateRoute><AIModels /></PrivateRoute>} />
              <Route path="/counter-attack" element={<PrivateRoute><CounterAttack /></PrivateRoute>} />
              <Route path="/file-scanner" element={<PrivateRoute><FileScanner /></PrivateRoute>} />
              <Route 
               path="/settings/2fa" 
                element={
                  <PrivateRoute>
                   <TwoFASettings />
                  </PrivateRoute>
                } 
              />
            </Routes>
          </Box>
      </AuthProvider>
    </ThemeProvider>
    </Router>
  );
}

export default App;