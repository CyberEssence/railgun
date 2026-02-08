import React, { createContext, useState, useContext, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

const AuthContext = createContext();

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      try {
        const parsedUser = JSON.parse(storedUser);
        setUser(parsedUser);
      } catch (error) {
        localStorage.removeItem('user');
      }
    }
    setLoading(false);
  }, []);

  // Функция для обычного логина
  const login = async (username, password) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Login failed');
      }

      const data = await response.json();
      
      // Если требуется 2FA
      if (data.requires_2fa) {
        return {
          requires2FA: true,
          userId: data.user_id,
          message: data.message || 'Please enter your 2FA code',
        };
      }
      
      // Если 2FA не требуется
      const userData = {
        id: data.user_id,
        username: username,
        email: data.email,
        token: data.access_token,
        refreshToken: data.refresh_token,
        expiresIn: data.expires_in,
      };
      
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      return { success: true, user: userData };
      
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  };

  // Верификация TOTP кода
  const verify2FA = async (userId, token) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/verify-2fa`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          user_id: userId, 
          token: token 
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Invalid 2FA code');
      }

      const data = await response.json();
      
      const userData = {
        id: userId,
        username: data.username || 'User',
        email: data.email,
        token: data.access_token,
        refreshToken: data.refresh_token,
        expiresIn: data.expires_in,
      };
      
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      return true;
      
    } catch (error) {
      console.error('2FA verification error:', error);
      throw error;
    }
  };

  // Включение 2FA
  const enable2FA = async () => {
    try {
      const token = user?.token;
      if (!token) throw new Error('Not authenticated');

      const response = await fetch(`${API_BASE_URL}/api/auth/2fa/enable`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to enable 2FA');
      }

      return await response.json();
    } catch (error) {
      console.error('Enable 2FA error:', error);
      throw error;
    }
  };

  // Подтверждение настройки 2FA
  const verify2FASetup = async (token) => {
    try {
      const userToken = user?.token;
      if (!userToken) throw new Error('Not authenticated');

      console.log('🔐 Sending 2FA setup verification request...', {
        url: `${API_BASE_URL}/api/auth/2fa/verify-setup`,
        tokenLength: token?.length,
        authToken: userToken.substring(0, 20) + '...'
      });

      const response = await fetch(`${API_BASE_URL}/api/auth/2fa/verify-setup`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${userToken}`,
        },
        body: JSON.stringify({ token }),
      });

      console.log('Response received:', {
        status: response.status,
        statusText: response.statusText,
        ok: response.ok
      });

      if (!response.ok) {
        // Пытаемся получить ошибку в JSON формате
        let errorMessage = `HTTP ${response.status}`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || errorData.message || errorMessage;
        } catch (e) {
          // Если не JSON, читаем как текст
          const text = await response.text();
          errorMessage = text || errorMessage;
        }
        throw new Error(errorMessage);
      }

      const data = await response.json();
      console.log('Verify setup success:', data);
      
      // Возвращаем полные данные для обновления состояния
      return {
        success: true,
        data: data,
        message: data.message || '2FA setup verified successfully!'
      };

    } catch (error) {
      console.error('Verify 2FA setup error:', error);
      throw error;
    }
  };

  // Отключение 2FA
  const disable2FA = async (password) => {
    try {
      const userToken = user?.token;
      if (!userToken) throw new Error('Not authenticated');

      const response = await fetch(`${API_BASE_URL}/api/auth/2fa/disable`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${userToken}`,
        },
        body: JSON.stringify({ password }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to disable 2FA');
      }

      return await response.json();
    } catch (error) {
      console.error('Disable 2FA error:', error);
      throw error;
    }
  };

  // Получение статуса 2FA
  const get2FAStatus = async () => {
    try {
      const token = user?.token;
      if (!token) throw new Error('Not authenticated');

      const response = await fetch(`${API_BASE_URL}/api/auth/2fa/status`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to get 2FA status');
      }

      return await response.json();
    } catch (error) {
      console.error('Get 2FA status error:', error);
      throw error;
    }
  };

  // Генерация новых резервных кодов
  const generateBackupCodes = async () => {
    try {
      const token = user?.token;
      if (!token) throw new Error('Not authenticated');

      const response = await fetch(`${API_BASE_URL}/api/auth/2fa/new-backup-codes`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to generate backup codes');
      }

      return await response.json();
    } catch (error) {
      console.error('Generate backup codes error:', error);
      throw error;
    }
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('user');
    navigate('/login');
  };

  const register = async (username, email, password) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, email, password }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Registration failed');
      }

      return await response.json();
    } catch (error) {
      console.error('Registration error:', error);
      throw error;
    }
  };

  const getToken = () => {
    return user?.token;
  };

  const isAuthenticated = useCallback(() => {
    return !!user?.token;
  }, [user]);

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        login,
        logout,
        verify2FA,
        register,
        enable2FA,
        verify2FASetup,
        disable2FA,
        get2FAStatus,
        generateBackupCodes,
        getToken,
        isAuthenticated,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};