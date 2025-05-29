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
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Button,
  Select,
  MenuItem,
  FormControl,
  InputLabel
} from '@mui/material';
import { useAuth } from '../context/AuthContext';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

export const ThreatAnalysis = () => {
  const [threatStats, setThreatStats] = useState(null);
  const [threats, setThreats] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [dateRange, setDateRange] = useState({
    from: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
    to: new Date()
  });
  const [severityFilter, setSeverityFilter] = useState('all');
  const { authToken } = useAuth();

  const fetchThreatData = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const fromISO = dateRange.from.toISOString();
      const toISO = dateRange.to.toISOString();
      
      // Fetch threat statistics
      const statsResponse = await fetch(
        `http://localhost:8080/api/dashboard/stats?from=${fromISO}&to=${toISO}`,
        {
          headers: {
            'Authorization': `Bearer ${authToken}`
          }
        }
      );
      
      if (!statsResponse.ok) throw new Error('Failed to fetch threat stats');
      const statsData = await statsResponse.json();
      setThreatStats(statsData.threats || {
        Total: 0,
        Critical: 0,
        High: 0,
        Medium: 0,
        Low: 0
      });

      // Fetch threat list
      const threatsResponse = await fetch(
        `http://localhost:8080/api/ai/threats?from=${fromISO}&to=${toISO}`,
        {
          headers: {
            'Authorization': `Bearer ${authToken}`
          }
        }
      );
      
      if (!threatsResponse.ok) throw new Error('Failed to fetch threats list');
      const threatsData = await threatsResponse.json();
      setThreats(threatsData);
    } catch (err) {
      setError(err.message);
      console.error('Error fetching threat data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (authToken) {
      fetchThreatData();
    }
  }, [authToken, dateRange]);

  const handleDateChange = (field) => (newValue) => {
    setDateRange(prev => ({
      ...prev,
      [field]: newValue
    }));
  };

  const filteredThreats = threats.filter(threat => 
    severityFilter === 'all' || threat.severity === severityFilter
  );

  const severityData = [
    { name: 'Critical', value: threatStats?.Critical || 0 },
    { name: 'High', value: threatStats?.High || 0 },
    { name: 'Medium', value: threatStats?.Medium || 0 },
    { name: 'Low', value: threatStats?.Low || 0 }
  ];

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Threat Analysis
      </Typography>
      
      <Paper sx={{ p: 3, mb: 3 }}>
        <Grid container spacing={3} alignItems="center">
          <Grid item xs={12} md={3}>
            <LocalizationProvider dateAdapter={AdapterDateFns}>
              <DatePicker
                label="From Date"
                value={dateRange.from}
                onChange={handleDateChange('from')}
                renderInput={(params) => <TextField {...params} fullWidth />}
              />
            </LocalizationProvider>
          </Grid>
          <Grid item xs={12} md={3}>
            <LocalizationProvider dateAdapter={AdapterDateFns}>
              <DatePicker
                label="To Date"
                value={dateRange.to}
                onChange={handleDateChange('to')}
                renderInput={(params) => <TextField {...params} fullWidth />}
              />
            </LocalizationProvider>
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth>
              <InputLabel>Severity</InputLabel>
              <Select
                value={severityFilter}
                onChange={(e) => setSeverityFilter(e.target.value)}
                label="Severity"
              >
                <MenuItem value="all">All Severities</MenuItem>
                <MenuItem value="critical">Critical</MenuItem>
                <MenuItem value="high">High</MenuItem>
                <MenuItem value="medium">Medium</MenuItem>
                <MenuItem value="low">Low</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <Button 
              variant="contained" 
              onClick={fetchThreatData}
              fullWidth
              sx={{ height: '56px' }}
            >
              Refresh Data
            </Button>
          </Grid>
        </Grid>
      </Paper>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Box display="flex" justifyContent="center" p={4}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          <Grid container spacing={3} sx={{ mb: 3 }}>
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Threat Severity Distribution
                  </Typography>
                  <Box sx={{ height: 300 }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart
                        data={severityData}
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
                  </Box>
                </CardContent>
              </Card>
            </Grid>
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Threat Statistics
                  </Typography>
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Paper sx={{ p: 2, textAlign: 'center' }}>
                        <Typography variant="h3">
                          {threatStats?.Total || 0}
                        </Typography>
                        <Typography variant="subtitle1">
                          Total Threats
                        </Typography>
                      </Paper>
                    </Grid>
                    <Grid item xs={6}>
                      <Paper sx={{ p: 2, textAlign: 'center' }}>
                        <Typography variant="h3" color="error">
                          {threatStats?.Critical || 0}
                        </Typography>
                        <Typography variant="subtitle1">
                          Critical Threats
                        </Typography>
                      </Paper>
                    </Grid>
                    <Grid item xs={6}>
                      <Paper sx={{ p: 2, textAlign: 'center' }}>
                        <Typography variant="h3" color="warning.main">
                          {threatStats?.High || 0}
                        </Typography>
                        <Typography variant="subtitle1">
                          High Threats
                        </Typography>
                      </Paper>
                    </Grid>
                    <Grid item xs={6}>
                      <Paper sx={{ p: 2, textAlign: 'center' }}>
                        <Typography variant="h3" color="info.main">
                          {threatStats?.Medium || 0}
                        </Typography>
                        <Typography variant="subtitle1">
                          Medium Threats
                        </Typography>
                      </Paper>
                    </Grid>
                  </Grid>
                </CardContent>
              </Card>
            </Grid>
          </Grid>

          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Recent Threats
              </Typography>
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Timestamp</TableCell>
                      <TableCell>Type</TableCell>
                      <TableCell>Severity</TableCell>
                      <TableCell>Host</TableCell>
                      <TableCell>Description</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {filteredThreats.length > 0 ? (
                      filteredThreats.map((threat) => (
                        <TableRow key={threat.id}>
                          <TableCell>
                            {new Date(threat.timestamp).toLocaleString()}
                          </TableCell>
                          <TableCell>{threat.threat_type}</TableCell>
                          <TableCell>
                            <SeverityBadge severity={threat.severity} />
                          </TableCell>
                          <TableCell>{threat.host_id}</TableCell>
                          <TableCell>{threat.description}</TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={5} align="center">
                          No threats found
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>
        </>
      )}
    </Box>
  );
};

const SeverityBadge = ({ severity }) => {
  let color;
  switch (severity) {
    case 'critical':
      color = 'error';
      break;
    case 'high':
      color = 'warning';
      break;
    case 'medium':
      color = 'info';
      break;
    default:
      color = 'success';
  }

  return (
    <Box
      component="span"
      sx={{
        px: 1,
        py: 0.5,
        borderRadius: 1,
        backgroundColor: `${color}.light`,
        color: `${color}.contrastText`,
        fontWeight: 'bold',
        textTransform: 'capitalize'
      }}
    >
      {severity}
    </Box>
  );
};