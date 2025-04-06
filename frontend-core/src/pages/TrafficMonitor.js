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
  Alert
} from '@mui/material';

export const TrafficMonitor = () => {
  const [trafficData, setTrafficData] = useState([]);
  const [hostId, setHostId] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchTraffic = async () => {
      if (!hostId) {
        setTrafficData([]);
        return;
      }
      
      setLoading(true);
      setError(null);
      try {
        const response = await fetch(`http://localhost:8080/api/traffic/host/${hostId}`);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        // Обработка случая, когда data = null
        setTrafficData(Array.isArray(data) ? data : []);
      } catch (err) {
        setError(err.message);
        console.error('Error fetching traffic data:', err);
        setTrafficData([]);
      } finally {
        setLoading(false);
      }
    };

    const timer = setTimeout(() => {
      fetchTraffic();
    }, 500);

    return () => clearTimeout(timer);
  }, [hostId]);

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Network Traffic Monitor
      </Typography>
      
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

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Box display="flex" justifyContent="center" p={4}>
          <CircularProgress />
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Timestamp</TableCell>
                <TableCell>Source</TableCell>
                <TableCell>Destination</TableCell>
                <TableCell>Protocol</TableCell>
                <TableCell>Traffic</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {trafficData && trafficData.length > 0 ? (
                trafficData.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      {item.timestamp ? new Date(item.timestamp).toLocaleString() : '-'}
                    </TableCell>
                    <TableCell>
                      {item.src_ip || '-'}:{item.src_port || '-'}
                    </TableCell>
                    <TableCell>
                      {item.dst_ip || '-'}:{item.dst_port || '-'}
                    </TableCell>
                    <TableCell>{item.protocol || '-'}</TableCell>
                    <TableCell>
                      {item.bytes_sent || 0} bytes sent / {item.bytes_recv || 0} bytes received
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={5} align="center">
                    {hostId ? 'No traffic data found' : 'Enter host ID to search'}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  );
};
