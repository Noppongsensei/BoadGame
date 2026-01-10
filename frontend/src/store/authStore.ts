import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import axios from 'axios';

// Define the user type
interface User {
  id: string;
  username: string;
}

// Define the authentication state
interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string) => Promise<void>;
  logout: () => void;
}

// Create the authentication store with persistence
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Login function
      login: async (username: string, password: string) => {
        try {
          set({ isLoading: true, error: null });
          const response = await axios.post('/api/auth/login', { username, password });
          const { user, token } = response.data;
          
          set({ 
            user, 
            token,
            isAuthenticated: true,
            isLoading: false 
          });
          
          // Set the token in axios default headers
          axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
        } catch (error) {
          set({ 
            isLoading: false, 
            error: error.response?.data?.error || 'Failed to login' 
          });
          throw error;
        }
      },

      // Register function
      register: async (username: string, password: string) => {
        try {
          set({ isLoading: true, error: null });
          const response = await axios.post('/api/auth/register', { username, password });
          const { user, token } = response.data;
          
          set({ 
            user, 
            token,
            isAuthenticated: true,
            isLoading: false 
          });
          
          // Set the token in axios default headers
          axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
        } catch (error) {
          set({ 
            isLoading: false, 
            error: error.response?.data?.error || 'Failed to register' 
          });
          throw error;
        }
      },

      // Logout function
      logout: () => {
        set({ 
          user: null, 
          token: null, 
          isAuthenticated: false 
        });
        
        // Remove the token from axios default headers
        delete axios.defaults.headers.common['Authorization'];
      }
    }),
    {
      name: 'auth-storage', // unique name for localStorage
      partialize: (state) => ({ user: state.user, token: state.token, isAuthenticated: state.isAuthenticated }),
    }
  )
);

// Initialize axios with the token from storage on app load
if (typeof window !== 'undefined') {
  const storedState = JSON.parse(localStorage.getItem('auth-storage') || '{}');
  if (storedState?.state?.token) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${storedState.state.token}`;
  }
}

// Set default axios base URL
axios.defaults.baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
