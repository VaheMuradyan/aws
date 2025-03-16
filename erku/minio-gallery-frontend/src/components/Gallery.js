// src/components/Gallery.js
import React, { useState, useEffect } from 'react';
import { Row, Col, Alert, Spinner } from 'react-bootstrap';
import { fetchImages } from '../services/api';
import ImageCard from './ImageCard';

const Gallery = ({ refreshTrigger }) => {
  const [images, setImages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Նկարների ցանկի ստացում API-ից
  const loadImages = async () => {
    try {
      setLoading(true);
      const data = await fetchImages();
      // Համոզվենք, որ data-ն զանգված է (array)
      setImages(Array.isArray(data) ? data : []);
      setError('');
    } catch (err) {
      setError('Նկարները բեռնելու ժամանակ առաջացել է սխալ');
      console.error('Error loading images:', err);
      // Սխալի դեպքում դատարկ զանգված սահմանենք images-ի համար
      setImages([]);
    } finally {
      setLoading(false);
    }
  };

  // Նկարների բեռնում էջի բեռնման և refreshTrigger-ի փոփոխման ժամանակ
  useEffect(() => {
    loadImages();
  }, [refreshTrigger]);

  // Նկարների բացակայության դեպքում ցուցադրվող բովանդակություն
  if (loading) {
    return (
      <div className="text-center py-5">
        <Spinner animation="border" variant="primary" />
        <p className="mt-3">Նկարները բեռնվում են...</p>
      </div>
    );
  }

  if (error) {
    return <Alert variant="danger">{error}</Alert>;
  }

  if (!images || images.length === 0) {
    return (
      <div className="text-center py-5">
        <Alert variant="info">
          Դեռևս նկարներ չկան: Վերբեռնեք Ձեր առաջին նկարը:
        </Alert>
      </div>
    );
  }

  return (
    <div className="gallery-container">
      <h3 className="gallery-title mb-4">Պատկերասրահ</h3>
      <Row>
        {images.map((image, index) => (
          <Col key={index} xs={12} sm={6} md={4} lg={3} className="mb-4">
            <ImageCard image={image} />
          </Col>
        ))}
      </Row>
    </div>
  );
};

export default Gallery;