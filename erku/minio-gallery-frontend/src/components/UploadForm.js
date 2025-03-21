import React, { useState, useCallback } from 'react';
import { Form, Button, Card, Alert, Spinner } from 'react-bootstrap';
import { useDropzone } from 'react-dropzone';
import { uploadImage } from '../services/api';

const UploadForm = ({ onUploadSuccess }) => {
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [tags, setTags] = useState('');

  const onDrop = useCallback((acceptedFiles) => {
    const selectedFile = acceptedFiles[0];
    
    if (selectedFile) {
      setFile(selectedFile);
      
      // Նախադիտման URL-ի ստեղծում
      const previewUrl = URL.createObjectURL(selectedFile);
      setPreview(previewUrl);
      
      // Սխալի և հաջողության հաղորդագրությունների մաքրում
      setError('');
      setSuccess('');
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({ 
    onDrop,
    accept: {
      'image/*': []
    },
    maxFiles: 1
  });

  const handleUpload = async (e) => {
    e.preventDefault();
    
    if (!file) {
      setError('Խնդրում ենք ընտրել նկար վերբեռնելու համար');
      return;
    }
    
    setUploading(true);
    setError('');
    
    try {
      await uploadImage(file, tags);
      setSuccess('Նկարը հաջողությամբ վերբեռնվել է');
      setFile(null);
      setPreview(null);
      setTags('');
      
      // Ծնող կոմպոնենտին տեղեկացում վերբեռնման հաջողության մասին
      if (onUploadSuccess) {
        onUploadSuccess();
      }
    } catch (err) {
      setError(`Վերբեռնման սխալ: ${err.message || 'Անհայտ սխալ'}`);
    } finally {
      setUploading(false);
    }
  };
  
  const handleClear = () => {
    setFile(null);
    setPreview(null);
    setError('');
    setSuccess('');
    setTags('');
  };

  return (
    <Card className="upload-card mb-4">
      <Card.Header as="h5">Նկարի վերբեռնում</Card.Header>
      <Card.Body>
        {error && <Alert variant="danger">{error}</Alert>}
        {success && <Alert variant="success">{success}</Alert>}
        
        <div 
          {...getRootProps()} 
          className={`dropzone mb-3 ${isDragActive ? 'active' : ''}`}
        >
          <input {...getInputProps()} />
          
          {preview ? (
            <div className="preview-container">
              <img 
                src={preview} 
                alt="Նախադիտում" 
                className="img-preview" 
              />
            </div>
          ) : (
            <div className="dropzone-content">
              <p>Քաշեք և գցեք նկարը այստեղ, կամ սեղմեք ընտրելու համար</p>
            </div>
          )}
        </div>
        
        {file && (
          <div className="selected-file mb-3">
            <strong>Ընտրված ֆայլ:</strong> {file.name} ({(file.size / 1024).toFixed(2)} KB)
          </div>
        )}
        
        <Form.Group className="mb-3">
          <Form.Label>Թեգեր (ստորակետով բաժանված)</Form.Label>
          <Form.Control
            type="text"
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="Օր.՝ բնապատկեր, արվեստ, կենդանիներ"
            disabled={uploading}
          />
          <Form.Text className="text-muted">
            Օգտագործեք ստորակետներ մի քանի թեգեր ավելացնելու համար
          </Form.Text>
        </Form.Group>
        
        <div className="d-flex">
          <Button 
            variant="primary"
            onClick={handleUpload}
            disabled={!file || uploading}
            className="me-2"
          >
            {uploading ? (
              <>
                <Spinner as="span" animation="border" size="sm" className="me-2" />
                Վերբեռնում...
              </>
            ) : 'Վերբեռնել'}
          </Button>
          
          {file && (
            <Button 
              variant="outline-secondary"
              onClick={handleClear}
              disabled={uploading}
            >
              Մաքրել
            </Button>
          )}
        </div>
      </Card.Body>
    </Card>
  );
};

export default UploadForm;