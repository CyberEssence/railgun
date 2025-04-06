import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { CssBaseline, ThemeProvider, createTheme } from '@mui/material';
import { Dashboard } from './pages/Dashboard';
import { TrafficMonitor } from './pages/TrafficMonitor';
import { ArtifactExplorer } from './pages/ArtifactExplorer';
import { Navbar } from './components/Navbar';

const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#2196f3',
    },
    secondary: {
      main: '#f50057',
    },
  },
});

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <Navbar />
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/traffic" element={<TrafficMonitor />} />
          <Route path="/artifacts" element={<ArtifactExplorer />} />
        </Routes>
      </Router>
    </ThemeProvider>
  );
}

export default App;