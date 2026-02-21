import React, { useState, useEffect } from 'react';
import {
  Box, Paper, Typography, Button, Alert, CircularProgress,
  Dialog, DialogTitle, DialogContent, DialogActions,
  TextField, List, ListItem, ListItemText, ListItemSecondaryAction,
  IconButton, Divider, Switch, Grid, Stepper, Step, StepLabel,
  InputAdornment
} from '@mui/material';
import {
  QrCode2, ContentCopy, Download,
  Security, ArrowBack, ArrowForward
} from '@mui/icons-material';
import { useAuth } from '../context/AuthContext';

export const TwoFASettings = () => {
  const { 
    get2FAStatus, 
    enable2FA, 
    verify2FASetup, 
    disable2FA, 
    generateBackupCodes, 
    user 
  } = useAuth();
  
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  
  // Состояние для включения 2FA
  const [setupDialog, setSetupDialog] = useState(false);
  const [activeStep, setActiveStep] = useState(0);
  const [qrCode, setQrCode] = useState(null);
  const [secret, setSecret] = useState(null);
  const [backupCodes, setBackupCodes] = useState([]);
  const [verificationCode, setVerificationCode] = useState('');
  
  // Состояние для отключения 2FA
  const [disableDialog, setDisableDialog] = useState(false);
  const [password, setPassword] = useState('');
  
  // Состояние для генерации кодов
  const [generatingCodes, setGeneratingCodes] = useState(false);

  const steps = ['Scan QR Code', 'Enter Verification Code', 'Save Backup Codes'];

  useEffect(() => {
    loadStatus();
  }, []);

  const loadStatus = async () => {
    try {
      setLoading(true);
      const data = await get2FAStatus();
      setStatus(data);
      if (data?.backup_codes) {
        setBackupCodes(data.backup_codes);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Обработчик для включения 2FA
  const handleEnable2FA = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await enable2FA();
      setQrCode(data.qr_code);
      setSecret(data.secret);
      setBackupCodes(data.backup_codes || []);
      setSetupDialog(true);
      setActiveStep(0);
      setVerificationCode('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Обработчик для подтверждения настройки
  const handleVerifySetup = async () => {
    if (verificationCode.length !== 6) {
      setError('Please enter a valid 6-digit code');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const result = await verify2FASetup(verificationCode);
      setSuccess('2FA has been successfully enabled!');
      setActiveStep(2);
      await loadStatus();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDisable2FA = async () => {
    if (!password) {
      setError('Password is required');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      await disable2FA(password);
      setSuccess('2FA has been disabled');
      setDisableDialog(false);
      setPassword('');
      await loadStatus();
      setBackupCodes([]);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateBackupCodes = async () => {
    try {
      setGeneratingCodes(true);
      setError(null);
      const data = await generateBackupCodes();
      setBackupCodes(data.backup_codes || []);
      setSuccess('New backup codes generated');
    } catch (err) {
      setError(err.message);
    } finally {
      setGeneratingCodes(false);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    setSuccess('Copied to clipboard');
  };

  const downloadBackupCodes = () => {
    const element = document.createElement('a');
    const file = new Blob([backupCodes.join('\n')], { type: 'text/plain' });
    element.href = URL.createObjectURL(file);
    element.download = `backup-codes-${user?.username || 'user'}.txt`;
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
  };

  const handleCloseSetupDialog = () => {
    setSetupDialog(false);
    setActiveStep(0);
    setVerificationCode('');
  };

  if (loading && !status) {
    return (
      <Box display="flex" justifyContent="center" p={3}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ maxWidth: 800, mx: 'auto', p: 3 }}>
      <Typography variant="h5" gutterBottom>
        Two-Factor Authentication Settings
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" sx={{ mb: 3 }} onClose={() => setSuccess(null)}>
          {success}
        </Alert>
      )}

      <Paper sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Security sx={{ mr: 2, color: status?.enabled ? 'success.main' : 'warning.main' }} />
          <Box sx={{ flex: 1 }}>
            <Typography variant="h6">
              {status?.enabled ? '2FA is Enabled' : '2FA is Disabled'}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {status?.enabled 
                ? 'Your account is protected with two-factor authentication'
                : 'Add an extra layer of security to your account'}
            </Typography>
          </Box>
          <Switch
            checked={status?.enabled || false}
            onChange={() => {
              if (status?.enabled) {
                setDisableDialog(true);
              } else {
                handleEnable2FA();
              }
            }}
            color="primary"
            disabled={loading}
          />
        </Box>

        {status?.enabled && (
          <>
            <Divider sx={{ my: 2 }} />
            
            <Typography variant="subtitle1" gutterBottom>
              Backup Codes
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              Save these codes in a secure place. Each code can be used only once.
            </Typography>

            {backupCodes.length > 0 ? (
              <Paper variant="outlined" sx={{ p: 2, mb: 2, bgcolor: 'background.default' }}>
                <Grid container spacing={1}>
                  {backupCodes.map((code, index) => (
                    <Grid item xs={6} key={index}>
                      <Typography
                        variant="body2"
                        sx={{
                          fontFamily: 'monospace',
                          p: 1,
                          bgcolor: 'background.paper',
                          borderRadius: 1,
                          border: 1,
                          borderColor: 'divider',
                          textAlign: 'center',
                        }}
                      >
                        {code}
                      </Typography>
                    </Grid>
                  ))}
                </Grid>
              </Paper>
            ) : (
              <Alert severity="info" sx={{ mb: 2 }}>
                No backup codes generated yet. Generate them to ensure account access if you lose your phone.
              </Alert>
            )}

            <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
              <Button
                startIcon={<Download />}
                onClick={downloadBackupCodes}
                disabled={backupCodes.length === 0}
                variant="outlined"
              >
                Download Codes
              </Button>
              <Button
                startIcon={<ContentCopy />}
                onClick={() => copyToClipboard(backupCodes.join('\n'))}
                disabled={backupCodes.length === 0}
                variant="outlined"
              >
                Copy All
              </Button>
              <Button
                onClick={handleGenerateBackupCodes}
                disabled={generatingCodes}
                variant="outlined"
              >
                {generatingCodes ? <CircularProgress size={20} /> : 'Generate New Codes'}
              </Button>
            </Box>
          </>
        )}
      </Paper>

      <Alert severity="info">
        <Typography variant="body2">
          <strong>How 2FA works:</strong> After enabling, you'll need to enter a 6-digit code 
          from your authenticator app (like Google Authenticator or Authy) every time you log in.
        </Typography>
      </Alert>

      {/* Диалог настройки 2FA */}
      <Dialog 
        open={setupDialog} 
        onClose={handleCloseSetupDialog} 
        maxWidth="sm" 
        fullWidth
        disableEscapeKeyDown={activeStep === 2}
      >
        <DialogTitle>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <QrCode2 sx={{ mr: 1 }} />
            Setup Two-Factor Authentication
          </Box>
        </DialogTitle>
        
        <DialogContent dividers>
          <Stepper activeStep={activeStep} sx={{ my: 2 }}>
            {steps.map((label) => (
              <Step key={label}>
                <StepLabel>{label}</StepLabel>
              </Step>
            ))}
          </Stepper>

          {activeStep === 0 && qrCode && (
            <Box sx={{ textAlign: 'center', py: 2 }}>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                1. Scan this QR code with Google Authenticator or Authy
              </Typography>
              
              <Box
                component="img"
                src={`data:image/png;base64,${qrCode}`}
                alt="2FA QR Code"
                sx={{
                  width: 200,
                  height: 200,
                  my: 2,
                  mx: 'auto',
                  display: 'block',
                  border: 1,
                  borderColor: 'divider',
                  borderRadius: 1,
                }}
              />
              
              <Typography variant="body2" color="text.secondary" gutterBottom>
                Or enter this secret key manually:
              </Typography>
              
              <TextField
                value={secret}
                InputProps={{
                  readOnly: true,
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton 
                        onClick={() => copyToClipboard(secret)} 
                        edge="end"
                        size="small"
                      >
                        <ContentCopy />
                      </IconButton>
                    </InputAdornment>
                  ),
                }}
                fullWidth
                size="small"
                sx={{ mb: 2 }}
              />
            </Box>
          )}

          {activeStep === 1 && (
            <Box sx={{ py: 2 }}>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                2. Enter the 6-digit code from your authenticator app
              </Typography>
              
              <TextField
                fullWidth
                label="Verification Code"
                value={verificationCode}
                onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, ''))}
                placeholder="000000"
                inputProps={{ 
                  maxLength: 6, 
                  pattern: '[0-9]*',
                  style: { fontSize: '1.5rem', letterSpacing: '0.5rem', textAlign: 'center' }
                }}
                sx={{ mt: 2 }}
              />
            </Box>
          )}

          {activeStep === 2 && (
            <Box sx={{ py: 2 }}>
              <Alert severity="success" sx={{ mb: 2 }}>
                <Typography variant="body2">
                  2FA has been successfully enabled!
                </Typography>
              </Alert>
              
              <Alert severity="warning" sx={{ mb: 2 }}>
                <Typography variant="body2" fontWeight="bold" gutterBottom>
                  Save these backup codes in a secure place!
                </Typography>
                <Typography variant="body2">
                  Each code can be used only once. You'll need them if you lose access to your authenticator app.
                </Typography>
              </Alert>

              <Paper variant="outlined" sx={{ p: 2, bgcolor: 'background.default' }}>
                <Grid container spacing={1}>
                  {backupCodes.map((code, index) => (
                    <Grid item xs={6} key={index}>
                      <Typography
                        variant="body2"
                        sx={{
                          fontFamily: 'monospace',
                          p: 1,
                          bgcolor: 'background.paper',
                          borderRadius: 1,
                          border: 1,
                          borderColor: 'divider',
                          textAlign: 'center',
                        }}
                      >
                        {code}
                      </Typography>
                    </Grid>
                  ))}
                </Grid>
              </Paper>

              <Box sx={{ display: 'flex', gap: 2, mt: 2 }}>
                <Button
                  startIcon={<Download />}
                  onClick={downloadBackupCodes}
                  variant="outlined"
                  fullWidth
                >
                  Download Codes
                </Button>
                <Button
                  startIcon={<ContentCopy />}
                  onClick={() => copyToClipboard(backupCodes.join('\n'))}
                  variant="outlined"
                  fullWidth
                >
                  Copy All
                </Button>
              </Box>
            </Box>
          )}
        </DialogContent>

        <DialogActions>
          {activeStep === 0 && (
            <>
              <Button onClick={handleCloseSetupDialog}>Cancel</Button>
              <Button 
                variant="contained" 
                onClick={() => setActiveStep(1)}
                endIcon={<ArrowForward />}
              >
                Next
              </Button>
            </>
          )}

          {activeStep === 1 && (
            <>
              <Button 
                onClick={() => setActiveStep(0)}
                startIcon={<ArrowBack />}
              >
                Back
              </Button>
              <Button
                variant="contained"
                onClick={handleVerifySetup}
                disabled={verificationCode.length !== 6 || loading}
              >
                {loading ? <CircularProgress size={20} /> : 'Verify'}
              </Button>
            </>
          )}

          {activeStep === 2 && (
            <Button 
              variant="contained" 
              onClick={handleCloseSetupDialog}
            >
              Done
            </Button>
          )}
        </DialogActions>
      </Dialog>

      {/* Диалог отключения 2FA */}
      <Dialog open={disableDialog} onClose={() => setDisableDialog(false)}>
        <DialogTitle>Disable Two-Factor Authentication</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" gutterBottom>
            To disable 2FA, please enter your password:
          </Typography>
          <TextField
            fullWidth
            type="password"
            label="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            margin="normal"
            autoFocus
          />
          <Alert severity="warning" sx={{ mt: 2 }}>
            Disabling 2FA will make your account less secure.
          </Alert>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDisableDialog(false)}>Cancel</Button>
          <Button 
            onClick={handleDisable2FA} 
            color="error"
            disabled={!password || loading}
            variant="contained"
          >
            {loading ? <CircularProgress size={20} /> : 'Disable 2FA'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};