import Link from 'next/link';
import { ArrowRight } from 'lucide-react';

export default function Home() {
  return (
    <main className="min-h-screen flex flex-col items-center justify-center p-4">
      <div className="max-w-lg w-full space-y-8 text-center">
        <h1 className="text-4xl font-bold text-primary-600">Avalon</h1>
        <p className="text-lg text-gray-600 dark:text-gray-300">
          A web-based implementation of the popular board game
        </p>
        
        <div className="mt-10 flex flex-col space-y-4">
          <Link 
            href="/auth/login"
            className="py-3 px-4 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 transition-colors flex items-center justify-center"
          >
            Login
            <ArrowRight className="ml-2 h-5 w-5" />
          </Link>
          
          <Link 
            href="/auth/register"
            className="py-3 px-4 bg-secondary-600 text-white font-medium rounded-lg hover:bg-secondary-700 transition-colors flex items-center justify-center"
          >
            Register
            <ArrowRight className="ml-2 h-5 w-5" />
          </Link>
          
          <div className="pt-4 text-sm text-gray-500 dark:text-gray-400">
            <p>Experience the classic game of deduction and deception</p>
            <p className="mt-2">Play with friends using our mobile-optimized interface</p>
          </div>
        </div>
      </div>
    </main>
  );
}
