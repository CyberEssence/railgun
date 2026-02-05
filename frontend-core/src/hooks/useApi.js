// hooks/useApi.js
import { useCallback } from 'react';
import { useAuth } from '../context/AuthContext';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export const useApi = () => {
  const { getToken, logout } = useAuth();

  const fetchWithAuth = useCallback(async (endpoint, options = {}) => {
    const token = getToken();
    
    if (!token) {
      console.error('No authentication token available');
      logout();
      throw new Error('Not authenticated');
    }

    const fullUrl = endpoint.startsWith('http') 
      ? endpoint 
      : `${API_BASE_URL}${endpoint}`;

    console.log('API Request:', {
      url: fullUrl,
      method: options.method || 'GET',
      hasToken: !!token,
      tokenPreview: token.substring(0, 20) + '...'
    });

    const response = await fetch(fullUrl, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        ...options.headers,
      },
    });

    console.log('API Response:', {
      status: response.status,
      statusText: response.statusText,
      url: fullUrl
    });

    if (response.status === 401) {
      console.log('Authentication failed (401)');
      logout();
      throw new Error('Session expired. Please login again.');
    }

    if (!response.ok) {
      const errorText = await response.text();
      console.error('API Error:', errorText);
      throw new Error(errorText || `Request failed with status ${response.status}`);
    }

    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      return await response.json();
    }

    return await response.text();
  }, [getToken, logout]);

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