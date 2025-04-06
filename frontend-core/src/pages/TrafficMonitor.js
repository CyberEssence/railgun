import React, { useState, useEffect } from 'react';
import { Box, Paper, Typography, TextField, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import axios from 'axios';

export const TrafficMonitor = () => {
  const [trafficData, setTrafficData] = useState([]);
  const [hostId, setHostId] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchTraffic = async () => {
      setLoading(true);
      try {
        const response = await axios.get(`/api/traffic/host/${hostId}`);
        setTrafficData(response.data);
      } catch (error) {
        console.error('Ошибка при загрузке данных трафика:', error);
        alert('Ошибка при загрузке данных трафика: ' + error.message);
      } finally {
        setLoading(false);
      }
    };

    if (hostId) {
      fetchTraffic();
    }
  }, [hostId]);

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Мониторинг сетевого трафика
      </Typography>
      <Paper sx={{ p: 2, mb: 3 }}>
        <TextField
          fullWidth
          label="ID хоста"
          value={hostId}
          onChange={(e) => setHostId(e.target.value)}
          margin="normal"
        />
      </Paper>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Время</TableCell>
              <TableCell>Источник</TableCell>
              <TableCell>Назначение</TableCell>
              <TableCell>Протокол</TableCell>
              <TableCell>Трафик</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {trafficData.map((item) => (
              <TableRow key={item.id}>
                <TableCell>{new Date(item.timestamp).toLocaleString()}</TableCell>
                <TableCell>{item.src_ip}:{item.src_port}</TableCell>
                <TableCell>{item.dst_ip}:{item.dst_port}</TableCell>
                <TableCell>{item.protocol}</TableCell>
                <TableCell>
                  {item.bytes_sent} байт отправлено / {item.bytes_recv} байт получено
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};
