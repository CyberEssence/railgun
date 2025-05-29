import React, { useState } from 'react';
import { 
  Box, Paper, Typography, TextField, Button, 
  Alert, Grid, Card, CardContent, Slider, 
  FormControl, InputLabel, Select, MenuItem 
} from '@mui/material';
import { Security, BugReport, Speed } from '@mui/icons-material';

export const CounterAttack = () => {
  const [targetIP, setTargetIP] = useState('');
  const [attackType, setAttackType] = useState('ddos');
  const [intensity, setIntensity] = useState(3);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  const attackTypes = [
    { value: 'ddos', label: 'DDoS' },
    { value: 'port_scan', label: 'Port Scan' },
    { value: 'brute_force', label: 'Brute Force' },
    { value: 'exploit', label: 'Exploit' },
    { value: 'custom', label: 'Custom Payload' }
  ];

  const handleCounterAttack = async () => {
    if (!targetIP) {
      setError('Target IP is required');
      return;
    }
    
    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      const response = await fetch('http://localhost:8080/api/ai/counter-attack', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target_ip: targetIP, attack_type: attackType, intensity })
      });
      
      if (!response.ok) throw new Error(await response.text());
      
      const result = await response.json();
      setSuccess(result.message);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Контратака
      </Typography>
      
      <Paper sx={{ p: 2, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Параметры контратаки
        </Typography>
        
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Целевой IP"
              value={targetIP}
              onChange={(e) => setTargetIP(e.target.value)}
              placeholder="192.168.1.1"
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <FormControl fullWidth>
              <InputLabel>Тип атаки</InputLabel>
              <Select
                value={attackType}
                onChange={(e) => setAttackType(e.target.value)}
                label="Тип атаки"
              >
                {attackTypes.map(type => (
                  <MenuItem key={type.value} value={type.value}>
                    {type.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12}>
            <Typography gutterBottom>Интенсивность: {intensity}</Typography>
            <Slider
              value={intensity}
              onChange={(e, newValue) => setIntensity(newValue)}
              min={1}
              max={5}
              marks={[
                { value: 1, label: 'Low' },
                { value: 3, label: 'Medium' },
                { value: 5, label: 'High' }
              ]}
            />
          </Grid>
        </Grid>
        
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
          <Button 
            variant="contained" 
            color="error"
            startIcon={<BugReport />}
            onClick={handleCounterAttack}
            disabled={loading || !targetIP}
          >
            Инициировать контратаку
          </Button>
        </Box>
      </Paper>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      
      {success && (
        <Alert severity="success" sx={{ mb: 2 }}>
          {success}
        </Alert>
      )}

      <Grid container spacing sx={{ mt: 2 }}>
        <Grid item xs={12} md={6}>
        <Card>
    <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
        <Security color="warning" sx={{ mr: 1 }} />
            <Typography variant="h6">Предупреждение</Typography>
                </Box>
            <Typography variant="body2">
                Контратака может нарушать законы и политики безопасности. Убедитесь, что у вас есть
                явное разрешение на проведение этих действий против указанной цели.
            </Typography>
    </CardContent>
</Card>
</Grid>
<Grid item xs={12} md={6}>
<Card>
<CardContent>
<Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
<Speed color="info" sx={{ mr: 1 }} />
<Typography variant="h6">Рекомендации</Typography>
</Box>
<Typography variant="body2">
Для тестирования используйте специально выделенные тестовые системы.
Начинайте с минимальной интенсивности. Все действия логируются и требуют
подтверждения.
</Typography>
</CardContent>
</Card>
</Grid>
</Grid>
</Box>
);
};