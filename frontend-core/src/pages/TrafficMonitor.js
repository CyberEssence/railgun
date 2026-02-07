import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Paper, 
  Typography, 
  TextField, 
  Table, 
  TableBody, 
  TableCell, 
  TableContainer, 
  TableHead, 
  TableRow,
  CircularProgress,
  Alert,
  Button,
  Chip,
  Grid,
  Card,
  CardContent,
  FormControl,
  InputLabel,
  Select,
  MenuItem
} from '@mui/material';
import { useApi } from '../hooks/useApi';
import { useAuth } from '../context/AuthContext';
import UploadIcon from '@mui/icons-material/Upload';
import DownloadIcon from '@mui/icons-material/Download';
import NetworkCheckIcon from '@mui/icons-material/NetworkCheck';
import TimerIcon from '@mui/icons-material/Timer';
import RefreshIcon from '@mui/icons-material/Refresh';

export const TrafficMonitor = () => {

  const api = useApi();
  const { isAuthenticated, logout } = useAuth();
  
  const [trafficStats, setTrafficStats] = useState(null);
  const [trafficData, setTrafficData] = useState([]);
  const [hostId, setHostId] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [timeRange, setTimeRange] = useState('24h');

  const fetchTrafficStats = async () => {
    if (!hostId.trim()) {
      setError('Please enter a host ID');
      return;
    }
    
    if (!isAuthenticated()) { 
      setError('Please login to access traffic data');
      return;
    }
    
    setLoading(true);
    setError(null);
    
    try {
      // Рассчитываем временной диапазон
      const now = new Date();
      let fromDate;
      
      switch (timeRange) {
        case '1h':
          fromDate = new Date(now.getTime() - 60 * 60 * 1000);
          break;
        case '24h':
          fromDate = new Date(now.getTime() - 24 * 60 * 60 * 1000);
          break;
        case '7d':
          fromDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
          break;
        case '30d':
          fromDate = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
          break;
        default:
          fromDate = new Date(now.getTime() - 24 * 60 * 60 * 1000);
      }
      
      const params = new URLSearchParams({
        from: fromDate.toISOString(),
        to: now.toISOString()
      });
      
      const url = `/api/traffic/stats/${hostId}?${params.toString()}`;
      console.log('Fetching traffic stats:', url);
      
      const data = await api.get(url);
      console.log('Traffic stats response:', data);
      
      if (data && typeof data === 'object') {
        setTrafficStats(data);
      } else {
        throw new Error('Invalid response format');
      }
      
      // Генерируем демо-данные для таблицы
      generateDemoTrafficData();
      
    } catch (err) {
      console.error('Traffic stats error:', err);
      
      // Обработка ошибок авторизации
      if (err.message.includes('Authorization') || 
          err.message.includes('401') || 
          err.message.includes('authenticated') ||
          err.message.includes('Session expired')) {
        setError('Session expired. Please login again.');
        setTimeout(() => logout(), 2000);
      } else if (err.message.includes('404')) {
        setError(`No traffic data found for host: ${hostId}. Using demo data.`);
        // Демо-статистика
        setTrafficStats({
          total_bytes_sent: 125000000,
          total_bytes_recv: 89000000,
          total_packets_sent: 15000,
          total_packets_recv: 12000,
          average_duration: 0.85
        });
        generateDemoTrafficData();
      } else {
        setError(`Error: ${err.message}`);
      }
    } finally {
      setLoading(false);
    }
  };

  const generateDemoTrafficData = () => {
    const protocols = ['TCP', 'UDP', 'ICMP', 'HTTP', 'HTTPS'];
    const ips = ['192.168.1.', '10.0.0.', '172.16.0.'];
    
    const demoData = Array.from({ length: 10 }, (_, i) => ({
      id: i + 1,
      timestamp: new Date(Date.now() - Math.random() * 24 * 60 * 60 * 1000).toISOString(),
      src_ip: `${ips[Math.floor(Math.random() * ips.length)]}${Math.floor(Math.random() * 254) + 1}`,
      src_port: Math.floor(Math.random() * 65535),
      dst_ip: `${ips[Math.floor(Math.random() * ips.length)]}${Math.floor(Math.random() * 254) + 1}`,
      dst_port: Math.floor(Math.random() * 65535),
      protocol: protocols[Math.floor(Math.random() * protocols.length)],
      bytes_sent: Math.floor(Math.random() * 1000000),
      bytes_recv: Math.floor(Math.random() * 500000),
      packets_sent: Math.floor(Math.random() * 1000),
      packets_recv: Math.floor(Math.random() * 800),
      duration: Math.random() * 5
    }));
    
    setTrafficData(demoData);
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      if (hostId.trim() && isAuthenticated()) {
        fetchTrafficStats();
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [hostId, timeRange]);

  const formatBytes = (bytes) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (!isAuthenticated()) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="warning" sx={{ mb: 2 }}>
          Please login to access network traffic monitoring.
        </Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Network Traffic Monitor
      </Typography>
      
      {/* Панель управления */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2} alignItems="center">
          <Grid item xs={12} md={5}>
            <TextField
              fullWidth
              label="Host ID"
              value={hostId}
              onChange={(e) => setHostId(e.target.value)}
              placeholder="e.g., host-001 or 192.168.1.100"
              variant="outlined"
              size="small"
            />
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth size="small">
              <InputLabel>Time Range</InputLabel>
              <Select
                value={timeRange}
                onChange={(e) => setTimeRange(e.target.value)}
                label="Time Range"
              >
                <MenuItem value="1h">Last hour</MenuItem>
                <MenuItem value="24h">Last 24 hours</MenuItem>
                <MenuItem value="7d">Last 7 days</MenuItem>
                <MenuItem value="30d">Last 30 days</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={2}>
            <Button 
              variant="contained" 
              onClick={fetchTrafficStats}
              disabled={loading || !hostId.trim()}
              fullWidth
              startIcon={loading ? <CircularProgress size={20} /> : <RefreshIcon />}
            >
              {loading ? 'Loading...' : 'Refresh'}
            </Button>
          </Grid>
        </Grid>
      </Paper>

      {error && (
        <Alert 
          severity={error.includes('demo data') ? 'info' : 'error'} 
          sx={{ mb: 2 }}
          onClose={() => setError(null)}
        >
          {error}
        </Alert>
      )}

      {/* Статистика трафика */}
      {trafficStats && (
        <Paper sx={{ p: 2, mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            Traffic Statistics ({timeRange})
          </Typography>
          
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} md={3}>
              <Card variant="outlined">
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                    <UploadIcon color="primary" sx={{ mr: 1 }} />
                    <Typography variant="subtitle1">Data Sent</Typography>
                  </Box>
                  <Typography variant="h5">
                    {formatBytes(trafficStats.total_bytes_sent)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            
            <Grid item xs={12} sm={6} md={3}>
              <Card variant="outlined">
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                    <DownloadIcon color="secondary" sx={{ mr: 1 }} />
                    <Typography variant="subtitle1">Data Received</Typography>
                  </Box>
                  <Typography variant="h5">
                    {formatBytes(trafficStats.total_bytes_recv)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            
            <Grid item xs={12} sm={6} md={3}>
              <Card variant="outlined">
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                    <NetworkCheckIcon color="info" sx={{ mr: 1 }} />
                    <Typography variant="subtitle1">Total Packets</Typography>
                  </Box>
                  <Typography variant="h5">
                    {(trafficStats.total_packets_sent || 0) + (trafficStats.total_packets_recv || 0)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            
            <Grid item xs={12} sm={6} md={3}>
              <Card variant="outlined">
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                    <TimerIcon color="success" sx={{ mr: 1 }} />
                    <Typography variant="subtitle1">Avg Duration</Typography>
                  </Box>
                  <Typography variant="h5">
                    {(trafficStats.average_duration || 0).toFixed(2)}s
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          </Grid>
        </Paper>
      )}

      {/* Таблица трафика */}
      {trafficData.length > 0 && (
        <>
          <Typography variant="h6" gutterBottom>
            Recent Network Connections
          </Typography>
          
          <TableContainer component={Paper}>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Time</TableCell>
                  <TableCell>Source</TableCell>
                  <TableCell>Destination</TableCell>
                  <TableCell>Protocol</TableCell>
                  <TableCell>Traffic</TableCell>
                  <TableCell>Duration</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {trafficData.map((item) => (
                  <TableRow key={item.id} hover>
                    <TableCell>
                      {new Date(item.timestamp).toLocaleTimeString()}
                    </TableCell>
                    <TableCell>
                      <Box>
                        <Typography variant="body2" fontWeight="medium">
                          {item.src_ip}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          Port: {item.src_port}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Box>
                        <Typography variant="body2" fontWeight="medium">
                          {item.dst_ip}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          Port: {item.dst_port}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Chip 
                        label={item.protocol} 
                        size="small"
                        color={
                          item.protocol === 'TCP' ? 'primary' : 
                          item.protocol === 'UDP' ? 'secondary' : 
                          'default'
                        }
                      />
                    </TableCell>
                    <TableCell>
                      <Box>
                        <Typography variant="body2" fontSize="0.8rem">
                          ↑ {formatBytes(item.bytes_sent)}
                        </Typography>
                        <Typography variant="body2" fontSize="0.8rem">
                          ↓ {formatBytes(item.bytes_recv)}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      {(item.duration || 0).toFixed(2)}s
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </>
      )}
    </Box>
  );
};

export default TrafficMonitor;