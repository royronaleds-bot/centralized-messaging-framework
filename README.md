# 🚀 CoreHub: High-Performance Centralized Messaging Framework

CoreHub is a scalable, real-time messaging server engineered with Go (Golang). It features a robust centralized architecture, secure user authentication, and persistent data storage using PostgreSQL. This project is designed to handle high-concurrency connections using modern software engineering patterns.

---

## 🏛 System Architecture
The project adopts a Modular Clean Architecture, separating concerns into distinct packages to ensure maintainability and testability.

### 📁 Directory Breakdown
* **main.go**: The application's entry point. Orchestrates the server lifecycle and routes.
* **/handlers**: The controller layer. Contains business logic for WebSockets (Chat engine) and User Authentication (Registration/Login).
* **/database**: Infrastructure layer. Handles the singleton connection to PostgreSQL and database mi/models **/models**: Data Access Objects (DAOs). Defines the structure of entities within the system.

---

## ✨ Core Features
### 1. Real-Time Communication EngineWebSockets (Full-Duplex)ll-Duplex)** for instantaneous message delivery.
* ImCentralized Hubalized Hub** pattern to manage active client connections.
* Thread-safe broadcasting using sync.Mutex to prevent race conditions.

### 2. Security & AutheBcrypt Hashing:t Hashing:** No raw passwords are stored. All credentials undergo multi-round cryptographicValidation:alidation:** Server-side checks for existing users and invalid payloads.

### 3. Data Persistence
* IntePostgreSQLPostgreSQL** to archive every transaction and message.
* Optimized SQL queries for rapid message retrieval (Stage 2 focus).

---

## 📡 API Endpoints & Routes

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| POST | /register | Creates a new user account with hashed credentials. |
| POST | /login | Authenticates users and initiates a session. |
| WS | /ws | Establhes a WebSocket connection for real-time chat. |

---

## 🛠 TBackend:**Backend:** Go (GolaDatabase:*Database:** PostgNetworking:etworking:** Gorilla WebSockeSecurity:*Security:** Golang Crypto (Bcrypt)

---

## ⚙️ Installation & Clone the Repository:epository:**
   
    git clone [https://github.com/royronaleds-bot/centralized-messaging-framework.git](https://github.com/royronaleds-bot/centralized-messaging-framework.git)
    
2.  **Install Dependencies:**
   
    go mod tidy
    
3.  **Environment Setup:**
    Update your PostgreSQL credentials in database/db.go.
4.  **Execute:**
   
    go run main.go
    
-Phase 1:admap
- [x] **Phase 1:** Core engine, DB integration, and modPhase 2:ring.
- [ ] **Phase 2:** JWT Authentication & PPhase 3:ging.
- [ ] **Phase 3:** React/Vue Frontend Integration.
