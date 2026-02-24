import React from 'react';
import {
  Box,
  Paper,
  Typography,
  Avatar,
  Grid,
  Button,
  Divider,
  Chip,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  Card,
  CardContent,
  Alert,
} from '@mui/material';
import {
  Person,
  Email,
  Security,
  QrCode2,
  ContentCopy,
  VpnKey,
  ArrowForward,
} from '@mui/icons-material';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';

export const Profile = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    // Можно добавить уведомление о успешном копировании
  };

  if (!user) {
    return (
      <Box display="flex" justifyContent="center" p={3}>
        <Typography>Загрузка...</Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ maxWidth: 1000, mx: 'auto', p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Профиль пользователя
      </Typography>

      <Grid container spacing={3}>
        {/* Левая колонка - аватар и основная информация */}
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3, textAlign: 'center' }}>
            <Avatar
              sx={{
                width: 120,
                height: 120,
                mx: 'auto',
                mb: 2,
                bgcolor: 'primary.main',
                fontSize: '3rem',
              }}
            >
              {user.username?.charAt(0).toUpperCase()}
            </Avatar>
            
            <Typography variant="h5" gutterBottom>
              {user.username}
            </Typography>
            
            <Typography variant="body2" color="text.secondary" gutterBottom>
              {user.email}
            </Typography>
            
            <Chip
              label={user.role || 'Пользователь'}
              color="primary"
              size="small"
              sx={{ mt: 1, mb: 2 }}
            />
            
            <Divider sx={{ my: 2 }} />
            
            <Button
              variant="outlined"
              color="error"
              onClick={logout}
              fullWidth
            >
              Выйти из аккаунта
            </Button>
          </Paper>
        </Grid>

        {/* Правая колонка - детальная информация */}
        <Grid item xs={12} md={8}>
          {/* Информация об аккаунте */}
          <Paper sx={{ p: 3, mb: 3 }}>
            <Typography variant="h6" gutterBottom>
              Информация об аккаунте
            </Typography>
            
            <List>
              <ListItem>
                <ListItemText
                  primary="Имя пользователя"
                  secondary={user.username}
                />
              </ListItem>
              <Divider />
              
              <ListItem>
                <ListItemText
                  primary="Email"
                  secondary={user.email}
                />
              </ListItem>
              <Divider />
              
              <ListItem>
                <ListItemText
                  primary="ID пользователя"
                  secondary={user.id}
                />
                <ListItemSecondaryAction>
                  <IconButton
                    edge="end"
                    onClick={() => copyToClipboard(user.id)}
                    size="small"
                  >
                    <ContentCopy />
                  </IconButton>
                </ListItemSecondaryAction>
              </ListItem>
              <Divider />
              
              <ListItem>
                <ListItemText
                  primary="Роль"
                  secondary={user.role || 'Пользователь'}
                />
              </ListItem>
            </List>
          </Paper>

          {/* Безопасность */}
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Безопасность
            </Typography>
            
            <List>
              {/* Смена пароля */}
              <ListItem>
                <ListItemText
                  primary="Пароль"
                  secondary="Последнее изменение: недавно"
                />
                <ListItemSecondaryAction>
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<VpnKey />}
                  >
                    Сменить
                  </Button>
                </ListItemSecondaryAction>
              </ListItem>
              <Divider />
              
              {/* Двухфакторная аутентификация - ссылка на существующую страницу */}
              <ListItem>
                <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                  <Security sx={{ mr: 2, color: 'primary.main' }} />
                  <ListItemText
                    primary="Двухфакторная аутентификация (2FA)"
                    secondary="Настройте двухфакторную аутентификацию для дополнительной защиты аккаунта"
                  />
                  <Button
                    variant="contained"
                    color="primary"
                    endIcon={<ArrowForward />}
                    onClick={() => navigate('/settings/2fa')}
                    sx={{ ml: 2 }}
                  >
                    Настроить 2FA
                  </Button>
                </Box>
              </ListItem>
            </List>

            {/* Информационный блок о 2FA */}
            <Alert severity="info" sx={{ mt: 2 }}>
              <Typography variant="body2">
                <strong>Двухфакторная аутентификация</strong> добавляет дополнительный уровень защиты. 
                После включения, при входе в аккаунт потребуется не только пароль, но и код из приложения-аутентификатора.
              </Typography>
            </Alert>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};