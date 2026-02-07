import React, { useState, useEffect } from 'react';
import { 
  Box, Paper, Typography, Table, TableBody, TableCell, 
  TableContainer, TableHead, TableRow, CircularProgress, 
  Alert, Button, Dialog, DialogTitle, DialogContent, 
  DialogActions, TextField, Grid, LinearProgress, 
  Chip, IconButton, Tooltip, MenuItem,
  Snackbar
} from '@mui/material';
import UpdateIcon from '@mui/icons-material/Update';
import TrainIcon from '@mui/icons-material/Train';
import CodeIcon from '@mui/icons-material/Code';
import DeleteIcon from '@mui/icons-material/Delete';
import InfoIcon from '@mui/icons-material/Info';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import PendingIcon from '@mui/icons-material/Pending';
import AddIcon from '@mui/icons-material/Add';
import { useApi } from '../hooks/useApi';
import { useAuth } from '../context/AuthContext';

export const AIModels = () => {
  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [modelType, setModelType] = useState('');
  const [openUpdateDialog, setOpenUpdateDialog] = useState(false);
  const [openTrainDialog, setOpenTrainDialog] = useState(false);
  const [openCreateDialog, setOpenCreateDialog] = useState(false);
  const [selectedModel, setSelectedModel] = useState(null);
  const [trainEpochs, setTrainEpochs] = useState(10);
  const [datasetPath, setDatasetPath] = useState('');

  const [newModelName, setNewModelName] = useState('');
  const [newModelType, setNewModelType] = useState('');

  const api = useApi();
  const { user, logout, isAuthenticated } = useAuth();

  const modelTypes = [
    '', 'classification', 'anomaly_detection', 
    'time_series', 'nlp', 'image_recognition'
  ];

  // Функция для дебаггинга - логирует текущее состояние
  const debugState = () => {
    console.log('=== DEBUG INFO ===');
    console.log('User:', user);
    console.log('Is authenticated:', isAuthenticated());
    console.log('User token:', user?.token ? 'YES' : 'NO');
    console.log('Token length:', user?.token?.length || 0);
    console.log('Models count:', models.length);
    console.log('Loading:', loading);
    console.log('Error:', error);
    console.log('==================');
  };

  const fetchModels = async () => {
    debugState(); 
    
    if (!isAuthenticated()) {
      setError('Not authenticated. Please login first.');
      return;
    }

    setLoading(true);
    setError(null);
    
    try {
      console.log('Fetching models...');
      
      let url = '/api/ai/models';
      const params = new URLSearchParams();
      
      if (modelType) params.append('type', modelType);
      if (params.toString()) url += `?${params.toString()}`;
      
      console.log('Request URL:', url);
      console.log('User token present:', !!user?.token);
      
      const data = await api.get(url);
      console.log('Response data:', data);
      
      // Обработка ответа
      if (Array.isArray(data)) {
        setModels(data);
      } else if (data && data.models) {
        setModels(data.models);
      } else if (data && data.data) {
        setModels(data.data);
      } else {
        console.warn('Unexpected response format:', data);
        setModels([]);
      }
      
      setSuccess(`Loaded ${models.length} models`);
      
    } catch (err) {
      console.error('Error fetching AI models:', err);
      setError(`Failed to load models: ${err.message}`);
      
      // Если ошибка авторизации
      if (err.message.includes('Authorization') || 
          err.message.includes('401') || 
          err.message.includes('authenticated')) {
        logout();
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isAuthenticated()) {
      fetchModels();
    } else {
      setError('Please login to access AI models');
    }
  }, [modelType, isAuthenticated()]);

  const handleUpdateModel = async () => {
    if (!selectedModel) return;
    
    try {
      const result = await api.post('/api/ai/models/update', {
        model_ids: [selectedModel.id]
      });
      
      setModels(prev => prev.map(m => 
        m.id === selectedModel.id ? { 
          ...m, 
          version: result.new_version || result.version || (m.version + 1) 
        } : m
      ));
      
      setOpenUpdateDialog(false);
      setSuccess(`Model ${selectedModel.name} updated successfully!`);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleTrainModel = async () => {
    if (!selectedModel) return;
    
    try {
      const result = await api.post('/api/ai/models/train', {
        model_id: selectedModel.id,
        dataset_path: datasetPath || `/datasets/${selectedModel.type}/latest.csv`,
        epochs: trainEpochs,
        batch_size: 32
      });
      
      setModels(prev => prev.map(m => 
        m.id === selectedModel.id ? { 
          ...m, 
          training_status: 'in_progress',
          last_training_started: new Date().toISOString()
        } : m
      ));
      
      setOpenTrainDialog(false);
      setSuccess(`Training started for model ${selectedModel.name}`);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleCreateModel = async () => {
    try {
      const result = await api.post('/api/ai/models', {
        name: newModelName,
        type: newModelType,
        description: `New ${newModelType} model`,
        config: {
          layers: [128, 64, 32],
          dropout: 0.2,
          activation: 'relu'
        }
      });
      
      const newModel = {
        id: result.id || Date.now(),
        name: newModelName,
        type: newModelType,
        version: 1,
        training_status: 'ready',
        accuracy: 0,
        created_at: new Date().toISOString()
      };
      
      setModels(prev => [...prev, newModel]);
      setOpenCreateDialog(false);
      setSuccess(`Model ${newModelName} created successfully!`);
      
      // Сброс формы
      setNewModelName('');
      setNewModelType('');
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDeleteModel = async (modelId, modelName) => {
    if (window.confirm(`Are you sure you want to delete model "${modelName}"?`)) {
      try {
        await api.delete(`/api/ai/models/${modelId}`);
        setModels(prev => prev.filter(m => m.id !== modelId));
        setSuccess(`Model ${modelName} deleted successfully!`);
      } catch (err) {
        setError(err.message);
      }
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'ready': return <CheckCircleIcon color="success" />;
      case 'in_progress': return <PendingIcon color="warning" />;
      case 'error': return <ErrorIcon color="error" />;
      default: return <InfoIcon color="info" />;
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return dateString;
    }
  };

  // Если не авторизован
  if (!isAuthenticated()) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="warning">
          Please login to access AI models management.
        </Alert>
        <Button 
          variant="contained" 
          onClick={debugState}
          sx={{ mt: 2 }}
        >
          Debug State
        </Button>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">
          AI Models Management
        </Typography>
        <Button 
          variant="outlined" 
          onClick={debugState}
          size="small"
        >
          Debug
        </Button>
      </Box>
      
      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2} alignItems="center">
          <Grid item xs={12} md={4}>
            <TextField
              select
              fullWidth
              label="Filter by Type"
              value={modelType}
              onChange={(e) => setModelType(e.target.value)}
              size="small"
            >
              {modelTypes.map(type => (
                <MenuItem key={type || 'all'} value={type}>
                  {type || 'All Types'}
                </MenuItem>
              ))}
            </TextField>
          </Grid>
          <Grid item xs={12} md={8} sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Button 
              variant="contained" 
              startIcon={<UpdateIcon />}
              onClick={fetchModels}
              disabled={loading}
            >
              {loading ? 'Loading...' : 'Refresh'}
            </Button>
            <Button 
              variant="outlined" 
              startIcon={<AddIcon />}
              onClick={() => setOpenCreateDialog(true)}
            >
              New Model
            </Button>
            <Typography variant="body2" color="text.secondary">
              Logged in as: {user?.username}
            </Typography>
          </Grid>
        </Grid>
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

      {success && (
        <Alert 
          severity="success" 
          sx={{ mb: 2 }}
          onClose={() => setSuccess(null)}
        >
          {success}
        </Alert>
      )}

      {loading ? (
        <Box display="flex" justifyContent="center" alignItems="center" p={4}>
          <CircularProgress />
          <Typography sx={{ ml: 2 }}>Loading models...</Typography>
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>Name</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Version</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Accuracy</TableCell>
                <TableCell>Created</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {models.length > 0 ? (
                models.map((model) => (
                  <TableRow key={model.id} hover>
                    <TableCell>{model.id}</TableCell>
                    <TableCell sx={{ fontWeight: 'bold' }}>{model.name}</TableCell>
                    <TableCell>
                      <Chip 
                        label={model.type} 
                        variant="outlined" 
                        size="small"
                      />
                    </TableCell>
                    <TableCell>v{model.version}</TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center' }}>
                        {getStatusIcon(model.training_status)}
                        <Typography sx={{ ml: 1, fontSize: '0.875rem' }}>
                          {model.training_status === 'ready' ? 'Ready' : 
                           model.training_status === 'in_progress' ? 'Training' : 
                           model.training_status === 'error' ? 'Error' : 'Unknown'}
                        </Typography>
                      </Box>
                      {model.training_status === 'in_progress' && (
                        <LinearProgress sx={{ mt: 1, height: 4 }} />
                      )}
                    </TableCell>
                    <TableCell>
                      {model.accuracy ? `${(model.accuracy * 100).toFixed(1)}%` : 'N/A'}
                    </TableCell>
                    <TableCell>
                      {formatDate(model.created_at)}
                    </TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', gap: 0.5 }}>
                        <Tooltip title="Update model">
                          <IconButton 
                            size="small"
                            onClick={() => {
                              setSelectedModel(model);
                              setOpenUpdateDialog(true);
                            }}
                            disabled={model.training_status === 'in_progress'}
                          >
                            <UpdateIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Train model">
                          <IconButton 
                            size="small"
                            onClick={() => {
                              setSelectedModel(model);
                              setDatasetPath(`/datasets/${model.type}/latest.csv`);
                              setOpenTrainDialog(true);
                            }}
                            disabled={model.training_status === 'in_progress'}
                          >
                            <TrainIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Delete model">
                          <IconButton 
                            size="small"
                            color="error"
                            onClick={() => handleDeleteModel(model.id, model.name)}
                            disabled={model.training_status === 'in_progress'}
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      </Box>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={8} align="center" sx={{ py: 4 }}>
                    <Typography color="text.secondary">
                      No models found. Create your first model.
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* Диалоги остаются без изменений */}
      <Dialog open={openUpdateDialog} onClose={() => setOpenUpdateDialog(false)}>
        <DialogTitle>Update Model {selectedModel?.name}</DialogTitle>
        <DialogContent>
          <Typography>Update this model to the latest version?</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenUpdateDialog(false)}>Cancel</Button>
          <Button onClick={handleUpdateModel} variant="contained">
            Update
          </Button>
        </DialogActions>
      </Dialog>

      {/* Остальные диалоги аналогично... */}

      <Snackbar
        open={!!success}
        autoHideDuration={3000}
        onClose={() => setSuccess(null)}
        message={success}
      />
    </Box>
  );
};

export default AIModels;
