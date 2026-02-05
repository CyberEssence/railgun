// components/Test2FA.js
import React, { useState } from 'react';
import { 
  Box, 
  Button, 
  TextField, 
  Paper, 
  Typography,
  Alert,
  CircularProgress 
} from '@mui/material';
import { useAuth } from '../context/AuthContext';

export const Test2FA = () => {
  const { verify2FA, testAuthAPI } = useAuth();
  const [token, setToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [testResult, setTestResult] = useState(null);

  const handleTest2FA = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      // Тестовый userId (должен совпадать с реальным ID из бэкенда)
      const userId = 1; // Замените на реальный ID пользователя
      const result = await verify2FA(userId, token);
      
      if (result) {
        setSuccess('2FA verified successfully!');
        // Тестируем API после успешной аутентификации
        const apiTest = await testAuthAPI();
        setTestResult(apiTest ? 'API test passed' : 'API test failed');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Paper sx={{ p: 3, maxWidth: 500, mx: 'auto' }}>
        <Typography variant="h5" gutterBottom>
          2FA Test Page
        </Typography>
        
        <Typography variant="body1" paragraph>
          Use this page to test 2FA verification directly.
          First login normally, then copy the 2FA token and test it here.
        </Typography>

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

        <TextField
          fullWidth
          label="2FA Token"
          value={token}
          onChange={(e) => setToken(e.target.value)}
          margin="normal"
          helperText="Enter the 2FA token from the login response"
        />

        <Button
          fullWidth
          variant="contained"
          onClick={handleTest2FA}
          disabled={loading || !token}
          sx={{ mt: 2, height: 48 }}
        >
          {loading ? <CircularProgress size={24} /> : 'Test 2FA Verification'}
        </Button>

        {testResult && (
          <Alert severity="info" sx={{ mt: 2 }}>
            {testResult}
          </Alert>
        )}
      </Paper>
    </Box>
  );
};