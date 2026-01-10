import React, { useState } from 'react';
import { Problem, Solution } from '../types';
import { X, Sparkles, Loader2 } from 'lucide-react';
import { LANGUAGES } from '../constants';
import { api } from '../services/apiService';

interface ProblemFormProps {
  problem?: Problem;
  patternId: string;
  onSave: (problem: Partial<Problem>) => Promise<void>;
  onClose: () => void;
}

const ProblemForm: React.FC<ProblemFormProps> = ({ problem, patternId, onSave, onClose }) => {
  const [title, setTitle] = useState(problem?.title || '');
  const [difficulty, setDifficulty] = useState<'Easy' | 'Medium' | 'Hard'>(problem?.difficulty || 'Easy');
  const [description, setDescription] = useState(problem?.description || '');
  const [input, setInput] = useState(problem?.input || '');
  const [output, setOutput] = useState(problem?.output || '');
  const [constraints, setConstraints] = useState(problem?.constraints || '');
  const [sampleInput, setSampleInput] = useState(problem?.sampleInput || '');
  const [sampleOutput, setSampleOutput] = useState(problem?.sampleOutput || '');
  const [explanation, setExplanation] = useState(problem?.explanation || '');
  const [notes, setNotes] = useState(problem?.notes || '');
  const [solutions, setSolutions] = useState<Solution[]>(problem?.solutions || []);
  const [loading, setLoading] = useState(false);
  const [aiQuery, setAiQuery] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [showAIModal, setShowAIModal] = useState(false);

  const updateSolution = (language: string, code: string) => {
    setSolutions(prev => {
      const existing = prev.find(s => s.language === language);
      if (existing) {
        return prev.map(s => s.language === language ? { ...s, code } : s);
      }
      return [...prev, { id: '', language: language as any, code, problemId: patternId }];
    });
  };

  const handleAIGenerate = async () => {
    if (!aiQuery.trim()) {
      alert('Please enter a problem description or name');
      return;
    }

    setIsGenerating(true);
    try {
      const generated = await api.generateProblem(aiQuery);
      // Populate form fields with generated data
      setTitle(generated.title);
      setDifficulty(generated.difficulty);
      setDescription(generated.description);
      setInput(generated.input);
      setOutput(generated.output);
      setConstraints(generated.constraints);
      setSampleInput(generated.sampleInput);
      setSampleOutput(generated.sampleOutput);
      setExplanation(generated.explanation);
      setShowAIModal(false);
      setAiQuery('');
    } catch (error: any) {
      console.error('Error generating problem:', error);
      alert('Failed to generate problem: ' + (error.message || 'Unknown error'));
    } finally {
      setIsGenerating(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await onSave({
        title,
        difficulty,
        description,
        input,
        output,
        constraints,
        sampleInput,
        sampleOutput,
        explanation,
        notes,
        solutions,
      });
      onClose();
    } catch (error) {
      console.error('Error saving problem:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4 overflow-y-auto">
      <div className="bg-slate-800 border border-slate-700 rounded-2xl w-full max-w-4xl shadow-2xl my-8">
        <div className="flex justify-between items-center p-6 border-b border-slate-700 sticky top-0 bg-slate-800">
          <h2 className="text-xl font-bold text-white">
            {problem ? 'Edit Problem' : 'New Problem'}
          </h2>
          <div className="flex items-center gap-3">
            <button
              onClick={() => setShowAIModal(true)}
              className="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-medium transition-colors"
              title="Generate problem with AI"
            >
              <Sparkles size={18} />
              AI Generate
            </button>
            <button
              onClick={onClose}
              className="text-slate-400 hover:text-white transition-colors"
            >
              <X size={24} />
            </button>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-6 max-h-[80vh] overflow-y-auto">
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Title</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                required
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="Problem title"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Difficulty</label>
              <select
                value={difficulty}
                onChange={(e) => setDifficulty(e.target.value as any)}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="Easy">Easy</option>
                <option value="Medium">Medium</option>
                <option value="Hard">Hard</option>
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Description (Markdown)</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              required
              rows={4}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
              placeholder="Problem description in markdown..."
            />
          </div>

          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Input (Markdown)</label>
              <textarea
                value={input}
                onChange={(e) => setInput(e.target.value)}
                required
                rows={3}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
                placeholder="Input format..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Output (Markdown)</label>
              <textarea
                value={output}
                onChange={(e) => setOutput(e.target.value)}
                required
                rows={3}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
                placeholder="Output format..."
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Constraints (Markdown)</label>
            <textarea
              value={constraints}
              onChange={(e) => setConstraints(e.target.value)}
              required
              rows={3}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
              placeholder="Constraints..."
            />
          </div>

          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Sample Input (Markdown)</label>
              <textarea
                value={sampleInput}
                onChange={(e) => setSampleInput(e.target.value)}
                required
                rows={3}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
                placeholder="Sample input..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-300 mb-2">Sample Output (Markdown)</label>
              <textarea
                value={sampleOutput}
                onChange={(e) => setSampleOutput(e.target.value)}
                required
                rows={3}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
                placeholder="Sample output..."
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Explanation (Markdown)</label>
            <textarea
              value={explanation}
              onChange={(e) => setExplanation(e.target.value)}
              required
              rows={4}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Explanation..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Notes (Markdown)</label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={4}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Your notes..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">Solutions</label>
            <div className="space-y-4">
              {LANGUAGES.map(lang => {
                const solution = solutions.find(s => s.language === lang.value);
                return (
                  <div key={lang.value} className="bg-slate-900 rounded-lg p-4 border border-slate-700">
                    <label className="block text-sm font-medium text-slate-300 mb-2">{lang.label}</label>
                    <textarea
                      value={solution?.code || ''}
                      onChange={(e) => updateSolution(lang.value, e.target.value)}
                      rows={8}
                      className="w-full bg-slate-950 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono text-sm"
                      placeholder={`${lang.label} solution...`}
                    />
                  </div>
                );
              })}
            </div>
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

      {/* AI Generation Modal */}
      {showAIModal && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-[60] p-4">
          <div className="bg-slate-800 border border-slate-700 rounded-2xl w-full max-w-md shadow-2xl">
            <div className="flex justify-between items-center p-6 border-b border-slate-700">
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                <Sparkles size={20} className="text-indigo-400" />
                AI Problem Generator
              </h3>
              <button
                onClick={() => {
                  setShowAIModal(false);
                  setAiQuery('');
                }}
                className="text-slate-400 hover:text-white transition-colors"
              >
                <X size={20} />
              </button>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Describe the problem you want to generate
                </label>
                <textarea
                  value={aiQuery}
                  onChange={(e) => setAiQuery(e.target.value)}
                  placeholder="e.g., Two Sum problem, Binary Tree Inorder Traversal, Find maximum subarray sum..."
                  rows={4}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && e.ctrlKey) {
                      handleAIGenerate();
                    }
                  }}
                />
                <p className="text-xs text-slate-500 mt-1">Press Ctrl+Enter to generate</p>
              </div>
              <div className="flex gap-3">
                <button
                  onClick={() => {
                    setShowAIModal(false);
                    setAiQuery('');
                  }}
                  className="flex-1 px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg font-medium transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleAIGenerate}
                  disabled={isGenerating || !aiQuery.trim()}
                  className="flex-1 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {isGenerating ? (
                    <>
                      <Loader2 size={18} className="animate-spin" />
                      Generating...
                    </>
                  ) : (
                    <>
                      <Sparkles size={18} />
                      Generate
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ProblemForm;
