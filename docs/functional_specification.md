# DMARC Report Analyzer - Functional Specification

## 1. Overview

DMARC Report Analyzer is a self-contained local server application designed to parse and visualize DMARC (Domain-based Message Authentication, Reporting, and Conformance) aggregate reports. Users can upload DMARC report files (XML, ZIP, GZ formats) and view detailed information about email authentication results, source IP addresses, and domains. The application aims to assist in monitoring and improving DMARC policies.

## 2. Technology Stack

*   **Frontend Framework:** React (TypeScript)
*   **Build Tool:** Vite
*   **Styling:** Tailwind CSS v3
*   **Backend Language:** Go
*   **Backend Web Framework:** Gorilla Mux
*   **Database:** SQLite
*   **File Compression/Decompression:** JSZip (for frontend ZIP handling), pako (for frontend GZ handling)
*   **Graph Drawing:** Chart.js, chartjs-adapter-date-fns (Planned)
*   **UI Operations:** SortableJS (for panel drag & drop) (Planned)

## 3. Key Features

### 3.1. DMARC Report Ingestion

*   **Supported File Formats:** `.xml`, `.zip`, `.gz`
*   **Ingestion Method:**
    *   **File Upload:** Users can select and upload DMARC report files via a web interface.
    *   **Drag & Drop:** (Planned) Users can drag and drop files directly onto a designated area for processing.
    *   **File Selection Dialog:** (Planned) Users can click an area to open a file selection dialog.
*   **Processing Status Display:** (Planned) Display real-time processing status and completion messages.
*   **Duplicate Report Handling:** Reports with identical report IDs that have already been processed will be skipped, ensuring data integrity and preventing redundant entries.

### 3.2. Data Management and Persistence

*   **Data Storage:** All parsed DMARC report data, including associated IP information, is stored persistently in a SQLite database file (`dmarc_reports.db`) managed by the backend.
*   **IP Information Import:** IPInfo.io MMDB files (e.g., `ipinfo-city.mmdb`) can be manually imported via a command-line interface (CLI) option (`--import-ip-db`). This updates the IP geolocation and ASN data used by the application.
*   **Data Import (ZIP):** (Planned) Users can import previously exported data from a ZIP file.
*   **Data Export (ZIP):** (Planned) Users can export current report data as a JSON-formatted ZIP file.
*   **Clear All Data:** (Planned) Users can clear all stored data from the database.

### 3.3. Filtering Functionality (Planned)

*   **Time Range Filter:** Filter reports by start and end dates, with preset options (e.g., Today, Last 7 Days, All Time).
*   **DMARC Authentication Result Filters:** Filter by DMARC, SPF, and DKIM authentication results (Pass/Fail/None).
*   **Disposition Filter:** Filter by DMARC policy disposition (None, Quarantine, Reject).
*   **Keyword Filter:** Search for keywords within IP addresses, AS names, and reverse domain names.
*   **Active Filter Display:** Show currently applied filters and provide options to clear them.
*   **URL Hash State Management:** Manage filter settings via URL hash for bookmarking and sharing.

### 3.4. Dashboard Display (Planned)

*   **Summary Cards:** Display total reports, total emails, and analysis period.
*   **Panel Types:**
    *   **Disposition Breakdown (Pie Chart):** Visualize DMARC policy disposition percentages.
    *   **Time Series Trend (Line Chart):** Show daily email trends by disposition.
    *   **Top 10 Source IPs (Bar Chart):** Display top source IPs by email count, with DMARC Pass/Fail breakdown.
    *   **Summary Tables:** Aggregate data by Country, AS (Autonomous System), and Reverse Domain.
    *   **Detailed Records Table:** List individual DMARC records with options for aggregation and sorting.
*   **Panel Customization:** Users can show/hide and reorder dashboard panels via drag & drop.

### 3.5. Detailed Analysis Modal (Planned)

*   **Display Content:** Show detailed information for a selected DMARC record, including report origin, IP analysis, authentication details (SPF/DKIM results, alignment), and policy evaluation.
*   **Improvement Advice:** Provide context-sensitive advice based on DMARC authentication results.
*   **Gemini Prompt Generation:** Automatically generate a detailed prompt for consulting Gemini based on the analysis results.

### 3.6. Confirmation Modal (Planned)

*   A generic confirmation dialog for destructive operations (e.g., clearing data).

## 4. External Integration

*   **ipinfo.io:** Used for IP geolocation and ASN information. MMDB files are manually imported via CLI.

## 5. UI/UX Considerations (Planned)

*   **Responsive Design:** Adapt to various screen sizes.
*   **Custom Scrollbars:** Enhanced visual consistency.
*   **Custom Toggle Switches:** For interactive controls.
*   **Fonts:** Use Inter and Noto Sans JP for multilingual support.
*   **Interactive Element Feedback:** Visual cues for clickable and selected elements.
*   **Confirmation Dialogs:** Prevent accidental destructive operations.
*   **Empty State Display:** Provide clear messages when no data is available or filtered.