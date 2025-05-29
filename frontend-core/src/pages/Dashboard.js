import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  CircularProgress,
  Alert,
  Button,
  Divider
} from '@mui/material';
import { 
  Refresh,
  Security,
  NetworkCheck,
  Storage,
  Warning,
  CheckCircle,
  Error
} from '@mui/icons-material';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

export const Dashboard = () => {
  const [dashboardData, setDashboardData] = useState(null);
  const [trafficData, setTrafficData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const { authToken } = useAuth();
  const navigate = useNavigate();

  const fetchDashboardData = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Fetch dashboard stats
      const statsResponse = await fetch('http://localhost:8080/api/dashboard/stats', {
        headers: {
          'Authorization': `Bearer ${authToken}`
        }
      });
      
      if (!statsResponse.ok) throw new Error('Failed to fetch dashboard stats');
      const stats = await statsResponse.json();
      setDashboardData(stats);

      // Fetch recent traffic
      const trafficResponse = await fetch('http://localhost:8080/api/traffic/recent', {
        headers: {
          'Authorization': `Bearer ${authToken}`
        }
      });
      
      if (!trafficResponse.ok) throw new Error('Failed to fetch traffic data');
      const traffic = await trafficResponse.json();
      setTrafficData(traffic);
    } catch (err) {
      setError(err.message);
      console.error('Error fetching dashboard data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (authToken) {
      fetchDashboardData();
    }
  }, [authToken]);

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]);
  };

  const getHealthIcon = () => {
    if (!dashboardData) return <CircularProgress size={20} />;
    
    const health = dashboardData.traffic?.system_health || 'healthy';
    switch (health) {
      case 'critical':
        return <Error color="error" fontSize="large" />;
      case 'warning':
        return <Warning color="warning" fontSize="large" />;
      default:
        return <CheckCircle color="success" fontSize="large" />;
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">Dashboard Overview</Typography>
        <Button 
          variant="contained" 
          startIcon={<Refresh />}
          onClick={fetchDashboardData}
          disabled={loading}
        >
          Refresh
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {loading && !dashboardData ? (
        <Box display="flex" justifyContent="center" p={4}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <Grid container spacing={3} sx={{ mb: 3 }}>
            {/* System Health */}
            <Grid item xs={12} md={4}>
              <Card sx={{ height: '100%' }}>
                <CardContent sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                  {getHealthIcon()}
                  <Typography variant="h6" sx={{ mt: 1 }}>
                    System Status
                  </Typography>
                  <Typography variant="h4" sx={{ mt: 1 }}>
                    {dashboardData?.traffic?.system_health?.toUpperCase() || 'HEALTHY'}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>

            {/* Events */}
            <Grid item xs={12} md={4}>
              <Card sx={{ height: '100%' }}>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Events
                  </Typography>
                  <Typography variant="h3">
                    {dashboardData?.traffic?.total_events?.toLocaleString() || 0}
                  </Typography>
                  <Typography variant="subtitle2" color="text.secondary">
                    Total events in last 24h
                  </Typography>
                </CardContent>
              </Card>
            </Grid>

            {/* Active Connections */}
            <Grid item xs={12} md={4}>
              <Card sx={{ height: '100%' }}>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Active Connections
                  </Typography>
                  <Typography variant="h3">
                    {dashboardData?.traffic?.active_connections?.toLocaleString() || 0}
                  </Typography>
                  <Typography variant="subtitle2" color="text.secondary">
                    Current active connections
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          </Grid>

          {/* Threat Overview */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Threat Overview
              </Typography>
              <Grid container spacing={3}>
                <Grid item xs={12} md={6}>
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart
                      data={[
                        { name: 'Critical', value: dashboardData?.threats?.Critical || 0 },
                        { name: 'High', value: dashboardData?.threats?.High || 0 },
                        { name: 'Medium', value: dashboardData?.threats?.Medium || 0 },
                        { name: 'Low', value: dashboardData?.threats?.Low || 0 }
                      ]}
                      margin={{
                        top: 5,
                        right: 30,
                        left: 20,
                        bottom: 5,
                      }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="name" />
                      <YAxis />
                      <Tooltip />
                      <Legend />
                      <Bar dataKey="value" fill="#8884d8" name="Threat Count" />
                    </BarChart>
                  </ResponsiveContainer>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Paper sx={{ p: 2, textAlign: 'center' }}>
                        <Typography variant="h3" color="error">
                          {dashboardData?.threats?.Critical || 0}
                        </Typography>
                        <Typography variant="subtitle1">
                          Critical Threats
                        </Typography>
                      </Paper>
                    </Grid>
                    <Grid item xs={6}>
                      <Paper sx={{ p: 2, textAlign: 'center' }}>
                        <Typography variant="h3" color="warning.main">
                          {dashboardData?.threats?.High || 0}
                        </Typography>
                        <Typography variant="subtitle1">
                          High Threats
                        </Typography>
                      </Paper>
                    </Grid>
                    <Grid item xs={6}>
                      <Paper sx={{ p: 2, textAlign: 'center' }}>
                        <Typography variant="h3" color="info.main">
                          {dashboardData?.threats?.Medium || 0}
                        </Typography>
                        <Typography variant="subtitle1">
                          Medium Threats
                        </Typography>
                      </Paper>
                    </Grid>
                    <Grid item xs={6}>
                      <Paper sx={{ p: 2, textAlign: 'center' }}>
                        <Typography variant="h3" color="success.main">
                          {dashboardData?.threats?.Low || 0}
                        </Typography>
                        <Typography variant="subtitle1">
                          Low Threats
                        </Typography>
                      </Paper>
                    </Grid>
                  </Grid>
                </Grid>
              </Grid>
            </CardContent>
          </Card>

          {/* Network Traffic */}
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Network Traffic (Last 24h)
              </Typography>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart
                  data={trafficData}
                  margin={{
                    top: 5,
                    right: 30,
                    left: 20,
                    bottom: 5,
                  }}
                >
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="timestamp" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Line type="monotone" dataKey="bytes_sent" stroke="#8884d8" name="Bytes Sent" />
                  <Line type="monotone" dataKey="bytes_recv" stroke="#82ca9d" name="Bytes Received" />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          {/* Quick Actions */}
          <Typography variant="h6" gutterBottom>
            Quick Actions
          </Typography>
          <Grid container spacing={2} sx={{ mb: 3 }}>
            <Grid item xs={12} sm={4}>
              <Button 
                variant="contained" 
                startIcon={<Security />}
                fullWidth
                onClick={() => navigate('/threats')}
                sx={{ py: 2 }}
              >
                View Threats
              </Button>
            </Grid>
            <Grid item xs={12} sm={4}>
              <Button 
                variant="contained" 
                startIcon={<NetworkCheck />}
                fullWidth
                onClick={() => navigate('/traffic')}
                sx={{ py: 2 }}
              >
                Analyze Traffic
              </Button>
            </Grid>
            <Grid item xs={12} sm={4}>
              <Button 
                variant="contained" 
                startIcon={<Storage />}
                fullWidth
                onClick={() => navigate('/artifacts')}
                sx={{ py: 2 }}
              >
                Explore Artifacts
              </Button>
            </Grid>
          </Grid>
        </>
      )}
    </Box>
  );
};