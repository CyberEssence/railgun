import React, { useState, useEffect } from 'react';
import { Box, Paper, Typography, TextField, Button, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import axios from 'axios';

export const ArtifactExplorer = () => {
  const [artifacts, setArtifacts] = useState([]);
  const [hostId, setHostId] = useState('');
  const [artifactType, setArtifactType] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchArtifacts = async () => {
      setLoading(true);
      try {
        const response = await axios.get(`/api/artifacts/host/${hostId}`, {
          params: { type: artifactType }
        });
        setArtifacts(response.data);
      } catch (error) {
        console.error('Error fetching artifacts:', error);
      } finally {
        setLoading(false);
      }
    };

    if (hostId) {
      fetchArtifacts();
    }
  }, [hostId, artifactType]);

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Исследование артефактов
      </Typography>
      <Paper sx={{ p: 2, mb: 3 }}>
        <TextField
          fullWidth
          label="ID хоста"
          value={hostId}
          onChange={(e) => setHostId(e.target.value)}
          margin="normal"
        />
        <TextField
          fullWidth
          label="Тип артефакта"
          value={artifactType}
          onChange={(e) => setArtifactType(e.target.value)}
          margin="normal"
        />
      </Paper>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Время</TableCell>
              <TableCell>Тип</TableCell>
              <TableCell>Путь</TableCell>
              <TableCell>Значение</TableCell>
              <TableCell>Размер</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {artifacts.map((artifact) => (
              <TableRow key={artifact.id}>
                <TableCell>{new Date(artifact.timestamp).toLocaleString()}</TableCell>
                <TableCell>{artifact.type}</TableCell>
                <TableCell>{artifact.path}</TableCell>
                <TableCell>{artifact.value}</TableCell>
                <TableCell>{artifact.size} байт</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};