import React, { useState, useEffect } from 'react';
import {
  Box, Paper, Typography, Button, Alert, CircularProgress,
  Dialog, DialogTitle, DialogContent, DialogActions,
  TextField, List, ListItem, ListItemText, ListItemSecondaryAction,
  IconButton, Divider, Switch, Grid
} from '@mui/material';
import {
  QrCode2, Delete, ContentCopy, Download,
  Security, Lock, LockOpen
} from '@mui/icons-material';
import { useAuth } from '../context/AuthContext';

export const TwoFASettings = () => {
  const { get2FAStatus, disable2FA, generateBackupCodes, user } = useAuth();
  const [status, setStatus] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [disableDialog, setDisableDialog] = useState(false);
  const [password, setPassword] = useState('');
  const [backupCodes, setBackupCodes] = useState([]);
  const [generatingCodes, setGeneratingCodes] = useState(false);

  useEffect(() => {
    loadStatus();
  }, []);

  const loadStatus = async () => {
    try {
      setLoading(true);
      const data = await get2FAStatus();
      setStatus(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDisable2FA = async () => {
    try {
      await disable2FA(password);
      setDisableDialog(false);
      setPassword('');
      await loadStatus();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleGenerateBackupCodes = async () => {
    try {
      setGeneratingCodes(true);
      const data = await generateBackupCodes();
      setBackupCodes(data.backup_codes || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setGeneratingCodes(false);
    }
  };

  if (loading) {
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
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <Paper sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Security sx={{ mr: 2, color: status?.enabled ? 'success.main' : 'warning.main' }} />
          <Box>
            <Typography variant="h6">
              {status?.enabled ? '2FA is Enabled' : '2FA is Disabled'}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {status?.enabled 
                ? 'Your account is protected with two-factor authentication'
                : 'Add an extra layer of security to your account'}
            </Typography>
          </Box>
          <Box sx={{ ml: 'auto' }}>
            <Switch
              checked={status?.enabled}
              onChange={() => setDisableDialog(true)}
              color="primary"
            />
          </Box>
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
              <Paper variant="outlined" sx={{ p: 2, mb: 2 }}>
                <List dense>
                  {backupCodes.map((code, index) => (
                    <ListItem key={index}>
                      <ListItemText 
                        primary={code}
                        primaryTypographyProps={{ fontFamily: 'monospace' }}
                      />
                      <ListItemSecondaryAction>
                        <IconButton
                          edge="end"
                          onClick={() => navigator.clipboard.writeText(code)}
                          size="small"
                        >
                          <ContentCopy />
                        </IconButton>
                      </ListItemSecondaryAction>
                    </ListItem>
                  ))}
                </List>
              </Paper>
            ) : (
              <Alert severity="info" sx={{ mb: 2 }}>
                No backup codes generated yet.
              </Alert>
            )}

            <Button
              startIcon={<Download />}
              onClick={handleGenerateBackupCodes}
              disabled={generatingCodes}
              variant="outlined"
              sx={{ mr: 2 }}
            >
              {generatingCodes ? <CircularProgress size={20} /> : 'Generate New Codes'}
            </Button>

            <Button
              startIcon={<ContentCopy />}
              onClick={() => navigator.clipboard.writeText(backupCodes.join('\n'))}
              disabled={backupCodes.length === 0}
              variant="outlined"
            >
              Copy All
            </Button>
          </>
        )}
      </Paper>

      <Alert severity="info">
        <Typography variant="body2">
          <strong>How 2FA works:</strong> After enabling, you'll need to enter a 6-digit code 
          from your authenticator app (like Google Authenticator or Authy) every time you log in.
        </Typography>
      </Alert>

      {/* Диалог отключения 2FA */}
      <Dialog open={disableDialog} onClose={() => setDisableDialog(false)}>
        <DialogTitle>Disable Two-Factor Authentication</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" gutterBottom>
            To disable 2FA, please confirm your password:
          </Typography>
          <TextField
            fullWidth
            type="password"
            label="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            margin="normal"
          />
          <Alert severity="warning" sx={{ mt: 2 }}>
            Disabling 2FA reduces your account security.
          </Alert>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDisableDialog(false)}>Cancel</Button>
          <Button 
            onClick={handleDisable2FA} 
            color="error"
            disabled={!password}
          >
            Disable 2FA
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};