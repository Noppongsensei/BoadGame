import { create } from 'zustand';
import axios from 'axios';

// Define room and player types
interface Player {
  id: string;
  username: string;
}

interface Room {
  id: string;
  name: string;
  host_id: string;
  status: string;
  max_players: number;
  players?: Player[];
  created_at: string;
  updated_at: string;
}

// Define the room state
interface RoomState {
  rooms: Room[];
  currentRoom: Room | null;
  isLoading: boolean;
  error: string | null;
  fetchRooms: () => Promise<void>;
  fetchOpenRooms: () => Promise<void>;
  createRoom: (name: string, maxPlayers: number) => Promise<Room>;
  joinRoom: (roomId: string) => Promise<void>;
  leaveRoom: (roomId: string) => Promise<void>;
  startGame: (roomId: string) => Promise<void>;
  fetchRoom: (roomId: string) => Promise<void>;
}

// Create the room store
export const useRoomStore = create<RoomState>((set, get) => ({
  rooms: [],
  currentRoom: null,
  isLoading: false,
  error: null,

  // Fetch all rooms
  fetchRooms: async () => {
    try {
      set({ isLoading: true, error: null });
      const response = await axios.get('/api/rooms');
      set({ rooms: response.data.rooms, isLoading: false });
    } catch (error: any) {
      set({ 
        isLoading: false, 
        error: error.response?.data?.error || 'Failed to fetch rooms' 
      });
    }
  },

  // Fetch open rooms
  fetchOpenRooms: async () => {
    try {
      set({ isLoading: true, error: null });
      const response = await axios.get('/api/rooms/open');
      set({ rooms: response.data.rooms, isLoading: false });
    } catch (error: any) {
      set({ 
        isLoading: false, 
        error: error.response?.data?.error || 'Failed to fetch open rooms' 
      });
    }
  },

  // Create a new room
  createRoom: async (name: string, maxPlayers: number) => {
    try {
      set({ isLoading: true, error: null });
      const response = await axios.post('/api/rooms', { name, max_players: maxPlayers });
      const newRoom = response.data;
      
      set(state => ({ 
        rooms: [...state.rooms, newRoom],
        currentRoom: newRoom,
        isLoading: false 
      }));
      
      return newRoom;
    } catch (error: any) {
      set({ 
        isLoading: false, 
        error: error.response?.data?.error || 'Failed to create room' 
      });
      throw error;
    }
  },

  // Join a room
  joinRoom: async (roomId: string) => {
    try {
      set({ isLoading: true, error: null });
      const response = await axios.post(`/api/rooms/${roomId}/join`);
      set({ currentRoom: response.data, isLoading: false });
    } catch (error: any) {
      set({ 
        isLoading: false, 
        error: error.response?.data?.error || 'Failed to join room' 
      });
      throw error;
    }
  },

  // Leave a room
  leaveRoom: async (roomId: string) => {
    try {
      set({ isLoading: true, error: null });
      await axios.post(`/api/rooms/${roomId}/leave`);
      set({ currentRoom: null, isLoading: false });
    } catch (error: any) {
      set({ 
        isLoading: false, 
        error: error.response?.data?.error || 'Failed to leave room' 
      });
      throw error;
    }
  },

  // Start a game
  startGame: async (roomId: string) => {
    try {
      set({ isLoading: true, error: null });
      const response = await axios.post(`/api/rooms/${roomId}/start`);
      set({ currentRoom: response.data, isLoading: false });
    } catch (error: any) {
      set({ 
        isLoading: false, 
        error: error.response?.data?.error || 'Failed to start game' 
      });
      throw error;
    }
  },

  // Fetch a specific room
  fetchRoom: async (roomId: string) => {
    try {
      set({ isLoading: true, error: null });
      const response = await axios.get(`/api/rooms/${roomId}`);
      set({ currentRoom: response.data, isLoading: false });
    } catch (error: any) {
      set({ 
        isLoading: false, 
        error: error.response?.data?.error || 'Failed to fetch room' 
      });
      throw error;
    }
  }
}));
