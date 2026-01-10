"use client";

import { useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Users, PlusCircle, ArrowRight, Play } from 'lucide-react';

import { useAuthStore } from '@/store/authStore';
import { useRoomStore } from '@/store/roomStore';

export default function RoomsPage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const { rooms, isLoading, error, fetchRooms } = useRoomStore();
  const safeRooms = Array.isArray(rooms) ? rooms : [];
  
  useEffect(() => {
    // Redirect to login if not authenticated
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    
    // Fetch all rooms (both open and playing)
    fetchRooms();
  }, [isAuthenticated, router, fetchRooms]);
  
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-6 md:p-8">
      <div className="max-w-4xl mx-auto">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
          <h1 className="text-3xl font-medium tracking-tight text-gray-900 dark:text-gray-50">Game Rooms</h1>
          <Link 
            href="/rooms/create"
            className="btn btn-primary flex items-center justify-center gap-2 px-4 py-2 rounded-lg shadow-sm hover:shadow transition-all duration-200"
          >
            <PlusCircle className="h-5 w-5" />
            <span>Create Room</span>
          </Link>
        </div>
        
        {isLoading ? (
          <div className="flex justify-center p-12">
            <div className="animate-spin rounded-full h-10 w-10 border-2 border-primary-300 border-t-primary-600"></div>
          </div>
        ) : error ? (
          <div className="bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-lg shadow-sm">
            {error}
          </div>
        ) : safeRooms.length === 0 ? (
          <div className="bg-white dark:bg-gray-800 p-12 rounded-xl shadow-sm text-center">
            <div className="max-w-md mx-auto">
              <p className="text-gray-600 dark:text-gray-300 mb-6 text-lg">No game rooms found. Start a new adventure!</p>
              <Link 
                href="/rooms/create"
                className="btn btn-primary inline-flex items-center px-6 py-3 rounded-lg shadow-sm hover:shadow transition-all duration-200"
              >
                Create a Room
                <ArrowRight className="h-5 w-5 ml-2" />
              </Link>
            </div>
          </div>
        ) : (
          <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
            {safeRooms.map((room) => (
              <Link 
                key={room.id} 
                href={`/rooms/${room.id}`}
                className="card card-hover bg-white dark:bg-gray-800 rounded-xl overflow-hidden border border-gray-100 dark:border-gray-700 shadow-sm hover:shadow-md transition-all duration-200"
              >
                <div className="p-5">
                  <div className="flex justify-between items-start mb-3">
                    <h2 className="text-xl font-medium text-gray-900 dark:text-gray-100">{room.name}</h2>
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${room.status === 'playing' ? 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400' : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'}`}>
                      {room.status}
                    </span>
                  </div>
                  
                  <div className="flex items-center mt-4 text-sm text-gray-500 dark:text-gray-400">
                    <Users className="h-4 w-4 mr-2" />
                    <span className="flex-1">{room.players?.length || 0} / {room.max_players} players</span>
                    
                    <div className={`flex items-center ml-2 font-medium text-sm ${room.status === 'playing' ? 'text-orange-600 dark:text-orange-400' : 'text-primary-600 dark:text-primary-400'}`}>
                      {room.status === 'playing' ? (
                        <>
                          View Game
                          <Play className="h-4 w-4 ml-1" />
                        </>
                      ) : (
                        <>
                          Join Room
                          <ArrowRight className="h-4 w-4 ml-1" />
                        </>
                      )}
                    </div>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// Convert page to client component
export const dynamic = 'force-dynamic';
