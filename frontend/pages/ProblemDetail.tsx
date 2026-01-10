import React, { useState, useEffect } from 'react';
import { Problem, Solution } from '../types';
import { LANGUAGES } from '../constants';
import { generateNote } from '../services/geminiService';
import { api } from '../services/apiService';
import MarkdownRenderer from '../components/MarkdownRenderer';
import DeleteConfirmationModal from '../components/DeleteConfirmationModal';
import { 
  Clipboard, 
  BookText, 
  Code, 
  FileEdit, 
  Sparkles, 
  Check, 
  ExternalLink,
  ChevronDown,
  Edit,
  Trash2,
  Save,
  X,
  Loader2
} from 'lucide-react';

interface ProblemDetailProps {
  problem: Problem;
  onUpdate?: (problem: Problem) => void;
  onDelete?: () => void;
  isDemoUser?: boolean;
}

const ProblemDetail: React.FC<ProblemDetailProps> = ({ problem: initialProblem, onUpdate, onDelete, isDemoUser = false }) => {
  const [problem, setProblem] = useState<Problem>(initialProblem);
  const [isEditing, setIsEditing] = useState(false);
  
  // Wrapper function to prevent setting isEditing to true for demo users
  const setEditingState = (value: boolean) => {
    if (value && isDemoUser) {
      return; // Don't allow editing for demo users
    }
    setIsEditing(value);
  };
  
  // Prevent editing for demo users - ensure isEditing is always false for demo users
  useEffect(() => {
    if (isDemoUser) {
      setEditingState(false);
    }
  }, [isDemoUser]);
  
  // Additional check to prevent editing if user becomes demo user while editing
  useEffect(() => {
    if (isDemoUser && isEditing) {
      setEditingState(false);
    }
  }, [isDemoUser, isEditing]);
  const [activeTab, setActiveTab] = useState<'description' | 'solution' | 'notes'>('description');
  const [selectedLanguage, setSelectedLanguage] = useState((problem.solutions && problem.solutions.length > 0) ? problem.solutions[0].language : 'cpp');
  const [isCopied, setIsCopied] = useState(false);
  const [isGeneratingNote, setIsGeneratingNote] = useState(false);
  const [dynamicNote, setDynamicNote] = useState(problem.notes);
  const [saving, setSaving] = useState(false);
  const [aiQuery, setAiQuery] = useState('');
  const [isGeneratingProblem, setIsGeneratingProblem] = useState(false);
  const [showAIModal, setShowAIModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  // Edit state
  const [editTitle, setEditTitle] = useState(problem.title);
  const [editDifficulty, setEditDifficulty] = useState<'Easy' | 'Medium' | 'Hard'>(problem.difficulty);
  const [editDescription, setEditDescription] = useState(problem.description);
  const [editInput, setEditInput] = useState(problem.input);
  const [editOutput, setEditOutput] = useState(problem.output);
  const [editConstraints, setEditConstraints] = useState(problem.constraints);
  const [editSampleInput, setEditSampleInput] = useState(problem.sampleInput);
  const [editSampleOutput, setEditSampleOutput] = useState(problem.sampleOutput);
  const [editExplanation, setEditExplanation] = useState(problem.explanation);
  const [editNotes, setEditNotes] = useState(problem.notes);
  const [editSolutions, setEditSolutions] = useState<Solution[]>(problem.solutions || []);

  // Update local state when prop changes
  useEffect(() => {
    setProblem(initialProblem);
    setDynamicNote(initialProblem.notes);
    setSelectedLanguage((initialProblem.solutions && initialProblem.solutions.length > 0) ? initialProblem.solutions[0].language : 'cpp');
    // Reset edit state
    setEditTitle(initialProblem.title);
    setEditDifficulty(initialProblem.difficulty);
    setEditDescription(initialProblem.description);
    setEditInput(initialProblem.input);
    setEditOutput(initialProblem.output);
    setEditConstraints(initialProblem.constraints);
    setEditSampleInput(initialProblem.sampleInput);
    setEditSampleOutput(initialProblem.sampleOutput);
    setEditExplanation(initialProblem.explanation);
    setEditNotes(initialProblem.notes);
    // Deep copy solutions to avoid reference issues
    setEditSolutions(initialProblem.solutions ? initialProblem.solutions.map(s => ({ ...s })) : []);
  }, [initialProblem]);

  // Use problem.solutions for view mode, editSolutions for edit mode
  // Find solution by exact language match - don't fallback to first solution
  const currentSolution = isEditing 
    ? editSolutions.find(s => s.language === selectedLanguage)
    : (problem.solutions || []).find(s => s.language === selectedLanguage);

  const handleCopy = async () => {
    if (!currentSolution || !currentSolution.code) return;
    
    try {
      // Check if clipboard API is available
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(currentSolution.code);
        setIsCopied(true);
        setTimeout(() => setIsCopied(false), 2000);
      } else {
        // Fallback for browsers without clipboard API
        const textArea = document.createElement('textarea');
        textArea.value = currentSolution.code;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        try {
          document.execCommand('copy');
          setIsCopied(true);
          setTimeout(() => setIsCopied(false), 2000);
        } catch (err) {
          console.error('Failed to copy:', err);
          alert('Failed to copy to clipboard. Please copy manually.');
        }
        document.body.removeChild(textArea);
      }
    } catch (err) {
      console.error('Failed to copy:', err);
      alert('Failed to copy to clipboard. Please copy manually.');
    }
  };

  const handleGenerateAINote = async () => {
    setIsGeneratingNote(true);
    const note = await generateNote(problem.title, problem.solutions);
    setDynamicNote(note || "Failed to generate AI note.");
    setIsGeneratingNote(false);
  };

  const updateSolution = (language: string, code: string) => {
    setEditSolutions(prev => {
      const existing = prev.find(s => s.language === language);
      if (existing) {
        // Preserve the existing ID when updating
        return prev.map(s => s.language === language ? { ...s, code } : s);
      }
      // Create new solution, but try to find existing one from problem to preserve ID
      const existingFromProblem = (problem.solutions || []).find(s => s.language === language);
      return [...prev, { 
        id: existingFromProblem?.id || '', 
        language: language as any, 
        code, 
        problemId: problem.id 
      }];
    });
  };

  // Handle tab key for proper indentation in code editor
  const handleTabKey = (e: React.KeyboardEvent<HTMLTextAreaElement>, language: string) => {
    if (e.key === 'Tab') {
      e.preventDefault();
      const textarea = e.currentTarget;
      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      const value = textarea.value;
      const newValue = value.substring(0, start) + '    ' + value.substring(end);
      // Update the solution state
      updateSolution(language, newValue);
      // Set cursor position after state update
      setTimeout(() => {
        textarea.selectionStart = textarea.selectionEnd = start + 4;
      }, 0);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      // Ensure all solutions have proper structure - language is the key identifier
      const solutionsToSave = editSolutions
        .filter(sol => sol.code && sol.code.trim().length > 0) // Only save non-empty solutions
        .map(sol => {
          const existing = (problem.solutions || []).find(s => s.language === sol.language);
          return {
            ...sol,
            id: existing?.id || sol.id || '', // Preserve ID if exists
            language: sol.language, // Ensure language is set
            problemId: problem.id,
          };
        });
      
      const updated = await api.updateProblem(problem.id, {
        title: editTitle,
        difficulty: editDifficulty,
        description: editDescription,
        input: editInput,
        output: editOutput,
        constraints: editConstraints,
        sampleInput: editSampleInput,
        sampleOutput: editSampleOutput,
        explanation: editExplanation,
        notes: editNotes,
        solutions: solutionsToSave,
      });
      setProblem(updated);
      setEditSolutions(updated.solutions || []); // Update edit state with saved solutions
      setIsEditing(false);
      if (onUpdate) {
        onUpdate(updated);
      }
    } catch (error) {
      console.error('Error updating problem:', error);
      alert('Failed to save changes. Please try again.');
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    // Reset to original values
    setEditTitle(problem.title);
    setEditDifficulty(problem.difficulty);
    setEditDescription(problem.description);
    setEditInput(problem.input);
    setEditOutput(problem.output);
    setEditConstraints(problem.constraints);
    setEditSampleInput(problem.sampleInput);
    setEditSampleOutput(problem.sampleOutput);
    setEditExplanation(problem.explanation);
    setEditNotes(problem.notes);
    setEditSolutions(problem.solutions || []);
    setEditingState(false);
  };

  const handleDelete = async () => {
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    setIsDeleting(true);
    try {
      await api.deleteProblem(problem.id);
      if (onDelete) {
        onDelete();
      }
    } catch (error) {
      console.error('Error deleting problem:', error);
    } finally {
      setIsDeleting(false);
      setShowDeleteModal(false);
    }
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      {/* Left Sidebar - Navigation / Overview */}
      <div className="lg:col-span-3 space-y-4">
        <div className="bg-slate-800 border border-slate-700 rounded-2xl p-6 shadow-xl sticky top-24 overflow-hidden">
          <div className="flex items-center justify-between mb-6 gap-2 min-w-0">
            <div className="flex items-center gap-2 flex-1 min-w-0 overflow-hidden">
              <span className={`w-3 h-3 rounded-full flex-shrink-0 ${
                (isEditing ? editDifficulty : problem.difficulty) === 'Easy' ? 'bg-emerald-500' :
                (isEditing ? editDifficulty : problem.difficulty) === 'Medium' ? 'bg-amber-500' : 'bg-red-500'
              }`} />
              {isEditing && !isDemoUser ? (
                <input
                  type="text"
                  value={editTitle}
                  onChange={(e) => setEditTitle(e.target.value)}
                  className="flex-1 min-w-0 bg-slate-900 border border-slate-700 rounded-lg px-3 py-1 text-white font-bold text-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              ) : (
                <span className="font-bold text-lg truncate block min-w-0">{problem.title}</span>
              )}
            </div>
            {!isDemoUser && (
              <div className="flex gap-2 flex-shrink-0">
                {isEditing ? (
                  <>
                    <button
                      onClick={() => setShowAIModal(true)}
                      className="p-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-white transition-colors"
                      title="Generate with AI"
                    >
                      <Sparkles size={16} />
                    </button>
                    <button
                      onClick={handleSave}
                      disabled={saving}
                      className="p-2 bg-emerald-600 hover:bg-emerald-500 rounded-lg text-white transition-colors disabled:opacity-50"
                      title="Save Changes"
                    >
                      <Save size={16} />
                    </button>
                    <button
                      onClick={handleCancel}
                      className="p-2 bg-slate-700 hover:bg-slate-600 rounded-lg text-white transition-colors"
                      title="Cancel"
                    >
                      <X size={16} />
                    </button>
                  </>
                ) : (
                  <>
                    <button
                      onClick={() => {
                        setEditingState(true);
                      }}
                      className="p-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-white transition-colors"
                      title="Edit Problem"
                    >
                      <Edit size={16} />
                    </button>
                    <button
                      onClick={handleDelete}
                      className="p-2 bg-red-500/20 hover:bg-red-500/30 rounded-lg text-red-400 hover:text-red-300 transition-colors"
                      title="Delete Problem"
                    >
                      <Trash2 size={16} />
                    </button>
                  </>
                )}
              </div>
            )}
          </div>

          {isEditing && !isDemoUser && (
            <div className="mb-4">
              <label className="block text-xs font-medium text-slate-400 mb-2">Difficulty</label>
              <select
                value={editDifficulty}
                onChange={(e) => setEditDifficulty(e.target.value as any)}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="Easy">Easy</option>
                <option value="Medium">Medium</option>
                <option value="Hard">Hard</option>
              </select>
            </div>
          )}

          <nav className="space-y-1">
            <button 
              onClick={() => setActiveTab('description')}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl text-left transition-all ${
                activeTab === 'description' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:bg-slate-700 hover:text-white'
              }`}
            >
              <BookText size={18} />
              <span className="font-semibold">Problem</span>
            </button>
            <button 
              onClick={() => setActiveTab('solution')}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl text-left transition-all ${
                activeTab === 'solution' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:bg-slate-700 hover:text-white'
              }`}
            >
              <Code size={18} />
              <span className="font-semibold">Solutions</span>
            </button>
            <button 
              onClick={() => setActiveTab('notes')}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl text-left transition-all ${
                activeTab === 'notes' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:bg-slate-700 hover:text-white'
              }`}
            >
              <FileEdit size={18} />
              <span className="font-semibold">Notes & Tips</span>
            </button>
          </nav>

          <div className="mt-8 pt-8 border-t border-slate-700 space-y-4">
            <div className="flex justify-between items-center text-xs">
              <span className="text-slate-500 uppercase font-bold">Difficulty</span>
              <span className={
                (isEditing ? editDifficulty : problem.difficulty) === 'Easy' ? 'text-emerald-500' :
                (isEditing ? editDifficulty : problem.difficulty) === 'Medium' ? 'text-amber-500' : 'text-red-500'
              }>{isEditing ? editDifficulty : problem.difficulty}</span>
            </div>
            <button className="w-full flex items-center justify-between px-4 py-2.5 bg-slate-900 border border-slate-700 rounded-xl text-xs hover:border-slate-500 transition-colors">
              <div className="flex items-center gap-2">
                <ExternalLink size={14} className="text-slate-500" />
                <span>LeetCode Link</span>
              </div>
              <ChevronDown size={14} className="text-slate-500" />
            </button>
          </div>
        </div>
      </div>

      {/* Main Content View */}
      <div className="lg:col-span-9 bg-slate-800 border border-slate-700 rounded-2xl shadow-2xl flex flex-col min-h-[70vh] overflow-hidden">
        {activeTab === 'description' && (
          <div className="p-8 space-y-6 overflow-y-auto flex-1 min-h-0">
            {/* Description Section */}
            <section className="mb-8">
              <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
                Description
              </h2>
              {isEditing && !isDemoUser ? (
                <textarea
                  value={editDescription}
                  onChange={(e) => setEditDescription(e.target.value)}
                  rows={6}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-3 text-slate-300 leading-relaxed focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-y"
                  placeholder="Problem description in markdown..."
                />
              ) : (
                <div className="bg-slate-900/50 p-4 rounded-lg border border-slate-700">
                  {problem.description ? (
                    <MarkdownRenderer content={problem.description} />
                  ) : (
                    <div className="text-slate-400 italic">No description provided.</div>
                  )}
                </div>
              )}
            </section>

            {/* Input Section */}
            <section className="bg-slate-900/50 p-6 rounded-2xl border border-slate-700">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-bold text-emerald-500 uppercase tracking-widest">INPUT</h3>
                {isEditing && !isDemoUser && (
                  <button className="text-slate-400 hover:text-white transition-colors">
                    <X size={16} />
                  </button>
                )}
              </div>
              {isEditing && !isDemoUser ? (
                <textarea
                  value={editInput}
                  onChange={(e) => setEditInput(e.target.value)}
                  rows={8}
                  className="w-full bg-slate-950 border border-slate-600 rounded-lg px-4 py-3 text-slate-300 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-y"
                  placeholder="Input format..."
                />
              ) : (
                <div className="bg-slate-950 p-4 rounded-lg min-h-[100px]">
                  {problem.input ? (
                    <MarkdownRenderer content={problem.input} className="font-mono text-sm" />
                  ) : (
                    <div className="text-slate-400 italic font-mono text-sm">No input format specified.</div>
                  )}
                </div>
              )}
            </section>

            {/* Output Section */}
            <section className="bg-slate-900/50 p-6 rounded-2xl border border-slate-700">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-bold text-amber-500 uppercase tracking-widest">OUTPUT</h3>
                {isEditing && !isDemoUser && (
                  <button className="text-slate-400 hover:text-white transition-colors">
                    <X size={16} />
                  </button>
                )}
              </div>
              {isEditing && !isDemoUser ? (
                <textarea
                  value={editOutput}
                  onChange={(e) => setEditOutput(e.target.value)}
                  rows={8}
                  className="w-full bg-slate-950 border border-slate-600 rounded-lg px-4 py-3 text-slate-300 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-y"
                  placeholder="Output format..."
                />
              ) : (
                <div className="bg-slate-950 p-4 rounded-lg min-h-[100px]">
                  {problem.output ? (
                    <MarkdownRenderer content={problem.output} className="font-mono text-sm" />
                  ) : (
                    <div className="text-slate-400 italic font-mono text-sm">No output format specified.</div>
                  )}
                </div>
              )}
            </section>

            {/* Explanation Section */}
            <section className="bg-slate-900/50 p-6 rounded-2xl border border-slate-700">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-bold text-indigo-400 uppercase tracking-widest">EXPLANATION</h3>
                {isEditing && !isDemoUser && (
                  <button className="text-slate-400 hover:text-white transition-colors">
                    <X size={16} />
                  </button>
                )}
              </div>
              {isEditing && !isDemoUser ? (
                <textarea
                  value={editExplanation}
                  onChange={(e) => setEditExplanation(e.target.value)}
                  rows={8}
                  className="w-full bg-slate-950 border border-slate-600 rounded-lg px-4 py-3 text-slate-300 leading-relaxed focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-y"
                  placeholder="Explanation..."
                />
              ) : (
                <div className="bg-slate-950 p-4 rounded-lg min-h-[100px]">
                  {problem.explanation ? (
                    <MarkdownRenderer content={problem.explanation} />
                  ) : (
                    <div className="text-slate-400 italic">No explanation provided.</div>
                  )}
                </div>
              )}
            </section>

            {/* Constraints Section (Collapsible) */}
            <section className="bg-slate-900/30 p-6 rounded-2xl border border-slate-700">
              <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest mb-3">Constraints</h3>
              {isEditing && !isDemoUser ? (
                <textarea
                  value={editConstraints}
                  onChange={(e) => setEditConstraints(e.target.value)}
                  rows={4}
                  className="w-full bg-slate-950 border border-slate-600 rounded-lg px-4 py-3 text-slate-300 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-y"
                  placeholder="Constraints..."
                />
              ) : (
                <div className="bg-slate-950 p-4 rounded-xl">
                  {problem.constraints ? (
                    <MarkdownRenderer content={problem.constraints} className="font-mono text-sm" />
                  ) : (
                    <div className="text-slate-400 italic font-mono text-sm">No constraints specified.</div>
                  )}
                </div>
              )}
            </section>

            {/* Sample Input/Output Section */}
            {(problem.sampleInput || problem.sampleOutput || isEditing) && (
              <section className="space-y-4">
                <h2 className="text-xl font-bold text-white">Sample Cases</h2>
                <div className="bg-slate-950 rounded-2xl border border-slate-700 overflow-hidden">
                  <div className="bg-slate-900 px-6 py-3 border-b border-slate-800">
                    <span className="text-xs font-bold text-slate-400 uppercase">Sample Case #1</span>
                  </div>
                  <div className="p-6 space-y-4">
                    <div>
                      <span className="text-xs font-bold text-emerald-500 uppercase block mb-2">Sample Input</span>
                      {isEditing && !isDemoUser ? (
                        <textarea
                          value={editSampleInput}
                          onChange={(e) => setEditSampleInput(e.target.value)}
                          rows={4}
                          className="w-full bg-slate-900 border border-slate-600 rounded-lg px-4 py-3 text-slate-300 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-y"
                          placeholder="Sample input..."
                        />
                      ) : (
                        <div className="bg-slate-900 p-3 rounded-lg">
                          {problem.sampleInput ? (
                            <MarkdownRenderer content={problem.sampleInput} className="font-mono text-sm" />
                          ) : (
                            <div className="text-slate-400 italic font-mono text-sm">No sample input provided.</div>
                          )}
                        </div>
                      )}
                    </div>
                    <div>
                      <span className="text-xs font-bold text-amber-500 uppercase block mb-2">Sample Output</span>
                      {isEditing && !isDemoUser ? (
                        <textarea
                          value={editSampleOutput}
                          onChange={(e) => setEditSampleOutput(e.target.value)}
                          rows={4}
                          className="w-full bg-slate-900 border border-slate-600 rounded-lg px-4 py-3 text-slate-300 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-y"
                          placeholder="Sample output..."
                        />
                      ) : (
                        <div className="bg-slate-900 p-3 rounded-lg">
                          {problem.sampleOutput ? (
                            <MarkdownRenderer content={problem.sampleOutput} className="font-mono text-sm" />
                          ) : (
                            <div className="text-slate-400 italic font-mono text-sm">No sample output provided.</div>
                          )}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </section>
            )}
          </div>
        )}

        {activeTab === 'solution' && (
          <div className="flex flex-col flex-1 overflow-hidden min-h-0">
            {isEditing && !isDemoUser ? (
              <div className="flex-1 p-6 space-y-6 overflow-y-auto min-h-0">
                <h2 className="text-2xl font-bold text-white mb-4">Solutions</h2>
                {LANGUAGES.map(lang => {
                  const solution = editSolutions.find(s => s.language === lang.value);
                  return (
                    <div key={lang.value} className="bg-slate-900 rounded-lg p-4 border border-slate-700">
                      <label className="block text-sm font-medium text-slate-300 mb-2">{lang.label}</label>
                      <textarea
                        value={solution?.code || ''}
                        onChange={(e) => updateSolution(lang.value, e.target.value)}
                        onKeyDown={(e) => handleTabKey(e, lang.value)}
                        rows={16}
                        spellCheck={false}
                        className="w-full bg-slate-950 border border-slate-700 rounded-lg px-4 py-3 text-indigo-300 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-y"
                        placeholder={`${lang.label} solution...`}
                        style={{ 
                          fontFamily: 'Fira Code, Consolas, Monaco, monospace',
                          tabSize: 4,
                          whiteSpace: 'pre',
                          overflowWrap: 'normal',
                          wordWrap: 'normal',
                        }}
                      />
                    </div>
                  );
                })}
              </div>
            ) : (
              <>
                <div className="bg-slate-900/50 px-6 py-4 border-b border-slate-700 flex items-center justify-between">
                  <div className="flex gap-2">
                    {LANGUAGES.map(lang => {
                      const hasSolution = (problem.solutions || []).some(s => s.language === lang.value && s.code && s.code.trim().length > 0);
                      return (
                        <button
                          key={lang.value}
                          onClick={() => setSelectedLanguage(lang.value as any)}
                          className={`px-4 py-1.5 rounded-lg text-sm font-semibold transition-all relative ${
                            selectedLanguage === lang.value ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/20' : 'text-slate-400 hover:text-white hover:bg-slate-800'
                          }`}
                          title={hasSolution ? `${lang.label} solution available` : `No ${lang.label} solution yet`}
                        >
                          {lang.label}
                          {hasSolution && (
                            <span className="absolute -top-1 -right-1 w-2 h-2 bg-emerald-500 rounded-full"></span>
                          )}
                        </button>
                      );
                    })}
                  </div>
                  <button 
                    onClick={handleCopy}
                    className="flex items-center gap-2 px-3 py-1.5 bg-slate-800 hover:bg-slate-700 rounded-lg text-xs font-bold text-slate-300 transition-colors"
                  >
                    {isCopied ? <Check size={14} className="text-emerald-500" /> : <Clipboard size={14} />}
                    {isCopied ? 'Copied!' : 'Copy Code'}
                  </button>
                </div>
                <div className="flex-1 bg-slate-950 p-6 overflow-auto">
                  <pre className="code-font text-indigo-300 text-sm leading-relaxed whitespace-pre" style={{ 
                    fontFamily: 'Fira Code, Consolas, Monaco, monospace',
                    tabSize: 4,
                  }}>
                    {currentSolution && currentSolution.code ? currentSolution.code : `// No solution found for ${LANGUAGES.find(l => l.value === selectedLanguage)?.label || selectedLanguage} language.\n// Switch to edit mode to add a solution.`}
                  </pre>
                </div>
              </>
            )}
          </div>
        )}

        {activeTab === 'notes' && (
          <div className="p-8 space-y-8 overflow-y-auto flex-1 min-h-0">
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-2xl font-bold text-white flex items-center gap-3">
                <Sparkles size={24} className="text-amber-400" />
                Problem Insights
              </h2>
              {!isEditing && (
                <button 
                  onClick={handleGenerateAINote}
                  disabled={isGeneratingNote}
                  className="flex items-center gap-2 px-4 py-2 bg-amber-600 hover:bg-amber-500 disabled:opacity-50 text-white rounded-xl font-bold text-sm transition-all shadow-lg shadow-amber-900/20"
                >
                  {isGeneratingNote ? 'Thinking...' : 'Generate AI Tip'}
                </button>
              )}
            </div>

            <div className="prose prose-invert max-w-none">
              {isEditing && !isDemoUser ? (
                <textarea
                  value={editNotes}
                  onChange={(e) => setEditNotes(e.target.value)}
                  rows={12}
                  className="w-full bg-slate-900/50 border border-slate-700 rounded-3xl px-8 py-6 text-slate-300 leading-relaxed focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-y"
                  placeholder="Your notes and insights..."
                />
              ) : (
                <div className="bg-slate-900/50 p-8 rounded-3xl border border-slate-700 shadow-inner">
                  {(dynamicNote || problem.notes) ? (
                    <MarkdownRenderer content={dynamicNote || problem.notes} />
                  ) : (
                    <div className="text-slate-400 italic">No notes yet. Add your insights here!</div>
                  )}
                </div>
              )}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-8">
              <div className="p-4 bg-emerald-500/5 border border-emerald-500/20 rounded-2xl">
                <h4 className="text-xs font-bold text-emerald-500 uppercase tracking-widest mb-2">Complexity</h4>
                <div className="text-sm font-mono text-slate-300">Time: O(N)</div>
                <div className="text-sm font-mono text-slate-300">Space: O(1)</div>
              </div>
              <div className="p-4 bg-indigo-500/5 border border-indigo-500/20 rounded-2xl">
                <h4 className="text-xs font-bold text-indigo-400 uppercase tracking-widest mb-2">Related Patterns</h4>
                <div className="text-sm text-slate-300">Palindrome, Strings</div>
              </div>
              <div className="p-4 bg-amber-500/5 border border-amber-500/20 rounded-2xl">
                <h4 className="text-xs font-bold text-amber-500 uppercase tracking-widest mb-2">Last Practiced</h4>
                <div className="text-sm text-slate-300">Today, 14:30</div>
              </div>
            </div>
          </div>
        )}
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
                  disabled={isGeneratingProblem || !aiQuery.trim()}
                  className="flex-1 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {isGeneratingProblem ? (
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

      <DeleteConfirmationModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={confirmDelete}
        title="Delete Problem"
        message="Are you sure you want to delete this problem? This action cannot be undone."
        itemName={problem.title}
        isLoading={isDeleting}
      />
    </div>
  );
};

export default ProblemDetail;
