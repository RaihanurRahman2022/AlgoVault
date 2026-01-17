import React, { useState, useEffect } from 'react';
import { X, Edit, Save, Sparkles, Loader2 } from 'lucide-react';
import MarkdownRenderer from './MarkdownRenderer';
import { api } from '../services/apiService';

interface TheoryModalProps {
  patternId: string;
  patternName: string;
  categoryName: string;
  initialTheory: string;
  isOpen: boolean;
  onClose: () => void;
  onUpdate: (theory: string) => Promise<void>;
  isDemoUser?: boolean;
}

const TheoryModal: React.FC<TheoryModalProps> = ({
  patternId,
  patternName,
  categoryName,
  initialTheory,
  isOpen,
  onClose,
  onUpdate,
  isDemoUser = false,
}) => {
  const [theory, setTheory] = useState(initialTheory);
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [showPromptModal, setShowPromptModal] = useState(false);
  const [aiPrompt, setAiPrompt] = useState('');

  useEffect(() => {
    if (isOpen) {
      setTheory(initialTheory);
      setIsEditing(false);
    }
  }, [isOpen, initialTheory]);

  if (!isOpen) return null;

  const handleSave = async () => {
    setIsSaving(true);
    try {
      await onUpdate(theory);
      setIsEditing(false);
    } catch (error) {
      console.error('Error saving theory:', error);
      alert('Failed to save theory. Please try again.');
    } finally {
      setIsSaving(false);
    }
  };

  const handleGenerate = () => {
    setShowPromptModal(true);
  };

  const handleGenerateWithPrompt = async (prompt: string) => {
    setIsGenerating(true);
    try {
      const response = await api.generatePatternContent(patternName, categoryName, 'theory', prompt);
      setTheory(response.content);
      setIsEditing(true);
      setShowPromptModal(false);
      setAiPrompt('');
    } catch (error) {
      console.error('Error generating theory:', error);
      alert('Failed to generate theory. Please try again.');
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-slate-800 border border-slate-700 rounded-2xl w-full max-w-4xl max-h-[90vh] flex flex-col shadow-2xl">
        <div className="flex justify-between items-center p-6 border-b border-slate-700">
          <div>
            <h2 className="text-xl font-bold text-white">Pattern Theory</h2>
            <p className="text-sm text-slate-400 mt-1">{patternName}</p>
          </div>
          <div className="flex items-center gap-2">
            {!isDemoUser && !isEditing && theory && (
              <button
                onClick={handleGenerate}
                disabled={isGenerating}
                className="flex items-center gap-2 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
                title="Generate with AI"
              >
                {isGenerating ? (
                  <>
                    <Loader2 size={16} className="animate-spin" />
                    Generating...
                  </>
                ) : (
                  <>
                    <Sparkles size={16} />
                    Regenerate
                  </>
                )}
              </button>
            )}
            {!isDemoUser && !isEditing && !theory && (
              <button
                onClick={handleGenerate}
                disabled={isGenerating}
                className="flex items-center gap-2 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
                title="Generate with AI"
              >
                {isGenerating ? (
                  <>
                    <Loader2 size={16} className="animate-spin" />
                    Generating...
                  </>
                ) : (
                  <>
                    <Sparkles size={16} />
                    Generate Theory
                  </>
                )}
              </button>
            )}
            {!isDemoUser && !isEditing && (
              <button
                onClick={() => setIsEditing(true)}
                className="flex items-center gap-2 px-3 py-1.5 bg-slate-700 hover:bg-slate-600 text-white rounded-lg text-sm font-medium transition-colors"
              >
                <Edit size={16} />
                Edit
              </button>
            )}
            {!isDemoUser && isEditing && (
              <button
                onClick={handleSave}
                disabled={isSaving}
                className="flex items-center gap-2 px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
              >
                {isSaving ? (
                  <>
                    <Loader2 size={16} className="animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <Save size={16} />
                    Save
                  </>
                )}
              </button>
            )}
            <button
              onClick={onClose}
              className="text-slate-400 hover:text-white transition-colors"
            >
              <X size={24} />
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-auto p-6">
          {isEditing && !isDemoUser ? (
            <textarea
              value={theory}
              onChange={(e) => setTheory(e.target.value)}
              className="w-full h-full min-h-[400px] bg-slate-900 border border-slate-700 rounded-lg px-4 py-3 text-white font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
              placeholder="Enter theory content in markdown format..."
            />
          ) : (
            <div className="min-h-[400px]">
              {theory ? (
                <MarkdownRenderer content={theory} />
              ) : (
                <div className="text-center text-slate-500 py-20">
                  <p className="text-lg mb-2">No theory content yet</p>
                  {!isDemoUser && (
                    <p className="text-sm">Click "Generate Theory" to create content with AI</p>
                  )}
                </div>
              )}
            </div>
          )}
        </div>

        {!isDemoUser && isEditing && (
          <div className="p-4 border-t border-slate-700 bg-slate-900/50">
            <button
              onClick={() => {
                setIsEditing(false);
                setTheory(initialTheory);
              }}
              className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg text-sm font-medium transition-colors"
            >
              Cancel
            </button>
          </div>
        )}
      </div>

      {/* AI Prompt Modal */}
      {showPromptModal && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-[60] p-4">
          <div className="bg-slate-800 border border-slate-700 rounded-2xl w-full max-w-md shadow-2xl">
            <div className="flex justify-between items-center p-6 border-b border-slate-700">
              <h3 className="text-lg font-bold text-white">AI Generation Prompt</h3>
              <button
                onClick={() => {
                  setShowPromptModal(false);
                  setAiPrompt('');
                }}
                className="text-slate-400 hover:text-white transition-colors"
              >
                <X size={20} />
              </button>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Additional Instructions (Optional)
                </label>
                <textarea
                  value={aiPrompt}
                  onChange={(e) => setAiPrompt(e.target.value)}
                  rows={4}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="e.g., Focus on competitive programming, include time complexity analysis, add more examples..."
                />
                <p className="text-xs text-slate-500 mt-1">
                  Add any specific instructions or context for the AI generation
                </p>
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowPromptModal(false);
                    setAiPrompt('');
                  }}
                  className="flex-1 px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg font-medium transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={() => handleGenerateWithPrompt(aiPrompt)}
                  disabled={isGenerating}
                  className="flex-1 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {isGenerating ? (
                    <>
                      <Loader2 size={16} className="animate-spin" />
                      Generating...
                    </>
                  ) : (
                    <>
                      <Sparkles size={16} />
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

export default TheoryModal;
