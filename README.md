# centralized-messaging-framework
high-concurrency centralized messaging framework built with Go, utilizing WebSockets for real-time communication and PostgreSQL for data persistence. Part of the Network Applications Programming course 2026
# CoreHub: High-Concurrency Messaging Framework

## 🚀 Overview
[span_28](start_span)CoreHub is an advanced messaging system built on a Centralized Client-Server architecture[span_28](end_span). [span_29](start_span)[span_30](start_span)Developed using Go, it is specifically engineered for high-concurrency environments, utilizing lightweight threads to manage thousands of simultaneous connections with minimal overhead[span_29](end_span)[span_30](end_span).

## ✨ Key Features
* [span_31](start_span)[span_32](start_span)Real-time Messaging: Full-duplex communication with minimal latency using WebSockets (TCP)[span_31](end_span)[span_32](end_span).
* [span_33](start_span)[span_34](start_span)High Concurrency: Powered by Go’s Goroutines and Channels for efficient parallel processing[span_33](end_span)[span_34](end_span).
* [span_35](start_span)[span_36](start_span)Offline Support: Messages are securely archived in the server and delivered immediately upon user reconnection[span_35](end_span)[span_36](end_span).
* [span_37](start_span)[span_38](start_span)Advanced Security: Secure user authentication via JWT, password encryption using Bcrypt, and data-in-transit protection via TLS/WSS[span_37](end_span)[span_38](end_span).
* [span_39](start_span)Interactive Indicators: Features real-time presence tracking (Online/Offline), typing indicators, and delivery receipts[span_39](end_span).
* [span_40](start_span)Group Management: Support for group chat rooms with real-time message broadcasting[span_40](end_span).

## 🛠 Tech Stack
* [span_41](start_span)Language: Go (Golang)[span_41](end_span).
* [span_42](start_span)Database: PostgreSQL (Relational persistence)[span_42](end_span).
* [span_43](start_span)Real-time Protocol: WebSockets (via Gorilla library)[span_43](end_span).
* [span_44](start_span)Security: JWT (JSON Web Tokens), Bcrypt, TLS[span_44](end_span).

## 🏗 System Architecture
[span_45](start_span)The system is divided into three functional layers[span_45](end_span):
1. [span_46](start_span)Central Hub: The core engine for connection management and message routing[span_46](end_span).
2. [span_47](start_span)Security Layer: Ensures data integrity, encryption, and authorized access[span_47](end_span).
3. [span_48](start_span)[span_49](start_span)Persistence Layer: Handles long-term storage and chat history retrieval[span_48](end_span)[span_49](end_span).

---
*Developed as part of the Network Applications Programming Course Project 2026.*
