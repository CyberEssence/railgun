import React from 'react';
import { 
  AppBar, 
  Toolbar, 
  Typography, 
  Button, 
  Box, 
  IconButton,
  Avatar,
  Menu,
  MenuItem
} from '@mui/material';
import { 
  Home, 
  NetworkCheck, 
  Storage, 
  Security,
  AccountCircle,
  Login,
  Logout
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export const Navbar = () => {
  const navigate = useNavigate();
  const { authToken, user, logout } = useAuth();
  const [anchorEl, setAnchorEl] = React.useState(null);

  const handleMenu = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <AppBar position="static" sx={{ mb: 4 }}>
      <Toolbar>
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          SIEM Dashboard
        </Typography>
        
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
          <Button 
            color="inherit" 
            startIcon={<Home />} 
            onClick={() => navigate('/')}
            sx={{ mr: 1 }}
          >
            Dashboard
          </Button>
          <Button 
            color="inherit" 
            startIcon={<NetworkCheck />} 
            onClick={() => navigate('/traffic')}
            sx={{ mr: 1 }}
          >
            Traffic
          </Button>
          <Button 
            color="inherit" 
            startIcon={<Storage />} 
            onClick={() => navigate('/artifacts')}
            sx={{ mr: 1 }}
          >
            Artifacts
          </Button>
          <Button 
            color="inherit" 
            startIcon={<Security />} 
            onClick={() => navigate('/threats')}
            sx={{ mr: 2 }}
          >
            Threats
          </Button>
          
          {authToken ? (
            <div>
              <IconButton
                size="large"
                aria-label="account of current user"
                aria-controls="menu-appbar"
                aria-haspopup="true"
                onClick={handleMenu}
                color="inherit"
              >
                <Avatar sx={{ width: 32, height: 32 }}>
                  {user?.username?.charAt(0).toUpperCase() || <AccountCircle />}
                </Avatar>
              </IconButton>
              <Menu
                id="menu-appbar"
                anchorEl={anchorEl}
                anchorOrigin={{
                  vertical: 'top',
                  horizontal: 'right',
                }}
                keepMounted
                transformOrigin={{
                  vertical: 'top',
                  horizontal: 'right',
                }}
                open={Boolean(anchorEl)}
                onClose={handleClose}
              >
                <MenuItem onClick={() => { navigate('/profile'); handleClose(); }}>
                  Profile
                </MenuItem>
                <MenuItem onClick={() => { logout(); handleClose(); }}>
                  <Logout sx={{ mr: 1 }} /> Logout
                </MenuItem>
              </Menu>
            </div>
          ) : (
            <Button 
              color="inherit" 
              startIcon={<Login />} 
              onClick={() => navigate('/login')}
            >
              Login
            </Button>
          )}
        </Box>
      </Toolbar>
    </AppBar>
  );
};