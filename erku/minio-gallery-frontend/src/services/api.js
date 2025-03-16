// src/services/api.js
import axios from 'axios';

const API_URL = 'http://localhost:8080/api';

// API հարցումների համար axios instance-ի ստեղծում
const apiClient = axios.create({
  baseURL: API_URL,
});

// Բոլոր նկարների ստացում API-ից
export const fetchImages = async () => {
  try {
    const response = await apiClient.get('/images');
    // Համոզվենք, որ վերադարձրած արժեքը զանգված է
    return Array.isArray(response.data) ? response.data : [];
  } catch (error) {
    console.error('Error fetching images:', error);
    // Սխալի դեպքում դատարկ զանգված վերադարձնենք
    return [];
  }
};

// Նկարի վերբեռնում API-ի միջոցով
export const uploadImage = async (file) => {
  try {
    // Կարգավորենք headers-ը միայն վերբեռնման հարցման համար
    const formData = new FormData();
    formData.append('image', file);
    
    const response = await apiClient.post('/upload', formData, {
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