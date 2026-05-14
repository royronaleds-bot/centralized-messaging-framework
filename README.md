# CoreHub: Centralized Messaging Framework

## 📖 Introduction
CoreHub is a high-concurrency real-time messaging system built on a Centralized Client-Server architecture. The primary goal of this framework is to provide a highly reliable, low-latency communication hub capable of managing thousands of concurrent users with minimal memory footprint.

## 🏗️ Technical Architecture
The system is engineered using **Go (Golang)** to leverage its native concurrency features (Goroutines and Channels), ensuring safe and fast message routing across the server without race conditions.

### Core Technologies:
* **Backend Language:** Go (Golang) - Chosen for distributed systems and high-performance networking.
* **Real-time Protocol:** WebSockets (via Gorilla WebSocket) - Provides a full-duplex, persistent TCP connection.
* **Database & Persistence:** PostgreSQL - Used to safely archive chat history and handle offline message storage.
* **Security Layer:** JWT (JSON Web Tokens) for session authentication and Bcrypt for secure password hashing.

## ✨ Implemented Features (Current Progress)

### Phase 1: Security & Identity Management ✅
- Secure user registration and login endpoints.
- Password hashing using `bcrypt` to protect user credentials.
- JWT-based authentication middleware to secure the WebSocket handshake.

### Phase 2: Instant Messaging & Persistence ✅
- Established stable WebSocket connections for real-time data streaming.
- Implemented full Data Persistence by archiving every message into PostgreSQL.
- Thread-safe broadcasting using Go Mutexes and Channels.

### Phase 3: Smart Message Routing (1-to-1 Chat) ✅
- Shifted from public broadcasting to private 1-to-1 message routing.
- Real-time filtering to ensure messages are delivered only to the designated receiver.
- Fetching and rendering structured Chat History from the database upon user login.

## 🛠️ Technical Challenges Solved
1. **High Concurrency:** Overcame the memory overhead of traditional servers by utilizing lightweight **Goroutines** to handle thousands of connections simultaneously.
2. **Data Synchronization:** Prevented data race conditions when multiple users send messages at the exact same millisecond by using **Go Channels and RWMutex**.

## 🚀 How to Run the Project
1. Ensure **PostgreSQL** is installed and running on port `5432`.
2. Create a database named `corehub_db` and run the necessary SQL migrations to create the `users` and `messages` tables (ensure the `receiver_username` column exists).
3. Clone this repository.
4. Run the backend server:
   ```bash
   go run .
