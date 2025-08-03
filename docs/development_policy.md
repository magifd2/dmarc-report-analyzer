# Development Policy for DMARC Report Analyzer (New Application)

## 1. Introduction

This document outlines the core development policy and architectural decisions for the new DMARC Report Analyzer application. This policy is established to address the limitations of the previous HTML-based application, specifically regarding data persistence and multi-user access, while adhering to strict constraints against cloud deployment or dedicated server infrastructure.

## 2. Core Architecture: Self-Contained Local Server Application

The application will be developed as a self-contained, local server application that runs directly on a user's personal computer. It will comprise a backend server component and a modern web-based frontend.

### 2.1. Frontend (User Interface)

*   **Technology Stack:** React (TypeScript) for the application logic and UI components, styled with Tailwind CSS for a modern and responsive design.
    *   **Tailwind CSS Versioning Strategy:** To ensure stability and avoid breaking changes due to version incompatibilities, Tailwind CSS and its related PostCSS dependencies will be strictly version-locked. This means:
        *   Exact versions (e.g., `3.4.3` instead of `^3.0.0`) will be specified in `package.json`.
        *   The `package-lock.json` (or `yarn.lock`) file will be committed to the repository to guarantee identical dependency trees across all development and build environments.
        *   Developers are encouraged to use `npm ci` (or `yarn install --frozen-lockfile`) for installing dependencies to ensure strict adherence to the locked versions.
*   **Delivery:** The frontend static files (HTML, CSS, JavaScript) will be served by the local backend server.
*   **User Experience:** Re-implement all features from the existing functional specification, including DMARC report parsing, filtering, sorting, charting, summary tables, detailed records, and the analysis modal.

### 2.2. Backend (Server Functionality)

*   **Technology Stack:** **Go** will be used for the backend. Go's strong typing, performance, and ease of packaging into a single executable make it an ideal choice for this self-contained application.
*   **Database:** **SQLite** will be used as the primary data store. SQLite is a file-based database, eliminating the need for a separate database server installation or management. The database file (e.g., `dmarc_reports.db`) will reside alongside the application executable.
*   **Key Backend Responsibilities:**
    *   Serving frontend static assets.
    *   Providing API endpoints for DMARC report upload, parsing, and storage into the SQLite database.
    *   Providing API endpoints for querying, filtering, aggregating, and sorting DMARC data from SQLite for the frontend.
    *   Managing IP information (GeoIP, ASN, Reverse DNS).

## 3. Data Persistence and Sharing

### 3.1. Data Persistence

*   All parsed DMARC report data, including associated IP information, will be stored persistently in the SQLite database file. This ensures data is retained across application restarts.

### 3.2. Multi-User Access (Local Network Sharing)

*   The application is designed to be run on a designated "host" PC within a local network.
*   Other users within the same local network can access the application and its data by navigating their web browsers to the host PC's IP address and the application's listening port (e.g., `http://[Host_IP_Address]:[Port]`).
*   This approach enables multiple users to simultaneously view the same centralized DMARC report data without requiring cloud deployment or dedicated server infrastructure.
*   **Note:** Network configuration (e.g., firewall rules on the host PC) may be necessary to allow access from other devices.

## 4. IP Information Handling

### 4.1. GeoIP and ASN Information

*   **Source:** **IPInfo.io free database (.mmdb format)**. This choice eliminates the need for users to obtain a MaxMind license key, simplifying setup.
*   **Database Update:** The application will **not** automatically download or update the IPInfo database. IPInfo API keys are **not** required by the application.
    *   **Manual Import via CLI:** Users are responsible for manually downloading the latest `.mmdb` file from IPInfo's website. This file can then be imported into the application via a command-line option (e.g., `--import-ip-db <path_to_mmdb_file>`). When this option is used, the specified file will be copied to the application's data directory, overriding any existing database. This provides full control and flexibility for users, especially in environments with limited internet access.

### 4.2. Reverse DNS (Hostname) Information

*   **Resolution Method:** During the DMARC report parsing and data ingestion process, the backend will perform a **local DNS PTR record lookup** for each unique `source_ip` address.
*   **Persistence:** The resolved hostname (reverse domain), its **reversed form (e.g., `com.example.www` for `www.example.com`)**, and its **apex domain (e.g., `example.com` for `www.example.com`)** will be stored in the SQLite database along with the GeoIP and ASN information. This eliminates the need for repeated DNS queries during subsequent data viewing and enables more efficient filtering and aggregation.
*   **Error Handling:** DNS lookup failures (e.g., no PTR record, timeout) will be handled gracefully, storing a placeholder like "N/A".

## 5. Application Distribution

*   The entire application (backend server, frontend static files, and SQLite database) will be packaged into a **single, self-contained executable file**.
*   Tools like PyInstaller (for Python) or Go's native build capabilities will be used for this purpose.
*   Users will be able to launch the application by simply executing this single file.

## 6. Development Principles

*   **Modularity:** The application will be built with a clear separation of concerns between frontend and backend, and within each, using modular components/functions.
*   **Robustness:** Emphasis on error handling, especially during file parsing, network operations (IPInfo downloads, DNS lookups), and database interactions.
*   **User Experience:** The frontend will prioritize an intuitive and responsive user interface, consistent with the detailed functional specification.
*   **Maintainability:** Code will be clean, well-structured, and documented to facilitate future maintenance and enhancements.

## 7. Authentication Strategy

The application will implement JWT (JSON Web Tokens) for authentication to secure access to its backend API endpoints. This approach is chosen for its stateless nature, improved security over basic authentication (by not transmitting credentials on every request), and better overall compatibility with modern web application architectures.

### 7.1. JWT Implementation Details

*   **Login Endpoint:** A dedicated API endpoint (`POST /api/login`) will be provided for users to authenticate with a username and password. Upon successful authentication, the backend will generate and return a signed JWT.
*   **Token Management:**
    *   The frontend will store the received JWT securely (e.g., in `localStorage` or `sessionStorage`).
    *   Subsequent API requests to protected endpoints will include the JWT in the `Authorization` header (e.g., `Authorization: Bearer <JWT>`).
*   **Backend Validation:** The backend will validate the authenticity, integrity, and expiration of the JWT for every protected API request.

### 7.2. Security Considerations and HTTPS

*   **HTTPS Requirement:** While JWTs offer advantages over basic authentication, they are base64-encoded and not encrypted. Therefore, **HTTPS (SSL/TLS) is absolutely critical** to protect JWTs from interception and ensure secure communication.
*   **Local HTTPS Challenges:** Implementing HTTPS in a local, self-contained application presents specific challenges:
    *   **Self-Signed Certificates:** The application will likely use self-signed SSL/TLS certificates. Browsers will typically display a "Not Secure" warning for connections using self-signed certificates, as they are not issued by a trusted Certificate Authority.
    *   **User Experience Impact:** Users will need to acknowledge or bypass these browser warnings. For a smoother experience, users might need to manually install the self-signed certificate as a trusted root certificate on their operating system, which requires technical steps.
*   **Documentation:** The application's user documentation will clearly explain the importance of HTTPS, the implications of self-signed certificates, and provide instructions for managing browser warnings or installing the certificate for a seamless experience.
