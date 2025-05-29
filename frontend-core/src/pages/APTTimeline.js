import React, { useState, useEffect } from 'react';
import { 
  Box, Paper, Typography, TextField, CircularProgress, 
  Alert, Button, FormControl, InputLabel, Select, MenuItem, Grid
} from '@mui/material';
import { Timeline, TimelineItem, TimelineSeparator, TimelineDot, 
  TimelineConnector, TimelineContent, TimelineOppositeContent } from '@mui/lab';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';

export const APTTimeline = () => {
  const [timelineData, setTimelineData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [hostId, setHostId] = useState('');
  const [fromDate, setFromDate] = useState(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000));
  const [toDate, setToDate] = useState(new Date());
  const [groupBy, setGroupBy] = useState('day');

  const fetchAPTTimeline = async () => {
    if (!hostId) return;
    
    setLoading(true);
    setError(null);
    try {
      const from = format(fromDate, 'yyyy-MM-dd');
      const to = format(toDate, 'yyyy-MM-dd');
      
      const response = await fetch(
        `http://localhost:8080/api/ai/apt-timeline?host_id=${hostId}&from=${from}&to=${to}&group_by=${groupBy}`
      );
      
      if (!response.ok) throw new Error(await response.text());
      
      const data = await response.json();
      setTimelineData(data);
    } catch (err) {
      setError(err.message);
      console.error('Error fetching APT timeline:', err);
    } finally {
      setLoading(false);
    }
  };

  const getEventColor = (eventType) => {
    switch (eventType) {
      case 'malware': return 'error';
      case 'suspicious': return 'warning';
      case 'exploit': return 'error';
      case 'recon': return 'info';
      case 'lateral': return 'secondary';
      default: return 'primary';
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        APT Timeline
      </Typography>
      
      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2}>
          <Grid item xs={12} md={4}>
            <TextField
              fullWidth
              label="Host ID"
              value={hostId}
              onChange={(e) => setHostId(e.target.value)}
              margin="normal"
              placeholder="Введите идентификатор хоста"
            />
          </Grid>
          <Grid item xs={12} md={2}>
            <FormControl fullWidth margin="normal">
              <InputLabel>Группировать по</InputLabel>
              <Select
                value={groupBy}
                onChange={(e) => setGroupBy(e.target.value)}
                label="Группировать по"
              >
                <MenuItem value="hour">Часам</MenuItem>
                <MenuItem value="day">Дням</MenuItem>
                <MenuItem value="week">Неделям</MenuItem>
                <MenuItem value="month">Месяцам</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <LocalizationProvider dateAdapter={AdapterDateFns} locale={ru}>
              <DatePicker
                label="С даты"
                value={fromDate}
                onChange={setFromDate}
                renderInput={(params) => <TextField {...params} fullWidth margin="normal" />}
              />
            </LocalizationProvider>
          </Grid>
          <Grid item xs={12} md={3}>
            <LocalizationProvider dateAdapter={AdapterDateFns} locale={ru}>
              <DatePicker
                label="По дату"
                value={toDate}
                onChange={setToDate}
                renderInput={(params) => <TextField {...params} fullWidth margin="normal" />}
              />
            </LocalizationProvider>
          </Grid>
        </Grid>
        
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
          <Button 
            variant="contained" 
            onClick={fetchAPTTimeline}
            disabled={!hostId || loading}
          >
            Загрузить
          </Button>
        </Box>
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
        <Timeline position="alternate">
          {timelineData.map((event, index) => (
            <TimelineItem key={index}>
              <TimelineOppositeContent color="text.secondary">
                {format(new Date(event.timestamp), 'dd MMM yyyy HH:mm', { locale: ru })}
              </TimelineOppositeContent>
              <TimelineSeparator>
                <TimelineDot color={getEventColor(event.type)} />
                {index < timelineData.length - 1 && <TimelineConnector />}
              </TimelineSeparator>
              <TimelineContent>
                <Paper elevation={3} sx={{ p: 2 }}>
                  <Typography variant="h6" component="h1">
                    {event.title}
                  </Typography>
                  <Typography>Тип: {event.type}</Typography>
                  <Typography>Уровень угрозы: {event.severity}</Typography>
                  <Typography>{event.description}</Typography>
                  {event.indicators && (
                    <Box sx={{ mt: 1 }}>
                      <Typography variant="subtitle2">Индикаторы:</Typography>
                      <ul>
                        {event.indicators.map((indicator, i) => (
                          <li key={i}>{indicator}</li>
                        ))}
                      </ul>
                    </Box>
                  )}
                </Paper>
              </TimelineContent>
            </TimelineItem>
          ))}
        </Timeline>
      )}
    </Box>
  );
};