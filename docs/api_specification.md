# DMARC Report Analyzer - API Specification

This document details the API endpoints provided by the DMARC Report Analyzer backend server. These APIs facilitate communication between the frontend (React application) and the backend (Go server), enabling DMARC report processing, data retrieval, and application management.

## 1. Authentication

All protected API endpoints require a valid JSON Web Token (JWT) in the `Authorization` header. The JWT is obtained via the login endpoint.

### 1.1. User Login

*   **Purpose:** Authenticates a user with a username and password, and returns a valid JWT.
*   **HTTP Method:** `POST`
*   **Path:** `/api/login`
*   **Request:**
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "username": "string",
            "password": "string"
        }
        ```
*   **Response:**
    *   **Status:** `200 OK`
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "status": "success",
            "message": "Authentication successful.",
            "token": "string", // The issued JWT
            "user": {
                "username": "string",
                "role": "string" // e.g., "admin", "viewer"
            }
        }
        ```
    *   **Status:** `401 Unauthorized` (Authentication failed: invalid username or password)
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "status": "error",
            "message": "Invalid credentials."
        }
        ```
*   **Remarks:**
    *   User credentials (passwords hashed) are stored in the SQLite database.
    *   JWTs are signed using a backend-configured secret key and include claims like username, role, and expiration.

### 1.2. Authenticated User's Password Change

*   **Purpose:** Allows a logged-in user to change their own password.
*   **HTTP Method:** `PUT`
*   **Path:** `/api/user/password`
*   **Request:**
    *   **Content-Type:** `application/json`
    *   **Header:** `Authorization: Bearer <JWT>` (Requires a valid JWT)
    *   **Body:**
        ```json
        {
            "current_password": "string",
            "new_password": "string"
        }
        ```
*   **Response:**
    *   **Status:** `200 OK`
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "status": "success",
            "message": "Password updated successfully."
        }
        ```
    *   **Status:** `401 Unauthorized` (Invalid JWT or current password incorrect)
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "status": "error",
            "message": "Authentication failed or current password incorrect."
        }
        ```
*   **Remarks:**
    *   The backend verifies the `current_password` against the stored hash before updating.

## 2. Core Application APIs

### 2.1. Static File Serving

*   **Purpose:** Serves the frontend React application's build artifacts (HTML, CSS, JavaScript, etc.) to the browser.
*   **HTTP Method:** `GET`
*   **Path:** `/` and `/*` (any path)
*   **Request:** None
*   **Response:** Frontend static files (e.g., `index.html`, `bundle.js`, `style.css`)
*   **Remarks:** Accessing the application's root (`/`) returns `index.html`, which then loads other assets.

### 2.2. DMARC Report Upload and Parsing

*   **Purpose:** Allows users to upload DMARC report files (XML, ZIP, GZ, TGZ) for backend parsing and storage in the database.
*   **HTTP Method:** `POST`
*   **Path:** `/api/upload`
*   **Header:** `Authorization: Bearer <JWT>` (Requires a valid JWT)
*   **Request:**
    *   **Content-Type:** `multipart/form-data`
    *   **Body:** `file` field containing the uploaded file(s) (multiple files allowed).
*   **Response:**
    *   **Status:** `200 OK`
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "status": "success",
            "message": "Reports processed successfully.",
            "processed_count": 5, // Number of reports successfully processed
            "skipped_count": 2,   // Number of reports skipped due to duplication
            "error_count": 1,     // Number of reports that caused errors during processing
            "errors": [           // Details of errors (subset of information recorded in ingestion_errors table)
                {
                    "filename": "string",
                    "error_type": "string",
                    "message": "string"
                }
            ]
        }
        ```
    *   **Status:** `400 Bad Request` (e.g., invalid file format, missing file)
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "status": "error",
            "message": "string"
        }
        ```
    *   **Status:** `401 Unauthorized` (Invalid JWT)
*   **Remarks:**
    *   The backend extracts individual DMARC XMLs from archives, parses them, and stores data in `reports`, `records`, and `ip_info` tables.
    *   Duplicate reports are skipped based on `xml_hash`.
    *   Parsing errors are logged to `ingestion_errors` and summarized in the response.
    *   IP information (MaxMind GeoLite2 and local DNS reverse lookup) is processed during this step.

### 2.3. DMARC Report Data Retrieval

*   **Purpose:** Retrieves DMARC report data for dashboard display, supporting filtering, aggregation, and sorting.
*   **HTTP Method:** `GET`
*   **Path:** `/api/reports`
*   **Header:** `Authorization: Bearer <JWT>` (Requires a valid JWT)
*   **Request:**
    *   **Query Parameters:**
        *   `start_date`: `YYYY-MM-DD` (Optional)
        *   `end_date`: `YYYY-MM-DD` (Optional)
        *   `keyword`: Search string for IP, AS, domain (Optional)
        *   `disposition`: `none`, `quarantine`, `reject`, `any` (Optional, default: `any`)
        *   `dmarc_result`: `pass`, `fail`, `any` (Optional, default: `any`)
        *   `spf_result`: `pass`, `fail`, `any` (Optional, default: `any`)
        *   `dkim_result`: `pass`, `fail`, `any` (Optional, default: `any`)
        *   `as_name`: Filter by AS name (Optional)
        *   `country_code`: Filter by country code (Optional)
        *   `domain`: Filter by reverse domain or From domain (Optional)
        *   `aggregate`: `true` or `false` (Aggregate detailed records, Optional, default: `true`)
        *   `sort_by`: Column name for sorting (e.g., `count`, `source_ip`, `disposition`) (Optional)
        *   `sort_order`: `asc` or `desc` (Optional, default: `desc`)
        *   `limit`: Maximum number of records to retrieve (Optional, default: 500)
        *   `offset`: Starting position for retrieval (Optional, default: 0)
*   **Response:**
    *   **Status:** `200 OK`
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "total_reports_count": 0, // Total reports before filtering
            "total_emails_count": 0,  // Total emails after filtering
            "total_domains_count": 0, // Unique domains after filtering
            "min_date": 0,            // Min date (Unix timestamp) after filtering
            "max_date": 0,            // Max date (Unix timestamp) after filtering
            "disposition_summary": {  // Disposition breakdown
                "none": 0,
                "quarantine": 0,
                "reject": 0
            },
            "timeseries_data": [      // Timeseries data (emails per day by disposition)
                {"date": "YYYY-MM-DD", "none": 0, "quarantine": 0, "reject": 0}
            ],
            "source_ip_summary": [    // Top 10 source IPs (with DMARC Pass/Fail breakdown)
                {"ip": "string", "total": 0, "pass": 0, "fail": 0, "info": {"as_name": "string", "reverse_domain": "string"}}
            ],
            "country_summary": [      // Country summary
                {"key": "string", "total": 0, "dmarc_pass": 0, "dmarc_fail": 0, "spf_pass": 0, "spf_fail": 0, "dkim_pass": 0, "dkim_fail": 0}
            ],
            "as_summary": [           // AS summary
                {"key": "string", "total": 0, "dmarc_pass": 0, "dmarc_fail": 0, "spf_pass": 0, "spf_fail": 0, "dkim_pass": 0, "dkim_fail": 0}
            ],
            "domain_summary": [       // Reverse domain summary
                {"key": "string", "total": 0, "dmarc_pass": 0, "dmarc_fail": 0, "spf_pass": 0, "spf_fail": 0, "dkim_pass": 0, "dkim_fail": 0}
            ],
            "records": [              // Detailed records (aggregated or individual)
                {
                    "id": "string",
                    "source_ip": "string",
                    "count": 0,
                    "disposition": "string",
                    "dkim_result": "string",
                    "spf_result": "string",
                    "header_from": "string",
                    "ip_info": {"country_name": "string", "as_name": "string", "reverse_domain": "string"},
                    "report_org_name": "string", // Only if aggregate=false
                    "report_id": "string"        // Only if aggregate=false
                }
            ]
        }
        ```
    *   **Status:** `401 Unauthorized` (Invalid JWT)
*   **Remarks:**
    *   The backend performs filtering, aggregation, and sorting based on query parameters.
    *   `records` array contains either aggregated or individual records based on the `aggregate` parameter.
    *   `ip_info` is enriched from the `ip_info` table.

### 2.4. Specific Record Analysis Data Retrieval

*   **Purpose:** Retrieves detailed information for a specific DMARC record, used for the analysis modal.
*   **HTTP Method:** `GET`
*   **Path:** `/api/records/{record_id}/analyze`
*   **Header:** `Authorization: Bearer <JWT>` (Requires a valid JWT)
*   **Request:**
    *   **Path Parameter:** `record_id` (ID of the detailed record)
*   **Response:**
    *   **Status:** `200 OK`
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "record": { // Data from records table and associated IP info
                "id": "string",
                "source_ip": "string",
                "count": 0,
                "disposition": "string",
                "dkim_evaluated_result": "string",
                "spf_evaluated_result": "string",
                "header_from": "string",
                "auth_dkim_domain": "string",
                "auth_dkim_result": "string",
                "auth_dkim_selector": "string",
                "auth_spf_domain": "string",
                "auth_spf_result": "string",
                "ip_info": {"country_name": "string", "as_name": "string", "reverse_domain": "string"}
            },
            "report": { // Data from associated reports table (policy info etc.)
                "report_id": "string",
                "org_name": "string",
                "policy_domain": "string",
                "policy_adkim": "string",
                "policy_aspf": "string",
                "policy_p": "string",
                "policy_sp": "string",
                "policy_pct": 0
            },
            "contributing_reports": [ // List of reports contributing to an aggregated record
                {"org_name": "string", "report_id": "string"}
            ]
        }
        ```
    *   **Status:** `404 Not Found` (If `record_id` is not found)
    *   **Status:** `401 Unauthorized` (Invalid JWT)
*   **Remarks:**
    *   This endpoint provides all necessary details for the frontend's analysis modal.
    *   If the requested `record_id` belongs to an aggregated record, `contributing_reports` will list the original reports that formed this aggregation.

### 2.5. MaxMind GeoLite2 Database Download and Update

*   **Purpose:** Triggers the download and update of MaxMind GeoLite2 databases.
*   **HTTP Method:** `POST`
*   **Path:** `/api/maxmind/update`
*   **Header:** `Authorization: Bearer <JWT>` (Requires a valid JWT)
*   **Request:**
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "maxmind_license_key": "string" // Required for initial setup or key change
        }
        ```
*   **Response:**
    *   **Status:** `200 OK`
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "status": "success",
            "message": "MaxMind databases updated successfully.",
            "city_db_updated": true,
            "asn_db_updated": true
        }
        ```
    *   **Status:** `400 Bad Request` (e.g., invalid license key, download failure)
    *   **Status:** `401 Unauthorized` (Invalid JWT)
*   **Remarks:**
    *   This endpoint is called on application startup or when a user manually requests an update.
    *   The license key should be securely stored by the backend (e.g., in a settings table in SQLite).

### 2.6. Application Settings (Get/Update)

*   **Purpose:** Retrieves or updates application configuration settings (e.g., listening port, MaxMind license key).
*   **HTTP Method:** `GET` / `POST`
*   **Path:** `/api/settings`
*   **Header:** `Authorization: Bearer <JWT>` (Requires a valid JWT)
*   **Request (GET):** None
*   **Request (POST):**
    *   **Content-Type:** `application/json`
    *   **Body:**
        ```json
        {
            "port": 0, // Integer
            "maxmind_license_key": "string"
        }
        ```
*   **Response:**
    *   **Status:** `200 OK`
    *   **Content-Type:** `application/json`
    *   **Body:** Current settings information
        ```json
        {
            "port": 8080,
            "maxmind_license_key": "string"
        }
        ```
    *   **Status:** `401 Unauthorized` (Invalid JWT)
*   **Remarks:**
    *   Settings are stored in a dedicated table within the SQLite database.
    *   Changing the listening port may require an application restart.

## 3. Common Error Responses

For all protected API endpoints, if authentication fails (e.g., missing, invalid, or expired JWT), the following response format will be used:

*   **Status:** `401 Unauthorized`
*   **Content-Type:** `application/json`
*   **Body:**
    ```json
    {
        "status": "error",
        "message": "Authentication required or token invalid."
    }
    ```
