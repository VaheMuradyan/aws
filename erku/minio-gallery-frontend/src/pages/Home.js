import React, { useState } from 'react';
import { Container } from 'react-bootstrap';
import Gallery from '../components/Gallery';
import UploadForm from '../components/UploadForm';

function Home() {
  // refreshTrigger-ի օգտագործում պատկերասրահը թարմացնելու համար
  const [refreshGallery, setRefreshGallery] = useState(0);

  // Վերբեռնումից հետո պատկերասրահի թարմացում
  const handleUploadSuccess = () => {
    setRefreshGallery(prev => prev + 1);
  };

  return (
    <Container className="py-4">
      <UploadForm onUploadSuccess={handleUploadSuccess} />
      <Gallery refreshTrigger={refreshGallery} />
    </Container>
  );
}

export default Home;