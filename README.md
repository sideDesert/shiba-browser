# Shiba Browser

Shiba Browser is a modern, full-stack web application that enables real-time chat and collaborative virtual browser sessions using WebRTC. It features a React-based frontend and a Go backend, supporting user authentication (with Google OAuth), chatrooms, friend management, notifications, and live media streaming.

## Features

- **Real-time Chat**: Join chatrooms and exchange messages instantly.
- **Virtual Browser Streaming**: Share and interact with a virtual browser session using WebRTC.
- **User Authentication**: Secure login with Google OAuth.
- **Dashboard**: Personalized dashboard for managing chatrooms, friends, and notifications.
- **Friend System**: Add, accept, or reject friend requests.
- **Notifications**: Stay updated with real-time notifications.
- **Modern UI**: Built with React, TailwindCSS, and Radix UI components.

## Tech Stack

- **Frontend**: React, React Router, TailwindCSS, TypeScript, Vite, WebRTC
- **Backend**: Go, Gorilla Mux, PostgreSQL, JWT, WebSockets, NATS
- **Streaming**: pion/webrtc, go-gst for media handling

## Getting Started

### Prerequisites

- Node.js (for the client)
- Go 1.24+ (for the server)
- PostgreSQL (for persistent storage)

### Setup

#### Client

```bash
cd client
npm install
npm run dev
