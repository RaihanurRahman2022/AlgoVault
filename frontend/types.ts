
export interface Category {
  id: string;
  name: string;
  icon: string;
  description: string;
  patternCount?: number;
}

export interface Pattern {
  id: string;
  categoryId: string;
  name: string;
  icon: string;
  description: string;
  theory?: string; // Markdown content
  problemCount?: number;
}

export interface Solution {
  id: string;
  language: 'cpp' | 'go' | 'python' | 'java' | 'javascript';
  code: string;
}

export interface Problem {
  id: string;
  patternId: string;
  title: string;
  difficulty: 'Easy' | 'Medium' | 'Hard';
  description: string; // Markdown
  input: string; // Markdown
  output: string; // Markdown
  constraints: string; // Markdown
  sampleInput: string; // Markdown
  sampleOutput: string; // Markdown
  explanation: string; // Markdown
  solutions: Solution[];
  notes: string;
}

export interface User {
  id: string;
  email: string;
  name: string;
  role?: string; // 'admin' or 'demo'
}

export interface LearningTopic {
  id: string;
  name: string;
  icon: string;
  description: string;
  slug: string;
}

export interface LearningResource {
  id: string;
  topicId: string;
  title: string;
  content: string;
  type: 'article' | 'video' | 'link';
  url?: string;
  orderIndex: number;
}

export interface RoadmapItem {
  id: string;
  topicId: string;
  title: string;
  description: string;
  orderIndex: number;
  status: 'todo' | 'in-progress' | 'completed';
}

export enum ViewState {
  CATEGORIES,
  PATTERNS,
  PROBLEMS,
  PROBLEM_DETAIL,
  LEARNING_TOPIC,
  LEARNING_RESOURCE
}

