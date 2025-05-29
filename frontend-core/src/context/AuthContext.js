import React, { createContext, useContext, useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [authToken, setAuthToken] = useState(null);
  const [user, setUser] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    const token = localStorage.getItem('authToken');
    const userData = localStorage.getItem('user');
    if (token && userData) {
      setAuthToken(token);
      setUser(JSON.parse(userData));
    }
  }, []);

  const login = async (username, password) => {
    try {
      const response = await fetch('http://localhost:8080/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      const data = await response.json();
      
      if (response.ok) {
        if (data.requiresTwoFA) {
          return { requiresTwoFA: true, userId: data.userId };
        }

        localStorage.setItem('authToken', data.access_token);
        localStorage.setItem('user', JSON.stringify({ id: data.userId }));
        setAuthToken(data.access_token);
        setUser({ id: data.userId });
        return { success: true };
      } else {
        throw new Error(data.error || 'Login failed');
      }
    } catch (error) {
      throw error;
    }
  };

  const verify2FA = async (userId, token) => {
    try {
      const response = await fetch('http://localhost:8080/api/auth/verify-2fa', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ user_id: userId, token }),
      });

      const data = await response.json();
      
      if (response.ok) {
        localStorage.setItem('authToken', data.access_token);
        localStorage.setItem('user', JSON.stringify({ id: userId }));
        setAuthToken(data.access_token);
        setUser({ id: userId });
        return true;
      } else {
        throw new Error(data.error || '2FA verification failed');
      }
    } catch (error) {
      throw error;
    }
  };

  const logout = () => {
    localStorage.removeItem('authToken');
    localStorage.removeItem('user');
    setAuthToken(null);
    setUser(null);
    navigate('/login');
  };

  return (
    <AuthContext.Provider value={{ authToken, user, login, verify2FA, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);