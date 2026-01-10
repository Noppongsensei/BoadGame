import { useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Users, PlusCircle, ArrowRight } from 'lucide-react';

import { useAuthStore } from '@/store/authStore';
import { useRoomStore } from '@/store/roomStore';

export default function RoomsPage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const { rooms, isLoading, error, fetchOpenRooms } = useRoomStore();
  
  useEffect(() => {
    // Redirect to login if not authenticated
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    
    // Fetch open rooms
    fetchOpenRooms();
  }, [isAuthenticated, router, fetchOpenRooms]);
  
  return (
    <div className="min-h-screen p-4">
      <div className="max-w-2xl mx-auto">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold">Game Rooms</h1>
          <Link 
            href="/rooms/create"
            className="flex items-center text-primary-600 hover:text-primary-700"
          >
            <PlusCircle className="h-5 w-5 mr-1" />
            <span>Create Room</span>
          </Link>
        </div>
        
        {isLoading ? (
          <div className="flex justify-center p-8">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-600"></div>
          </div>
        ) : error ? (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        ) : rooms.length === 0 ? (
          <div className="bg-gray-100 dark:bg-gray-800 p-8 rounded-lg text-center">
            <p className="text-gray-600 dark:text-gray-300 mb-4">No open game rooms found.</p>
            <Link 
              href="/rooms/create"
              className="inline-flex items-center px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
            >
              Create a Room
              <ArrowRight className="h-4 w-4 ml-2" />
            </Link>
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2">
            {rooms.map((room) => (
              <div 
                key={room.id}
                className="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden"
              >
                <div className="p-4">
                  <h2 className="text-lg font-medium">{room.name}</h2>
                  <div className="flex items-center text-sm text-gray-500 dark:text-gray-400 mt-2">
                    <Users className="h-4 w-4 mr-1" />
                    <span>{room.players?.length || 0} / {room.max_players} players</span>
                  </div>
                </div>
                <div className="bg-gray-50 dark:bg-gray-700 px-4 py-3 flex justify-between items-center">
                  <span className="text-sm text-gray-600 dark:text-gray-300">
                    Status: <span className="font-medium capitalize">{room.status}</span>
                  </span>
                  <Link 
                    href={`/rooms/${room.id}`}
                    className="inline-flex items-center text-sm text-primary-600 hover:text-primary-700 font-medium"
                  >
                    Join Room
                    <ArrowRight className="h-4 w-4 ml-1" />
                  </Link>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// Convert page to client component
export const dynamic = 'force-dynamic';
export const runtime = 'client';
