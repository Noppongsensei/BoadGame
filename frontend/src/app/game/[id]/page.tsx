import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Loader, Users, ShieldCheck, ShieldX, ThumbsUp, ThumbsDown } from 'lucide-react';

import { useAuthStore } from '@/store/authStore';
import { useRoomStore } from '@/store/roomStore';
import { useGameStore } from '@/store/gameStore';
import HoldToReveal from '@/components/game/HoldToReveal';

// Role icon mapping
const roleIcons = {
  merlin: '👑', // Crown
  percival: '🔍', // Magnifying Glass
  loyal: '⚔️',  // Crossed Swords
  assassin: '🗡️', // Dagger
  morgana: '🧙‍♀️', // Witch
  mordred: '💀', // Skull
  oberon: '👻', // Ghost
  minion: '👹', // Ogre
};

// Team color mapping
const teamColors = {
  good: 'bg-goodTeam text-white',
  evil: 'bg-evilTeam text-white',
};

export default function GamePage() {
  const router = useRouter();
  const params = useParams();
  const roomId = params?.id as string;
  
  const { user, isAuthenticated } = useAuthStore();
  const { currentRoom } = useRoomStore();
  const { 
    gameState, 
    isLoading, 
    error, 
    getGameState, 
    connectWebSocket,
    selectQuestTeam,
    voteForQuest,
    performQuest,
    assassinateMerlin
  } = useGameStore();
  
  const [selectedPlayers, setSelectedPlayers] = useState<string[]>([]);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  
  // Find current player in game state
  const currentPlayer = gameState?.players.find(p => p.id === user?.id);
  const isLeader = currentPlayer?.is_leader;
  const currentQuest = gameState?.current_quest;
  const isPlayerOnQuest = currentQuest?.selected_players?.includes(user?.id || '');
  
  useEffect(() => {
    // Redirect to login if not authenticated
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    
    if (roomId) {
      // Get initial game state
      getGameState(roomId);
      
      // Connect to WebSocket
      if (user?.token) {
        connectWebSocket(user.token, roomId);
      }
    }
    
    // Set up polling for game state (fallback if WebSocket fails)
    const interval = setInterval(() => {
      if (roomId) {
        getGameState(roomId);
      }
    }, 5000);
    
    return () => clearInterval(interval);
  }, [isAuthenticated, router, roomId, user, getGameState, connectWebSocket]);
  
  // Handle selecting a player for a quest
  const handleSelectPlayer = (playerId: string) => {
    if (selectedPlayers.includes(playerId)) {
      // Remove player if already selected
      setSelectedPlayers(prev => prev.filter(id => id !== playerId));
    } else {
      // Add player if not at max required players
      if (currentQuest && selectedPlayers.length < currentQuest.required_players) {
        setSelectedPlayers(prev => [...prev, playerId]);
      }
    }
  };
  
  // Handle submitting selected players for quest
  const handleSubmitTeam = () => {
    if (!roomId || !currentQuest) return;
    
    if (selectedPlayers.length !== currentQuest.required_players) {
      setErrorMessage(`You must select exactly ${currentQuest.required_players} players`);
      return;
    }
    
    selectQuestTeam(roomId, selectedPlayers);
    setErrorMessage(null);
  };
  
  // Handle voting for a quest
  const handleVote = (approve: boolean) => {
    if (!roomId) return;
    voteForQuest(roomId, approve);
  };
  
  // Handle performing a quest action
  const handleQuestAction = (success: boolean) => {
    if (!roomId) return;
    performQuest(roomId, success);
  };
  
  // Handle assassinating a player (for Assassin role)
  const handleAssassinate = (targetId: string) => {
    if (!roomId) return;
    assassinateMerlin(roomId, targetId);
  };
  
  if (isLoading && !gameState) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader className="h-12 w-12 text-primary-600 animate-spin" />
      </div>
    );
  }
  
  return (
    <div className="min-h-screen pb-24">
      {/* Header */}
      <div className="bg-gray-900 p-4 flex items-center justify-between">
        <Link 
          href={`/rooms/${roomId}`}
          className="text-white flex items-center"
        >
          <ArrowLeft className="h-5 w-5 mr-1" />
          <span>Back</span>
        </Link>
        
        <h1 className="text-lg font-bold text-white">Avalon Game</h1>
        
        <div className="w-8"></div> {/* Spacer for alignment */}
      </div>
      
      {/* Error message */}
      {(error || errorMessage) && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 m-4 rounded">
          {error || errorMessage}
        </div>
      )}
      
      {gameState && (
        <div className="p-4">
          {/* Game info */}
          <div className="mb-6">
            <div className="bg-gray-100 dark:bg-gray-800 rounded-lg p-4 flex justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Round</p>
                <p className="font-bold">{gameState.current_round} / 5</p>
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Phase</p>
                <p className="font-bold capitalize">{gameState.phase.replace('_', ' ')}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Vote Track</p>
                <p className="font-bold">{gameState.vote_track} / 5</p>
              </div>
            </div>
          </div>
          
          {/* Player's role (Hold to Reveal) */}
          <div className="mb-6 flex justify-center">
            <HoldToReveal revealText="Hold to see your role" hiddenText="Release to hide role">
              <div className="text-center">
                <div className="text-4xl mb-2">
                  {currentPlayer?.role && roleIcons[currentPlayer.role as keyof typeof roleIcons]}
                </div>
                <h3 className="text-lg font-bold capitalize mb-1">{currentPlayer?.role}</h3>
                <div className={`inline-block px-3 py-1 rounded-full text-sm ${
                  currentPlayer?.team === 'good' ? 'bg-goodTeam text-white' : 'bg-evilTeam text-white'
                }`}>
                  {currentPlayer?.team === 'good' ? 'Good Team' : 'Evil Team'}
                </div>
              </div>
            </HoldToReveal>
          </div>
          
          {/* Quest tracker */}
          <div className="mb-6">
            <h2 className="text-lg font-semibold mb-2">Quest Tracker</h2>
            <div className="flex justify-between">
              {gameState.quest_tracker.map((quest) => (
                <div 
                  key={quest.round} 
                  className={`w-16 h-16 rounded-full flex items-center justify-center font-bold text-lg ${
                    quest.status === 'success' ? 'bg-goodTeam text-white' :
                    quest.status === 'fail' ? 'bg-evilTeam text-white' :
                    quest.round === gameState.current_round ? 'bg-yellow-400 text-gray-900' :
                    'bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
                  }`}
                >
                  {quest.status === 'success' ? <ShieldCheck className="h-8 w-8" /> :
                   quest.status === 'fail' ? <ShieldX className="h-8 w-8" /> :
                   quest.round}
                </div>
              ))}
            </div>
          </div>
          
          {/* Game phase specific UI */}
          <div className="mb-6">
            {gameState.phase === 'quest_selection' && isLeader && (
              <div>
                <h2 className="text-lg font-semibold mb-2">Select Quest Team</h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                  You are the leader. Select {currentQuest?.required_players} players for the quest.
                </p>
                
                <div className="grid grid-cols-2 gap-2">
                  {gameState.players.map(player => (
                    <div
                      key={player.id}
                      onClick={() => handleSelectPlayer(player.id)}
                      className={`p-3 rounded-lg border-2 ${
                        selectedPlayers.includes(player.id) 
                          ? 'border-primary-600 bg-primary-50 dark:bg-primary-900/20'
                          : 'border-gray-200 dark:border-gray-700'
                      } cursor-pointer`}
                    >
                      <div className="flex items-center">
                        <div className="h-8 w-8 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-gray-700 dark:text-gray-300 font-medium">
                          {player.username.charAt(0).toUpperCase()}
                        </div>
                        <span className="ml-2 font-medium">{player.username}</span>
                      </div>
                    </div>
                  ))}
                </div>
                
                <div className="mt-4">
                  <button
                    onClick={handleSubmitTeam}
                    disabled={selectedPlayers.length !== currentQuest?.required_players}
                    className={`w-full py-3 px-4 rounded-md shadow-sm font-medium text-white ${
                      selectedPlayers.length === currentQuest?.required_players
                        ? 'bg-primary-600 hover:bg-primary-700'
                        : 'bg-gray-400 cursor-not-allowed'
                    }`}
                  >
                    Submit Team Selection
                  </button>
                </div>
              </div>
            )}
            
            {gameState.phase === 'quest_selection' && !isLeader && (
              <div className="text-center py-4">
                <p className="text-gray-600 dark:text-gray-400">
                  Waiting for {gameState.players[gameState.leader]?.username} to select the team for this quest.
                </p>
              </div>
            )}
            
            {gameState.phase === 'quest_voting' && currentQuest && (
              <div>
                <h2 className="text-lg font-semibold mb-2">Vote for Quest</h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                  The leader has selected the following team:
                </p>
                
                <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 mb-4">
                  <div className="flex items-center flex-wrap gap-2">
                    {currentQuest.selected_players?.map(playerId => {
                      const player = gameState.players.find(p => p.id === playerId);
                      return player ? (
                        <div key={player.id} className="bg-white dark:bg-gray-700 px-3 py-1 rounded-lg text-sm font-medium">
                          {player.username}
                        </div>
                      ) : null;
                    })}
                  </div>
                </div>
                
                {/* Only show voting if player hasn't voted yet */}
                {!currentQuest.votes?.[user?.id || ''] && (
                  <div className="thumb-zone">
                    <div className="flex justify-between w-full max-w-xs mx-auto">
                      <button
                        onClick={() => handleVote(true)}
                        className="bg-green-500 text-white rounded-lg p-4 flex flex-col items-center"
                      >
                        <ThumbsUp className="h-10 w-10 mb-1" />
                        <span>Approve</span>
                      </button>
                      
                      <button
                        onClick={() => handleVote(false)}
                        className="bg-red-500 text-white rounded-lg p-4 flex flex-col items-center"
                      >
                        <ThumbsDown className="h-10 w-10 mb-1" />
                        <span>Reject</span>
                      </button>
                    </div>
                  </div>
                )}
                
                {currentQuest.votes?.[user?.id || ''] !== undefined && (
                  <div className="text-center py-4">
                    <p className="text-gray-600 dark:text-gray-400">
                      You have voted. Waiting for all players to vote.
                    </p>
                    <p className="font-medium mt-2">
                      Your vote: {currentQuest.votes[user?.id || ''] ? 'Approve' : 'Reject'}
                    </p>
                  </div>
                )}
              </div>
            )}
            
            {gameState.phase === 'quest_performance' && isPlayerOnQuest && (
              <div>
                <h2 className="text-lg font-semibold mb-2">Perform Quest</h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                  You are on the quest. Choose whether to make the quest succeed or fail.
                </p>
                
                <div className="thumb-zone">
                  <div className="flex justify-between w-full max-w-xs mx-auto">
                    <button
                      onClick={() => handleQuestAction(true)}
                      className="bg-green-500 text-white rounded-lg p-4 flex flex-col items-center"
                    >
                      <ShieldCheck className="h-10 w-10 mb-1" />
                      <span>Success</span>
                    </button>
                    
                    {/* Only evil players can choose to fail the quest */}
                    {currentPlayer?.team === 'evil' && (
                      <button
                        onClick={() => handleQuestAction(false)}
                        className="bg-red-500 text-white rounded-lg p-4 flex flex-col items-center"
                      >
                        <ShieldX className="h-10 w-10 mb-1" />
                        <span>Fail</span>
                      </button>
                    )}
                  </div>
                </div>
              </div>
            )}
            
            {gameState.phase === 'quest_performance' && !isPlayerOnQuest && (
              <div className="text-center py-4">
                <p className="text-gray-600 dark:text-gray-400">
                  Waiting for the quest team to complete their mission.
                </p>
              </div>
            )}
            
            {gameState.phase === 'assassination' && currentPlayer?.role === 'assassin' && (
              <div>
                <h2 className="text-lg font-semibold mb-2">Assassinate Merlin</h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                  As the Assassin, you must identify and assassinate Merlin to win the game.
                </p>
                
                <div className="grid grid-cols-2 gap-2">
                  {gameState.players
                    // Only show good team players as potential targets
                    .filter(player => player.team === 'good')
                    .map(player => (
                      <div
                        key={player.id}
                        onClick={() => handleAssassinate(player.id)}
                        className="p-3 rounded-lg border-2 border-gray-200 dark:border-gray-700 cursor-pointer hover:border-evilTeam hover:bg-red-50 dark:hover:bg-red-900/20"
                      >
                        <div className="flex items-center">
                          <div className="h-8 w-8 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-gray-700 dark:text-gray-300 font-medium">
                            {player.username.charAt(0).toUpperCase()}
                          </div>
                          <span className="ml-2 font-medium">{player.username}</span>
                        </div>
                      </div>
                    ))
                  }
                </div>
              </div>
            )}
            
            {gameState.phase === 'assassination' && currentPlayer?.role !== 'assassin' && (
              <div className="text-center py-4">
                <p className="text-gray-600 dark:text-gray-400">
                  Waiting for the Assassin to try to identify Merlin.
                </p>
              </div>
            )}
            
            {gameState.phase === 'game_over' && (
              <div className="text-center py-4">
                <h2 className="text-2xl font-bold mb-2">
                  Game Over
                </h2>
                <p className={`text-xl font-medium ${
                  gameState.winning_team === 'good' ? 'text-goodTeam' : 'text-evilTeam'
                }`}>
                  {gameState.winning_team === 'good' ? 'Good Team Wins!' : 'Evil Team Wins!'}
                </p>
                
                <div className="mt-6">
                  <h3 className="font-medium mb-2">Player Roles:</h3>
                  <div className="grid grid-cols-2 gap-2">
                    {gameState.players.map(player => (
                      <div key={player.id} className={`p-3 rounded-lg ${
                        player.team === 'good' ? 'bg-blue-50 dark:bg-blue-900/20' : 'bg-red-50 dark:bg-red-900/20'
                      }`}>
                        <p className="font-medium">{player.username}</p>
                        <p className="text-sm capitalize flex items-center">
                          <span className="mr-1">
                            {roleIcons[player.role as keyof typeof roleIcons]}
                          </span>
                          {player.role}
                        </p>
                      </div>
                    ))}
                  </div>
                  
                  <div className="mt-6">
                    <Link
                      href={`/rooms/${roomId}`}
                      className="w-full block text-center py-3 px-4 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700"
                    >
                      Back to Room
                    </Link>
                  </div>
                </div>
              </div>
            )}
          </div>
          
          {/* Players list */}
          <div className="mb-6">
            <h2 className="text-lg font-semibold mb-2">Players</h2>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden">
              {gameState.players.map(player => (
                <div 
                  key={player.id}
                  className={`p-3 flex items-center justify-between ${
                    player.is_leader ? 'bg-yellow-50 dark:bg-yellow-900/20' : ''
                  }`}
                >
                  <div className="flex items-center">
                    <div className="h-8 w-8 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-gray-700 dark:text-gray-300 font-medium">
                      {player.username.charAt(0).toUpperCase()}
                    </div>
                    <span className="ml-2 font-medium">{player.username}</span>
                  </div>
                  {player.is_leader && (
                    <div className="bg-yellow-100 dark:bg-yellow-800 px-2 py-1 rounded-full text-xs font-medium text-yellow-800 dark:text-yellow-200">
                      Leader
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export const dynamic = 'force-dynamic';
export const runtime = 'client';
