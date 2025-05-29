import React, { useState } from 'react';
import { 
  Box, Paper, Typography, Button, Alert, 
  LinearProgress, Card, CardContent, List, 
  ListItem, ListItemText, Divider, Chip, Grid
} from '@mui/material';
import { CloudUpload, Security, CheckCircle, Error } from '@mui/icons-material';

export const FileScanner = () => {
  const [file, setFile] = useState(null);
  const [scanResult, setScanResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [progress, setProgress] = useState(0);

  const handleFileChange = (e) => {
    setFile(e.target.files[0]);
    setScanResult(null);
  };

  const handleFileUpload = async () => {
    if (!file) {
      setError('Please select a file first');
      return;
    }
    
    setLoading(true);
    setError(null);
    setProgress(0);
    
    const formData = new FormData();
    formData.append('file', file);
    
    try {
      const response = await fetch('http://localhost:8080/api/integration/scan', {
        method: 'POST',
        body: formData,
      });
      
      if (!response.ok) throw new Error(await response.text());
      
      const result = await response.json();
      setScanResult(result);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
      setProgress(0);
    }
  };

  const getThreatLevelColor = (level) => {
    switch (level) {
      case 'high': return 'error';
      case 'medium': return 'warning';
      case 'low': return 'info';
      default: return 'default';
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Сканер файлов
      </Typography>
      
      <Paper sx={{ p: 2, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Загрузите файл для проверки
        </Typography>
        
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Button
            variant="contained"
            component="label"
            startIcon={<CloudUpload />}
            sx={{ mr: 2 }}
          >
            Выбрать файл
            <input
              type="file"
              hidden
              onChange={handleFileChange}
            />
          </Button>
          
          <Typography>
            {file ? file.name : 'Файл не выбран'}
          </Typography>
        </Box>
        
        {loading && (
          <Box sx={{ width: '100%', mb: 2 }}>
            <LinearProgress variant="determinate" value={progress} />
            <Typography variant="caption" display="block" textAlign="center">
              {progress}% завершено
            </Typography>
          </Box>
        )}
        
        <Button
          variant="contained"
          color="primary"
          onClick={handleFileUpload}
          disabled={!file || loading}
          fullWidth
        >
          Сканировать файл
        </Button>
      </Paper>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {scanResult && (
        <Card>
          <CardContent>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
              {scanResult.malicious ? (
                <Error color="error" sx={{ fontSize: 40, mr: 2 }} />
              ) : (
                <CheckCircle color="success" sx={{ fontSize: 40, mr: 2 }} />
              )}
              <Typography variant="h5">
                {scanResult.malicious ? 'Обнаружены угрозы!' : 'Угроз не обнаружено'}
              </Typography>
            </Box>
            
            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle1">Детали файла:</Typography>
                <List dense>
                  <ListItem>
                    <ListItemText primary="Имя файла" secondary={scanResult.file_info.name} />
                  </ListItem>
                  <ListItem>
                    <ListItemText primary="Тип" secondary={scanResult.file_info.type} />
                  </ListItem>
                  <ListItem>
                    <ListItemText primary="Размер" secondary={`${(scanResult.file_info.size / 1024).toFixed(2)} KB`} />
                  </ListItem>
                  <ListItem>
                    <ListItemText primary="Хэш (MD5)" secondary={scanResult.file_info.md5} />
                  </ListItem>
                </List>
              </Grid>
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle1">Результаты сканирования:</Typography>
                <List dense>
                  <ListItem>
                    <ListItemText 
                      primary="Статус" 
                      secondary={
                        <Chip 
                          label={scanResult.malicious ? 'Опасный' : 'Безопасный'} 
                          color={scanResult.malicious ? 'error' : 'success'} 
                          size="small" 
                        />
                      } 
                    />
                  </ListItem>
                  {scanResult.scan_results.map((result, index) => (
                    <React.Fragment key={index}>
                      <ListItem>
                        <ListItemText
                          primary={result.engine}
                          secondary={
                            <>
                              <Chip 
                                label={result.threat_level} 
                                color={getThreatLevelColor(result.threat_level)} 
                                size="small" 
                                sx={{ mr: 1 }}
                              />
                              {result.threat_name || 'No threats detected'}
                            </>
                          }
                        />
                      </ListItem>
                      {index < scanResult.scan_results.length - 1 && <Divider component="li" />}
                    </React.Fragment>
                  ))}
                </List>
              </Grid>
            </Grid>
          </CardContent>
        </Card>
      )}
    </Box>
  );
};