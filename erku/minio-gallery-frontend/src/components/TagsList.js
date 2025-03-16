import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container, Row, Col, Alert, Spinner, Card, Badge } from 'react-bootstrap';
import { fetchAllTags } from '../services/api';

const TagsList = () => {
  const navigate = useNavigate();
  const [tags, setTags] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const loadTags = async () => {
      try {
        setLoading(true);
        const data = await fetchAllTags();
        setTags(data);
        setError('');
      } catch (err) {
        setError('Թեգերը բեռնելու ժամանակ առաջացել է սխալ');
      } finally {
        setLoading(false);
      }
    };

    loadTags();
  }, []);

  if (loading) {
    return (
      <Container className="text-center py-5">
        <Spinner animation="border" variant="primary" />
        <p className="mt-3">Թեգերը բեռնվում են...</p>
      </Container>
    );
  }

  if (error) {
    return (
      <Container className="py-4">
        <Alert variant="danger">{error}</Alert>
      </Container>
    );
  }

  if (tags.length === 0) {
    return (
      <Container className="py-4">
        <Alert variant="info">Թեգեր չեն գտնվել</Alert>
      </Container>
    );
  }

  return (
    <Container className="py-4">
      <h3 className="mb-4">Բոլոր թեգերը</h3>
      
      <Card>
        <Card.Body>
          <div className="d-flex flex-wrap gap-2">
            {tags.map((tag) => (
              <Badge 
                key={tag.id} 
                bg="primary" 
                style={{ fontSize: '1rem', cursor: 'pointer', padding: '8px 12px' }}
                onClick={() => navigate(`/tags/${tag.name}`)}
              >
                {tag.name}
              </Badge>
            ))}
          </div>
        </Card.Body>
      </Card>
    </Container>
  );
};

export default TagsList;