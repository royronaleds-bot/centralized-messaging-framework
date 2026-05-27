```python
readme_content = """# CoreHub: Centralized Concurrency-Driven Messaging Framework

An enterprise-grade, high-concurrency real-time messaging framework engineered using a centralized Client-Server architecture. Built with **Go (Golang)**, **PostgreSQL**, and **WebSockets**, CoreHub is architected to deliver ultra-low latency data routing, persistent multi-session message state preservation, and dynamic network discoverability with a highly optimized memory footprint.

---

## 🏗️ System Architecture & Concurrency Model

CoreHub follows a strictly decoupled **Centralized Client-Server Model**. The Go backend acts as the authoritative communication hub controlling session state, user authentication, security isolation, relational database persistence, and synchronized web socket connection topologies.


```

```text
README.md written successfully.


```

```
   ┌────────────────────────┐               ┌────────────────────────┐
   │   Client Laptop (A)    │               │   Client Laptop (B)    │
   │  (Vanilla HTML/CSS/JS) │               │  (Vanilla HTML/CSS/JS) │
   └───────────┬────────────┘               └───────────┬────────────┘
               │                                        │
               │  Persistent Full-Duplex TCP Link       │
               │  (Authenticated WebSocket - WSS/WS)    │
               ▼                                        ▼
 ======================================================================
                     CENTRAL SERVER BACKEND (GO / GOLANG)
 ======================================================================
 │                                                                    │
 │  ┌──────────────────┐      Goroutine Pool      ┌────────────────┐  │
 │  │  JWT Auth & TLS  ├─────────────────────────►│  Central Hub   │  │
 │  │  Security Layer  │                          │ (Presence Map) │  │
 │  └──────────────────┘                          └───────┬────────┘  │
 │                                                        │           │
 │                                       Go Channels      │           │
 │                                   ┌────────────────────┘           │
 │                                   ▼                                │
 │                           ┌──────────────┐                         │
 │                           │  Broadcast   │                         │
 │                           │  Engine      │                         │
 │                           └──────┬───────┘                         │
 ===================================│==================================
                                    │ Relational Persistence
                                    ▼
                       ┌────────────────────────┐
                       │   PostgreSQL Database  │
                       │   (Persistent Storage) │
                       └────────────────────────┘

```

```

### 1. Concurrency Management
* **Goroutines:** Unlike traditional threaded environments where each TCP socket consumes heavy OS-level thread stacks, CoreHub spawns ultra-lightweight Go runtimes (Goroutines) consuming only a few kilobytes per connection. This allows the framework to easily scale to thousands of simultaneous connections.
* **Go Channels:** Active messages are handled asynchronously via thread-safe FIFO data channels (`chan Message`). This completely decouples message reading from message writing loops, avoiding I/O bottlenecks.
* **RWMutex Synchronization:** Concurrent maps holding active WebSocket context pointers are shielded using synchronized mutual exclusion locks (`sync.RWMutex`). This systematically mitigates **Race Conditions** during sudden connection drops or simultaneous multi-branch message spikes.

---

## ✨ Framework Features & Requirements Mapping

The system natively implements all standard and advanced components defined within the official project specifications:

### 1. Security & Cryptographic Identity Management
* **Cryptographic Hashing:** Passwords are mathematically hashed with a highly secure salt using **Bcrypt** before committing to the `users` relational index. Plaintext credentials never exist inside the persistent layer.
* **Stateful Handshake Verification:** User sessions are verified via stateless **JWT (JSON Web Tokens)**. The token is issued upon successful login and seamlessly parsed during the WebSocket compilation handshake to validate access rights.

### 2. Advanced Real-Time Chat Engine & Interactive UX
* **Live Presence Architecture:** Real-time state machines seamlessly track client connectivity. When a TCP connection drops or opens, the server captures the lifecycle hook and dynamically broadcasts `online`/`offline` state changes to all corresponding network nodes.
* **Database-Bypassing Ephemeral Notifications:** Typing states (*"User is typing..."*) are explicitly routed as ephemeral packet frames that bypass the database engine entirely, reducing disk I/O overhead and accelerating immediate delivery.
* **Multi-State Delivery Receipts:** Complete tracking of the message lifecycle via asynchronous acknowledgment protocols:
  * `🕒 Queued/Pending`: Message produced and held inside local volatile storage (Offline Queue).
  * `✓ Sent`: Message accepted by the central hub and successfully committed to PostgreSQL.
  * `✓✓ Delivered`: Acknowledged by the target node's active layout engine.
  * `✓✓ Read (Cyan)`: Explicit read acknowledgement dispatched automatically when the user actively mounts the corresponding viewport.

### 3. Isolated Group Routing (Rooms)
* Group channels are fully managed inside synchronized server memory structures. Messages addressed to a group ID query a cached membership relation map, ensuring data frames are distributed exclusively to matching active connections.

### 4. Network Health Orchestration & Local Discoverability
* **Heartbeat Protocol (Ping/Pong):** Prevents memory leaks and zombie sockets by enforcing a strict 60-second read deadline on the server. Clients push silent ping events every 30 seconds to maintain connection health across intermediate NAT routers.
* **Dynamic Network Discovery:** Outfitted with an elegant frontend abstraction layer utilizing `window.location.host`. This replaces rigid `localhost` hooks with dynamic contextual address binding. It allows the centralized instance to operate smoothly over a Local Area Network (LAN) or wireless Hotspot, adapting instantly to dynamically assigned network IP blocks or local mDNS Hostnames (`.local`).

---

## 🗄️ Relational Database Schema

CoreHub utilizes **PostgreSQL** for strict transactional integrity and structural historical preservation. The relational model is implemented as follows:

```sql
-- Users Entity Index
CREATE TABLE users (
    username TEXT PRIMARY KEY,
    password TEXT NOT NULL
);

-- Groups/Rooms Index
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Group Membership Resolution Mapping (Many-to-Many Bridge)
CREATE TABLE group_members (
    group_id INT REFERENCES groups(id) ON DELETE CASCADE,
    username TEXT REFERENCES users(username) ON DELETE CASCADE,
    PRIMARY KEY (group_id, username)
);

-- Transactional Historical Message Store
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL REFERENCES users(username),
    receiver_username TEXT REFERENCES users(username), -- Nullable for group rooms
    group_id INT DEFAULT 0,                            -- 0 specifies peer-to-peer chat
    content TEXT NOT NULL,
    client_msg_id TEXT UNIQUE,                         -- Enforces idempotent packet ingestion
    status TEXT DEFAULT 'sent',                        -- [sent, delivered, read]
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP     -- System clock timestamp for chronological sorting
);

```

---

## 🚀 Execution & Deployment Walkthrough

### 1. Database Initialization

Ensure PostgreSQL is up and running, and configure your target connection parameters inside `database/db.go`. Execute the SQL schema script provided above to set up the corresponding indices.

### 2. Launching the Central Hub Backend

Compile and launch the Go server framework directly from your primary repository directory:

```bash
go run main.go

```

The console will log server assembly initialization:
`Successfully connected to database!`
`🚀 Server launched on: http://localhost:8080`

### 3. Dynamic Multi-Device Cross-Network Testing

Thanks to the integration of dynamic hostname binding, deploying the framework across various laptop terminals inside a branch setup (e.g., simulating connectivity across Latakia, Jableh, Baniyas, or Tartus regional node infrastructures) requires zero manual asset modifications:

1. Connect all target computing nodes (Laptops) to a single shared network point or local hot-spot connection.
2. Determine the authoritative server's current address parameter. You can locate this via your OS networking panel or using the terminal:
```bash
# Windows Command Prompt
ipconfig

```


Locate your IPv4 block (e.g., `192.168.1.45`) or obtain the native Hostname designation (e.g., `DESKTOP-823OF7G`).
3. Distribute the network access parameter to your testing nodes. They can simply input the following into their browser location bar:
```
[http://192.168.1.45:8080](http://192.168.1.45:8080)   OR   [http://DESKTOP-823OF7G.local:8080](http://DESKTOP-823OF7G.local:8080)

```


4. The central Go server instantly serves the compiled frontend. The application dynamics will instantly bind to the host address configuration, seamlessly executing WebSocket pipelines across all distributed multi-node devices.
"""

with open("README.md", "w", encoding="utf-8") as f:
f.write(readme_content).

انسخي الرسالة، وأرسليها مع الصور، وأنا هنا جاهز لأي خطوة قادمة أو تعديل تطلبه منكِ!

```
