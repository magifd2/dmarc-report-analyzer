# DMARC Report Analyzer - Functional Specification

## 1. Overview

DMARC Report Analyzer is a self-contained local server application designed to parse and visualize DMARC (Domain-based Message Authentication, Reporting, and Conformance) aggregate reports. Users can upload DMARC report files (XML, ZIP, GZ formats) and view detailed information about email authentication results, source IP addresses, and domains. The application aims to assist in monitoring and improving DMARC policies.

## 2. Technology Stack

*   **Frontend Framework:** React (TypeScript) (Implemented)
*   **Build Tool:** Vite (Implemented)
*   **Styling:** Tailwind CSS v3 (Implemented)
*   **Backend Language:** Go (Implemented)
*   **Backend Web Framework:** Gorilla Mux (Implemented)
*   **Database:** SQLite (Implemented)
*   **File Compression/Decompression:**
    *   JSZip (for frontend ZIP handling) (Planned - frontend side)
    *   pako (for frontend GZ handling) (Planned - frontend side)
*   **Graph Drawing:** Chart.js, chartjs-adapter-date-fns (Planned)
*   **UI Operations:** SortableJS (for panel drag & drop) (Planned)
*   **IP Information Source:** ipinfo.io (MMDB via CLI import) (Implemented - backend side)

## 3. Key Features

### 3.1. DMARC Report Ingestion

*   **Supported File Formats:** `.xml`, `.zip`, `.gz` (Implemented - backend parsing)
*   **Ingestion Method:**
    *   **File Upload:** Users can select and upload DMARC report files via a web interface. (Implemented - backend API, basic frontend UI)
    *   **Drag & Drop:** (Planned) Users can drag and drop files directly onto a designated area for processing. During drag-over, the area's border changes to blue (`border-blue-400`) and background to dark gray (`bg-gray-600`) for visual feedback.
    *   **File Selection Dialog:** (Planned) Clicking the DMARC report area triggers a hidden file input (`#file-input`) to open a file selection dialog. The `accept` attribute restricts selection to supported file formats.
*   **Processing Status Display:** (Planned)
    *   During file processing, a "Processing..." message is displayed in the sidebar status area (`#processing-status`).
    *   Upon completion, a message like "Completed: X new reports added, Y duplicates skipped." is shown for 5 seconds.
*   **Duplicate Report Handling:** Files with report IDs already loaded will be skipped, and only new reports will be added. (Implemented - backend)

### 3.2. Data Management and Persistence

*   **Data Storage:** All parsed DMARC report data, including associated IP information, is stored persistently in a SQLite database file (`dmarc_reports.db`) managed by the backend. (Implemented - backend)
*   **IP Information Import (CLI):** IPInfo.io MMDB files (e.g., `ipinfo-city.mmdb`) can be manually imported via a command-line interface (CLI) option (`--import-ip-db`). This updates the IP geolocation and ASN data used by the application. (Implemented - backend)
*   **Data Import (ZIP):** (Planned) Clicking the "Load Data (ZIP)" button (`#import-btn`) triggers a hidden file input (`#import-input`) to open a ZIP file selection dialog. If the selected ZIP contains `dmarc-analyzer-data.json`, its content is read and merged with existing data, skipping duplicates. Alerts are shown on error.
*   **Data Export (ZIP):** (Planned) Clicking the "Save Data (ZIP)" button (`#export-btn`) exports all current report data as a JSON-formatted ZIP file named `dmarc-analyzer-data_YYYY-MM-DD.zip`. An alert is shown if no data is available for export.
*   **Clear All Data:** (Planned) Clicking the "Clear All Data" button (`#clear-data-btn`) displays a confirmation dialog (`#confirm-modal`). If confirmed, all local storage data is cleared, and the page reloads. This action is irreversible.

### 3.3. Filtering Functionality (Planned)

Filters are located in the dashboard header area (`#main-header`) to narrow down the analyzed report data. The dashboard updates in real-time after filter application.

*   **Time Range Filter:**
    *   **Start/End Date Inputs (`#start-date`, `#end-date`):** Use HTML5 `type="date"` to specify a period.
    *   **Preset Buttons (`.preset-btn`):** "Today", "Yesterday", "7 Days", "30 Days", "90 Days", "1 Year", "This Year", "Last Year", "All Time" presets for quick filtering. Clicking a preset updates date inputs automatically. The active preset button has a blue background (`bg-blue-500`) and white text (`text-white`). Manual date changes clear the active preset.
*   **DMARC Authentication Result Filters:**
    *   **DMARC (`#dmarc-filter`):** Dropdown for "All", "Pass", "Fail".
    *   **SPF (Align) (`#spf-filter`):** Dropdown for "All", "Pass", "Fail".
    *   **DKIM (Align) (`#dkim-filter`):** Dropdown for "All", "Pass", "Fail".
*   **Disposition Filter (`#disposition-filter`):** Dropdown for "All", "None", "Quarantine", "Reject".
*   **Keyword Filter (`#keyword-filter`):** Text input for searching IP addresses, AS names, and reverse domain names. Input is converted to lowercase for searching.
*   **Action Buttons:**
    *   **Apply (`#filter-btn`):** Applies current filter settings and updates the dashboard. Primarily used after manual date changes.
    *   **Reset (`#reset-filter-btn`):** Resets all filters to default (default period, all authentication results, no keyword) and updates the UI.
*   **Active Filter Display (`#active-filter-display`):** Displays currently applied filter conditions (keyword, AS, domain, country, DMARC/SPF/DKIM/Disposition) compactly below the header. A "Clear" button next to displayed conditions clears all drill-down filters.
*   **URL Hash State Management:** Filter settings are encoded as JSON in the URL hash, supporting browser back/forward buttons and URL sharing. `history.pushState` and `history.replaceState` are used for proper browser history management. `popstate` event listener detects URL hash changes and automatically applies filters.

### 3.4. Dashboard Display (Planned)

Analyzed DMARC report aggregates are displayed in multiple panels (`.panel`). Panels can be reordered via drag & drop and toggled visible/hidden.

*   **Summary Cards:**
    *   `#total-reports`, `#total-emails`, `#total-domains`, `#date-range` elements dynamically display "Total Reports", "Total Emails", "Target Domains", and "Analysis Period" respectively. Email counts are formatted with `toLocaleString()`.
*   **Panel Types:**
    *   **Disposition Breakdown (Pie Chart) (`#dispositionChart`):** Displays the percentage of DMARC policy results (None, Quarantine, Reject) in a pie chart. Colors are green for None, red for Reject, orange for Quarantine. Clicking a segment filters the dashboard by that disposition.
    *   **Time Series Trend (Line Chart) (`#timeseriesChart`):** Shows daily email counts by disposition as a stacked line chart. X-axis uses a time scale (`type: 'time', unit: 'day'`).
    *   **Top 10 Source IPs (Bar Chart) (`#sourceIpChart`):** Displays the top 10 source IP addresses by email count as a horizontal bar chart. Each bar shows DMARC Pass/Fail breakdown. Clicking a bar sets that IP as a keyword filter. Labels include IP address, reverse hostname, and AS name.
    *   **Summary Tables:**
        *   **Country Summary (`#country-summary-table-body`):** Aggregation by country of source IP addresses.
        *   **AS (Organization) Summary (`#as-summary-table-body`):** Aggregation by Autonomous System (AS) of source IP addresses.
        *   **Reverse Domain Summary (`#domain-summary-table-body`):** Aggregation by reverse domain of source IP addresses.
        *   Each summary table displays key (country name, AS name, domain), total email count, and DMARC/SPF/DKIM Pass/Fail percentages and counts. Pass is green (`pass` class), Fail is red (`fail` class), 0% is gray (`zero` class). Clicking a table row filters the dashboard by that country/AS/domain. Selected rows have a light blue background (`bg-blue-100`). Key text (`.drilldown-key`) is blue and underlined on hover.
    *   **Detailed Records (Table) (`#records`):**
        *   Lists individual DMARC record details.
        *   **Record Aggregation Toggle (`#aggregation-toggle`):** A custom toggle switch (`.toggle-switch`) to switch between aggregated (default) and individual record display. Aggregated records combine email counts for identical source IP, authentication results, and From domains. Aggregated records show an icon indicating they are combined from multiple reports. Toggle state is saved to local storage.
        *   **Sorting:** Clicking column headers (`.sortable-th`) sorts the column ascending/descending. Sort direction is indicated by "▲" (ascending) or "▼" (descending). Sort state is saved to local storage.
        *   **Displayed Columns:** Source Info (IP, reverse hostname, AS name), Email Count, Disposition (colored and clickable), DKIM/SPF (colored), From Domain (clickable), Report Organization Name/Report ID (only when aggregation is off).
        *   **Action:** "Analyze" button (`.analyze-btn`) on each row opens the detailed analysis modal.
        *   For performance, displayed records are limited to a maximum of 500.

### 3.5. Dashboard Customization (Planned)

*   **Panel Visibility:**
    *   The sidebar's "Panel Display Settings" area (`#panel-visibility-controls`) shows each panel's title with a custom toggle switch.
    *   Toggling controls the `display` style of the corresponding panel (`.panel`) to `none` or empty, controlling visibility.
    *   Clicking the eye icon (`.visibility-toggle`) in each panel header also hides that panel.
*   **Panel Reordering:**
    *   Panels can be reordered within the dashboard grid (`#dashboard-grid`) by dragging and dropping their grid icons (`.panel-handle`).
    *   The SortableJS library provides this functionality.
*   **Layout Persistence:** Panel visibility settings and order are saved to local storage (`dmarcAnalyzerLayout`) and restored on subsequent visits.

### 3.6. Detailed Analysis Modal (`#analysis-modal`) (Planned)

Displayed when clicking the "Analyze" button in the "Detailed Records" panel. Appears centered with a semi-transparent overlay (`bg-opacity-50`).

*   **Display Content:**
    *   **Report Origin Information:** For aggregated records, lists all contributing report organization names and report IDs in a scrollable area (`max-h-40`).
    *   **Analyzed IP:** IP address, reverse hostname, AS information.
    *   **Judgment Summary:** From header, DMARC evaluation (Pass/Fail with color coding), final disposition.
    *   **Authentication Details:** SPF and DKIM authentication results, alignment results, related domains, DKIM selector. Results are color-coded (green for `pass`, red for `fail`, gray for `N/A`).
*   **Improvement Advice:** Context-specific advice generated based on DMARC authentication results. Background and border colors vary by severity (success: green `bg-green-100`, warning: yellow `bg-yellow-100`, danger: red `bg-red-100`).
*   **Gemini Consultation Prompt:**
    *   A detailed prompt for querying Gemini is automatically generated based on the current analysis results (IP, From domain, DMARC evaluation, disposition, SPF/DKIM details, source info, DMARC policy).
    *   The prompt is displayed in a read-only textarea.
    *   Clicking the "Copy Prompt for Consultation" button (`#copy-prompt-btn`) copies the prompt to the clipboard. The button text changes to "Copied!" and is temporarily disabled after copying.
*   **Close Button (`#close-modal-btn`):** Closes the modal.

### 3.7. Confirmation Modal (`#confirm-modal`) (Planned)

A generic confirmation dialog displayed before destructive operations (e.g., "Clear All Data"). Appears centered with a semi-transparent overlay.

*   **Display Content:**
    *   Warning icon (red SVG).
    *   Dynamically set title (`#confirm-modal-title`) and message (`#confirm-modal-message`).
    *   "Execute" button (`#confirm-modal-confirm-btn`) and "Cancel" button (`#confirm-modal-cancel-btn`).
*   **Behavior:**
    *   Clicking "Execute" runs the registered callback function and closes the modal.
    *   Clicking "Cancel" closes the modal.

## 4. External Integration

*   **ipinfo.io:** Used for obtaining geographical information (country, AS name, reverse hostname) for source IP addresses.
    *   **Manual MMDB Import:** The application does not automatically download or update the IPInfo database. Users are responsible for manually downloading the latest `.mmdb` file from IPInfo's website and importing it via a CLI option (`--import-ip-db <path_to_mmdb_file>`). This file is then copied to the application's data directory. (Implemented - backend)
    *   **Frontend Caching (Planned):** API calls (if any are made directly from frontend) will be cached in browser local storage (`dmarcIpInfoCache`) to avoid duplicate requests. A 200ms delay between requests will be implemented to respect API rate limits.

## 5. UI/UX Considerations (Planned)

*   **Responsive Design:** Layouts adapt to various screen sizes using Tailwind CSS utility classes (`md:`, `lg:`, `xl:`).
*   **Custom Scrollbars:** `::-webkit-scrollbar` pseudo-elements are used to customize scrollbar width, track, and thumb for improved visibility and design consistency.
*   **Custom Toggle Switches:**
    *   Used for "Aggregate Records" and "Panel Display Settings". Standard checkboxes are hidden, and custom `slider` elements with `before` pseudo-elements provide visual representation.
    *   Background color and knob position change based on checked state.
*   **Fonts:** `Inter` (English) and `Noto Sans JP` (Japanese) are loaded from Google Fonts and applied to the `body` for multilingual support and high readability.
*   **Interactive Element Feedback:**
    *   Table rows with `summary-table-row` class change to light blue (`hover:bg-blue-50`) on hover and darker blue (`active`, `bg-blue-100`) when selected.
    *   Elements with `clickable` class (IP addresses, Disposition) show an underline on hover to indicate clickability.
    *   Elements with `drilldown-key` class (summary table keys) are blue (`text-blue-600`) and bold (`font-semibold`), with an underline on hover.
*   **Confirmation Dialogs:** Destructive operations (e.g., data clearing) require confirmation dialogs to prevent accidental user actions and enhance safety.
*   **Empty State Display:** When no data is loaded or filters result in zero data, each dashboard element (summary cards, charts, tables) displays a message like "No data available."
