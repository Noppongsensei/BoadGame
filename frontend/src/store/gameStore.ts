import { create } from 'zustand';
import axios from 'axios';
import { useRoomStore } from './roomStore';

// Define game state types
interface Player {
  id: string;
  username: string;
  role?: string;
  team?: string;
  is_leader?: boolean;
}

interface Quest {
  round: number;
  required_players: number;
  selected_players?: string[];
  votes?: Record<string, boolean>;
  results?: boolean[];
  status: string;
}

interface GameState {
  players: Player[];
  current_round: number;
  leader: number;
  quest_tracker: Quest[];
  current_quest?: Quest;
  vote_track: number;
  game_over: boolean;
  winning_team?: string;
  winners?: string[];
  phase: string;
}

// Define the WebSocket message type
export interface WSMessage {
  type: string;
  room_id?: string;
  user_id?: string;
  payload?: any;
}

// Define the game store state
interface GameStoreState {
  gameState: GameState | null;
  isLoading: boolean;
  error: string | null;
  socket: WebSocket | null;
  isConnected: boolean;
  // Game actions
  initGame: (roomId: string, gameType: string) => Promise<void>;
  getGameState: (roomId: string) => Promise<void>;
  getGameHistory: (roomId: string) => Promise<any>;
  // WebSocket actions
  connectWebSocket: (token: string, roomId?: string) => void;
  disconnectWebSocket: () => void;
  sendMessage: (message: WSMessage) => void;
  // Game logic actions
  selectQuestTeam: (roomId: string, selectedPlayers: string[]) => void;
  voteForQuest: (roomId: string, approve: boolean) => void;
  performQuest: (roomId: string, success: boolean) => void;
  assassinateMerlin: (roomId: string, targetId: string) => void;
}

// Create the game store
export const useGameStore = create<GameStoreState>((set, get) => ({
  gameState: null,
  isLoading: false,
  error: null,
  socket: null,
  isConnected: false,

  // Initialize a game
  initGame: async (roomId: string, gameType: string = 'avalon') => {
    try {
      set({ isLoading: true, error: null });
      await axios.post(`/api/games/${roomId}/init`, { game_type: gameType });
      set({ isLoading: false });
    } catch (error: any) {
      set({
        isLoading: false,
        error: error.response?.data?.error || 'Failed to initialize game'
      });
      throw error;
    }
  },

  // Get the current game state
  getGameState: async (roomId: string) => {
    try {
      set({ isLoading: true, error: null });
      const response = await axios.get(`/api/games/${roomId}/state`);
      set({ gameState: response.data, isLoading: false });
    } catch (error: any) {
      set({
        isLoading: false,
        error: error.response?.data?.error || 'Failed to get game state'
      });
      throw error;
    }
  },

  // Get game history
  getGameHistory: async (roomId: string) => {
    try {
      set({ isLoading: true, error: null });
      const response = await axios.get(`/api/games/${roomId}/history`);
      set({ isLoading: false });
      return response.data.history;
    } catch (error: any) {
      set({
        isLoading: false,
        error: error.response?.data?.error || 'Failed to get game history'
      });
      throw error;
    }
  },

  // Connect to WebSocket
  connectWebSocket: (token: string, roomId?: string) => {
    // Close existing connection if any
    const currentSocket = get().socket;
    if (currentSocket && currentSocket.readyState === WebSocket.OPEN) {
      currentSocket.close();
    }

    // Build WebSocket URL
    let wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws?token=${token}`;
    if (roomId) {
      wsUrl += `&room_id=${roomId}`;
    }

    // Create new WebSocket connection
    const socket = new WebSocket(wsUrl);

    // Set up event handlers
    socket.onopen = () => {
      set({ socket, isConnected: true, error: null });
      console.log('WebSocket connected');
    };

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        
        // Handle different message types
        switch (message.type) {
          case 'game.state_update':
            // Update game state when server sends an update
            set({ gameState: message.payload });
            break;
            
          case 'room.player_joined':
          case 'room.player_left':
          case 'room.updated':
            // Fetch room data to update player list or room status
            if (message.room_id) {
              // Use roomStore to update currentRoom
              const roomStore = useRoomStore.getState();
              roomStore.fetchRoom(message.room_id);
              console.log(`Room updated: ${message.type}`);
            }
            break;
            
          case 'room.game_started':
            // Handle game start notification (redirect non-host players)
            if (message.room_id) {
              const roomStore = useRoomStore.getState();
              roomStore.fetchRoom(message.room_id);
              
              // If this is sent to all players, they can navigate to game
              if (typeof window !== 'undefined') {
                window.location.href = `/game/${message.room_id}`;
              }
            }
            break;
          
          case 'system.error':
            set({ error: message.payload?.message || 'Server error' });
            break;
            
          // Add more message type handlers as needed
        }
      } catch (error) {
        console.error('Error processing WebSocket message:', error);
      }
    };

    socket.onclose = () => {
      set({ isConnected: false });
      console.log('WebSocket disconnected');
    };

    socket.onerror = (error) => {
      set({ error: 'WebSocket error' });
      console.error('WebSocket error:', error);
    };

    set({ socket });
  },

  // Disconnect WebSocket
  disconnectWebSocket: () => {
    const { socket } = get();
    if (socket) {
      socket.close();
      set({ socket: null, isConnected: false });
    }
  },

  // Send a message through WebSocket
  sendMessage: (message: WSMessage) => {
    const { socket, isConnected } = get();
    if (socket && isConnected) {
      socket.send(JSON.stringify(message));
    } else {
      set({ error: 'WebSocket not connected' });
    }
  },

  // Select players for a quest
  selectQuestTeam: (roomId: string, selectedPlayers: string[]) => {
    const message: WSMessage = {
      type: 'game.action',
      room_id: roomId,
      payload: {
        type: 'select_quest',
        selected_players: selectedPlayers
      }
    };
    get().sendMessage(message);
  },

  // Vote for a quest
  voteForQuest: (roomId: string, approve: boolean) => {
    const message: WSMessage = {
      type: 'game.action',
      room_id: roomId,
      payload: {
        type: 'vote_quest',
        approve
      }
    };
    get().sendMessage(message);
  },

  // Perform a quest action
  performQuest: (roomId: string, success: boolean) => {
    const message: WSMessage = {
      type: 'game.action',
      room_id: roomId,
      payload: {
        type: 'perform_quest',
        success
      }
    };
    get().sendMessage(message);
  },

  // Assassinate Merlin
  assassinateMerlin: (roomId: string, targetId: string) => {
    const message: WSMessage = {
      type: 'game.action',
      room_id: roomId,
      payload: {
        type: 'assassinate_merlin',
        target_id: targetId
      }
    };
    get().sendMessage(message);
  }
}));
