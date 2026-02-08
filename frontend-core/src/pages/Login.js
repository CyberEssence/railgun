import React, { useState } from 'react';
import { 
  Box, Paper, Typography, TextField, Button, 
  Alert, CircularProgress, Grid, Link,
  Stepper, Step, StepLabel
} from '@mui/material';
import { 
  Lock, Person, VerifiedUser, QrCode2, 
  ContentCopy, Download, Visibility, VisibilityOff,
  ArrowBack
} from '@mui/icons-material';
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
  const [setup2FAData, setSetup2FAData] = useState(null);
  const [setupStep, setSetupStep] = useState(0);
  const [showPassword, setShowPassword] = useState(false);
  const [verificationCode, setVerificationCode] = useState('');
  const [enable2FALoading, setEnable2FALoading] = useState(false);
  
  const { login, verify2FA, enable2FA, verify2FASetup, user } = useAuth();
  const navigate = useNavigate();

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await login(username, password);
      
      if (result.requires2FA) {
        setTwoFARequired(true);
        setUserId(result.userId);
      } else {
        navigate('/');
      }
    } catch (err) {
      setError(err.message || 'Login failed');
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
      setError(err.message || 'Invalid 2FA code');
    } finally {
      setLoading(false);
    }
  };

  const handleEnable2FA = async () => {
    setEnable2FALoading(true);
    setError(null);
    
    try {
      const data = await enable2FA();
      setSetup2FAData(data);
      setSetupStep(1);
    } catch (err) {
      setError(err.message || 'Failed to enable 2FA');
    } finally {
      setEnable2FALoading(false);
    }
  };

  const handleVerifySetup = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await verify2FASetup(verificationCode);
      
      if (result.success || result.status === 'success') {
        setSetupStep(3);
        setError(null);
      } else {
        setError(result.message || 'Verification failed');
      }
    } catch (err) {
      setError(err.message || 'Invalid verification code. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    // Можно добавить toast уведомление
    alert('Copied to clipboard!');
  };

  const downloadQRCode = () => {
    if (!setup2FAData?.qr_code) return;
    
    const link = document.createElement('a');
    link.href = `data:image/png;base64,${setup2FAData.qr_code}`;
    link.download = 'railgun-2fa-qrcode.png';
    link.click();
  };

  const downloadBackupCodes = () => {
    if (!setup2FAData?.backup_codes) return;
    
    const blob = new Blob([setup2FAData.backup_codes.join('\n')], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'railgun-2fa-backup-codes.txt';
    link.click();
  };

  const steps = ['Enable 2FA', 'Scan QR Code', 'Verify Setup', 'Complete'];

  const resetState = () => {
    setSetup2FAData(null);
    setSetupStep(0);
    setVerificationCode('');
    setTwoFARequired(false);
    setTwoFAToken('');
    setError(null);
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
        
        {setup2FAData ? (
          <>
            <Stepper activeStep={setupStep} sx={{ mb: 3 }}>
              {steps.map((label) => (
                <Step key={label}>
                  <StepLabel>{label}</StepLabel>
                </Step>
              ))}
            </Stepper>
            
            {setupStep === 1 && (
              <>
                <Typography variant="h6" align="center" gutterBottom>
                  Setup Two-Factor Authentication
                </Typography>
                
                <Box sx={{ textAlign: 'center', mb: 3 }}>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Scan this QR code with Google Authenticator, Authy, or Microsoft Authenticator
                  </Typography>
                  
                  {setup2FAData.qr_code && (
                    <Box sx={{ mb: 2 }}>
                      <img 
                        src={`data:image/png;base64,${setup2FAData.qr_code}`} 
                        alt="QR Code"
                        style={{ 
                          width: 200, 
                          height: 200, 
                          margin: '0 auto',
                          border: '1px solid #333',
                          borderRadius: '8px'
                        }}
                      />
                    </Box>
                  )}
                  
                  <Box sx={{ display: 'flex', justifyContent: 'center', gap: 1, mb: 3 }}>
                    <Button
                      startIcon={<Download />}
                      onClick={downloadQRCode}
                      variant="outlined"
                      size="small"
                    >
                      Download QR
                    </Button>
                    
                    {setup2FAData.url && (
                      <Button
                        startIcon={<ContentCopy />}
                        onClick={() => copyToClipboard(setup2FAData.url)}
                        variant="outlined"
                        size="small"
                      >
                        Copy URL
                      </Button>
                    )}
                  </Box>
                  
                  <Typography variant="body2" sx={{ mt: 2, mb: 1 }}>
                    <strong>Or enter this secret key manually:</strong>
                  </Typography>
                  
                  <Paper sx={{ 
                    p: 2, 
                    mb: 2, 
                    bgcolor: 'grey.900',
                    borderRadius: 1
                  }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                      <Typography variant="body1" fontFamily="monospace">
                        {setup2FAData.secret}
                      </Typography>
                      <Button
                        startIcon={<ContentCopy />}
                        onClick={() => copyToClipboard(setup2FAData.secret)}
                        size="small"
                      >
                        Copy
                      </Button>
                    </Box>
                  </Paper>
                  
                  <Alert severity="warning" sx={{ mb: 2 }}>
                    <strong>Important:</strong> Save this secret key in a secure place! 
                    You won't see it again.
                  </Alert>
                  
                  <Box sx={{ display: 'flex', gap: 2, mt: 2 }}>
                    <Button
                      fullWidth
                      variant="outlined"
                      startIcon={<ArrowBack />}
                      onClick={resetState}
                    >
                      Cancel
                    </Button>
                    <Button
                      fullWidth
                      variant="contained"
                      onClick={() => setSetupStep(2)}
                    >
                      Next: Verify Setup
                    </Button>
                  </Box>
                </Box>
              </>
            )}
            
            {setupStep === 2 && (
              <>
                <Typography variant="h6" align="center" gutterBottom>
                  Verify Setup
                </Typography>
                
                <Typography variant="body1" align="center" sx={{ mb: 3 }}>
                  Enter the 6-digit code from your authenticator app
                </Typography>
                
                {error && (
                  <Alert severity="error" sx={{ mb: 3 }}>
                    {error}
                  </Alert>
                )}
                
                <TextField
                  fullWidth
                  label="Verification Code"
                  value={verificationCode}
                  onChange={(e) => {
                    const value = e.target.value.replace(/\D/g, '').slice(0, 6);
                    setVerificationCode(value);
                    if (error) setError(null);
                  }}
                  margin="normal"
                  InputProps={{
                    startAdornment: <VerifiedUser sx={{ mr: 1, color: 'action.active' }} />,
                    inputProps: {
                      maxLength: 6,
                      pattern: '[0-9]*',
                      inputMode: 'numeric'
                    }
                  }}
                  helperText="Enter exactly 6 digits from your authenticator app"
                  error={!!error}
                  autoFocus
                />
                
                <Box sx={{ display: 'flex', gap: 2, mt: 3 }}>
                  <Button
                    fullWidth
                    variant="outlined"
                    onClick={() => setSetupStep(1)}
                  >
                    Back
                  </Button>
                  <Button
                    fullWidth
                    variant="contained"
                    onClick={handleVerifySetup}
                    disabled={loading || verificationCode.length !== 6}
                    sx={{ height: 48 }}
                  >
                    {loading ? (
                      <CircularProgress size={24} color="inherit" />
                    ) : (
                      'Verify & Continue'
                    )}
                  </Button>
                </Box>
                
                <Typography variant="body2" color="text.secondary" align="center" sx={{ mt: 2 }}>
                  <strong>Tip:</strong> Make sure the time on your device is synchronized
                </Typography>
              </>
            )}
            
            {setupStep === 3 && (
              <>
                <Box sx={{ textAlign: 'center', mb: 3 }}>
                  <VerifiedUser sx={{ fontSize: 60, color: 'success.main', mb: 2 }} />
                  <Typography variant="h5" gutterBottom color="success.main">
                    ✓ 2FA Successfully Enabled!
                  </Typography>
                </Box>
                
                <Alert severity="success" sx={{ mb: 3 }}>
                  Your account is now protected with two-factor authentication.
                  You'll need to enter a verification code each time you log in.
                </Alert>
                
                {setup2FAData?.backup_codes && (
                  <>
                    <Typography variant="subtitle1" gutterBottom>
                      <strong>Important: Save Your Backup Codes</strong>
                    </Typography>
                    
                    <Paper sx={{ 
                      p: 2, 
                      mb: 2, 
                      bgcolor: 'grey.900',
                      border: '1px solid',
                      borderColor: 'warning.main'
                    }}>
                      <Typography variant="body2" sx={{ mb: 1 }}>
                        Each code can be used only once. Save them in a secure place:
                      </Typography>
                      
                      {setup2FAData.backup_codes.map((code, index) => (
                        <Box 
                          key={index} 
                          sx={{ 
                            display: 'flex', 
                            alignItems: 'center', 
                            justifyContent: 'space-between',
                            mb: 1,
                            p: 1,
                            bgcolor: 'grey.800',
                            borderRadius: 1
                          }}
                        >
                          <Typography variant="body2" fontFamily="monospace">
                            {code}
                          </Typography>
                          <Button
                            size="small"
                            onClick={() => copyToClipboard(code)}
                          >
                            <ContentCopy fontSize="small" />
                          </Button>
                        </Box>
                      ))}
                    </Paper>
                    
                    <Grid container spacing={1} sx={{ mb: 3 }}>
                      <Grid item xs={6}>
                        <Button
                          fullWidth
                          startIcon={<ContentCopy />}
                          onClick={() => copyToClipboard(setup2FAData.backup_codes.join('\n'))}
                          variant="outlined"
                        >
                          Copy All
                        </Button>
                      </Grid>
                      <Grid item xs={6}>
                        <Button
                          fullWidth
                          startIcon={<Download />}
                          onClick={downloadBackupCodes}
                          variant="outlined"
                        >
                          Download
                        </Button>
                      </Grid>
                    </Grid>
                  </>
                )}
                
                <Button
                  fullWidth
                  variant="contained"
                  onClick={() => {
                    resetState();
                    navigate('/');
                  }}
                  sx={{ mt: 2 }}
                >
                  Continue to Dashboard
                </Button>
              </>
            )}
          </>
        ) : !twoFARequired ? (
          <>
            <Typography variant="h6" align="center" gutterBottom>
              Sign In
            </Typography>
            
            {error && (
              <Alert severity="error" sx={{ mb: 3 }}>
                {error}
              </Alert>
            )}
            
            <TextField
              fullWidth
              label="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              margin="normal"
              InputProps={{
                startAdornment: <Person sx={{ mr: 1, color: 'action.active' }} />
              }}
              disabled={loading}
            />
            
            <TextField
              fullWidth
              label="Password"
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              margin="normal"
              InputProps={{
                startAdornment: <Lock sx={{ mr: 1, color: 'action.active' }} />,
                endAdornment: (
                  <Button
                    onClick={() => setShowPassword(!showPassword)}
                    size="small"
                    sx={{ minWidth: 'auto' }}
                  >
                    {showPassword ? <VisibilityOff /> : <Visibility />}
                  </Button>
                )
              }}
              disabled={loading}
            />
            
            <Button
              fullWidth
              variant="contained"
              onClick={handleLogin}
              disabled={loading || !username || !password}
              sx={{ mt: 2, mb: 2, height: 48 }}
            >
              {loading ? (
                <CircularProgress size={24} color="inherit" />
              ) : (
                'Sign In'
              )}
            </Button>
            
            <Grid container sx={{ mt: 1 }}>
              <Grid item xs>
                <Link href="#" variant="body2" sx={{ textDecoration: 'none' }}>
                  Forgot password?
                </Link>
              </Grid>
              <Grid item>
                <Link href="/register" variant="body2" sx={{ textDecoration: 'none' }}>
                  Create account
                </Link>
              </Grid>
            </Grid>
            
            {user && (
              <Box sx={{ mt: 3, pt: 2, borderTop: '1px solid rgba(255,255,255,0.1)' }}>
                <Typography variant="body2" align="center" color="text.secondary" sx={{ mb: 1 }}>
                  Enhance your account security
                </Typography>
                <Button
                  fullWidth
                  startIcon={<QrCode2 />}
                  onClick={handleEnable2FA}
                  disabled={enable2FALoading}
                  variant="outlined"
                  size="small"
                >
                  {enable2FALoading ? (
                    <CircularProgress size={16} />
                  ) : (
                    'Enable Two-Factor Authentication'
                  )}
                </Button>
              </Box>
            )}
          </>
        ) : (
          <>
            <Typography variant="h6" align="center" gutterBottom>
              Two-Factor Authentication
            </Typography>
            
            <Typography variant="body1" align="center" sx={{ mb: 3 }}>
              Please enter the 6-digit code from your authenticator app
            </Typography>
            
            {error && (
              <Alert severity="error" sx={{ mb: 3 }}>
                {error}
              </Alert>
            )}
            
            <TextField
              fullWidth
              label="Verification Code"
              value={twoFAToken}
              onChange={(e) => {
                const value = e.target.value.replace(/\D/g, '').slice(0, 6);
                setTwoFAToken(value);
                if (error) setError(null);
              }}
              margin="normal"
              InputProps={{
                startAdornment: <VerifiedUser sx={{ mr: 1, color: 'action.active' }} />,
                inputProps: {
                  maxLength: 6,
                  pattern: '[0-9]*',
                  inputMode: 'numeric'
                }
              }}
              helperText="Enter the 6-digit code from Google Authenticator or similar app"
              error={!!error}
              autoFocus
            />
            
            <Box sx={{ display: 'flex', gap: 2, mt: 3 }}>
              <Button
                fullWidth
                variant="outlined"
                onClick={() => {
                  setTwoFARequired(false);
                  setTwoFAToken('');
                  setError(null);
                }}
              >
                Back to Login
              </Button>
              <Button
                fullWidth
                variant="contained"
                onClick={handleVerify2FA}
                disabled={loading || twoFAToken.length !== 6}
                sx={{ height: 48 }}
              >
                {loading ? (
                  <CircularProgress size={24} color="inherit" />
                ) : (
                  'Verify & Continue'
                )}
              </Button>
            </Box>
            
            <Typography variant="body2" color="text.secondary" align="center" sx={{ mt: 2 }}>
              Having trouble? Try using a backup code or contact support.
            </Typography>
          </>
        )}
      </Paper>
    </Box>
  );
};