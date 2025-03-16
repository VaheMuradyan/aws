import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Button, Container, Row, Col, Badge, Form, Alert, Spinner } from 'react-bootstrap';
import { fetchImageById, addTagsToImage, deleteImage } from '../services/api';

const ImageDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [image, setImage] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [newTag, setNewTag] = useState('');
  const [addingTag, setAddingTag] = useState(false);
  const [tagError, setTagError] = useState('');
  const [deleteConfirm, setDeleteConfirm] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    loadImage();
  }, [id]);

  const loadImage = async () => {
    try {
      setLoading(true);
      const data = await fetchImageById(id);
      setImage(data);
      setError('');
    } catch (err) {
      setError('Նկարը բեռնելու ժամանակ առաջացել է սխալ');
    } finally {
      setLoading(false);
    }
  };

  const handleAddTag = async (e) => {
    e.preventDefault();
    
    if (!newTag.trim()) {
      setTagError('Խնդրում ենք մուտքագրել թեգ');
      return;
    }
    
    try {
      setAddingTag(true);
      setTagError('');
      
      await addTagsToImage(id, [newTag.trim()]);
      
      // Վերաբեռնել նկարի տվյալները
      await loadImage();
      
      setNewTag('');
    } catch (err) {
      setTagError('Թեգի ավելացման ժամանակ առաջացել է սխալ');
    } finally {
      setAddingTag(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteConfirm) {
      setDeleteConfirm(true);
      return;
    }
    
    try {
      setDeleting(true);
      await deleteImage(id);
      navigate('/');
    } catch (err) {
      setError('Նկարը ջնջելու ժամանակ առաջացել է սխալ');
      setDeleting(false);
      setDeleteConfirm(false);
    }
  };

  if (loading) {
    return (
      <Container className="text-center py-5">
        <Spinner animation="border" variant="primary" />
        <p className="mt-3">Նկարը բեռնվում է...</p>
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

  if (!image) {
    return (
      <Container className="py-4">
        <Alert variant="warning">Նկարը չի գտնվել</Alert>
        <Button variant="primary" onClick={() => navigate('/')}>Վերադառնալ գլխավոր էջ</Button>
      </Container>
    );
  }

  return (
    <Container className="py-4">
      <Row>
        <Col md={8}>
          <Card className="mb-4">
            <Card.Img 
              variant="top" 
              src={image.url} 
              alt={image.name} 
              className="img-fluid"
            />
            <Card.Body>
              <Card.Title>{image.name}</Card.Title>
              <Card.Text>
                <small className="text-muted">
                  Վերբեռնվել է: {new Date(image.uploadedAt).toLocaleString()}
                </small>
              </Card.Text>
              <div className="mb-3">
                <strong>Չափը:</strong> {(image.size / 1024).toFixed(2)} KB
              </div>
              <div className="mb-3">
                <strong>Տեսակը:</strong> {image.contentType}
              </div>
              <div className="d-flex flex-wrap gap-2 mb-3">
                {image.tags && image.tags.map((tag, index) => (
                  <Badge key={index} bg="primary" onClick={() => navigate(`/tags/${tag}`)} style={{ cursor: 'pointer' }}>
                    {tag}
                  </Badge>
                ))}
                {(!image.tags || image.tags.length === 0) && <span className="text-muted">Թեգեր չկան</span>}
              </div>
              
              <Form onSubmit={handleAddTag} className="mb-3">
                <Form.Group className="mb-2">
                  <Form.Label>Նոր թեգի ավելացում</Form.Label>
                  <div className="d-flex">
                    <Form.Control 
                      type="text" 
                      value={newTag} 
                      onChange={(e) => setNewTag(e.target.value)} 
                      placeholder="Մուտքագրեք թեգը"
                      disabled={addingTag}
                    />
                    <Button 
                      type="submit" 
                      variant="success" 
                      className="ms-2" 
                      disabled={addingTag}
                    >
                      {addingTag ? <Spinner as="span" animation="border" size="sm" /> : 'Ավելացնել'}
                    </Button>
                  </div>
                  {tagError && <small className="text-danger">{tagError}</small>}
                </Form.Group>
              </Form>
              
              <div className="mt-4">
                <Button 
                  variant={deleteConfirm ? "danger" : "outline-danger"} 
                  onClick={handleDelete}
                  disabled={deleting}
                >
                  {deleting ? (
                    <Spinner as="span" animation="border" size="sm" />
                  ) : deleteConfirm ? (
                    'Հաստատել ջնջումը'
                  ) : (
                    'Ջնջել նկարը'
                  )}
                </Button>
                {deleteConfirm && !deleting && (
                  <Button variant="secondary" className="ms-2" onClick={() => setDeleteConfirm(false)}>
                    Չեղարկել
                  </Button>
                )}
              </div>
            </Card.Body>
          </Card>
        </Col>
        <Col md={4}>
          <Card>
            <Card.Body>
              <Button variant="primary" className="w-100 mb-3" onClick={() => navigate('/')}>
                Վերադառնալ պատկերասրահ
              </Button>
              <Button variant="outline-primary" className="w-100" href={image.url} download={image.name} target="_blank">
                Ներբեռնել նկարը
              </Button>
            </Card.Body>
          </Card>
        </Col>
      </Row>
    </Container>
  );
};

export default ImageDetails;