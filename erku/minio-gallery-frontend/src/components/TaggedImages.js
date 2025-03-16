import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Container, Row, Col, Alert, Spinner, Card, Button } from 'react-bootstrap';
import { fetchImagesByTag } from '../services/api';
import ImageCard from './ImageCard';

const TaggedImages = () => {
  const { tag } = useParams();
  const navigate = useNavigate();
  const [images, setImages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const loadImages = async () => {
      try {
        setLoading(true);
        const data = await fetchImagesByTag(tag);
        setImages(Array.isArray(data) ? data : []);
        setError('');
      } catch (err) {
        setError(`Թեգով "${tag}" նկարները բեռնելու ժամանակ առաջացել է սխալ`);
      } finally {
        setLoading(false);
      }
    };

    loadImages();
  }, [tag]);

  if (loading) {
    return (
      <Container className="text-center py-5">
        <Spinner animation="border" variant="primary" />
        <p className="mt-3">Նկարները բեռնվում են...</p>
      </Container>
    );
  }

  if (error) {
    return (
      <Container className="py-4">
        <Alert variant="danger">{error}</Alert>
        <Button variant="primary" onClick={() => navigate('/')}>Վերադառնալ գլխավոր էջ</Button>
      </Container>
    );
  }

  return (
    <Container className="py-4">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <h3>Նկարներ "{tag}" թեգով</h3>
        <Button variant="primary" onClick={() => navigate('/')}>Վերադառնալ գլխավոր էջ</Button>
      </div>

      {images.length === 0 ? (
        <Alert variant="info">Այս թեգով նկարներ չեն գտնվել</Alert>
      ) : (
        <Row>
          {images.map((image, index) => (
            <Col key={index} xs={12} sm={6} md={4} lg={3} className="mb-4">
              <Card className="h-100 image-card">
                <Card.Img 
                  variant="top" 
                  src={image.url} 
                  alt={image.name} 
                  style={{ height: '200px', objectFit: 'cover' }}
                />
                <Card.Body className="d-flex flex-column">
                  <Card.Title className="text-truncate">{image.name}</Card.Title>
                  <div className="mt-auto">
                    <Button 
                      variant="primary" 
                      onClick={() => navigate(`/images/${image.id}`)}
                      className="w-100"
                    >
                      Դիտել մանրամասները
                    </Button>
                  </div>
                </Card.Body>
              </Card>
            </Col>
          ))}
        </Row>
      )}
    </Container>
  );
};

export default TaggedImages;