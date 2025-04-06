import React from 'react';
import { Box, AppBar, Toolbar, Typography, Button } from '@mui/material';
import { Home, NetworkCheck, Storage } from '@mui/icons-material';

export const Navbar = () => {
  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            SIEM Dashboard
          </Typography>
          <Button color="inherit" startIcon={<Home />}>
            Обзор
          </Button>
          <Button color="inherit" startIcon={<NetworkCheck />}>
            Трафик
          </Button>
          <Button color="inherit" startIcon={<Storage />}>
            Артефакты
          </Button>
        </Toolbar>
      </AppBar>
    </Box>
  );
};