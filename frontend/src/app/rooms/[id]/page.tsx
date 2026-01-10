"use client";

import { useEffect, useState } from 'react';
import { useParams, useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Users, Play, Loader, UserPlus } from 'lucide-react';

import { useAuthStore } from '@/store/authStore';
import { useRoomStore } from '@/store/roomStore';
import { useGameStore, type WSMessage } from '@/store/gameStore';

export default function RoomPage() {
  const router = useRouter();
  const params = useParams();
  const pathname = usePathname();
  const roomId = params?.id as string;
  
  const { user, token, isAuthenticated } = useAuthStore();
  const { currentRoom, isLoading: isRoomLoading, error: roomError, fetchRoom, joinRoom, leaveRoom, startGame } = useRoomStore();
  const { initGame, connectWebSocket } = useGameStore();
  
  const [isJoining, setIsJoining] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Check if user is in the room
  const isInRoom = currentRoom?.players?.some(p => p.id === user?.id);
  // Check if user is the host
  const isHost = currentRoom?.host_id === user?.id;
  // Check if game is in progress
  const isGameInProgress = currentRoom?.status === 'playing';
  // Check if the room has enough players
  const hasEnoughPlayers = currentRoom?.players && currentRoom.players.length >= 5;
  // Check if room is full
  const isRoomFull = currentRoom?.players && currentRoom.max_players && 
                     currentRoom.players.length >= currentRoom.max_players;

  useEffect(() => {
    // Redirect to login if not authenticated
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    
    // Fetch room data
    if (roomId) {
      fetchRoom(roomId);
      
      // Connect to WebSocket if authenticated
      if (token) {
        connectWebSocket(token, roomId);
      }

      // Save room ID to sessionStorage to handle page refresh
      sessionStorage.setItem('currentRoomId', roomId);
    } else if (typeof window !== 'undefined') {
      // If roomId is missing but we have it in sessionStorage, restore it
      const storedRoomId = sessionStorage.getItem('currentRoomId');
      if (storedRoomId && pathname === '/rooms') {
        router.replace(`/rooms/${storedRoomId}`);
      }
    }
  }, [isAuthenticated, router, roomId, fetchRoom, token, connectWebSocket, pathname]);
  
  // Handle joining room
  const handleJoinRoom = async () => {
    if (!roomId) return;
    
    setIsJoining(true);
    setError(null);
    
    try {
      await joinRoom(roomId);
    } catch (err: any) {
      setError(err.message || 'Failed to join room');
    } finally {
      setIsJoining(false);
    }
  };
  
  // Handle leaving room
  const handleLeaveRoom = async () => {
    if (!roomId) return;
    
    try {
      await leaveRoom(roomId);
      router.push('/rooms');
    } catch (err: any) {
      setError(err.message || 'Failed to leave room');
    }
  };
  
  // Handle starting game
  const handleStartGame = async () => {
    if (!roomId) return;
    
    setIsStarting(true);
    setError(null);
    
    try {
      console.log('Starting game for room:', roomId);
      // Start the game in the room
      await startGame(roomId);
      console.log('Game started in backend');

      // Initialize the game session
      await initGame(roomId, 'avalon');
      console.log('Game session initialized');
      
      // Send WebSocket notification that game started (for all players)
      if (token) {
        console.log('Sending WebSocket notification');
        const gameStore = useGameStore.getState();
        gameStore.sendMessage({
          type: 'room.game_started',
          room_id: roomId,
          payload: { status: 'playing' }
        });
      }
      
      // Navigate to game page - use direct navigation instead of router
      // This ensures we definitely navigate even if there's an issue with Next.js router
      console.log('Redirecting to game page');
      
      // Short timeout to ensure WebSocket message is sent before navigation
      setTimeout(() => {
        if (typeof window !== 'undefined') {
          window.location.href = `/game/${roomId}`;
        }
      }, 500);
    } catch (err: any) {
      console.error('Error starting game:', err);
      setError(err.message || 'Failed to start game');
      setIsStarting(false);
    }
  };
  
  // If loading, show loading spinner
  if (isRoomLoading) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-2 border-primary-300 border-t-primary-600"></div>
      </div>
    );
  }
  
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-6 md:p-8">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <Link 
            href="/rooms"
            className="flex items-center text-primary-600 hover:text-primary-700 transition-colors duration-200 font-medium"
          >
            <ArrowLeft className="h-5 w-5 mr-2" />
            <span>Back to Rooms</span>
          </Link>
        </div>
        
        {(error || roomError) && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-lg shadow-sm mb-6">
            {error || roomError}
          </div>
        )}
        
        {currentRoom && (
          <div className="bg-white dark:bg-gray-800 shadow-sm rounded-xl overflow-hidden border border-gray-100 dark:border-gray-700">
            {/* Room header */}
            <div className="p-8 border-b border-gray-100 dark:border-gray-700">
              <div className="flex justify-between items-start mb-4">
                <h1 className="text-3xl font-medium tracking-tight text-gray-900 dark:text-white">{currentRoom.name}</h1>
                <span className={`px-3 py-1 rounded-full text-xs font-medium ${currentRoom.status === 'playing' ? 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400' : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'}`}>
                  {currentRoom.status}
                </span>
              </div>
              
              <div className="flex items-center text-sm text-gray-500 dark:text-gray-400">
                <Users className="h-4 w-4 mr-2" />
                <span className="font-medium">
                  {currentRoom.players?.length || 0} / {currentRoom.max_players} players
                </span>
                <span className="mx-2 text-gray-300 dark:text-gray-600">•</span>
                <span>Created by {currentRoom.players?.find(p => p.id === currentRoom.host_id)?.username || 'Unknown'}</span>
              </div>
            </div>
            
            {/* Player list */}
            <div className="p-8">
              <h2 className="text-xl font-medium mb-6 text-gray-900 dark:text-gray-100">Players</h2>
              
              <div className="space-y-3">
                {currentRoom.players?.map(player => (
                  <div 
                    key={player.id}
                    className={`flex items-center justify-between p-4 rounded-lg transition-all ${player.id === user?.id ? 'bg-primary-50 dark:bg-primary-900/20 border border-primary-100 dark:border-primary-800' : 'bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700'}`}
                  >
                    <div className="flex items-center gap-3">
                      <div className={`h-10 w-10 rounded-full flex items-center justify-center text-white font-medium ${player.id === currentRoom.host_id ? 'bg-yellow-500 dark:bg-yellow-600' : 'bg-primary-500 dark:bg-primary-600'}`}>
                        {player.username.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <span className="font-medium text-gray-900 dark:text-gray-100">{player.username}</span>
                        {player.id === user?.id && (
                          <span className="ml-2 text-xs text-gray-500 dark:text-gray-400">(You)</span>
                        )}
                      </div>
                    </div>
                    {player.id === currentRoom.host_id && (
                      <span className="text-xs px-3 py-1 rounded-full bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-400">
                        Host
                      </span>
                    )}
                  </div>
                ))}
              </div>
              
              {/* Actions */}
              <div className="mt-10 space-y-4">
                {/* Join button for non-members */}
                {!isInRoom && !isGameInProgress && !isRoomFull && (
                  <button
                    onClick={handleJoinRoom}
                    disabled={isJoining}
                    className={`w-full flex justify-center items-center py-3.5 px-6 rounded-lg shadow-sm text-base font-medium text-white bg-primary-500 hover:bg-primary-600 transition-all duration-200 ${
                      isJoining ? 'opacity-70 cursor-not-allowed' : 'hover:shadow'
                    }`}
                  >
                    {isJoining ? (
                      <>
                        <div className="h-5 w-5 mr-3 rounded-full border-2 border-white border-t-transparent animate-spin"></div>
                        Joining...
                      </>
                    ) : (
                      <>
                        <UserPlus className="h-5 w-5 mr-3" />
                        Join Room
                      </>
                    )}
                  </button>
                )}
                
                {/* Start game button for host */}
                {isHost && !isGameInProgress && hasEnoughPlayers && (
                  <button
                    onClick={handleStartGame}
                    disabled={isStarting}
                    className={`w-full flex justify-center items-center py-3.5 px-6 rounded-lg shadow-sm text-base font-medium text-white bg-green-500 hover:bg-green-600 transition-all duration-200 ${
                      isStarting ? 'opacity-70 cursor-not-allowed' : 'hover:shadow'
                    }`}
                  >
                    {isStarting ? (
                      <>
                        <div className="h-5 w-5 mr-3 rounded-full border-2 border-white border-t-transparent animate-spin"></div>
                        Starting Game...
                      </>
                    ) : (
                      <>
                        <Play className="h-5 w-5 mr-3" />
                        Start Game
                      </>
                    )}
                  </button>
                )}
                
                {/* Join game in progress button */}
                {isInRoom && isGameInProgress && (
                  <Link
                    href={`/game/${roomId}`}
                    className="w-full flex justify-center items-center py-3.5 px-6 rounded-lg shadow-sm text-base font-medium text-white bg-orange-500 hover:bg-orange-600 transition-all duration-200 hover:shadow"
                  >
                    <Play className="h-5 w-5 mr-3" />
                    Join Game in Progress
                  </Link>
                )}
                
                {/* Leave room button for members */}
                {isInRoom && !isGameInProgress && (
                  <button
                    onClick={handleLeaveRoom}
                    className="w-full flex justify-center items-center py-3.5 px-6 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm text-base font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 transition-all duration-200"
                  >
                    Leave Room
                  </button>
                )}
                
                {/* Message for full rooms */}
                {!isInRoom && isRoomFull && (
                  <div className="mt-6 bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-800 rounded-lg px-4 py-3 text-center text-red-700 dark:text-red-400">
                    <p>This room is full and cannot accept more players.</p>
                  </div>
                )}
                
                {/* Message for not enough players */}
                {isHost && !isGameInProgress && !hasEnoughPlayers && (
                  <div className="mt-6 bg-amber-50 dark:bg-amber-900/20 border border-amber-100 dark:border-amber-800 rounded-lg px-4 py-3 text-center text-amber-700 dark:text-amber-400">
                    <p>Need at least 5 players to start the game</p>
                    <p className="text-sm mt-1 text-amber-600 dark:text-amber-500">{5 - (currentRoom.players?.length || 0)} more player{(5 - (currentRoom.players?.length || 0)) !== 1 ? 's' : ''} needed</p>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export const dynamic = 'force-dynamic';
