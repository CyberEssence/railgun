import { useCallback } from 'react';
import { useAuth } from '../context/AuthContext';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export const useApi = () => {
  const { getToken, logout, user } = useAuth();

  const fetchWithAuth = useCallback(async (endpoint, options = {}) => {
    const token = getToken();
    
    console.log('useApi - Token check:', { 
      hasToken: !!token, 
      tokenPreview: token ? token.substring(0, 20) + '...' : null,
      user: user ? 'present' : 'absent'
    });

    if (!token) {
      console.error('No authentication token available');
      throw new Error('Not authenticated. Please login again.');
    }

    const fullUrl = endpoint.startsWith('http') 
      ? endpoint 
      : `${API_BASE_URL}${endpoint}`;

    console.log('Making request to:', fullUrl);
    console.log('With token:', token.substring(0, 20) + '...');

    try {
      const response = await fetch(fullUrl, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
          ...options.headers,
        },
      });

      console.log('Response status:', response.status);

      if (response.status === 401) {
        const errorText = await response.text();
        console.error('401 Unauthorized:', errorText);
        logout();
        throw new Error('Session expired. Please login again.');
      }

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Request failed with status ${response.status}`);
      }

      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return await response.json();
      }

      return await response.text();
    } catch (error) {
      console.error('Fetch error:', error);
      throw error;
    }
  }, [getToken, logout, user]);

  return {
    get: (endpoint, options) => fetchWithAuth(endpoint, { method: 'GET', ...options }),
    post: (endpoint, data, options) => 
      fetchWithAuth(endpoint, { 
        method: 'POST', 
        body: JSON.stringify(data), 
        ...options 
      }),
    put: (endpoint, data, options) => 
      fetchWithAuth(endpoint, { 
        method: 'PUT', 
        body: JSON.stringify(data), 
        ...options 
      }),
    delete: (endpoint, options) => 
      fetchWithAuth(endpoint, { 
        method: 'DELETE', 
        ...options 
      }),
    fetch: fetchWithAuth,
  };
};