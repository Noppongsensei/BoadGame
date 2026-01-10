# Avalon Web-Based Board Game Platform

A scalable, mobile-first implementation of the popular Avalon board game using a modern tech stack with Go (Golang) backend and Next.js frontend.

## Technology Stack

### Backend
- **Go (Golang)** with **Fiber** framework for high concurrency & WebSockets
- **PostgreSQL** database with hybrid schema (relational + JSONB)
- Hexagonal Architecture (Ports and Adapters) for clean separation of concerns

### Frontend
- **Next.js** (React) with **Zustand** for state management
- **TailwindCSS** for responsive, mobile-first design
- WebSocket for real-time communication

### Infrastructure
- **Docker** and **Docker Compose** for containerization and easy setup

## Features

- **Mobile-first UX design** optimized for handheld devices
- **Hold-to-Reveal pattern** for hidden roles to prevent screen peeking
- **Server-authoritative architecture** where all logic happens on the server
- **Anti-cheat (Fog of War)** system that filters game state per player
- **WebSocket Hub Pattern** for real-time updates
- **Responsive UI** that works on all device sizes
- **Extensible game engine** that can support other games in the future

## Setup Instructions

### Prerequisites
- Docker and Docker Compose
- Node.js and npm/yarn (for frontend development)

### Backend Setup

1. Clone the repository:
   ```
   git clone <repository-url>
   cd Boadgame
   ```

2. Configure environment variables:
   - Copy the `.env.example` file to `.env` and adjust settings as needed

3. Start the backend services using Docker:
   ```
   docker-compose up -d
   ```
   This will start both the PostgreSQL database and the Go Fiber backend.

### Frontend Setup

1. Navigate to the frontend directory:
   ```
   cd frontend
   ```

2. Install dependencies:
   ```
   npm install
   # or
   yarn install
   ```

3. Configure the frontend environment:
   - Create a `.env.local` file with:
   ```
   NEXT_PUBLIC_API_URL=http://localhost:8080
   ```

4. Run the development server:
   ```
   npm run dev
   # or
   yarn dev
   ```

5. Access the frontend at `http://localhost:3000`

## Game Rules

Avalon is a social deduction game where players are secretly divided into two teams: the loyal servants of Arthur (Good) and the minions of Mordred (Evil).

### Key Roles
- **Merlin**: Knows all evil players except Mordred
- **Percival**: Knows who Merlin is (but can be confused by Morgana)
- **Loyal Servants**: Regular good team members
- **Assassin**: Must identify Merlin at the end of the game
- **Mordred**: Evil player unknown to Merlin
- **Morgana**: Evil player who appears as Merlin to Percival
- **Oberon**: Evil player who doesn't know the other evil players
- **Minions**: Regular evil team members

### Game Flow
1. Players are assigned secret roles
2. A leader selects players for a quest
3. All players vote to approve/reject the team
4. If approved, the quest team members secretly choose success/fail
5. Evil players can choose to fail quests
6. Good team wins after 3 successful quests
7. Evil team wins after 3 failed quests or if the Assassin correctly identifies Merlin

## Project Structure

### Backend
- `/cmd/avalon`: Main application entry point
- `/internal/core`: Domain logic (pure game rules)
- `/internal/handlers`: HTTP & WebSocket handlers
- `/internal/repositories`: Database access
- `/internal/services`: Business logic
- `/pkg`: Utility libraries

### Frontend
- `/frontend/src/app`: Next.js app routes and pages
- `/frontend/src/components`: Reusable UI components
- `/frontend/src/store`: Zustand state management
- `/frontend/src/lib`: Utility functions and types

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
