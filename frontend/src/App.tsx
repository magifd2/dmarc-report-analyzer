import React, { useState } from 'react';
import ReportList from './components/ReportList'; // Import the new component

function App() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState<boolean>(false);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      setSelectedFile(event.target.files[0]);
      setMessage(null); // Clear previous messages
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) {
      setMessage('Please select a file first!');
      return;
    }

    setIsUploading(true);
    setMessage('Uploading...');

    const formData = new FormData();
    formData.append('file', selectedFile);

    try {
      const response = await fetch('/api/reports/upload', {
        method: 'POST',
        body: formData,
      });

      const data = await response.json();

      if (response.ok) {
        if (data.status === 'success') {
          let successMessage = `Upload successful! Processed: ${data.processed_count}, Skipped: ${data.skipped_count}.`;
          if (data.failed_files_count > 0) {
            successMessage += ` Failed: ${data.failed_files_count}.`;
          }
          setMessage(successMessage);
          // Optionally, refresh the report list after successful upload
          // You might need to pass a refresh function down to ReportList or use a global state management
        } else {
          setMessage(`Upload failed: ${data.message || 'Unknown error'}`);
          if (data.file_errors && data.file_errors.length > 0) {
            data.file_errors.forEach((err: any) => {
              setMessage(prev => `${prev}\nFile: ${err.filename}, Type: ${err.error_type}, Message: ${err.message}`);
            });
          }
        }
      } else {
        setMessage(`Server error: ${response.status} ${response.statusText} - ${data.message || 'Unknown error'}`);
      }
    } catch (error) {
      setMessage(`Network error: ${error instanceof Error ? error.message : String(error)}`);
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 p-4">
      <div className="max-w-4xl mx-auto bg-white p-8 rounded-lg shadow-md">
        <h1 className="text-3xl font-bold text-center text-gray-800 mb-8">DMARC Report Analyzer</h1>
        
        {/* Upload Section */}
        <div className="mb-8 border-b pb-6">
          <h2 className="text-2xl font-bold text-gray-800 mb-4">Upload New Report</h2>
          <div className="mb-4">
            <label htmlFor="file-upload" className="block text-sm font-medium text-gray-700 mb-2">
              Select DMARC XML/ZIP/GZ File:
            </label>
            <input
              id="file-upload"
              type="file"
              accept=".xml,.zip,.gz"
              onChange={handleFileChange}
              className="block w-full text-sm text-gray-500
                file:mr-4 file:py-2 file:px-4
                file:rounded-full file:border-0
                file:text-sm file:font-semibold
                file:bg-blue-50 file:text-blue-700
                hover:file:bg-blue-100"
              disabled={isUploading}
            />
            {selectedFile && (
              <p className="mt-2 text-sm text-gray-600">Selected: {selectedFile.name}</p>
            )}
          </div>
          <button
            onClick={handleUpload}
            className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            disabled={isUploading}
          >
            {isUploading ? 'Uploading...' : 'Upload Report'}
          </button>
          {message && (
            <div className={`mt-4 p-3 rounded-md text-sm ${message.startsWith('Upload successful') ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
              {message.split('\n').map((line, index) => (
                <p key={index}>{line}</p>
              ))}
            </div>
          )}
        </div>

        {/* Report List Section */}
        <ReportList />
      </div>
    </div>
  );
}

export default App;
