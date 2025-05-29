import React from 'react';
import { Box, AppBar, Toolbar, Typography, Button, IconButton, Menu, MenuItem, Avatar } from '@mui/material';
import { 
  Home, NetworkCheck, Storage, Security, Timeline, 
  Science, BugReport, CloudUpload, Logout 
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export const Navbar = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [anchorEl, setAnchorEl] = React.useState(null);
  const open = Boolean(anchorEl);

  const handleMenu = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static" elevation={0}>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1, fontWeight: 'bold' }}>
            SIEM Dashboard
          </Typography>
          
          <Button color="inherit" startIcon={<Home />} onClick={() => navigate('/')}>
            Обзор
          </Button>
          <Button color="inherit" startIcon={<NetworkCheck />} onClick={() => navigate('/traffic')}>
            Трафик
          </Button>
          <Button color="inherit" startIcon={<Storage />} onClick={() => navigate('/artifacts')}>
            Артефакты
          </Button>
          
          <Button color="inherit" startIcon={<Security />} onClick={() => navigate('/attack-patterns')}>
            Шаблоны атак
          </Button>
          <Button color="inherit" startIcon={<Timeline />} onClick={() => navigate('/apt-timeline')}>
            APT Timeline
          </Button>
          <Button color="inherit" startIcon={<Science />} onClick={() => navigate('/ai-models')}>
            AI Модели
          </Button>
          <Button color="inherit" startIcon={<BugReport />} onClick={() => navigate('/counter-attack')}>
            Контратака
          </Button>
          <Button color="inherit" startIcon={<CloudUpload />} onClick={() => navigate('/file-scanner')}>
            Сканер файлов
          </Button>

          {user && (
            <div>
              <IconButton
                size="large"
                aria-label="account of current user"
                aria-controls="menu-appbar"
                aria-haspopup="true"
                onClick={handleMenu}
                color="inherit"
              >
                <Avatar sx={{ width: 32, height: 32 }}>{user.username ? user.username.charAt(0).toUpperCase() : 'U'}</Avatar>
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
                open={open}
                onClose={handleClose}
              >
                <MenuItem onClick={() => { navigate('/profile'); handleClose(); }}>Профиль</MenuItem>
                <MenuItem onClick={() => { logout(); handleClose(); }}>
                  <Logout sx={{ mr: 1 }} /> Выйти
                </MenuItem>
              </Menu>
            </div>
          )}
        </Toolbar>
      </AppBar>
    </Box>
  );
};