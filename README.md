# CoreHub — Engineering a Production-Ready Real-Time System

<div align="center">
  <img src="https://img.icons8.com/color/120/000000/server.png" alt="CoreHub Logo"/>
  <p><b>A high-performance, real-time messaging and administration platform built with Go (Golang) and PostgreSQL.</b></p>
</div>

---

## 🎯 The Philosophy

CoreHub is not just another chat application; it is a case study in **Backend Software Engineering**. The primary goal of this project was to demonstrate how to design, architect, and evolve a scalable real-time system using **Clean Architecture principles**, **Concurrency (Goroutines)**, and **Vanilla Reactive UI**. The emphasis is on *how* the system operates under the hood—handling WebSocket multi-session edge cases, ensuring real-time data consistency, and memory management—rather than merely stacking features.

---

## 🏗️ System Architecture

CoreHub strictly enforces a layered architecture. The frontend UI knows nothing about the database, and the business logic is entirely decoupled into modular handlers.

```text
📱 Presentation Layer (Vanilla JS / HTML / CSS)
   ↓ listens to / triggers
🌐 API & Routing Layer (Gorilla Mux / JWT Middleware)
   ↓ calls
⚙️ Business Logic Layer (Isolated Go Handlers / WebSocket Engine)
   ↓ communicates with
🗄️ Data Layer (PostgreSQL / Parameterized Queries)
```

### Core Architectural Principles:
*   **Separation of Concerns:** Authentication, real-time chat, group management, and file uploads are abstracted into dedicated, isolated Go files (e.g., `auth.go`, `websocket.go`, `upload.go`).
*   **Zero-Dependency Frontend:** The UI is built using pure Vanilla JavaScript and CSS. No heavy frameworks were used, ensuring lightning-fast load times and a highly reactive DOM manipulation engine.
*   **Secure by Design:** Hardcoded secrets are eliminated. Passwords are cryptographically hashed using `bcrypt`, and endpoints are secured via JWT Middlewares.

---

## 🔬 Engineering Masterpieces (Under the Hood)

CoreHub tackles real-world distributed system challenges with innovative solutions:

### 1. The Multi-Session Presence Engine (Solving the "Ghost Offline" Bug)
Relying solely on a single WebSocket connection drop to mark a user as "Offline" leads to inaccurate data if a user is logged in from multiple devices (e.g., Mobile and Laptop).
*   **The Solution:** Implemented a smart Goroutine session-checker. When a socket drops, the server locks the state using `sync.RWMutex`, scans the active connections map, and verifies if the user has other active instances before broadcasting an "Offline" status to the network.

### 2. Deep-System Admin Dashboard
Standard dashboards only read database rows. CoreHub's dashboard talks directly to the operating system's hardware.
*   **The Solution:** Integrated Go's `runtime` package to fetch and broadcast live server metrics. The Admin interface dynamically renders Active Goroutines, actual RAM memory allocation, and CPU Core utilization in real-time.

### 3. mDNS Local Area Network (LAN) Resolution
To ensure a smooth testing and deployment environment without exposing explicit IP addresses or ports.
*   **The Solution:** Configured the Go server to bind to the default Port `80` and leverage native mDNS capabilities. Users can instantly connect to the application over the network using `http://<machine-name>.local` without any extra configurations.

---

## 📱 Features & Capabilities

| Category | Features | Tech Stack |
| :--- | :--- | :--- |
| **Messaging** | 1-on-1 Real-time chat, Group Chat, Read receipts (Ticks), Instant delivery | Gorilla WebSockets + Go Channels |
| **Admin Controls** | Live Server Stats, Instant User Banning/Unbanning, System-wide Broadcasts | Go `runtime` + Secure JWT roles |
| **Media** | File attachments, Inline image rendering, Dynamic file downloads | Go `multipart/form-data` + Local I/O |
| **Presence** | Reliable Online/Offline tracking, Multi-device session awareness | Concurrent Maps + `sync.RWMutex` |
| **Security** | JWT Authentication, Bcrypt Password Hashing, Route Protection | `golang-jwt/jwt/v5` + `x/crypto/bcrypt` |

---

## ⛺ The Agile Journey (Sprints)

This project was built iteratively over multiple Sprints, reflecting a professional SDLC:

*   **Sprint 1:** Architecture foundation, Database schema design (PostgreSQL), and JWT Auth setup.
*   **Sprint 2:** The Communication Layer (WebSocket engine, 1-on-1 message delivery).
*   **Sprint 3:** Group Chat implementation and real-time user discovery.
*   **Sprint 4:** Advanced Messaging (Delivery/Read ticks, File Uploads handling).
*   **Sprint 5:** System Behavior (Presence Engine, Offline queues, Multi-session logic).
*   **Sprint 6:** The Admin Panel (Quarantine zones, Banning system, Broadcasts, Live Dashboard).
*   **Sprint 7:** Production Polish (Clean Architecture refactoring, Code Splitting, Theming).

---

## 🛠️ Tech Stack

*   **Backend:** Go (Golang)
*   **Router & Sockets:** `gorilla/mux`, `gorilla/websocket`
*   **Database:** PostgreSQL (`lib/pq`)
*   **Frontend:** HTML5, CSS3, Vanilla JavaScript
*   **Security:** `golang.org/x/crypto/bcrypt`, `github.com/golang-jwt/jwt/v5`

---

## 🚀 Getting Started

### Prerequisites
1.  Go (>= 1.20) installed on your machine.
2.  PostgreSQL installed and running.
3.  A database named `corehub_db`.

### Installation
1. Clone the repository:
```bash
   git clone [https://github.com/your-username/CoreHub.git](https://github.com/your-username/CoreHub.git)
   ```
2. Install Go dependencies:
```bash
   go mod tidy
   ```
3. Set your environment variables (or use the defaults in `database.go`):
```bash
   DB_USER=postgres
   DB_PASSWORD=yourpassword
   DB_NAME=corehub_db
   ```
4. Run the server:
```bash
   go run main.go
   ```
5. Access the application on `http://localhost` or `http://<your-pc-name>.local`.

---

## 👨‍💻 Author

**Ali Ahmad**  
*Software Engineer*  
Passionate about building scalable, maintainable, and highly resilient backend systems.
