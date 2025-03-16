import React from 'react';
import { Card } from 'react-bootstrap';

// Առանձին նկարի քարտ կոմպոնենտ
const ImageCard = ({ image }) => {
  return (
    <Card className="image-card shadow mb-4">
      <Card.Img 
        variant="top" 
        src={image.url} 
        alt={image.name} 
        className="image-preview"
      />
      <Card.Body>
        <Card.Title className="image-title">{image.name}</Card.Title>
      </Card.Body>
    </Card>
  );
};

export default ImageCard;