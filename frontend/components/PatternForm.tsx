import React, { useState } from 'react';
import { Pattern } from '../types';
import { api } from '../services/apiService';
import { X, Sparkles, Loader2 } from 'lucide-react';

interface PatternFormProps {
  pattern?: Pattern;
  categoryId: string;
  categoryName: string;
  onSave: (pattern: Partial<Pattern>) => Promise<void>;
  onClose: () => void;
}

const PatternForm: React.FC<PatternFormProps> = ({ pattern, categoryId, categoryName, onSave, onClose }) => {
  const [name, setName] = useState(pattern?.name || '');
  const [icon, setIcon] = useState(pattern?.icon || 'Code2');
  const [description, setDescription] = useState(pattern?.description || '');
  const [loading, setLoading] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await onSave({ name, icon, description });
      onClose();
    } catch (error) {
      console.error('Error saving pattern:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-slate-800 border border-slate-700 rounded-2xl w-full max-w-md shadow-2xl">
        <div className="flex justify-between items-center p-6 border-b border-slate-700">
          <h2 className="text-xl font-bold text-white">
            {pattern ? 'Edit Pattern' : 'New Pattern'}
          </h2>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-white transition-colors"
          >
            <X size={24} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="e.g., Two Pointers"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Icon</label>
            <input
              type="text"
              value={icon}
              onChange={(e) => setIcon(e.target.value)}
              required
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="e.g., Target, Zap, etc."
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="block text-sm font-medium text-slate-300">Description</label>
              <button
                type="button"
                onClick={async () => {
                  if (!name.trim()) {
                    alert('Please enter a pattern name first');
                    return;
                  }
                  setIsGenerating(true);
                  try {
                    const response = await api.generatePatternContent(name, categoryName, 'description');
                    setDescription(response.content);
                  } catch (error) {
                    console.error('Error generating description:', error);
                    alert('Failed to generate description. Please try again.');
                  } finally {
                    setIsGenerating(false);
                  }
                }}
                disabled={isGenerating || !name.trim()}
                className="flex items-center gap-1.5 px-2.5 py-1 bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-400 rounded text-xs font-medium transition-colors disabled:opacity-50"
              >
                {isGenerating ? (
                  <>
                    <Loader2 size={12} className="animate-spin" />
                    Generating...
                  </>
                ) : (
                  <>
                    <Sparkles size={12} />
                    Generate with AI
                  </>
                )}
              </button>
            </div>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              required
              rows={4}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Describe this pattern..."
            />
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg font-medium transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-medium transition-colors disabled:opacity-50"
            >
              {loading ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default PatternForm;
