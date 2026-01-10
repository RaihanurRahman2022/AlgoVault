
import { Category, Pattern, Problem, User } from '../types';

// Use environment variable for API URL, fallback to localhost for development
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL 
  ? `${import.meta.env.VITE_API_BASE_URL}/api`
  : 'http://localhost:8080/api';

const getAuthHeaders = () => {
  const token = localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    'Authorization': token ? `Bearer ${token}` : '',
  };
};

const handleResponse = async (response: Response) => {
  if (response.status === 401) {
    // If we're on the login page, don't trigger a reload loop
    if (!window.location.pathname.includes('login')) {
      localStorage.removeItem('token');
      window.location.reload();
    }
    // Try to get error message from response
    try {
      const error = await response.json();
      throw new Error(error.error || 'Unauthorized');
    } catch {
      throw new Error('Unauthorized');
    }
  }
  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.error || `API Request failed with status ${response.status}`);
  }
  return response.json();
};

export const api = {
  // Auth
  login: async (email: string, password: string): Promise<{ token: string, user: User }> => {
    const response = await fetch(`${API_BASE_URL}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    return handleResponse(response);
  },

  register: async (email: string, password: string, name: string): Promise<{ token: string, user: User }> => {
    const response = await fetch(`${API_BASE_URL}/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password, name }),
    });
    return handleResponse(response);
  },
  
  // Categories
  getCategories: async (): Promise<Category[]> => {
    const response = await fetch(`${API_BASE_URL}/categories`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  createCategory: async (category: Partial<Category>): Promise<Category> => {
    const response = await fetch(`${API_BASE_URL}/categories`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(category),
    });
    return handleResponse(response);
  },

  updateCategory: async (id: string, category: Partial<Category>): Promise<Category> => {
    const response = await fetch(`${API_BASE_URL}/categories/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(category),
    });
    return handleResponse(response);
  },

  deleteCategory: async (id: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/categories/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  // Patterns
  getPatterns: async (categoryId: string): Promise<Pattern[]> => {
    const response = await fetch(`${API_BASE_URL}/categories/${categoryId}/patterns`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  createPattern: async (categoryId: string, pattern: Partial<Pattern>): Promise<Pattern> => {
    const response = await fetch(`${API_BASE_URL}/categories/${categoryId}/patterns`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(pattern),
    });
    return handleResponse(response);
  },

  updatePattern: async (id: string, pattern: Partial<Pattern>): Promise<Pattern> => {
    const response = await fetch(`${API_BASE_URL}/patterns/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(pattern),
    });
    return handleResponse(response);
  },

  deletePattern: async (id: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/patterns/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },
  
  // Problems
  getProblems: async (patternId: string): Promise<Problem[]> => {
    const response = await fetch(`${API_BASE_URL}/patterns/${patternId}/problems`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  getProblem: async (id: string): Promise<Problem> => {
    const response = await fetch(`${API_BASE_URL}/problems/${id}`, {
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  createProblem: async (patternId: string, problem: Partial<Problem>): Promise<Problem> => {
    const response = await fetch(`${API_BASE_URL}/patterns/${patternId}/problems`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(problem),
    });
    return handleResponse(response);
  },

  updateProblem: async (id: string, problem: Partial<Problem>): Promise<Problem> => {
    const response = await fetch(`${API_BASE_URL}/problems/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(problem),
    });
    return handleResponse(response);
  },

  deleteProblem: async (id: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/problems/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse(response);
  },

  // AI Generation
  generateProblem: async (query: string): Promise<{
    title: string;
    difficulty: 'Easy' | 'Medium' | 'Hard';
    description: string;
    input: string;
    output: string;
    constraints: string;
    sampleInput: string;
    sampleOutput: string;
    explanation: string;
  }> => {
    const response = await fetch(`${API_BASE_URL}/ai/generate-problem`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ query }),
    });
    return handleResponse(response);
  }
};
