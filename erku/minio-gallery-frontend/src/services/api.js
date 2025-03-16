import axios from 'axios';

const API_URL = 'http://localhost:8080/api';

// API հարցումների համար axios instance-ի ստեղծում
const apiClient = axios.create({
  baseURL: API_URL,
});

// Բոլոր նկարների ստացում
export const fetchImages = async () => {
  try {
    const response = await apiClient.get('/images');
    return Array.isArray(response.data) ? response.data : [];
  } catch (error) {
    console.error('Error fetching images:', error);
    return [];
  }
};

// Կոնկրետ նկարի ստացում ID-ով
export const fetchImageById = async (id) => {
  try {
    const response = await apiClient.get(`/images/${id}`);
    return response.data;
  } catch (error) {
    console.error(`Error fetching image with ID ${id}:`, error);
    throw error;
  }
};

// Նկարների ստացում թեգով
export const fetchImagesByTag = async (tag) => {
  try {
    const response = await apiClient.get(`/images/tags/${tag}`);
    return Array.isArray(response.data) ? response.data : [];
  } catch (error) {
    console.error(`Error fetching images with tag ${tag}:`, error);
    return [];
  }
};

// Բոլոր թեգերի ստացում
export const fetchAllTags = async () => {
  try {
    const response = await apiClient.get('/tags');
    return Array.isArray(response.data) ? response.data : [];
  } catch (error) {
    console.error('Error fetching tags:', error);
    return [];
  }
};

// Նկարին թեգերի ավելացում
export const addTagsToImage = async (imageId, tags) => {
  try {
    const response = await apiClient.post(`/images/${imageId}/tags`, { tags });
    return response.data;
  } catch (error) {
    console.error(`Error adding tags to image ${imageId}:`, error);
    throw error;
  }
};

// Նկարի վերբեռնում API-ի միջոցով
export const uploadImage = async (file, tags = '') => {
  try {
    const formData = new FormData();
    formData.append('image', file);
    
    if (tags) {
      formData.append('tags', tags);
    }
    
    const response = await apiClient.post('/images', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  } catch (error) {
    console.error('Error uploading image:', error);
    throw error;
  }
};

// Նկարի ջնջում
export const deleteImage = async (id) => {
  try {
    const response = await apiClient.delete(`/images/${id}`);
    return response.data;
  } catch (error) {
    console.error(`Error deleting image ${id}:`, error);
    throw error;
  }
};