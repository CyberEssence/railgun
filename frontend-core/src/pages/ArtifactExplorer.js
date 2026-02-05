// ArtifactExplorer.js
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
  InputLabel,
  Button
} from '@mui/material';
import { useAuth } from '../context/AuthContext';
import { useApi } from '../hooks/useApi'; 
import { useNavigate } from 'react-router-dom';

export const ArtifactExplorer = () => {
  const api = useApi(); 
  const { isAuthenticated, logout, token } = useAuth();
  const navigate = useNavigate();
  const [artifacts, setArtifacts] = useState([]);
  const [pagination, setPagination] = useState({
    total: 0,
    page: 1,
    per_page: 20,
    total_pages: 0
  });
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

  const fetchArtifacts = async (page = 1) => {
    if (!hostId) {
      setArtifacts([]);
      return;
    }
    
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams();
      if (artifactType) params.append('type', artifactType);
      params.append('page', page.toString());
      params.append('per_page', pagination.per_page.toString());
      
      const data = await api.get(`/api/artifacts/host/${hostId}?${params.toString()}`);
      
      // Обновляем данные и пагинацию
      setArtifacts(Array.isArray(data.data) ? data.data : []);
      setPagination(data.meta || {
        total: 0,
        page: page,
        per_page: pagination.per_page,
        total_pages: 0
      });
      
    } catch (err) {
      setError(err.message);
      setArtifacts([]);
      setPagination({
        total: 0,
        page: 1,
        per_page: pagination.per_page,
        total_pages: 0
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      if (hostId && token) {
        fetchArtifacts(1);
      } else if (hostId && !token) {
        setError('Authentication required. Please login.');
        setArtifacts([]);
      } else {
        setArtifacts([]);
        setError(null);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [hostId, artifactType, token]);

  const handlePageChange = (newPage) => {
    fetchArtifacts(newPage);
  };

  const formatValue = (value) => {
    if (value === null || value === undefined) return '-';
    
    if (typeof value === 'object') {
      try {
        return JSON.stringify(value, null, 2);
      } catch (e) {
        return String(value);
      }
    }
    
    return String(value);
  };

  // Проверяем авторизацию
  if (!isAuthenticated) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert 
          severity="warning" 
          action={
            <Button 
              color="inherit" 
              size="small"
              onClick={() => navigate('/login')}
            >
              Login
            </Button>
          }
        >
          Please login to access the Artifact Explorer.
        </Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Artifact Explorer
      </Typography>
      
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="body2" color="text.secondary">
          Logged in with token: {token ? `${token.substring(0, 20)}...` : 'No token'}
        </Typography>
        <Button 
          variant="outlined" 
          size="small" 
          onClick={logout}
        >
          Logout
        </Button>
      </Box>
      
      <Paper sx={{ p: 2, mb: 3 }}>
        <TextField
          fullWidth
          label="Host ID"
          value={hostId}
          onChange={(e) => setHostId(e.target.value)}
          margin="normal"
          placeholder="Enter host identifier"
          helperText="Enter a host ID to search for artifacts"
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
        <Alert 
          severity="error" 
          sx={{ mb: 2 }}
          onClose={() => setError(null)}
        >
          {error}
        </Alert>
      )}

      {loading ? (
        <Box display="flex" justifyContent="center" p={4}>
          <CircularProgress />
        </Box>
      ) : (
        <>
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
                    <TableRow key={artifact.id} hover>
                      <TableCell>
                        {artifact.timestamp ? new Date(artifact.timestamp).toLocaleString() : '-'}
                      </TableCell>
                      <TableCell>{artifact.type || '-'}</TableCell>
                      <TableCell>
                        <Typography 
                          variant="body2" 
                          sx={{ 
                            fontFamily: 'monospace',
                            wordBreak: 'break-all'
                          }}
                        >
                          {artifact.path || '-'}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Box
                          sx={{
                            maxHeight: '100px',
                            overflow: 'auto',
                            fontFamily: 'monospace',
                            fontSize: '0.875rem'
                          }}
                        >
                          {formatValue(artifact.value)}
                        </Box>
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

          {/* Пагинация */}
          {pagination.total_pages > 1 && (
            <Box display="flex" justifyContent="center" mt={2} gap={1}>
              <Button
                variant="outlined"
                onClick={() => handlePageChange(pagination.page - 1)}
                disabled={pagination.page === 1}
              >
                Previous
              </Button>
              
              <Box display="flex" alignItems="center" mx={2}>
                Page {pagination.page} of {pagination.total_pages}
              </Box>
              
              <Button
                variant="outlined"
                onClick={() => handlePageChange(pagination.page + 1)}
                disabled={pagination.page === pagination.total_pages}
              >
                Next
              </Button>
            </Box>
          )}
        </>
      )}
    </Box>
  );
};