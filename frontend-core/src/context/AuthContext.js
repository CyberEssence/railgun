// context/AuthContext.js
import React, { createContext, useState, useContext, useEffect } from 'react';
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

  const login = async (username, password) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Login failed');
      }

      const data = await response.json();
      console.log('Login response:', data);
      
      // Проверяем требуется ли 2FA (используем разные варианты написания)
      if (data.requires_2fa || data.requiresTwoFA || data.requires2FA) {
        return {
          requires2FA: true,
          userId: data.user_id || data.userId,
          message: data.message,
          twoFAToken: data.two_fa_token || data.twoFAToken,
        };
      }
      
      // Если 2FA не требуется, сразу сохраняем токены
      if (data.access_token || data.accessToken) {
        const userData = {
          id: data.user_id || data.userId || 1,
          username: username,
          email: data.email,
          token: data.access_token || data.accessToken,
          refreshToken: data.refresh_token || data.refreshToken,
          expiresIn: data.expires_in,
        };
        
        setUser(userData);
        localStorage.setItem('user', JSON.stringify(userData));
        return { success: true, user: userData };
      }
      
      throw new Error('Unexpected login response format');
      
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  };

  const verify2FA = async (userId, token) => {
    try {
      console.log('Sending 2FA verification:', { userId, token });
      
      const response = await fetch(`${API_BASE_URL}/api/auth/verify-2fa`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          user_id: userId, 
          token: token 
        }),
      });

      console.log('2FA response status:', response.status);
      
      if (!response.ok) {
        const errorText = await response.text();
        console.error('2FA error response:', errorText);
        throw new Error(errorText || '2FA verification failed');
      }

      const data = await response.json();
      console.log('2FA success response:', data);
      
      // Проверяем наличие токенов в ответе
      if (!data.access_token && !data.accessToken) {
        throw new Error('No access token received from server');
      }
      
      const userData = {
        id: userId,
        username: data.username || 'User',
        email: data.email,
        token: data.access_token || data.accessToken,
        refreshToken: data.refresh_token || data.refreshToken,
        expiresIn: data.expires_in || 3600,
      };
      
      console.log('User data to save:', userData);
      
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      
      return true;
      
    } catch (error) {
      console.error('2FA verification error:', error);
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
        const errorText = await response.text();
        throw new Error(errorText || 'Registration failed');
      }

      const data = await response.json();
      return { success: true };
    } catch (error) {
      console.error('Registration error:', error);
      throw error;
    }
  };

  const getToken = () => {
    return user?.token;
  };

  const isAuthenticated = () => {
    const token = getToken();
    return !!token;
  };

  const token = getToken();

  // Тестовая функция для проверки API
  const testAuthAPI = async () => {
    try {
      const testResponse = await fetch(`${API_BASE_URL}/api/ai/models`, {
        headers: {
          'Authorization': `Bearer ${getToken()}`,
        },
      });
      console.log('Test API response:', await testResponse.text());
      return testResponse.ok;
    } catch (error) {
      console.error('Test API error:', error);
      return false;
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        login,
        logout,
        verify2FA,
        register,
        getToken,
        isAuthenticated: isAuthenticated(),
        token,
        testAuthAPI,
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