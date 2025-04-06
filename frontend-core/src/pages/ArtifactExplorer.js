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

export const ArtifactExplorer = () => {
  const [artifacts, setArtifacts] = useState([]);
  const [hostId, setHostId] = useState('');
  const [artifactType, setArtifactType] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchArtifacts = async () => {
      if (!hostId) {
        setArtifacts([]);
        return;
      }
      
      setLoading(true);
      setError(null);
      try {
        const url = `http://localhost:8080/api/artifacts/host/${hostId}${artifactType ? `?type=${artifactType}` : ''}`;
        const response = await fetch(url);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        // Обработка случая, когда data = null
        setArtifacts(Array.isArray(data) ? data : []);
      } catch (err) {
        setError(err.message);
        console.error('Error fetching artifacts:', err);
        setArtifacts([]);
      } finally {
        setLoading(false);
      }
    };

    const timer = setTimeout(() => {
      fetchArtifacts();
    }, 500);

    return () => clearTimeout(timer);
  }, [hostId, artifactType]);

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Artifact Explorer
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
        <TextField
          fullWidth
          label="Artifact Type"
          value={artifactType}
          onChange={(e) => setArtifactType(e.target.value)}
          margin="normal"
          placeholder="e.g. registry, file, process"
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
                <TableCell>Type</TableCell>
                <TableCell>Path</TableCell>
                <TableCell>Value</TableCell>
                <TableCell>Size</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {artifacts && artifacts.length > 0 ? (
                artifacts.map((artifact) => (
                  <TableRow key={artifact.id}>
                    <TableCell>
                      {artifact.timestamp ? new Date(artifact.timestamp).toLocaleString() : '-'}
                    </TableCell>
                    <TableCell>{artifact.type || '-'}</TableCell>
                    <TableCell>{artifact.path || '-'}</TableCell>
                    <TableCell>
                      {artifact.value || '-'}
                    </TableCell>
                    <TableCell>
                      {artifact.size ? `${artifact.size} bytes` : '-'}
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={5} align="center">
                    {hostId ? 'No artifacts found' : 'Enter host ID to search'}
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
