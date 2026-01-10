
import React from 'react';
import { 
  Code2, 
  Binary, 
  GitBranch, 
  Hash, 
  Layers, 
  Database, 
  Cpu, 
  Search, 
  BookOpen, 
  Zap,
  Terminal,
  FileText,
  MessageSquare,
  Lightbulb
} from 'lucide-react';

export const ICON_MAP: Record<string, React.ReactNode> = {
  'code': <Code2 size={20} />,
  'tree': <GitBranch size={20} />,
  'graph': <Hash size={20} />,
  'binary': <Binary size={20} />,
  'layers': <Layers size={20} />,
  'database': <Database size={20} />,
  'cpu': <Cpu size={20} />,
  'search': <Search size={20} />,
  'book': <BookOpen size={20} />,
  'zap': <Zap size={20} />,
  'terminal': <Terminal size={20} />,
  'file': <FileText size={20} />,
  'message': <MessageSquare size={20} />,
  'idea': <Lightbulb size={20} />
};

export const LANGUAGES = [
  { value: 'cpp', label: 'C++' },
  { value: 'go', label: 'Golang' },
  { value: 'python', label: 'Python' },
  { value: 'javascript', label: 'JavaScript' }
];
