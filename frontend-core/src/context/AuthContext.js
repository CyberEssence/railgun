import React, { createContext, useState, useContext, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    // Проверка аутентификации при загрузке
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      setUser(JSON.parse(storedUser));
    }
    setLoading(false);
  }, []);

  const login = async (username, password) => {
    try {
      const response = await fetch('http://localhost:8080/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });
      
      if (!response.ok) throw new Error(await response.text());
      
      const data = await response.json();
      
      // Если требуется 2FA
      if (data.requiresTwoFA) {
        return { requires2FA: true, userId: data.userId, token: data.twoFAToken };
      }
      
      // Если аутентификация завершена
      const userData = { 
        username, 
        token: data.accessToken,
        refreshToken: data.refreshToken
      };
      
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      return { success: true };
    } catch (err) {
      throw err;
    }
  };

  const verify2FA = async (userId, token) => {
    try {
      const response = await fetch('http://localhost:8080/api/auth/verify-2fa', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: userId, token })
      });
      
      if (!response.ok) throw new Error(await response.text());
      
      const data = await response.json();
      const userData = { 
        username: data.username, 
        token: data.accessToken,
        refreshToken: data.refreshToken
      };
      
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      return true;
    } catch (err) {
      throw err;
    }
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem('user');
    navigate('/login');
  };

  const refreshToken = async () => {
    if (!user?.refreshToken) {
      logout();
      return;
    }
    
    try {
      const response = await fetch('http://localhost:8080/api/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: user.refreshToken })
      });
      
      if (!response.ok) throw new Error(await response.text());
      
      const data = await response.json();
      const userData = { 
        ...user, 
        token: data.accessToken,
        refreshToken: data.refreshToken
      };
      
      setUser(userData);
      localStorage.setItem('user', JSON.stringify(userData));
      return data.accessToken;
    } catch (err) {
      logout();
      throw err;
    }
  };

  return (
    <AuthContext.Provider value={{ 
      user, 
      loading, 
      login, 
      logout, 
      verify2FA,
      refreshToken
    }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);