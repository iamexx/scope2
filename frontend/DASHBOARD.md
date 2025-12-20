# Vue 3 Dashboard Foundation

This document describes the Vue 3 dashboard implementation for the DayZ Server Manager.

## Architecture Overview

### Project Structure

```
frontend/src/
├── components/              # Reusable UI components
│   ├── NavBar.vue          # Top navigation bar
│   ├── SideBar.vue         # Navigation sidebar
│   ├── ServerCard.vue      # Server status card with action buttons
│   ├── StatusBadge.vue     # Server status indicator (running/stopped)
│   ├── Toast.vue           # Toast notification system
│   ├── LoadingSpinner.vue  # Loading indicator
│   └── Modal.vue           # Generic modal wrapper
├── layouts/                # Layout templates
│   ├── AuthLayout.vue      # Login/setup pages layout
│   └── MainLayout.vue      # Dashboard layout with navbar/sidebar
├── pages/                  # Page components
│   ├── LoginPage.vue       # Admin login form
│   ├── FirstRunSetup.vue   # Admin user creation for first load
│   └── ServerStatus.vue    # Main dashboard with server status
├── router/                 # Vue Router configuration
│   └── index.js            # Route definitions and navigation guards
├── stores/                 # Pinia state management
│   ├── authStore.js        # JWT token, login state, user info
│   ├── serverStore.js      # Server list, current server, status
│   └── uiStore.js          # Toast messages, loading states
├── utils/                  # Utility functions
│   └── apiClient.js        # Axios instance with JWT interceptor
├── App.vue                 # Root component
└── main.js                 # Application entry point
```

## Features Implemented

### Authentication
- **Login Page** (`LoginPage.vue`): Admin login form with username/password
- **FirstRunSetup Page** (`FirstRunSetup.vue`): Admin user creation for first-run
- **JWT Token Management**: Stores JWT in localStorage, auto-expires after 24 hours
- **API Interceptor**: Automatically adds Authorization header to requests
- **401 Error Handling**: Clears auth state and redirects to login on token expiration

### Dashboard
- **Server Status Page** (`ServerStatus.vue`): Main dashboard displaying:
  - Total server count
  - Running/stopped server counts
  - SteamCMD status section
  - Server grid with cards
- **Server Card** (`ServerCard.vue`): Individual server cards with:
  - Server name and ID
  - Current status (running/stopped)
  - Player count / max players
  - Port number
  - Action buttons: Start, Stop, Restart
  - Confirmation modal for stop action

### UI Components
- **NavBar**: Top navigation with:
  - App branding and title
  - Current username display
  - Logout button
- **SideBar**: Navigation sidebar with:
  - Dashboard link
  - Responsive design
  - Active route highlighting
- **Toast Notifications**: System for showing:
  - Success messages (green, 3s duration)
  - Error messages (red, 5s duration)
  - Info messages (blue, 3s duration)
  - Warning messages (amber, 4s duration)
- **Loading Spinner**: Animated loading indicator
- **Status Badge**: Color-coded status indicators
  - Green for running
  - Gray for stopped
  - Blue for loading
  - Red for error

### State Management (Pinia)

#### Auth Store (`authStore.js`)
```javascript
// State
token                 // JWT token
user                  // User object with id, username, createdAt
isFirstRun           // First-run detection flag
loading              // Loading state
error                // Error message

// Actions
login(username, password)        // Authenticate user
setup(username, password)        // Create admin user
fetchUser()                      // Fetch current user info
logout()                         // Clear auth state
checkFirstRun()                  // Check if first run
```

#### Server Store (`serverStore.js`)
```javascript
// State
servers              // Array of server objects
currentServer        // Currently selected server
loading              // Loading state
error                // Error message

// Computed
serverCount          // Total number of servers
runningServers       // Count of running servers

// Actions
fetchServers()                   // Get all servers
fetchServerDetails(id)           // Get server by ID
fetchServerStatus(id)            // Get server status
startServer(id)                  // Start server
stopServer(id)                   // Stop server
restartServer(id)                // Restart server
```

#### UI Store (`uiStore.js`)
```javascript
// State
toasts               // Array of toast messages
loading              // Global loading state

// Actions
addToast(msg, type, duration)    // Add toast notification
removeToast(id)                  // Remove toast by ID
showSuccess(msg)                 // Show success toast
showError(msg)                   // Show error toast
showInfo(msg)                    // Show info toast
showWarning(msg)                 // Show warning toast
```

### Router Configuration (`router/index.js`)

Routes:
- `/login` - Login page (public)
- `/setup` - First-run setup page (public)
- `/` - Server status dashboard (protected)

Navigation Guards:
- Redirects unauthenticated users to `/login`
- Redirects authenticated users away from login/setup pages
- Protects `/` route with JWT authentication

### API Integration (`apiClient.js`)

Axios instance with:
- Base URL: `/api` (configurable via `VITE_API_BASE_URL`)
- Request Interceptor: Adds JWT token from localStorage
- Response Interceptor: Handles 401 errors
- CORS support with credentials

### Styling

- **Tailwind CSS v4**: Modern utility-first CSS framework
- **Color Scheme**:
  - Primary: Blue (500, 600)
  - Success: Green (500, 600)
  - Error: Red (500, 600)
  - Warning: Amber (500)
  - Background: Slate (50, 900)
- **Responsive Design**: Mobile-first approach
- **Animations**: Smooth transitions and spinner animations

## API Integration Points

The dashboard integrates with backend API endpoints:

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/setup` - First-run admin user creation
- `GET /api/auth/me` - Get current user info (requires JWT)

### Servers
- `GET /api/servers` - List all servers (requires JWT)
- `GET /api/servers/{id}` - Get server details (requires JWT)
- `GET /api/servers/{id}/status` - Get server status (requires JWT)
- `POST /api/servers/{id}/start` - Start server (requires JWT)
- `POST /api/servers/{id}/stop` - Stop server (requires JWT)
- `POST /api/servers/{id}/restart` - Restart server (requires JWT)

### SteamCMD (Ready for Integration)
- `GET /api/steamcmd/status` - Get SteamCMD status (requires JWT)
- `POST /api/steamcmd/sync` - Start SteamCMD sync (requires JWT)
- `GET /api/steamcmd/sync/progress` - Get sync progress (requires JWT)

## Data Flow

### Initial Load
1. App.vue checks if user has token in localStorage
2. If token exists, fetches current user info
3. Router guards check authentication state
4. Unauthenticated users redirected to login
5. On first run (no token), directed to setup page

### Login Flow
1. User enters credentials on LoginPage
2. Form validates input
3. Submits POST request to `/api/auth/login`
4. API returns JWT token
5. Store saves token to localStorage
6. Router redirects to dashboard
7. Dashboard fetches server list

### Server Actions
1. User clicks Start/Stop/Restart on ServerCard
2. Component shows loading state
3. API request sent to appropriate endpoint
4. On success: status updated, toast shown
5. On error: error toast shown
6. List refreshes automatically

## Mock Data

For development, ServerStatus.vue includes mock server data when API is unavailable:
```javascript
[
  { id: '1', name: 'Main Server', status: 'running', ... },
  { id: '2', name: 'Test Server', status: 'stopped', ... },
  { id: '3', name: 'Backup Server', status: 'running', ... },
]
```

## Ready for Future Enhancement

### Planned Features
- Real-time server status updates (WebSocket)
- Server logs viewer
- Server configuration editor
- FTP user management
- Player management
- Backup/restore functionality
- Server crash detection
- Admin activity logs

### API Endpoints Ready
All major CRUD operations are prepared to integrate with API endpoints as they become available.

## Development

### Start Development Server
```bash
npm run dev
```

### Build for Production
```bash
npm run build
```

### Project Structure Notes
- Components are organized by function (layout, page, UI component)
- Stores follow the Pinia composition API pattern
- All API calls go through the centralized `apiClient.js`
- Authentication state is automatically persisted to localStorage
- Token expiration is handled via 401 response interception
