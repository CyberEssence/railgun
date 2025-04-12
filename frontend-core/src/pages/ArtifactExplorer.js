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
  MenuItem,
  Select,
  FormControl,
  InputLabel
} from '@mui/material';

export const ArtifactExplorer = () => {
  const [artifacts, setArtifacts] = useState([]);
  const [hostId, setHostId] = useState('');
  const [artifactType, setArtifactType] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const artifactTypes = [
    { value: '', label: 'All Types' },
    { value: 'registry', label: 'Registry' },
    { value: 'file', label: 'File' },
    { value: 'process', label: 'Process' }
  ];

  useEffect(() => {
    const fetchArtifacts = async () => {
      if (!hostId) {
        setArtifacts([]);
        return;
      }
      
      setLoading(true);
      setError(null);
      try {
        const url = `/api/artifacts/host/${hostId}${artifactType ? `?type=${artifactType}` : ''}`;
        const response = await fetch(url);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        // Handle case when data = null
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

  const formatValue = (value) => {
    if (value === null || value === undefined) return '-';
    
    if (typeof value === 'object') {
      try {
        return JSON.stringify(value);
      } catch (e) {
        return String(value);
      }
    }
    
    return String(value);
  };

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
        
        <FormControl fullWidth margin="normal">
          <InputLabel id="artifact-type-label">Artifact Type</InputLabel>
          <Select
            labelId="artifact-type-label"
            value={artifactType}
            onChange={(e) => setArtifactType(e.target.value)}
            label="Artifact Type"
          >
            {artifactTypes.map(option => (
              <MenuItem key={option.value} value={option.value}>
                {option.label}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
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
                <TableCell>Owner</TableCell>
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
                      {formatValue(artifact.value)}
                    </TableCell>
                    <TableCell>
                      {artifact.size ? `${artifact.size} bytes` : '-'}
                    </TableCell>
                    <TableCell>{artifact.owner || '-'}</TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={6} align="center">
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