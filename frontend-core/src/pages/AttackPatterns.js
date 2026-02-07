import React, { useState, useEffect } from 'react';
import { 
  Box, Paper, Typography, Table, TableBody, TableCell, 
  TableContainer, TableHead, TableRow, CircularProgress, 
  Alert, TextField, Grid, Card, CardContent, Select, 
  MenuItem, InputLabel, FormControl, Pagination, Chip,
  Button
} from '@mui/material';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { useApi } from '../hooks/useApi';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';

export const AttackPatterns = () => {
  const [patterns, setPatterns] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [category, setCategory] = useState('');
  const [severity, setSeverity] = useState('');
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(20);
  const [total, setTotal] = useState(0);
  const [stats, setStats] = useState({ byCategory: [], bySeverity: [] });

  const api = useApi();
  const { isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const categories = [
    '', 'Reconnaissance', 'Resource Development', 'Initial Access', 
    'Execution', 'Persistence', 'Privilege Escalation', 'Defense Evasion',
    'Credential Access', 'Discovery', 'Lateral Movement', 'Collection',
    'Command and Control', 'Exfiltration', 'Impact'
  ];

  const severities = ['', 'Low', 'Medium', 'High', 'Critical'];

  const fetchAttackPatterns = async () => {
    if (!isAuthenticated()) {
      setError('Please login to access attack patterns');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      let url = `/api/ai/patterns?page=${page}&per_page=${perPage}`;
      if (category) url += `&category=${encodeURIComponent(category)}`;
      if (severity) url += `&severity=${encodeURIComponent(severity)}`;
      
      console.log('Fetching attack patterns from:', url);
      const data = await api.get(url);
      
      // Обработка ответа в разных форматах
      if (data.data && data.meta) {
        setPatterns(data.data);
        setTotal(data.meta.total);
      } else if (Array.isArray(data)) {
        setPatterns(data);
        setTotal(data.length);
      } else {
        setPatterns([]);
        setTotal(0);
      }
      
      // Получение статистики
      try {
        const statsData = await api.get('/api/ai/patterns/stats');
        if (statsData) {
          setStats({
            byCategory: statsData.byCategory || [],
            bySeverity: statsData.bySeverity || []
          });
        }
      } catch (statsError) {
        console.warn('Could not fetch stats:', statsError);
        // Используем демо-статистику если API не доступен
        setStats({
          byCategory: [
            { category: 'Initial Access', count: 15 },
            { category: 'Execution', count: 22 },
            { category: 'Persistence', count: 18 },
            { category: 'Privilege Escalation', count: 12 },
            { category: 'Defense Evasion', count: 25 },
          ],
          bySeverity: [
            { severity: 'Low', count: 10 },
            { severity: 'Medium', count: 35 },
            { severity: 'High', count: 28 },
            { severity: 'Critical', count: 5 },
          ]
        });
      }
    } catch (err) {
      console.error('Error fetching attack patterns:', err);
      
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
      } else {
        setError(`Failed to load data: ${err.message}`);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAttackPatterns();
  }, [category, severity, page, perPage, isAuthenticated]);

  const handlePageChange = (event, value) => {
    setPage(value);
  };

  const handlePerPageChange = (event) => {
    setPerPage(event.target.value);
    setPage(1); // Сбрасываем на первую страницу при изменении количества элементов
  };

  const handleRetry = () => {
    fetchAttackPatterns();
  };

  const getSeverityColor = (severity) => {
    switch (severity?.toLowerCase()) {
      case 'low': return 'success';
      case 'medium': return 'warning';
      case 'high': return 'error';
      case 'critical': return 'error';
      default: return 'default';
    }
  };

  // Если не авторизован, показываем сообщение
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
          Please login to access attack patterns.
        </Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Attack Patterns
      </Typography>
      
      {/* Графики статистики */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>By Category</Typography>
              {stats.byCategory.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={stats.byCategory}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="category" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey="count" fill="#8884d8" name="Count" />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <Box sx={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Typography color="text.secondary">No statistics available</Typography>
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>By Severity Level</Typography>
              {stats.bySeverity.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={stats.bySeverity}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="severity" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey="count" fill="#82ca9d" name="Count" />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <Box sx={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Typography color="text.secondary">No statistics available</Typography>
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Фильтры и пагинация */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2} alignItems="center">
          <Grid item xs={12} md={3}>
            <FormControl fullWidth size="small">
              <InputLabel>Category</InputLabel>
              <Select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                label="Category"
              >
                {categories.map(cat => (
                  <MenuItem key={cat || 'all'} value={cat}>
                    {cat || 'All Categories'}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth size="small">
              <InputLabel>Severity</InputLabel>
              <Select
                value={severity}
                onChange={(e) => setSeverity(e.target.value)}
                label="Severity"
              >
                {severities.map(sev => (
                  <MenuItem key={sev || 'all'} value={sev}>
                    {sev || 'All Severities'}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth size="small">
              <InputLabel>Items per page</InputLabel>
              <Select
                value={perPage}
                onChange={handlePerPageChange}
                label="Items per page"
              >
                <MenuItem value={10}>10</MenuItem>
                <MenuItem value={20}>20</MenuItem>
                <MenuItem value={50}>50</MenuItem>
                <MenuItem value={100}>100</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <Button 
              variant="contained" 
              onClick={fetchAttackPatterns}
              disabled={loading}
              fullWidth
            >
              {loading ? 'Loading...' : 'Refresh'}
            </Button>
          </Grid>
        </Grid>
      </Paper>

      {/* Сообщения об ошибках и загрузке */}
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

      {loading ? (
        <Box display="flex" flexDirection="column" alignItems="center" p={4}>
          <CircularProgress />
          <Typography sx={{ mt: 2 }}>Loading attack patterns...</Typography>
        </Box>
      ) : (
        <>
          {/* Таблица с шаблонами атак */}
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>ID</TableCell>
                  <TableCell>Name</TableCell>
                  <TableCell>Category</TableCell>
                  <TableCell>Severity</TableCell>
                  <TableCell>Description</TableCell>
                  <TableCell>Mitigations</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {patterns.length > 0 ? (
                  patterns.map((pattern) => (
                    <TableRow key={pattern.id} hover>
                      <TableCell>{pattern.id}</TableCell>
                      <TableCell sx={{ fontWeight: 'bold' }}>{pattern.name}</TableCell>
                      <TableCell>
                        <Chip 
                          label={pattern.category} 
                          variant="outlined"
                          size="small"
                        />
                      </TableCell>
                      <TableCell>
                        <Chip 
                          label={pattern.severity} 
                          color={getSeverityColor(pattern.severity)} 
                          size="small" 
                        />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" sx={{ maxWidth: 300 }}>
                          {pattern.description || 'No description available'}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        {pattern.mitigations && pattern.mitigations.length > 0 ? (
                          <Box sx={{ maxWidth: 300 }}>
                            {pattern.mitigations.slice(0, 2).map((mitigation, idx) => (
                              <Chip 
                                key={idx}
                                label={mitigation}
                                size="small"
                                variant="outlined"
                                color="primary"
                                sx={{ mr: 0.5, mb: 0.5 }}
                              />
                            ))}
                            {pattern.mitigations.length > 2 && (
                              <Chip 
                                label={`+${pattern.mitigations.length - 2} more`}
                                size="small"
                                variant="outlined"
                              />
                            )}
                          </Box>
                        ) : (
                          <Typography variant="body2" color="text.secondary">
                            No mitigations available
                          </Typography>
                        )}
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={6} align="center" sx={{ py: 4 }}>
                      <Box sx={{ textAlign: 'center' }}>
                        <Typography color="text.secondary" gutterBottom>
                          No attack patterns found
                        </Typography>
                        <Button 
                          variant="outlined" 
                          onClick={fetchAttackPatterns}
                          sx={{ mt: 1 }}
                        >
                          Refresh
                        </Button>
                      </Box>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
          
          {/* Пагинация */}
          {total > 0 && (
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 3 }}>
              <Typography variant="body2" color="text.secondary">
                Showing {((page - 1) * perPage) + 1} to {Math.min(page * perPage, total)} of {total} patterns
              </Typography>
              <Pagination
                count={Math.ceil(total / perPage)}
                page={page}
                onChange={handlePageChange}
                color="primary"
                showFirstButton
                showLastButton
              />
            </Box>
          )}
        </>
      )}
    </Box>
  );
};

export default AttackPatterns;