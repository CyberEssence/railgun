import React, { useState } from 'react';
import { 
  Box, Paper, Typography, TextField, Button, 
  Alert, CircularProgress, Grid, Link 
} from '@mui/material';
import { Lock, Person, VerifiedUser } from '@mui/icons-material';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';

export const Login = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [twoFARequired, setTwoFARequired] = useState(false);
  const [twoFAToken, setTwoFAToken] = useState('');
  const [userId, setUserId] = useState(null);
  const { login, verify2FA } = useAuth();
  const navigate = useNavigate();

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await login(username, password);
      
      if (result.requires2FA) {
        setTwoFARequired(true);
        setUserId(result.userId);
        setTwoFAToken(result.token);
      } else {
        navigate('/');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleVerify2FA = async () => {
    setLoading(true);
    setError(null);
    
    try {
      await verify2FA(userId, twoFAToken);
      navigate('/');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #121212 0%, #1e1e1e 100%)',
        p: 3
      }}
    >
      <Paper elevation={10} sx={{ p: 4, width: '100%', maxWidth: 500 }}>
        <Typography variant="h4" align="center" gutterBottom>
          SIEM Dashboard
        </Typography>
        
        {!twoFARequired ? (
          <>
            <Typography variant="h6" align="center" gutterBottom>
              Вход в систему
            </Typography>
            
            {error && (
              <Alert severity="error" sx={{ mb: 3 }}>
                {error}
              </Alert>
            )}
            
            <TextField
              fullWidth
              label="Имя пользователя"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              margin="normal"
              InputProps={{
                startAdornment: <Person sx={{ mr: 1 }} />
              }}
            />
            
            <TextField
              fullWidth
              label="Пароль"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              margin="normal"
              InputProps={{
                startAdornment: <Lock sx={{ mr: 1 }} />
              }}
            />
            
            <Button
              fullWidth
              variant="contained"
              onClick={handleLogin}
              disabled={loading || !username || !password}
              sx={{ mt: 2, height: 48 }}
            >
              {loading ? (
                <CircularProgress size={24} color="inherit" />
              ) : (
                'Войти'
              )}
            </Button>
            
            <Grid container sx={{ mt: 2 }}>
              <Grid item xs>
                <Link href="#" variant="body2">
                  Забыли пароль?
                </Link>
              </Grid>
              <Grid item>
                <Link href="/register" variant="body2">
                  Регистрация
                </Link>
              </Grid>
            </Grid>
          </>
        ) : (
          <>
            <Typography variant="h6" align="center" gutterBottom>
              Двухфакторная аутентификация
            </Typography>
            
            <Typography variant="body1" align="center" sx={{ mb: 3 }}>
              Введите код из приложения аутентификатора
            </Typography>
            
            {error && (
              <Alert severity="error" sx={{ mb: 3 }}>
                {error}
              </Alert>
            )}
            
            <TextField
              fullWidth
              label="Код подтверждения"
              value={twoFAToken}
              onChange={(e) => setTwoFAToken(e.target.value)}
              margin="normal"
              InputProps={{
                startAdornment: <VerifiedUser sx={{ mr: 1 }} />
              }}
            />
            
            <Button
              fullWidth
              variant="contained"
              onClick={handleVerify2FA}
              disabled={loading || !twoFAToken}
              sx={{ mt: 2, height: 48 }}
            >
              {loading ? (
                <CircularProgress size={24} color="inherit" />
              ) : (
                'Подтвердить'
              )}
            </Button>
            
            <Button
              fullWidth
              variant="outlined"
              onClick={() => setTwoFARequired(false)}
              sx={{ mt: 2 }}
            >
              Назад
            </Button>
          </>
        )}
      </Paper>
    </Box>
  );
};