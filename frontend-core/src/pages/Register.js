// Register.js
import React, { useState } from 'react';
import { 
  Box, 
  Paper, 
  Typography, 
  TextField, 
  Button, 
  Alert, 
  CircularProgress, 
  Grid, 
  Link,
  Container
} from '@mui/material';
import { 
  Person, 
  Lock, 
  Email,
  ArrowBack
} from '@mui/icons-material';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export const Register = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
  });
  
  const [errors, setErrors] = useState({});
  const [loading, setLoading] = useState(false);
  const [serverError, setServerError] = useState(null);
  const [successMessage, setSuccessMessage] = useState(null);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Очищаем ошибку при изменении поля
    if (errors[name]) {
      setErrors(prev => ({
        ...prev,
        [name]: ''
      }));
    }
  };

  const validateForm = () => {
    const newErrors = {};
    
    if (!formData.username.trim()) {
      newErrors.username = 'Имя пользователя обязательно';
    } else if (formData.username.length < 3) {
      newErrors.username = 'Имя пользователя должно содержать минимум 3 символа';
    }
    
    if (!formData.email.trim()) {
      newErrors.email = 'Email обязателен';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Некорректный формат email';
    }
    
    if (!formData.password) {
      newErrors.password = 'Пароль обязателен';
    } else if (formData.password.length < 8) {
      newErrors.password = 'Пароль должен содержать минимум 8 символов';
    }
    
    if (!formData.confirmPassword) {
      newErrors.confirmPassword = 'Подтверждение пароля обязательно';
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Пароли не совпадают';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }
    
    setLoading(true);
    setServerError(null);
    setSuccessMessage(null);
    
    try {
      const response = await fetch('http://localhost:8080/api/auth/register', {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          username: formData.username,
          email: formData.email,
          password: formData.password
        })
      });
      
      const data = await response.json();
      
      if (!response.ok) {
        throw new Error(data.error || 'Ошибка при регистрации');
      }
      
      // Успешная регистрация
      setSuccessMessage('Регистрация прошла успешно! Теперь вы можете войти в систему.');
      
      // Очищаем форму
      setFormData({
        username: '',
        email: '',
        password: '',
        confirmPassword: ''
      });
      
      // Автоматический вход после успешной регистрации (опционально)
      // setTimeout(() => {
      //   login(formData.username, formData.password);
      //   navigate('/');
      // }, 2000);
      
    } catch (err) {
      setServerError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          minHeight: '100vh',
          py: 4
        }}
      >
        <Paper 
          elevation={10}
          sx={{ 
            p: 4, 
            width: '100%',
            borderRadius: 2
          }}
        >
          <Button
            startIcon={<ArrowBack />}
            component={RouterLink}
            to="/login"
            sx={{ mb: 2 }}
          >
            Назад к входу
          </Button>
          
          <Typography variant="h4" align="center" gutterBottom>
            Регистрация
          </Typography>
          
          <Typography variant="body2" color="text.secondary" align="center" sx={{ mb: 3 }}>
            Создайте учетную запись для доступа к SIEM Dashboard
          </Typography>
          
          {serverError && (
            <Alert severity="error" sx={{ mb: 3 }}>
              {serverError}
            </Alert>
          )}
          
          {successMessage && (
            <Alert severity="success" sx={{ mb: 3 }}>
              {successMessage}
            </Alert>
          )}
          
          <form onSubmit={handleSubmit}>
            <TextField
              fullWidth
              name="username"
              label="Имя пользователя"
              value={formData.username}
              onChange={handleChange}
              margin="normal"
              error={!!errors.username}
              helperText={errors.username}
              disabled={loading}
              InputProps={{
                startAdornment: <Person sx={{ mr: 1, color: 'action.active' }} />
              }}
              autoComplete="username"
            />
            
            <TextField
              fullWidth
              name="email"
              label="Email"
              type="email"
              value={formData.email}
              onChange={handleChange}
              margin="normal"
              error={!!errors.email}
              helperText={errors.email}
              disabled={loading}
              InputProps={{
                startAdornment: <Email sx={{ mr: 1, color: 'action.active' }} />
              }}
              autoComplete="email"
            />
            
            <TextField
              fullWidth
              name="password"
              label="Пароль"
              type="password"
              value={formData.password}
              onChange={handleChange}
              margin="normal"
              error={!!errors.password}
              helperText={errors.password || 'Минимум 8 символов'}
              disabled={loading}
              InputProps={{
                startAdornment: <Lock sx={{ mr: 1, color: 'action.active' }} />
              }}
              autoComplete="new-password"
            />
            
            <TextField
              fullWidth
              name="confirmPassword"
              label="Подтвердите пароль"
              type="password"
              value={formData.confirmPassword}
              onChange={handleChange}
              margin="normal"
              error={!!errors.confirmPassword}
              helperText={errors.confirmPassword}
              disabled={loading}
              InputProps={{
                startAdornment: <Lock sx={{ mr: 1, color: 'action.active' }} />
              }}
              autoComplete="new-password"
            />
            
            <Button
              type="submit"
              fullWidth
              variant="contained"
              disabled={loading}
              sx={{ 
                mt: 3, 
                mb: 2,
                height: 48 
              }}
            >
              {loading ? (
                <CircularProgress size={24} color="inherit" />
              ) : (
                'Зарегистрироваться'
              )}
            </Button>
          </form>
          
          <Grid container justifyContent="center" sx={{ mt: 2 }}>
            <Grid item>
              <Typography variant="body2" color="text.secondary">
                Уже есть аккаунт?{' '}
                <Link 
                  component={RouterLink} 
                  to="/login"
                  sx={{ textDecoration: 'none' }}
                >
                  Войти
                </Link>
              </Typography>
            </Grid>
          </Grid>
          
          <Typography 
            variant="caption" 
            color="text.secondary" 
            align="center" 
            sx={{ 
              display: 'block', 
              mt: 3,
              fontSize: '0.75rem'
            }}
          >
            Регистрируясь, вы соглашаетесь с условиями использования и политикой конфиденциальности
          </Typography>
        </Paper>
      </Box>
    </Container>
  );
};