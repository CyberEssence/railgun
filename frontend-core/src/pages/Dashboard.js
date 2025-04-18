import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  TextField,
  CircularProgress,
  Alert,
  Grid,
  Card,
  CardContent,
  CardHeader,
  Button
} from '@mui/material';
import { Home, Storage, BarChart, Refresh } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

export const Dashboard = () => {
  const [stats, setStats] = useState({
    totalEvents: 0,
    activeConnections: 0,
    suspiciousActivity: 0,
    systemHealth: 'healthy'
  });
  const [trafficStats, setTrafficStats] = useState({
    total_bytes_sent: 0,
    total_bytes_recv: 0,
    total_packets_sent: 0,
    total_packets_recv: 0,
    by_protocol: {},
    traffic_over_time: []
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [hostId, setHostId] = useState('');
  const navigate = useNavigate();

  const fetchDashboardData = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch('http://localhost:8080/api/dashboard/stats');
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
      }
      const dashboardData = await response.json();
      setStats(dashboardData);
      
      if (hostId) {
        // Add date range parameters (e.g., last 7 days)
        const to = new Date();
        const from = new Date();
        from.setDate(from.getDate() - 7);
        
        const trafficResponse = await fetch(
          `http://localhost:8080/api/traffic/stats/host/${hostId}?from=${from.toISOString()}&to=${to.toISOString()}`
        );
        
        if (!trafficResponse.ok) {
          const errorText = await trafficResponse.text();
          throw new Error(`HTTP error! status: ${trafficResponse.status}, message: ${errorText}`);
        }
        const trafficData = await trafficResponse.json();
        setTrafficStats(trafficData);  // Fixed typo here
      }
    } catch (err) {
      setError(err.message);
      console.error('Error fetching dashboard data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDashboardData();
  }, [hostId]);


  const handleNavigate = (path) => {
    navigate(path);
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">SIEM System Overview</Typography>
        <Button 
          variant="contained" 
          startIcon={<Refresh />}
          onClick={fetchDashboardData}
        >
          Refresh
        </Button>
      </Box>

      {/* Search bar */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <TextField
          fullWidth
          label="Host ID"
          value={hostId}
          onChange={(e) => setHostId(e.target.value)}
          margin="normal"
          placeholder="Enter host identifier"
        />
      </Paper>

      {/* General statistics */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardHeader 
              title="Total Events" 
              titleTypographyProps={{ variant: 'h6' }}
            />
            <CardContent>
              <Typography variant="h6">{stats.totalEvents.toLocaleString()}</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardHeader 
              title="Active Connections" 
              titleTypographyProps={{ variant: 'h6' }}
            />
            <CardContent>
              <Typography variant="h6">{stats.activeConnections}</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardHeader 
              title="Suspicious Activity" 
              titleTypographyProps={{ variant: 'h6' }}
            />
            <CardContent>
              <Typography variant="h6">{stats.suspiciousActivity}</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardHeader 
              title="System Status" 
              titleTypographyProps={{ variant: 'h6' }}
            />
            <CardContent>
              <Typography variant="h6">
                {stats.systemHealth === 'healthy' ? '✅ Healthy' : '⚠️ Issues'}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Detailed statistics */}
      {hostId && (
        <Paper sx={{ p: 2, mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            Traffic Statistics for {hostId}
          </Typography>
          
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <Card>
                <CardHeader 
                  title="Data Transfer" 
                  titleTypographyProps={{ variant: 'subtitle1' }}
                />
                <CardContent>
                  <Typography variant="body1">
                    Sent: {formatBytes(trafficStats.total_bytes_sent || 0)}
                  </Typography>
                  <Typography variant="body1">
                    Received: {formatBytes(trafficStats.total_bytes_recv || 0)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={12} md={6}>
              <Card>
                <CardHeader 
                  title="Packet Count" 
                  titleTypographyProps={{ variant: 'subtitle1' }}
                />
                <CardContent>
                  <Typography variant="body1">
                    Sent: {(trafficStats.total_packets_sent || 0).toLocaleString()} packets
                  </Typography>
                  <Typography variant="body1">
                    Received: {(trafficStats.total_packets_recv || 0).toLocaleString()} packets
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        </Paper>
      )}

      {/* Quick access cards */}
      <Grid container spacing={2}>
        <Grid item xs={12} sm={4}>
          <Card sx={{ cursor: 'pointer' }} onClick={() => handleNavigate('/')}>
            <CardContent sx={{ pt: 2 }}>
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Home sx={{ fontSize: 48, mb: 2 }} />
                <Typography variant="subtitle1">Dashboard</Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Card sx={{ cursor: 'pointer' }} onClick={() => handleNavigate('/traffic')}>
            <CardContent sx={{ pt: 2 }}>
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <BarChart sx={{ fontSize: 48, mb: 2 }} />
                <Typography variant="subtitle1">Network Traffic</Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Card sx={{ cursor: 'pointer' }} onClick={() => handleNavigate('/artifacts')}>
            <CardContent sx={{ pt: 2 }}>
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Storage sx={{ fontSize: 48, mb: 2 }} />
                <Typography variant="subtitle1">Artifacts</Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Loading and error handling */}
      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {error}
        </Alert>
      )}
      {loading && (
        <Box display="flex" justifyContent="center" p={4}>
          <CircularProgress />
        </Box>
      )}
    </Box>
  );
};