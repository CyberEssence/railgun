import React, { useState, useEffect } from 'react';
import { 
  Box, Paper, Typography, Table, TableBody, TableCell, 
  TableContainer, TableHead, TableRow, CircularProgress, 
  Alert, Button, Dialog, DialogTitle, DialogContent, 
  DialogActions, TextField, Grid, Card, CardContent, 
  LinearProgress, Chip, IconButton, Tooltip, MenuItem
} from '@mui/material';
import { 
  Update, Train, Code, Delete, Info, 
  CheckCircle, Error, Pending, PlayArrow, Stop 
} from '@mui/icons-material';

export const AIModels = () => {
  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [modelType, setModelType] = useState('');
  const [openUpdateDialog, setOpenUpdateDialog] = useState(false);
  const [openTrainDialog, setOpenTrainDialog] = useState(false);
  const [selectedModel, setSelectedModel] = useState(null);
  const [trainProgress, setTrainProgress] = useState(null);
  const [trainEpochs, setTrainEpochs] = useState(10);
  const [datasetPath, setDatasetPath] = useState('');

  const modelTypes = [
    '', 'classification', 'anomaly_detection', 
    'time_series', 'nlp', 'image_recognition'
  ];

  const fetchModels = async () => {
    setLoading(true);
    setError(null);
    try {
      let url = 'http://localhost:8080/api/ai/models';
      if (modelType) url += `?type=${modelType}`;
      
      const response = await fetch(url);
      if (!response.ok) throw new Error(await response.text());
      
      const data = await response.json();
      setModels(data);
      
      // Check for active training jobs
      data.forEach(model => {
        if (model.training_status === 'in_progress') {
          checkTrainingProgress(model.id);
        }
      });
    } catch (err) {
      setError(err.message);
      console.error('Error fetching AI models:', err);
    } finally {
      setLoading(false);
    }
  };

  const checkTrainingProgress = async (modelId) => {
    try {
      const response = await fetch(`http://localhost:8080/api/ai/models/train/status?model_id=${modelId}`);
      if (response.ok) {
        const progress = await response.json();
        setModels(prev => prev.map(m => 
          m.id === modelId ? { ...m, training_status: progress.status, accuracy: progress.accuracy } : m
        ));
        
        if (progress.status === 'in_progress') {
          setTimeout(() => checkTrainingProgress(modelId), 5000);
        }
      }
    } catch (err) {
      console.error('Error checking training progress:', err);
    }
  };

  useEffect(() => {
    fetchModels();
  }, [modelType]);

  const handleUpdateModel = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/ai/models/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ model_ids: [selectedModel.id] })
      });
      
      if (!response.ok) throw new Error(await response.text());
      
      const result = await response.json();
      setModels(prev => prev.map(m => 
        m.id === selectedModel.id ? { ...m, version: result.new_version } : m
      ));
      setOpenUpdateDialog(false);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleTrainModel = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/ai/models/train', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          model_id: selectedModel.id,
          dataset_path: datasetPath,
          epochs: trainEpochs
        })
      });
      
      if (!response.ok) throw new Error(await response.text());
      
      const result = await response.json();
      setModels(prev => prev.map(m => 
        m.id === selectedModel.id ? { ...m, training_status: 'in_progress' } : m
      ));
      setOpenTrainDialog(false);
      checkTrainingProgress(selectedModel.id);
    } catch (err) {
      setError(err.message);
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'ready': return <CheckCircle color="success" />;
      case 'in_progress': return <Pending color="warning" />;
      case 'error': return <Error color="error" />;
      default: return <Info color="info" />;
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Управление AI моделями
      </Typography>
      
      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2}>
          <Grid item xs={12} md={4}>
            <TextField
              select
              fullWidth
              label="Тип модели"
              value={modelType}
              onChange={(e) => setModelType(e.target.value)}
            >
              {modelTypes.map(type => (
                <MenuItem key={type || 'all'} value={type}>
                  {type || 'Все типы'}
                </MenuItem>
              ))}
            </TextField>
          </Grid>
          <Grid item xs={12} md={8} sx={{ display: 'flex', alignItems: 'center' }}>
            <Button 
              variant="contained" 
              startIcon={<Update />}
              onClick={fetchModels}
              sx={{ mr: 2 }}
            >
              Обновить список
            </Button>
          </Grid>
        </Grid>
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
                <TableCell>ID</TableCell>
                <TableCell>Название</TableCell>
                <TableCell>Тип</TableCell>
                <TableCell>Версия</TableCell>
                <TableCell>Статус</TableCell>
                <TableCell>Точность</TableCell>
                <TableCell>Действия</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {models.length > 0 ? (
                models.map((model) => (
                  <TableRow key={model.id}>
                    <TableCell>{model.id}</TableCell>
                    <TableCell sx={{ fontWeight: 'bold' }}>{model.name}</TableCell>
                    <TableCell>
                      <Chip label={model.type} variant="outlined" />
                    </TableCell>
                    <TableCell>v{model.version}</TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center' }}>
                        {getStatusIcon(model.training_status)}
                        <Typography sx={{ ml: 1 }}>
                          {model.training_status === 'ready' ? 'Готов' : 
                           model.training_status === 'in_progress' ? 'Обучение' : 
                           model.training_status === 'error' ? 'Ошибка' : 'Неизвестно'}
                        </Typography>
                      </Box>
                      {model.training_status === 'in_progress' && (
                        <LinearProgress sx={{ mt: 1 }} />
                      )}
                    </TableCell>
                    <TableCell>
                      {model.accuracy ? `${(model.accuracy * 100).toFixed(2)}%` : 'N/A'}
                    </TableCell>
                    <TableCell>
                      <Tooltip title="Обновить модель">
                        <IconButton 
                          onClick={() => {
                            setSelectedModel(model);
                            setOpenUpdateDialog(true);
                          }}
                        >
                          <Update />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Обучить модель">
                        <IconButton 
                          onClick={() => {
                            setSelectedModel(model);
                            setDatasetPath(`/datasets/${model.type}/latest.csv`);
                            setOpenTrainDialog(true);
                          }}
                        >
                          <Train />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Показать код">
                        <IconButton>
                          <Code />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={7} align="center">
                    Нет моделей
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* Update Model Dialog */}
      <Dialog open={openUpdateDialog} onClose={() => setOpenUpdateDialog(false)}>
        <DialogTitle>Обновить модель {selectedModel?.name}</DialogTitle>
        <DialogContent>
          <Typography>Вы уверены, что хотите обновить модель до последней версии?</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
            Текущая версия: v{selectedModel?.version}
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenUpdateDialog(false)}>Отмена</Button>
          <Button 
            onClick={handleUpdateModel}
            variant="contained"
            startIcon={<Update />}
          >
            Обновить
          </Button>
        </DialogActions>
      </Dialog>

      {/* Train Model Dialog */}
      <Dialog open={openTrainDialog} onClose={() => setOpenTrainDialog(false)}>
        <DialogTitle>Обучить модель {selectedModel?.name}</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Путь к датасету"
                value={datasetPath}
                onChange={(e) => setDatasetPath(e.target.value)}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Количество эпох"
                type="number"
                value={trainEpochs}
                onChange={(e) => setTrainEpochs(parseInt(e.target.value))}
                inputProps={{ min: 1, max: 100 }}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenTrainDialog(false)}>Отмена</Button>
          <Button 
            onClick={handleTrainModel}
            variant="contained"
            startIcon={<Train />}
          >
            Начать обучение
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};