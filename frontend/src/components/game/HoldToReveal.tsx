import { useState, useRef, useEffect } from 'react';

interface HoldToRevealProps {
  children: React.ReactNode;
  revealText?: string;
  hiddenText?: string;
  revealTime?: number; // Time in ms to hold before revealing
}

export default function HoldToReveal({
  children,
  revealText = 'Hold to reveal',
  hiddenText = 'Release to hide',
  revealTime = 500, // Default 500ms hold time
}: HoldToRevealProps) {
  const [isRevealed, setIsRevealed] = useState(false);
  const [isHolding, setIsHolding] = useState(false);
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  
  // Clear timer on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);
  
  // Handle touch/mouse down
  const handleStart = () => {
    setIsHolding(true);
    
    timerRef.current = setTimeout(() => {
      setIsRevealed(true);
    }, revealTime);
  };
  
  // Handle touch/mouse up
  const handleEnd = () => {
    setIsHolding(false);
    setIsRevealed(false);
    
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  };
  
  return (
    <div 
      className={`role-card ${isRevealed ? 'reveal' : ''}`}
      onTouchStart={handleStart}
      onTouchEnd={handleEnd}
      onMouseDown={handleStart}
      onMouseUp={handleEnd}
      onMouseLeave={handleEnd}
    >
      <div className="role-card-front">
        <span className="text-sm text-white/80">{isHolding ? hiddenText : revealText}</span>
      </div>
      <div className="role-content">
        {children}
      </div>
    </div>
  );
}
