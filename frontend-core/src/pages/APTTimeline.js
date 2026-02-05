// APTTimeline.js
import React, { useState, useEffect } from 'react';
import { 
  Box, Paper, Typography, TextField, CircularProgress, 
  Alert, Button, FormControl, InputLabel, Select, MenuItem, Grid,
  Card, CardContent, Chip, Divider
} from '@mui/material';
import { 
  Timeline, TimelineItem, TimelineSeparator, TimelineDot, 
  TimelineConnector, TimelineContent, TimelineOppositeContent 
} from '@mui/lab';
import { format, subDays, startOfDay, endOfDay } from 'date-fns';
import { ru } from 'date-fns/locale';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import EventIcon from '@mui/icons-material/Event';
import WarningIcon from '@mui/icons-material/Warning';
import SecurityIcon from '@mui/icons-material/Security';
import TimelineIcon from '@mui/icons-material/Timeline';
import { useApi } from '../hooks/useApi';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';

export const APTTimeline = () => {
  const [timelineData, setTimelineData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [hostId, setHostId] = useState('');
  const [fromDate, setFromDate] = useState(startOfDay(subDays(new Date(), 30)));
  const [toDate, setToDate] = useState(endOfDay(new Date()));
  const [groupBy, setGroupBy] = useState('day');
  const [totalEvents, setTotalEvents] = useState(0);
  const [severityCounts, setSeverityCounts] = useState({});

  const api = useApi();
  const { isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const fetchAPTTimeline = async () => {
    if (!isAuthenticated()) {
      setError('Please login to access APT timeline');
      return;
    }

    if (!hostId.trim()) {
      setError('Please enter a host ID');
      return;
    }

    setLoading(true);
    setError(null);
    
    try {
      const params = new URLSearchParams({
        host_id: hostId.trim(),
        from: fromDate.toISOString(), // или format(fromDate, "yyyy-MM-dd'T'HH:mm:ss'Z'")
        to: toDate.toISOString(),     // или format(toDate, "yyyy-MM-dd'T'HH:mm:ss'Z'")
        group_by: groupBy,
      });

      const url = `/api/ai/apt-timeline?${params.toString()}`;
      console.log('Fetching APT timeline from:', url);
      
      const data = await api.get(url);
      
      // Обработка данных в разных форматах
      if (Array.isArray(data)) {
        setTimelineData(data);
        setTotalEvents(data.length);
        calculateSeverityStats(data);
      } else if (data.data && Array.isArray(data.data)) {
        setTimelineData(data.data);
        setTotalEvents(data.data.length);
        calculateSeverityStats(data.data);
      } else if (data.events && Array.isArray(data.events)) {
        setTimelineData(data.events);
        setTotalEvents(data.events.length);
        calculateSeverityStats(data.events);
      } else {
        setTimelineData([]);
        setTotalEvents(0);
        setSeverityCounts({});
      }
      
    } catch (err) {
      console.error('Error fetching APT timeline:', err);
      
      // Обработка ошибок авторизации
      if (err.message.includes('Authorization') || 
          err.message.includes('401') || 
          err.message.includes('authenticated') ||
          err.message.includes('Session expired')) {
        setError('Session expired. Please login again.');
        setTimeout(() => {
          logout();
          navigate('/login');
        }, 2000);
      } else if (err.message.includes('400') || err.message.includes('Bad Request')) {
        setError('Invalid request. Please check your parameters.');
      } else if (err.message.includes('404')) {
        setError('No data found for the specified host ID and time range.');
        setTimelineData([]); // Сбрасываем данные если ничего не найдено
      } else {
        setError(`Failed to load timeline: ${err.message}`);
      }
    } finally {
      setLoading(false);
    }
  };

  const calculateSeverityStats = (events) => {
    const counts = {
      critical: 0,
      high: 0,
      medium: 0,
      low: 0,
      info: 0
    };

    events.forEach(event => {
      const severity = event.severity?.toLowerCase() || 'info';
      if (counts[severity] !== undefined) {
        counts[severity]++;
      } else {
        counts.info++;
      }
    });

    setSeverityCounts(counts);
  };

  useEffect(() => {
    // Автоматически загружаем данные если есть hostId и пользователь авторизован
    if (hostId.trim() && isAuthenticated()) {
      fetchAPTTimeline();
    }
  }, [hostId, isAuthenticated]);

  const getEventColor = (eventType) => {
    switch (eventType?.toLowerCase()) {
      case 'malware':
      case 'exploit':
      case 'ransomware': return 'error';
      case 'suspicious':
      case 'anomaly': return 'warning';
      case 'reconnaissance':
      case 'scanning': return 'info';
      case 'lateral_movement':
      case 'persistence': return 'secondary';
      case 'initial_access':
      case 'access': return 'primary';
      default: return 'default';
    }
  };

  const getEventIcon = (eventType) => {
    switch (eventType?.toLowerCase()) {
      case 'malware':
      case 'exploit': return <WarningIcon />;
      case 'suspicious': return <SecurityIcon />;
      case 'reconnaissance': return <TimelineIcon />;
      default: return <EventIcon />;
    }
  };

  const getSeverityChip = (severity) => {
    const severityLower = severity?.toLowerCase();
    
    switch (severityLower) {
      case 'critical':
        return <Chip label="CRITICAL" color="error" size="small" />;
      case 'high':
        return <Chip label="HIGH" color="error" size="small" variant="outlined" />;
      case 'medium':
        return <Chip label="MEDIUM" color="warning" size="small" />;
      case 'low':
        return <Chip label="LOW" color="success" size="small" />;
      default:
        return <Chip label="INFO" color="default" size="small" />;
    }
  };

  const handleRetry = () => {
    fetchAPTTimeline();
  };

  const handleClear = () => {
    setHostId('');
    setTimelineData([]);
    setTotalEvents(0);
    setSeverityCounts({});
    setError(null);
  };

  const DemoEvents = [
    {
      id: 1,
      timestamp: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000),
      title: 'Suspicious PowerShell Execution',
      type: 'suspicious',
      severity: 'high',
      description: 'Unusual PowerShell command execution detected with obfuscation techniques.',
      indicators: ['powershell.exe -enc', 'Invoke-Expression', 'Base64 encoded payload'],
      source_ip: '192.168.1.100',
      destination_ip: '10.0.0.5'
    },
    {
      id: 2,
      timestamp: new Date(Date.now() - 6 * 24 * 60 * 60 * 1000),
      title: 'Lateral Movement via SMB',
      type: 'lateral_movement',
      severity: 'medium',
      description: 'Attempt to move laterally using SMB protocol to multiple hosts.',
      indicators: ['SMB connection to multiple hosts', 'Pass-the-hash attempt'],
      source_ip: '10.0.0.5',
      destination_ip: '192.168.1.101'
    },
    {
      id: 3,
      timestamp: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
      title: 'Malware Download',
      type: 'malware',
      severity: 'critical',
      description: 'Download of known malware from external IP address.',
      indicators: ['Known bad hash: abc123...', 'C2 communication'],
      source_ip: '185.100.85.101',
      destination_ip: '192.168.1.100'
    },
    {
      id: 4,
      timestamp: new Date(Date.now() - 4 * 24 * 60 * 60 * 1000),
      title: 'Reconnaissance Scanning',
      type: 'reconnaissance',
      severity: 'low',
      description: 'Port scanning activity detected from internal host.',
      indicators: ['Multiple port connection attempts', 'SYN flood pattern'],
      source_ip: '192.168.1.100',
      destination_ip: '10.0.0.1'
    },
    {
      id: 5,
      timestamp: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000),
      title: 'Data Exfiltration Attempt',
      type: 'exfiltration',
      severity: 'high',
      description: 'Large volume of data transfer to external server.',
      indicators: ['Unusual outbound traffic', 'Compressed data transfer'],
      source_ip: '192.168.1.100',
      destination_ip: '45.33.32.156'
    }
  ];

  // Если не авторизован
  if (!isAuthenticated()) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert 
          severity="warning" 
          action={
            <Button color="inherit" size="small" onClick={() => navigate('/login')}>
              Login
            </Button>
          }
        >
          Please login to access APT timeline.
        </Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        APT Attack Timeline
      </Typography>
      
      {/* Панель управления */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          Timeline Configuration
        </Typography>
        
        <Grid container spacing={2}>
          <Grid item xs={12} md={4}>
            <TextField
              fullWidth
              label="Host ID / IP Address"
              value={hostId}
              onChange={(e) => setHostId(e.target.value)}
              placeholder="e.g., 192.168.1.100 or HOST-001"
              helperText="Enter host identifier or IP address"
              variant="outlined"
              size="small"
            />
          </Grid>
          
          <Grid item xs={12} md={3}>
            <LocalizationProvider dateAdapter={AdapterDateFns} locale={ru}>
              <DatePicker
                label="Start Date"
                value={fromDate}
                onChange={(newValue) => setFromDate(startOfDay(newValue))}
                maxDate={toDate}
                renderInput={(params) => (
                  <TextField {...params} fullWidth size="small" />
                )}
              />
            </LocalizationProvider>
          </Grid>
          
          <Grid item xs={12} md={3}>
            <LocalizationProvider dateAdapter={AdapterDateFns} locale={ru}>
              <DatePicker
                label="End Date"
                value={toDate}
                onChange={(newValue) => setToDate(endOfDay(newValue))}
                minDate={fromDate}
                maxDate={new Date()}
                renderInput={(params) => (
                  <TextField {...params} fullWidth size="small" />
                )}
              />
            </LocalizationProvider>
          </Grid>
          
          <Grid item xs={12} md={2}>
            <FormControl fullWidth size="small">
              <InputLabel>Group By</InputLabel>
              <Select
                value={groupBy}
                onChange={(e) => setGroupBy(e.target.value)}
                label="Group By"
              >
                <MenuItem value="hour">Hour</MenuItem>
                <MenuItem value="day">Day</MenuItem>
                <MenuItem value="week">Week</MenuItem>
                <MenuItem value="month">Month</MenuItem>
              </Select>
            </FormControl>
          </Grid>
        </Grid>
        
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1, mt: 2 }}>
          <Button 
            variant="outlined" 
            onClick={handleClear}
            disabled={loading}
          >
            Clear
          </Button>
          <Button 
            variant="contained" 
            onClick={fetchAPTTimeline}
            disabled={!hostId.trim() || loading}
            startIcon={loading ? <CircularProgress size={20} /> : null}
          >
            {loading ? 'Loading...' : 'Load Timeline'}
          </Button>
        </Box>
      </Paper>

      {/* Статистика */}
      {totalEvents > 0 && (
        <Paper sx={{ p: 2, mb: 3 }}>
          <Grid container spacing={2}>
            <Grid item xs={12} md={3}>
              <Card variant="outlined">
                <CardContent sx={{ textAlign: 'center' }}>
                  <Typography variant="h4" color="primary">
                    {totalEvents}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Total Events
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
            
            {Object.entries(severityCounts).map(([severity, count]) => (
              count > 0 && (
                <Grid item xs={12} md={2} key={severity}>
                  <Card variant="outlined">
                    <CardContent sx={{ textAlign: 'center' }}>
                      <Typography variant="h4">
                        {count}
                      </Typography>
                      <Typography variant="body2" color="text.secondary" sx={{ textTransform: 'capitalize' }}>
                        {severity}
                      </Typography>
                    </CardContent>
                  </Card>
                </Grid>
              )
            ))}
          </Grid>
        </Paper>
      )}

      {/* Сообщения об ошибках */}
      {error && (
        <Alert 
          severity="error" 
          sx={{ mb: 2 }}
          action={
            <Button color="inherit" size="small" onClick={handleRetry}>
              Retry
            </Button>
          }
        >
          {error}
        </Alert>
      )}

      {/* Временная шкала */}
      {loading ? (
        <Box display="flex" flexDirection="column" alignItems="center" p={4}>
          <CircularProgress />
          <Typography sx={{ mt: 2 }}>Loading attack timeline...</Typography>
        </Box>
      ) : timelineData.length > 0 ? (
        <Timeline position="alternate">
          {timelineData.map((event, index) => (
            <TimelineItem key={event.id || index}>
              <TimelineOppositeContent 
                color="text.secondary" 
                sx={{ py: 2, display: 'flex', alignItems: 'center' }}
              >
                <Box>
                  <Typography variant="body2">
                    {format(new Date(event.timestamp), 'dd MMM yyyy', { locale: ru })}
                  </Typography>
                  <Typography variant="caption">
                    {format(new Date(event.timestamp), 'HH:mm:ss', { locale: ru })}
                  </Typography>
                </Box>
              </TimelineOppositeContent>
              
              <TimelineSeparator>
                <TimelineDot color={getEventColor(event.type)}>
                  {getEventIcon(event.type)}
                </TimelineDot>
                {index < timelineData.length - 1 && <TimelineConnector />}
              </TimelineSeparator>
              
              <TimelineContent sx={{ py: 2 }}>
                <Card elevation={2}>
                  <CardContent>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1 }}>
                      <Typography variant="h6" component="h2">
                        {event.title}
                      </Typography>
                      <Box sx={{ display: 'flex', gap: 1 }}>
                        {getSeverityChip(event.severity)}
                        <Chip 
                          label={event.type?.replace('_', ' ') || 'Event'} 
                          size="small" 
                          variant="outlined"
                        />
                      </Box>
                    </Box>
                    
                    <Typography variant="body2" paragraph>
                      {event.description}
                    </Typography>
                    
                    {event.indicators && event.indicators.length > 0 && (
                      <>
                        <Divider sx={{ my: 1 }} />
                        <Typography variant="subtitle2" gutterBottom>
                          Indicators of Compromise (IoC):
                        </Typography>
                        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mb: 1 }}>
                          {event.indicators.map((indicator, i) => (
                            <Chip 
                              key={i}
                              label={indicator}
                              size="small"
                              variant="outlined"
                              color="primary"
                            />
                          ))}
                        </Box>
                      </>
                    )}
                    
                    {(event.source_ip || event.destination_ip) && (
                      <Box sx={{ display: 'flex', gap: 2, mt: 1 }}>
                        {event.source_ip && (
                          <Typography variant="caption">
                            <strong>Source:</strong> {event.source_ip}
                          </Typography>
                        )}
                        {event.destination_ip && (
                          <Typography variant="caption">
                            <strong>Destination:</strong> {event.destination_ip}
                          </Typography>
                        )}
                      </Box>
                    )}
                  </CardContent>
                </Card>
              </TimelineContent>
            </TimelineItem>
          ))}
        </Timeline>
      ) : hostId.trim() ? (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No attack events found
          </Typography>
          <Typography variant="body2" color="text.secondary" paragraph>
            No Advanced Persistent Threat events were detected for host "{hostId}" 
            in the specified time range ({format(fromDate, 'dd MMM yyyy')} - {format(toDate, 'dd MMM yyyy')}).
          </Typography>
          <Button 
            variant="outlined" 
            onClick={handleRetry}
          >
            Try Again
          </Button>
        </Paper>
      ) : (
        <Paper sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="h6" color="text.secondary" gutterBottom>
            Enter a host ID to view APT timeline
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Enter a host identifier or IP address in the field above to load the attack timeline.
          </Typography>
        </Paper>
      )}
    </Box>
  );
};

export default APTTimeline;