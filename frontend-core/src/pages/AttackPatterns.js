import React, { useState, useEffect } from 'react';
import { 
  Box, Paper, Typography, Table, TableBody, TableCell, 
  TableContainer, TableHead, TableRow, CircularProgress, 
  Alert, TextField, Grid, Card, CardContent, Select, 
  MenuItem, InputLabel, FormControl, Pagination, Chip
} from '@mui/material';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

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

  const categories = [
    '', 'Reconnaissance', 'Resource Development', 'Initial Access', 
    'Execution', 'Persistence', 'Privilege Escalation', 'Defense Evasion',
    'Credential Access', 'Discovery', 'Lateral Movement', 'Collection',
    'Command and Control', 'Exfiltration', 'Impact'
  ];

  const severities = ['', 'Low', 'Medium', 'High', 'Critical'];

  const fetchAttackPatterns = async () => {
    setLoading(true);
    setError(null);
    try {
      let url = `http://localhost:8080/api/ai/patterns?page=${page}&per_page=${perPage}`;
      if (category) url += `&category=${category}`;
      if (severity) url += `&severity=${severity}`;
      
      const response = await fetch(url);
      if (!response.ok) throw new Error(await response.text());
      
      const data = await response.json();
      setPatterns(data.data);
      setTotal(data.meta.total);
      
      // Fetch statistics
      const statsResponse = await fetch('http://localhost:8080/api/ai/patterns/stats');
      if (statsResponse.ok) {
        const statsData = await statsResponse.json();
        setStats(statsData);
      }
    } catch (err) {
      setError(err.message);
      console.error('Error fetching attack patterns:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAttackPatterns();
  }, [category, severity, page, perPage]);

  const handlePageChange = (event, value) => {
    setPage(value);
  };

  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'Low': return 'success';
      case 'Medium': return 'warning';
      case 'High': return 'error';
      case 'Critical': return 'error';
      default: return 'default';
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Шаблоны атак
      </Typography>
      
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>По категориям</Typography>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={stats.byCategory}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="category" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="count" fill="#8884d8" name="Количество" />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>По уровню угрозы</Typography>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={stats.bySeverity}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="severity" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="count" fill="#82ca9d" name="Количество" />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2}>
          <Grid item xs={12} md={4}>
            <FormControl fullWidth>
              <InputLabel>Категория</InputLabel>
              <Select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                label="Категория"
              >
                {categories.map(cat => (
                  <MenuItem key={cat || 'all'} value={cat}>
                    {cat || 'Все категории'}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={4}>
            <FormControl fullWidth>
              <InputLabel>Уровень угрозы</InputLabel>
              <Select
                value={severity}
                onChange={(e) => setSeverity(e.target.value)}
                label="Уровень угрозы"
              >
                {severities.map(sev => (
                  <MenuItem key={sev || 'all'} value={sev}>
                    {sev || 'Все уровни'}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={4}>
            <FormControl fullWidth>
              <InputLabel>Элементов на странице</InputLabel>
              <Select
                value={perPage}
                onChange={(e) => setPerPage(e.target.value)}
                label="Элементов на странице"
              >
                <MenuItem value={10}>10</MenuItem>
                <MenuItem value={20}>20</MenuItem>
                <MenuItem value={50}>50</MenuItem>
                <MenuItem value={100}>100</MenuItem>
              </Select>
            </FormControl>
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
        <>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>ID</TableCell>
                  <TableCell>Название</TableCell>
                  <TableCell>Категория</TableCell>
                  <TableCell>Уровень угрозы</TableCell>
                  <TableCell>Описание</TableCell>
                  <TableCell>Методы защиты</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {patterns.length > 0 ? (
                  patterns.map((pattern) => (
                    <TableRow key={pattern.id}>
                      <TableCell>{pattern.id}</TableCell>
                      <TableCell sx={{ fontWeight: 'bold' }}>{pattern.name}</TableCell>
                      <TableCell>{pattern.category}</TableCell>
                      <TableCell>
                        <Chip 
                          label={pattern.severity} 
                          color={getSeverityColor(pattern.severity)} 
                          size="small" 
                        />
                      </TableCell>
                      <TableCell>{pattern.description}</TableCell>
                      <TableCell>{pattern.mitigations.join(', ')}</TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={6} align="center">
                      Нет данных
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
          
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 3 }}>
            <Pagination
              count={Math.ceil(total / perPage)}
              page={page}
              onChange={handlePageChange}
              color="primary"
              showFirstButton
              showLastButton
            />
          </Box>
        </>
      )}
    </Box>
  );
};