
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

export enum ViewState {
  CATEGORIES,
  PATTERNS,
  PROBLEMS,
  PROBLEM_DETAIL
}
