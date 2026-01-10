import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';

import { useAuthStore } from '@/store/authStore';
import { useRoomStore } from '@/store/roomStore';

export default function CreateRoomPage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const { createRoom, isLoading, error } = useRoomStore();
  
  const [name, setName] = useState('');
  const [maxPlayers, setMaxPlayers] = useState(7); // Default to 7 players
  const [formError, setFormError] = useState('');
  
  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    router.push('/auth/login');
    return null;
  }
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');
    
    // Validate form
    if (!name.trim()) {
      setFormError('Room name is required');
      return;
    }
    
    try {
      // Create room
      const room = await createRoom(name, maxPlayers);
      
      // Navigate to the room page
      router.push(`/rooms/${room.id}`);
    } catch (err: any) {
      setFormError(err.message || 'Failed to create room');
    }
  };
  
  return (
    <div className="min-h-screen p-4">
      <div className="max-w-md mx-auto">
        <div className="mb-6">
          <Link 
            href="/rooms"
            className="flex items-center text-primary-600 hover:text-primary-700"
          >
            <ArrowLeft className="h-5 w-5 mr-1" />
            <span>Back to Rooms</span>
          </Link>
        </div>
        
        <div className="bg-white dark:bg-gray-800 shadow-md rounded-lg p-6">
          <h1 className="text-2xl font-bold mb-6">Create New Room</h1>
          
          {(error || formError) && (
            <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
              {error || formError}
            </div>
          )}
          
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Room Name
              </label>
              <input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                type="text"
                required
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                placeholder="Enter a room name"
              />
            </div>
            
            <div>
              <label htmlFor="maxPlayers" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Maximum Players
              </label>
              <select
                id="maxPlayers"
                value={maxPlayers}
                onChange={(e) => setMaxPlayers(Number(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-primary-500 focus:border-primary-500"
              >
                <option value={5}>5 players</option>
                <option value={6}>6 players</option>
                <option value={7}>7 players</option>
                <option value={8}>8 players</option>
                <option value={9}>9 players</option>
                <option value={10}>10 players</option>
              </select>
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
                Avalon works best with 7-10 players
              </p>
            </div>
            
            <div className="pt-2">
              <button
                type="submit"
                disabled={isLoading}
                className={`w-full flex justify-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 ${isLoading ? 'opacity-70 cursor-not-allowed' : ''}`}
              >
                {isLoading ? 'Creating...' : 'Create Room'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export const dynamic = 'force-dynamic';
export const runtime = 'client';
