import React, { useState } from 'react';
import { Container, Navbar } from 'react-bootstrap';
import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';
import Gallery from './components/Gallery';
import UploadForm from './components/UploadForm';

function App() {
  // refreshTrigger-ի օգտագործում պատկերասրահը թարմացնելու համար
  const [refreshGallery, setRefreshGallery] = useState(0);

  // Վերբեռնումից հետո պատկերասրահի թարմացում
  const handleUploadSuccess = () => {
    setRefreshGallery(prev => prev + 1);
  };

  return (
    <div className="App">
      <Navbar bg="dark" variant="dark" expand="lg">
        <Container>
          <Navbar.Brand>MinIO Նկարների Պատկերասրահ</Navbar.Brand>
        </Container>
      </Navbar>

      <Container className="py-4">
        <UploadForm onUploadSuccess={handleUploadSuccess} />
        <Gallery refreshTrigger={refreshGallery} />
      </Container>

      <footer className="footer bg-light py-3">
        <Container className="text-center">
          <p className="mb-0">MinIO Պատկերասրահ &copy; {new Date().getFullYear()}</p>
        </Container>
      </footer>
    </div>
  );
}

export default App;