import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Users, Play, Loader, UserPlus } from 'lucide-react';

import { useAuthStore } from '@/store/authStore';
import { useRoomStore } from '@/store/roomStore';
import { useGameStore } from '@/store/gameStore';

export default function RoomPage() {
  const router = useRouter();
  const params = useParams();
  const roomId = params?.id as string;
  
  const { user, isAuthenticated } = useAuthStore();
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
      if (user?.token) {
        connectWebSocket(user.token, roomId);
      }
    }
  }, [isAuthenticated, router, roomId, fetchRoom, user, connectWebSocket]);
  
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
      // Start the game in the room
      await startGame(roomId);
      // Initialize the game session
      await initGame(roomId, 'avalon');
      // Navigate to game page
      router.push(`/game/${roomId}`);
    } catch (err: any) {
      setError(err.message || 'Failed to start game');
      setIsStarting(false);
    }
  };
  
  // If loading, show loading spinner
  if (isRoomLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader className="h-12 w-12 text-primary-600 animate-spin" />
      </div>
    );
  }
  
  return (
    <div className="min-h-screen p-4">
      <div className="max-w-2xl mx-auto">
        <div className="mb-6">
          <Link 
            href="/rooms"
            className="flex items-center text-primary-600 hover:text-primary-700"
          >
            <ArrowLeft className="h-5 w-5 mr-1" />
            <span>Back to Rooms</span>
          </Link>
        </div>
        
        {(error || roomError) && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
            {error || roomError}
          </div>
        )}
        
        {currentRoom && (
          <div className="bg-white dark:bg-gray-800 shadow-md rounded-lg overflow-hidden">
            {/* Room header */}
            <div className="p-6 border-b border-gray-200 dark:border-gray-700">
              <h1 className="text-2xl font-bold">{currentRoom.name}</h1>
              <div className="flex items-center text-sm text-gray-500 dark:text-gray-400 mt-2">
                <Users className="h-4 w-4 mr-1" />
                <span>
                  {currentRoom.players?.length || 0} / {currentRoom.max_players} players
                </span>
                <span className="mx-2">•</span>
                <span className="capitalize">{currentRoom.status}</span>
              </div>
            </div>
            
            {/* Player list */}
            <div className="p-6">
              <h2 className="text-lg font-medium mb-4">Players</h2>
              
              <div className="space-y-2">
                {currentRoom.players?.map(player => (
                  <div 
                    key={player.id}
                    className={`flex items-center justify-between p-3 rounded-lg ${
                      player.id === currentRoom.host_id 
                        ? 'bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800'
                        : 'bg-gray-50 dark:bg-gray-700/50'
                    }`}
                  >
                    <div className="flex items-center">
                      <div className="h-8 w-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center text-primary-700 dark:text-primary-300 font-medium">
                        {player.username.charAt(0).toUpperCase()}
                      </div>
                      <span className="ml-3 font-medium">{player.username}</span>
                    </div>
                    {player.id === currentRoom.host_id && (
                      <span className="text-xs px-2 py-1 rounded-full bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200">
                        Host
                      </span>
                    )}
                  </div>
                ))}
              </div>
              
              {/* Actions */}
              <div className="mt-8 space-y-4">
                {/* Join button for non-members */}
                {!isInRoom && !isGameInProgress && !isRoomFull && (
                  <button
                    onClick={handleJoinRoom}
                    disabled={isJoining}
                    className={`w-full flex justify-center items-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 ${
                      isJoining ? 'opacity-70 cursor-not-allowed' : ''
                    }`}
                  >
                    {isJoining ? (
                      <>
                        <Loader className="h-4 w-4 animate-spin mr-2" />
                        Joining...
                      </>
                    ) : (
                      <>
                        <UserPlus className="h-4 w-4 mr-2" />
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
                    className={`w-full flex justify-center items-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 ${
                      isStarting ? 'opacity-70 cursor-not-allowed' : ''
                    }`}
                  >
                    {isStarting ? (
                      <>
                        <Loader className="h-4 w-4 animate-spin mr-2" />
                        Starting Game...
                      </>
                    ) : (
                      <>
                        <Play className="h-4 w-4 mr-2" />
                        Start Game
                      </>
                    )}
                  </button>
                )}
                
                {/* Join game in progress button */}
                {isInRoom && isGameInProgress && (
                  <Link
                    href={`/game/${roomId}`}
                    className="w-full flex justify-center items-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
                  >
                    <Play className="h-4 w-4 mr-2" />
                    Join Game in Progress
                  </Link>
                )}
                
                {/* Leave room button for members */}
                {isInRoom && !isGameInProgress && (
                  <button
                    onClick={handleLeaveRoom}
                    className="w-full flex justify-center items-center py-3 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
                  >
                    Leave Room
                  </button>
                )}
                
                {/* Message for full rooms */}
                {!isInRoom && isRoomFull && (
                  <div className="text-center text-red-500 p-2">
                    This room is full.
                  </div>
                )}
                
                {/* Message for not enough players */}
                {isHost && !isGameInProgress && !hasEnoughPlayers && (
                  <div className="text-center text-amber-500 p-2">
                    Need at least 5 players to start the game.
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
export const runtime = 'client';
