import React, { useState } from 'react';
import { 
  Box, 
  Paper, 
  Typography, 
  TextField, 
  Button, 
  CircularProgress,
  Alert,
  Grid,
  Link
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export const LoginPage = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [twoFARequired, setTwoFARequired] = useState(false);
  const [userId, setUserId] = useState(null);
  const [token, setToken] = useState('');
  const navigate = useNavigate();
  const { login, verify2FA } = useAuth();

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await login(username, password);
      
      if (result.requiresTwoFA) {
        setTwoFARequired(true);
        setUserId(result.userId);
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
      const success = await verify2FA(userId, token);
      if (success) {
        navigate('/');
      }
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
      <Paper 
        sx={{ 
          p: 4, 
          width: '100%', 
          maxWidth: 500,
          boxShadow: '0px 10px 25px rgba(0, 0, 0, 0.5)'
        }}
      >
        <Typography variant="h4" gutterBottom align="center">
          SIEM Dashboard Login
        </Typography>
        
        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}

        {!twoFARequired ? (
          <>
            <TextField
              fullWidth
              label="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              margin="normal"
              disabled={loading}
            />
            <TextField
              fullWidth
              label="Password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              margin="normal"
              disabled={loading}
            />
            <Button
              fullWidth
              variant="contained"
              onClick={handleLogin}
              disabled={loading || !username || !password}
              sx={{ mt: 2, height: 48 }}
            >
              {loading ? <CircularProgress size={24} /> : 'Login'}
            </Button>
          </>
        ) : (
          <>
            <Typography variant="body1" align="center" sx={{ mb: 2 }}>
              Please enter the 2FA token sent to your device
            </Typography>
            <TextField
              fullWidth
              label="2FA Token"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              margin="normal"
              disabled={loading}
            />
            <Button
              fullWidth
              variant="contained"
              onClick={handleVerify2FA}
              disabled={loading || !token}
              sx={{ mt: 2, height: 48 }}
            >
              {loading ? <CircularProgress size={24} /> : 'Verify Token'}
            </Button>
          </>
        )}

        <Grid container justifyContent="space-between" sx={{ mt: 2 }}>
          <Grid item>
            <Link href="#" variant="body2">
              Forgot password?
            </Link>
          </Grid>
          <Grid item>
            <Link href="#" variant="body2">
              Create account
            </Link>
          </Grid>
        </Grid>
      </Paper>
    </Box>
  );
};