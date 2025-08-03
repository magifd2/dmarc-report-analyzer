# DMARC Report Analyzer

A self-contained local server application for analyzing DMARC aggregate reports.

## Features

- **DMARC Report Ingestion**: Upload DMARC aggregate reports (XML, ZIP, GZ) for parsing and storage.
- **IP Geolocation & ASN Lookup**: Enriches DMARC records with IP geographical and Autonomous System Number (ASN) information using IPInfo.io MMDB files.
- **Local DNS PTR Lookup**: Performs reverse DNS lookups for source IPs and stores reversed hostnames and apex domains.
- **Report Listing & Details**: View a list of ingested DMARC reports and drill down into detailed report information.
- **User Authentication**: Secure API access using JWT (JSON Web Tokens).
- **User Management**: CLI-based user creation and management. Web UI for user's own password change.
- **Single Executable**: Packaged as a single, self-contained executable for easy deployment.

## Technologies

- **Backend**: Go
  - Web Framework: Gorilla Mux
  - Database: SQLite
  - Authentication: JWT (golang-jwt/jwt/v5)
  - IP Geo: MaxMindDB (oschwald/maxminddb-golang)
  - CORS: rs/cors
- **Frontend**: React (TypeScript)
  - Build Tool: Vite
  - Styling: Tailwind CSS v3

## Getting Started

### Prerequisites

- Go (1.21 or later)
- Node.js (18.x or later)
- npm (or yarn/pnpm)
- `sqlite3` CLI tool (for database inspection, optional)

### Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/magifd2/dmarc-report-analyzer.git
    cd dmarc-report-analyzer
    ```

2.  **Set JWT Secret:**
    The backend requires a JWT secret for authentication. Set it as an environment variable.
    ```bash
    export JWT_SECRET="your_super_secret_jwt_key_at_least_32_chars_long"
    ```
    **Important**: Replace `your_super_secret_jwt_key_at_least_32_chars_long` with a strong, unique secret.

3.  **Create an Admin User:**
    The application uses CLI for initial user creation.
    ```bash
    cd backend
    ./start.sh # This will build the backend executable
    ./bin/dmarc-report-analyzer-backend --create-user admin --password your_admin_password
    cd ..
    ```
    **Important**: Replace `your_admin_password` with a strong password.

4.  **Import IP Geo Database (Optional but Recommended):**
    Download an IPInfo.io MMDB file (e.g., `ipinfo-city.mmdb`) and import it.
    ```bash
    # Example: Download ipinfo-city.mmdb
    # curl -o ipinfo-city.mmdb https://ipinfo.io/data/ipinfo-city.mmdb?token=YOUR_IPINFO_TOKEN # (Requires IPInfo token for direct download)
    # Or manually download from ipinfo.io and place it somewhere accessible.

    cd backend
    ./bin/dmarc-report-analyzer-backend --import-ip-db /path/to/your/ipinfo-city.mmdb
    cd ..
    ```
    **Note**: Only manual import is supported. Automatic download is not implemented.

### Running the Application

The application is designed to run as a single executable. The `start.sh` script in the `backend` directory will build both the frontend and backend, then start the server.

```bash
cd backend
./start.sh
```

This will:
1.  Build the React frontend into `backend/src/static_frontend_dist`.
2.  Build the Go backend executable, embedding the frontend assets.
3.  Start the Go server, listening on `http://localhost:8080` (default).

To stop the server:
```bash
cd backend
./stop.sh
```

### Development

#### Backend

To run the backend in development mode (without embedding frontend):

```bash
cd backend
go run src/main.go
```

#### Frontend

To run the frontend development server (requires backend to be running separately for API calls):

```bash
cd frontend
npm install # if not already installed
npm run dev
```
The frontend development server will typically run on `http://localhost:5173` (or another available port). API calls will be proxied to `http://localhost:8080` as configured in `frontend/vite.config.ts`.

## Contributing

(To be added)

## License

(To be added)
