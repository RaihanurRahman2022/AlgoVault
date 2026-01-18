
import React, { useEffect, useState, useMemo } from 'react';
import { api } from '../services/apiService';
import { Pattern, Problem } from '../types';
import ProblemForm from '../components/ProblemForm';
import DeleteConfirmationModal from '../components/DeleteConfirmationModal';
import TheoryModal from '../components/TheoryModal';
import { Plus, Filter, SortAsc, SortDesc, CheckCircle2, Trash2, X, BookOpen } from 'lucide-react';

interface PatternDetailProps {
  pattern: Pattern;
  categoryName: string;
  onSelectProblem: (prob: Problem) => void;
  onProblemUpdated?: () => void;
  onPatternUpdated?: () => void;
  isDemoUser?: boolean;
}

type DifficultyFilter = 'All' | 'Easy' | 'Medium' | 'Hard';
type SortOrder = 'asc' | 'desc' | null;

const PatternDetail: React.FC<PatternDetailProps> = ({ pattern, categoryName, onSelectProblem, onPatternUpdated, isDemoUser = false }) => {
  const [problems, setProblems] = useState<Problem[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingProblem, setEditingProblem] = useState<Problem | undefined>();
  const [difficultyFilter, setDifficultyFilter] = useState<DifficultyFilter>('All');
  const [sortOrder, setSortOrder] = useState<SortOrder>(null);
  const [showFilterMenu, setShowFilterMenu] = useState(false);
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; problem: Problem | null }>({
    isOpen: false,
    problem: null,
  });
  const [isDeleting, setIsDeleting] = useState(false);
  const [showTheoryModal, setShowTheoryModal] = useState(false);
  const [currentPattern, setCurrentPattern] = useState<Pattern>(pattern);

  const loadProblems = () => {
    setLoading(true);
    api.getProblems(pattern.id)
      .then(data => {
        setProblems(data || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Error loading problems:', err);
        setProblems([]); // Ensure it's always an array
        setLoading(false);
      });
  };

  useEffect(() => {
    loadProblems();
    setCurrentPattern(pattern);
  }, [pattern.id, pattern]);

  // Close filter menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (!target.closest('.filter-menu-container')) {
        setShowFilterMenu(false);
      }
    };

    if (showFilterMenu) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [showFilterMenu]);

  // Filter and sort problems
  const filteredAndSortedProblems = useMemo(() => {
    let filtered = problems;

    // Apply difficulty filter
    if (difficultyFilter !== 'All') {
      filtered = filtered.filter(p => p.difficulty === difficultyFilter);
    }

    // Apply sorting
    if (sortOrder) {
      const difficultyOrder = { 'Easy': 1, 'Medium': 2, 'Hard': 3 };
      filtered = [...filtered].sort((a, b) => {
        const aOrder = difficultyOrder[a.difficulty];
        const bOrder = difficultyOrder[b.difficulty];
        return sortOrder === 'asc' ? aOrder - bOrder : bOrder - aOrder;
      });
    }

    return filtered;
  }, [problems, difficultyFilter, sortOrder]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-3xl font-bold text-white tracking-tight">{pattern.name}</h1>
            <button
              onClick={() => setShowTheoryModal(true)}
              className="p-2 text-slate-400 hover:text-indigo-400 hover:bg-slate-700 rounded-lg transition-colors"
              title="View Pattern Theory"
            >
              <BookOpen size={20} />
            </button>
          </div>
          <p className="text-slate-400 mt-1 max-w-2xl">{pattern.description}</p>
        </div>
        {!isDemoUser && (
          <button
            onClick={() => {
              setEditingProblem(undefined);
              setShowForm(true);
            }}
            className="flex items-center gap-2 px-5 py-2.5 bg-emerald-600 hover:bg-emerald-500 rounded-xl font-bold transition-all shadow-lg shadow-emerald-900/20"
          >
            <Plus size={20} />
            New Problem
          </button>
        )}
      </div>

      <div className="flex items-center gap-4 py-4 border-y border-slate-800">
        {/* Filter Dropdown */}
        <div className="relative filter-menu-container">
          <button
            onClick={() => setShowFilterMenu(!showFilterMenu)}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm border transition-colors ${difficultyFilter !== 'All'
              ? 'bg-indigo-600/20 text-indigo-400 border-indigo-500/30'
              : 'bg-slate-800 text-slate-300 border-slate-700 hover:bg-slate-700'
              }`}
          >
            <Filter size={16} />
            Filter
            {difficultyFilter !== 'All' && (
              <span className="ml-1 text-xs">({difficultyFilter})</span>
            )}
          </button>
          {showFilterMenu && (
            <div className="absolute top-full left-0 mt-2 bg-slate-800 border border-slate-700 rounded-lg shadow-xl z-10 min-w-[120px]">
              {(['All', 'Easy', 'Medium', 'Hard'] as DifficultyFilter[]).map(diff => (
                <button
                  key={diff}
                  onClick={() => {
                    setDifficultyFilter(diff);
                    setShowFilterMenu(false);
                  }}
                  className={`w-full text-left px-4 py-2 text-sm transition-colors ${difficultyFilter === diff
                    ? 'bg-indigo-600 text-white'
                    : 'text-slate-300 hover:bg-slate-700'
                    }`}
                >
                  {diff}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Sort Button */}
        <button
          onClick={() => {
            if (sortOrder === null) {
              setSortOrder('asc');
            } else if (sortOrder === 'asc') {
              setSortOrder('desc');
            } else {
              setSortOrder(null);
            }
          }}
          className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm border transition-colors ${sortOrder
            ? 'bg-indigo-600/20 text-indigo-400 border-indigo-500/30'
            : 'bg-slate-800 text-slate-300 border-slate-700 hover:bg-slate-700'
            }`}
        >
          {sortOrder === 'asc' ? <SortAsc size={16} /> : sortOrder === 'desc' ? <SortDesc size={16} /> : <SortAsc size={16} />}
          Difficulty
        </button>

        {/* Clear filters */}
        {(difficultyFilter !== 'All' || sortOrder !== null) && (
          <button
            onClick={() => {
              setDifficultyFilter('All');
              setSortOrder(null);
            }}
            className="flex items-center gap-1 px-3 py-1.5 bg-slate-800 rounded-lg text-sm text-slate-400 border border-slate-700 hover:bg-slate-700 hover:text-white transition-colors"
          >
            <X size={14} />
            Clear
          </button>
        )}

        <div className="ml-auto text-sm text-slate-500">
          Showing {filteredAndSortedProblems.length} of {problems.length} problems
        </div>
      </div>

      <div className="overflow-hidden bg-slate-800 border border-slate-700 rounded-2xl shadow-xl">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-slate-900/50 text-slate-400 text-xs font-bold uppercase tracking-widest border-b border-slate-700">
              <th className="px-6 py-4 w-12">Status</th>
              <th className="px-6 py-4">Problem Name</th>
              <th className="px-6 py-4">Difficulty</th>
              <th className="px-6 py-4">Solutions</th>
              <th className="px-6 py-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-700/50">
            {loading ? (
              [1, 2, 3, 4, 5, 6].map(i => (
                <tr key={i} className="animate-pulse">
                  <td className="px-6 py-5"><div className="w-5 h-5 bg-slate-700/50 rounded-full" /></td>
                  <td className="px-6 py-5"><div className="h-5 bg-slate-700/50 rounded w-1/2" /></td>
                  <td className="px-6 py-5"><div className="h-6 bg-slate-700/30 rounded-full w-20" /></td>
                  <td className="px-6 py-5">
                    <div className="flex gap-2">
                      <div className="w-12 h-4 bg-slate-700/30 rounded" />
                      <div className="w-12 h-4 bg-slate-700/30 rounded" />
                    </div>
                  </td>
                  <td className="px-6 py-5 text-right"><div className="h-5 bg-slate-700/50 rounded w-8 ml-auto" /></td>
                </tr>
              ))
            ) : filteredAndSortedProblems.map(prob => (
              <tr
                key={prob.id}
                className="hover:bg-slate-750 transition-colors group cursor-pointer"
                onClick={() => onSelectProblem(prob)}
              >
                <td className="px-6 py-4">
                  <CheckCircle2 size={20} className="text-emerald-500" />
                </td>
                <td className="px-6 py-4">
                  <span className="text-white font-medium group-hover:text-indigo-400 transition-colors">{prob.title}</span>
                </td>
                <td className="px-6 py-4">
                  <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border ${prob.difficulty === 'Easy' ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' :
                    prob.difficulty === 'Medium' ? 'bg-amber-500/10 text-amber-500 border-amber-500/20' :
                      'bg-red-500/10 text-red-500 border-red-500/20'
                    }`}>
                    {prob.difficulty}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <div className="flex gap-2">
                    {(prob.solutions || []).map(sol => (
                      <span key={sol.id} className="px-2 py-0.5 bg-slate-900 border border-slate-700 rounded text-[10px] font-mono text-slate-400 uppercase">
                        {sol.language}
                      </span>
                    ))}
                  </div>
                </td>
                <td className="px-6 py-4 text-right">
                  {!isDemoUser && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        setDeleteModal({ isOpen: true, problem: prob });
                      }}
                      className="p-1.5 text-slate-400 hover:text-red-400 transition-colors"
                      title="Delete"
                    >
                      <Trash2 size={16} />
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {filteredAndSortedProblems.length === 0 && !loading && (
          <div className="p-12 text-center text-slate-500 italic">
            {difficultyFilter !== 'All' || sortOrder !== null
              ? 'No problems match the current filter.'
              : 'No problems added yet. Time to start practicing!'}
          </div>
        )}
      </div>

      {showForm && (
        <ProblemForm
          problem={editingProblem}
          patternId={pattern.id}
          onSave={async (problem) => {
            if (editingProblem) {
              await api.updateProblem(editingProblem.id, problem);
            } else {
              await api.createProblem(pattern.id, problem);
            }
            loadProblems();
          }}
          onClose={() => {
            setShowForm(false);
            setEditingProblem(undefined);
          }}
        />
      )}

      <DeleteConfirmationModal
        isOpen={deleteModal.isOpen}
        onClose={() => setDeleteModal({ isOpen: false, problem: null })}
        onConfirm={async () => {
          if (deleteModal.problem) {
            setIsDeleting(true);
            try {
              await api.deleteProblem(deleteModal.problem.id);
              loadProblems();
              setDeleteModal({ isOpen: false, problem: null });
            } catch (error) {
              console.error('Error deleting problem:', error);
            } finally {
              setIsDeleting(false);
            }
          }
        }}
        title="Delete Problem"
        message="Are you sure you want to delete this problem? This action cannot be undone."
        itemName={deleteModal.problem?.title}
        isLoading={isDeleting}
      />

      <TheoryModal
        patternId={currentPattern.id}
        patternName={currentPattern.name}
        categoryName={categoryName}
        initialTheory={currentPattern.theory || ''}
        isOpen={showTheoryModal}
        onClose={() => setShowTheoryModal(false)}
        onUpdate={async (theory) => {
          await api.updatePatternTheory(currentPattern.id, theory);
          setCurrentPattern({ ...currentPattern, theory });
          if (onPatternUpdated) {
            onPatternUpdated();
          }
        }}
        isDemoUser={isDemoUser}
      />
    </div>
  );
};

export default PatternDetail;
