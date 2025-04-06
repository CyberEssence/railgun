import React from 'react';
import { Box, AppBar, Toolbar, Typography, Button } from '@mui/material';
import { Home, NetworkCheck, Storage } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

export const Navbar = () => {
    const navigate = useNavigate(); // Хук для навигации

    return (
        <Box sx={{ flexGrow: 1 }}>
            <AppBar position="static">
                <Toolbar>
                    <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
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
                </Toolbar>
            </AppBar>
        </Box>
    );
};
