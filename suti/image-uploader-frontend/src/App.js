// src/App.js
import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import './App.css';

// Home component - Upload form
function Home() {
  const [selectedFile, setSelectedFile] = useState(null);
  const [preview, setPreview] = useState('');
  const [uploadedImage, setUploadedImage] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleFileChange = (event) => {
    const file = event.target.files[0];
    setSelectedFile(file);
    
    if (file) {
      const reader = new FileReader();
      reader.onloadend = () => {
        setPreview(reader.result);
      };
      reader.readAsDataURL(file);
    } else {
      setPreview('');
    }
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (!selectedFile) {
      setError('Please select an image to upload');
      return;
    }

    setLoading(true);
    setError('');
    
    const formData = new FormData();
    formData.append('image', selectedFile);

    try {
      const response = await fetch('http://localhost:8080/', {
        method: 'POST',
        body: formData,
      });

      if (response.ok) {
        const data = await response.json();
        // Extract image URL from the HTML
        if (data.success && data.url) {
          setUploadedImage(data.url)
        }else{
          setError('Image was uploaded but URL could not be determined')
        }
      } else {
        setError('Failed to upload image');
      }
    } catch (err) {
      setError('Error connecting to server: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container">
      <h1>Image Uploader</h1>
      
      <form onSubmit={handleSubmit} className="upload-form">
        <div className="file-input-container">
          <input 
            type="file" 
            onChange={handleFileChange} 
            accept="image/*"
            id="file-input"
            className="file-input"
          />
          <label htmlFor="file-input" className="file-input-label">
            Choose file
          </label>
          <span className="file-name">
            {selectedFile ? selectedFile.name : 'No file chosen'}
          </span>
        </div>
        
        {preview && (
          <div className="preview-container">
            <h3>Preview:</h3>
            <img src={preview} alt="Preview" className="preview-image" />
          </div>
        )}
        
        <button 
          type="submit" 
          className="upload-button"
          disabled={loading || !selectedFile}
        >
          {loading ? 'Uploading...' : 'Upload'}
        </button>
      </form>
      
      {error && <div className="error-message">{error}</div>}
      
      {uploadedImage && (
        <div className="uploaded-container">
          <h3>Uploaded Image:</h3>
          <img src={uploadedImage} alt="Uploaded" className="uploaded-image" />
          <p>Image URL: <a href={uploadedImage} target="_blank" rel="noopener noreferrer">{uploadedImage}</a></p>
        </div>
      )}
      
      <div className="navigation">
        <Link to="/gallery" className="nav-link">View Gallery</Link>
      </div>
    </div>
  );
}

// Gallery component
function Gallery() {
  const [images, setImages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchImages = async () => {
      try {
        const response = await fetch('http://localhost:8080/images');
        if (response.ok) {
          const html = await response.text();
          const parser = new DOMParser();
          const doc = parser.parseFromString(html, 'text/html');
          
          // Extract image data from HTML
          const imageElements = doc.querySelectorAll('.image-card');
          const imageData = Array.from(imageElements).map(el => {
            const url = el.querySelector('img').src;
            const name = el.querySelector('.image-name').textContent;
            const size = el.querySelector('.image-size').textContent;
            const date = el.querySelector('.image-date').textContent;
            return { url, name, size, date };
          });
          
          setImages(imageData);
        } else {
          setError('Failed to load images');
        }
      } catch (err) {
        setError('Error connecting to server: ' + err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchImages();
  }, []);

  return (
    <div className="container">
      <h1>Image Gallery</h1>
      
      {loading && <div className="loading">Loading images...</div>}
      {error && <div className="error-message">{error}</div>}
      
      <div className="gallery-grid">
        {images.map((image, index) => (
          <div key={index} className="gallery-item">
            <Link to={`/view/${image.name}`}>
              <img src={image.url} alt={image.name} className="gallery-image" />
            </Link>
            <div className="image-info">
              <p className="image-name">{image.name}</p>
              <p className="image-size">{image.size}</p>
              <p className="image-date">{image.date}</p>
            </div>
          </div>
        ))}
      </div>
      
      {images.length === 0 && !loading && !error && (
        <div className="no-images">No images found. Upload some images first.</div>
      )}
      
      <div className="navigation">
        <Link to="/" className="nav-link">Back to Upload</Link>
      </div>
    </div>
  );
}

// ImageView component
function ImageView() {
  const [imageUrl, setImageUrl] = useState('');
  const [filename, setFilename] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchImageDetails = async () => {
      const pathname = window.location.pathname;
      const filename = pathname.substring(pathname.lastIndexOf('/') + 1);
      
      try {
        const response = await fetch(`http://localhost:8080/view/${filename}`);
        if (response.ok) {
          const html = await response.text();
          const parser = new DOMParser();
          const doc = parser.parseFromString(html, 'text/html');
          
          // Extract image URL from HTML
          const imageElement = doc.querySelector('.main-image');
          if (imageElement) {
            setImageUrl(imageElement.src);
            setFilename(filename);
          } else {
            setError('Could not find image details');
          }
        } else {
          setError('Failed to load image');
        }
      } catch (err) {
        setError('Error connecting to server: ' + err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchImageDetails();
  }, []);

  return (
    <div className="container">
      <h1>Image Details</h1>
      
      {loading && <div className="loading">Loading image...</div>}
      {error && <div className="error-message">{error}</div>}
      
      {imageUrl && (
        <div className="image-detail-container">
          <h2>{filename}</h2>
          <img src={imageUrl} alt={filename} className="detail-image" />
          <div className="image-actions">
            <a href={imageUrl} download={filename} className="download-link">Download</a>
          </div>
        </div>
      )}
      
      <div className="navigation">
        <Link to="/gallery" className="nav-link">Back to Gallery</Link>
        <Link to="/" className="nav-link">Upload New Image</Link>
      </div>
    </div>
  );
}

// Main App component
function App() {
  return (
    <Router>
      <div className="app">
        <header className="app-header">
          <div className="logo">Image Uploader</div>
          <nav className="main-nav">
            <Link to="/">Home</Link>
            <Link to="/gallery">Gallery</Link>
          </nav>
        </header>
        
        <main className="app-main">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/gallery" element={<Gallery />} />
            <Route path="/view/:filename" element={<ImageView />} />
          </Routes>
        </main>
        
        <footer className="app-footer">
          <p>&copy; {new Date().getFullYear()} Image Uploader</p>
        </footer>
      </div>
    </Router>
  );
}

export default App;