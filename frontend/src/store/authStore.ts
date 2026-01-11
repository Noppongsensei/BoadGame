import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import apiClient from '../lib/axios';

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
          const response = await apiClient.post('/api/auth/login', { username, password });
          const { user, token } = response.data;

          set({
            user,
            token,
            isAuthenticated: true,
            isLoading: false
          });

          // Store token in localStorage for interceptor
          localStorage.setItem('token', token);
        } catch (error: unknown) {
          const errorMessage = (error as any)?.response?.data?.error || 'Failed to login';
          set({
            isLoading: false,
            error: errorMessage
          });
          throw error;
        }
      },

      // Register function
      register: async (username: string, password: string) => {
        try {
          set({ isLoading: true, error: null });
          const response = await apiClient.post('/api/auth/register', { username, password });
          const { user, token } = response.data;

          set({
            user,
            token,
            isAuthenticated: true,
            isLoading: false
          });

          // Store token in localStorage for interceptor
          localStorage.setItem('token', token);
        } catch (error: unknown) {
          const errorMessage = (error as any)?.response?.data?.error || 'Failed to register';
          set({
            isLoading: false,
            error: errorMessage
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

        // Remove the token from localStorage
        localStorage.removeItem('token');
      }
    }),
    {
      name: 'auth-storage', // unique name for localStorage
      partialize: (state) => ({ user: state.user, token: state.token, isAuthenticated: state.isAuthenticated }),
    }
  )
);

// Initialize token from storage on app load
if (typeof window !== 'undefined') {
  const storedState = JSON.parse(localStorage.getItem('auth-storage') || '{}');
  if (storedState?.state?.token) {
    localStorage.setItem('token', storedState.state.token);
  }
}
