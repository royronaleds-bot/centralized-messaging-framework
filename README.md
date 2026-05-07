# CoreHub - Centralized Messaging Framework

CoreHub is a real-time messaging server built with Go (Golang) and PostgreSQL. The project follows a modular architecture for high scalability and clean code standards.

## 🏗 Project Structure
To maintain a professional workflow, the project is organized into several packages:
- **/database**: Manages PostgreSQL connections and initialization.
- **/handlers**: Contains the logic for WebSockets (Chat) and User Authentication.
- **/models**: Defines data structures for Users and main.go **main.go**: The main entry point that starts the server and routes.

## 🚀 Stage 1 FeaModular Backend:r Backend:** Separated concerns for better maintainabiSecure Auth:cure Auth:** Integrated password haBcryptg **BcryReal-time Engine:me Engine:** Full-duplex communWebSocketsWebSockeDatabase Persistence:rsistence:** All messages are archived in a PostgreSQL database.

## 🛠 TLanguage:*Language:** GoDatabase:*Database:** PLibraries:Libraries:** Gorilla WebSockeArchitecture:hitecture:** Modular Clean Architecture

## 📋 How to Run
1. Clone the repository.
2. Run go mod tidy to sync dependencies.
3. Configure your DB credentials in database/db.go.
4. Start the engine:
   `bash
   go run main.go
