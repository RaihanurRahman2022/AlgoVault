import React from 'react';
import { marked } from 'marked';

interface MarkdownRendererProps {
  content: string;
  className?: string;
}

const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ content, className = '' }) => {
  // Configure marked options
  marked.setOptions({
    breaks: true,
    gfm: true,
  });

  // Convert markdown to HTML
  let html = '';
  try {
    if (typeof marked.parse === 'function') {
      html = marked.parse(content || '') as string;
    } else {
      // Fallback for older API
      html = marked(content || '') as string;
    }
  } catch (error) {
    console.error('Error parsing markdown:', error);
    html = content || '';
  }

  return (
    <div
      className={`prose prose-invert prose-slate max-w-none ${className}`}
      dangerouslySetInnerHTML={{ __html: html }}
      style={{
        color: '#cbd5e1',
      }}
    />
  );
};

export default MarkdownRenderer;
