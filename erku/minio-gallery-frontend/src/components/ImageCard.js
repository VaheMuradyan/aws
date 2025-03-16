import React from 'react';
import { Card, Button, Badge } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';

const ImageCard = ({ image }) => {
  const navigate = useNavigate();

  return (
    <Card className="h-100 image-card shadow">
      <Card.Img 
        variant="top" 
        src={image.url} 
        alt={image.name} 
        className="image-preview"
        style={{ height: '200px', objectFit: 'cover' }}
        onClick={() => navigate(`/images/${image.id}`)}
      />
      <Card.Body className="d-flex flex-column">
        <Card.Title className="image-title text-truncate">{image.name}</Card.Title>
        
        {image.tags && image.tags.length > 0 && (
          <div className="mb-2">
            {image.tags.slice(0, 3).map((tag, index) => (
              <Badge 
                key={index} 
                bg="secondary" 
                className="me-1" 
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(`/tags/${tag}`);
                }}
                style={{ cursor: 'pointer' }}
              >
                {tag}
              </Badge>
            ))}
            {image.tags.length > 3 && <Badge bg="light" text="dark">+{image.tags.length - 3}</Badge>}
          </div>
        )}
        
        <div className="mt-auto">
          <Button 
            variant="primary" 
            className="w-100"
            onClick={() => navigate(`/images/${image.id}`)}
          >
            Դիտել
          </Button>
        </div>
      </Card.Body>
    </Card>
  );
};

export default ImageCard;