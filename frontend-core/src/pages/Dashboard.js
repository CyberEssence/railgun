import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Paper, 
  Typography, 
  Grid, 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle 
} from '@mui/material';
import { Home, Storage } from '@mui/icons-material';
import axios from 'axios';

export const Dashboard = () => {
  const [stats, setStats] = useState({
    totalEvents: 0,
    activeConnections: 0,
    suspiciousActivity: 0,
    systemHealth: 'healthy'
  });

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        const response = await axios.get('/api/dashboard/stats');
        setStats(response.data);
      } catch (error) {
        console.error('Ошибка загрузки статистики:', error);
      }
    };

    fetchDashboardData();
    // Обновляем данные каждую минуту
    const interval = setInterval(fetchDashboardData, 60000);
    
    return () => clearInterval(interval);
  }, []);

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Обзор системы SIEM
      </Typography>
      
      {/* Панель статистики */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardHeader>
                title={
                    <Typography variant="h6">
                        Всего событий
                    </Typography>
                }
            </CardHeader>
            <CardContent>
              <Typography variant="h6">{stats.totalEvents.toLocaleString()}</Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardHeader>
                title={
                    <Typography variant="h6">
                        Активных подключений
                    </Typography>
                }
            </CardHeader>
            <CardContent>
              <Typography variant="h6">{stats.activeConnections}</Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardHeader>
                title={
                    <Typography variant="h6">
                        Подозрительной активности
                    </Typography>
                }
            </CardHeader>
            <CardContent>
              <Typography variant="h6">{stats.suspiciousActivity}</Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardHeader>
                title={
                    <Typography variant="h6">
                        Статус системы
                    </Typography>
                }
            </CardHeader>
            <CardContent>
              <Typography variant="h6">
                {stats.systemHealth === 'healthy' ? '✅ Здоровая' : '⚠️ Проблемы'}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Карточки быстрого доступа */}
      <Grid container spacing={2}>
        <Grid item xs={12} sm={4}>
          <Card>
            <CardContent sx={{ pt: 2 }}>
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Home sx={{ fontSize: 48, mb: 2 }} />
                <Typography variant="subtitle1">Обзор</Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={4}>
          <Card>
            <CardContent sx={{ pt: 2 }}>
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Typography variant="subtitle1">Трафик</Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={4}>
          <Card>
            <CardContent sx={{ pt: 2 }}>
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Storage sx={{ fontSize: 48, mb: 2 }} />
                <Typography variant="subtitle1">Артефакты</Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};