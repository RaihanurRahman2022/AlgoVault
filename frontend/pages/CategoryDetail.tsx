
import React, { useEffect, useState } from 'react';
import { api } from '../services/apiService';
import { Category, Pattern } from '../types';
import { ICON_MAP } from '../constants';
import PatternForm from '../components/PatternForm';
import DeleteConfirmationModal from '../components/DeleteConfirmationModal';
import { Plus, Search, ChevronRight, Code2, Trash2, Edit } from 'lucide-react';

interface CategoryDetailProps {
  category: Category;
  onSelectPattern: (pat: Pattern) => void;
  isDemoUser?: boolean;
}

const CategoryDetail: React.FC<CategoryDetailProps> = ({ category, onSelectPattern, isDemoUser = false }) => {
  const [patterns, setPatterns] = useState<Pattern[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [editingPattern, setEditingPattern] = useState<Pattern | undefined>();
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; pattern: Pattern | null }>({
    isOpen: false,
    pattern: null,
  });
  const [isDeleting, setIsDeleting] = useState(false);
  
  // Prevent form from showing for demo users
  useEffect(() => {
    if (isDemoUser && showForm) {
      setShowForm(false);
      setEditingPattern(undefined);
    }
  }, [isDemoUser, showForm]);

  const loadPatterns = () => {
    api.getPatterns(category.id)
      .then(data => setPatterns(data || []))
      .catch(err => {
        console.error('Error loading patterns:', err);
        setPatterns([]); // Ensure it's always an array
      });
  };

  useEffect(() => {
    loadPatterns();
  }, [category.id]);

  const filtered = patterns.filter(p => 
    p.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 text-indigo-400 mb-2">
            {ICON_MAP[category.icon]}
            <span className="font-semibold uppercase tracking-widest text-xs">Category</span>
          </div>
          <h1 className="text-4xl font-bold text-white tracking-tight">{category.name} Patterns</h1>
          <p className="text-slate-400 mt-2">{category.description}</p>
        </div>

        <div className="flex gap-3">
          <div className="relative">
            <Search className="absolute left-3 top-2.5 text-slate-500" size={18} />
            <input 
              type="text"
              placeholder="Filter patterns..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="bg-slate-800 border border-slate-700 rounded-xl pl-10 pr-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/50 w-64"
            />
          </div>
          {!isDemoUser && (
            <button 
              onClick={() => {
                setEditingPattern(undefined);
                setShowForm(true);
              }}
              className="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-xl font-semibold transition-all"
            >
              <Plus size={18} />
              Add Pattern
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {filtered.map(pat => (
          <div
            key={pat.id}
            className="relative flex items-center p-5 bg-slate-800 border border-slate-700 hover:border-slate-500 hover:bg-slate-750 rounded-2xl transition-all group"
          >
            <button 
              onClick={() => onSelectPattern(pat)}
              className="flex-1 flex items-center text-left"
            >
            <div className="w-12 h-12 flex-shrink-0 bg-slate-900 border border-slate-700 rounded-xl flex items-center justify-center text-indigo-400 mr-5 group-hover:scale-110 transition-transform">
              {/* Fallback to Code2 icon if pattern icon is not found in map */}
              {ICON_MAP[pat.icon] || <Code2 size={24} />}
            </div>
            <div className="flex-1">
              <h3 className="text-lg font-bold text-white group-hover:text-indigo-400 transition-colors">
                {pat.name}
              </h3>
              <p className="text-slate-400 text-sm line-clamp-1">{pat.description}</p>
            </div>
            <div className="flex items-center gap-4 ml-4">
              <span className="text-xs font-semibold px-2 py-1 bg-slate-900 border border-slate-700 rounded-lg text-slate-400">
                {pat.problemCount} Problems
              </span>
              <ChevronRight size={20} className="text-slate-600 group-hover:text-indigo-400 transition-colors" />
            </div>
            </button>
            {!isDemoUser && (
              <div className="flex gap-2 ml-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setEditingPattern(pat);
                    setShowForm(true);
                  }}
                  className="p-2 bg-slate-700 hover:bg-slate-600 rounded-lg text-slate-400 hover:text-white transition-colors"
                  title="Edit"
                >
                  <Edit size={16} />
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setDeleteModal({ isOpen: true, pattern: pat });
                  }}
                  className="p-2 bg-red-500/20 hover:bg-red-500/30 rounded-lg text-red-400 hover:text-red-300 transition-colors"
                  title="Delete"
                >
                  <Trash2 size={16} />
                </button>
              </div>
            )}
          </div>
        ))}
        {filtered.length === 0 && (
          <div className="col-span-full py-12 flex flex-col items-center justify-center bg-slate-800/20 rounded-2xl border-2 border-dashed border-slate-700">
            <div className="p-4 bg-slate-800 rounded-full mb-4">
              <Search size={32} className="text-slate-600" />
            </div>
            <p className="text-slate-400">No patterns found in this category.</p>
          </div>
        )}
      </div>

      {showForm && !isDemoUser && (
        <PatternForm
          pattern={editingPattern}
          categoryId={category.id}
          onSave={async (pattern) => {
            if (editingPattern) {
              await api.updatePattern(editingPattern.id, pattern);
            } else {
              await api.createPattern(category.id, pattern);
            }
            loadPatterns();
          }}
          onClose={() => {
            setShowForm(false);
            setEditingPattern(undefined);
          }}
        />
      )}

      <DeleteConfirmationModal
        isOpen={deleteModal.isOpen}
        onClose={() => setDeleteModal({ isOpen: false, pattern: null })}
        onConfirm={async () => {
          if (deleteModal.pattern) {
            setIsDeleting(true);
            try {
              await api.deletePattern(deleteModal.pattern.id);
              loadPatterns();
              setDeleteModal({ isOpen: false, pattern: null });
            } catch (error) {
              console.error('Error deleting pattern:', error);
            } finally {
              setIsDeleting(false);
            }
          }
        }}
        title="Delete Pattern"
        message="Are you sure you want to delete this pattern? All problems under this pattern will also be deleted."
        itemName={deleteModal.pattern?.name}
        isLoading={isDeleting}
      />
    </div>
  );
};

export default CategoryDetail;