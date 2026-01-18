
import React, { useEffect, useState } from 'react';
import { api } from '../services/apiService';
import { Category } from '../types';
import { ICON_MAP } from '../constants';
import CategoryForm from '../components/CategoryForm';
import DeleteConfirmationModal from '../components/DeleteConfirmationModal';
import { Plus, ChevronRight, Code2, Trash2, Edit } from 'lucide-react';

interface DashboardProps {
  onSelectCategory: (cat: Category) => void;
  isDemoUser?: boolean;
}

const Dashboard: React.FC<DashboardProps> = ({ onSelectCategory, isDemoUser = false }) => {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingCategory, setEditingCategory] = useState<Category | undefined>();
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; category: Category | null }>({
    isOpen: false,
    category: null,
  });
  const [isDeleting, setIsDeleting] = useState(false);

  // Prevent form from showing for demo users
  useEffect(() => {
    if (isDemoUser && showForm) {
      setShowForm(false);
      setEditingCategory(undefined);
    }
  }, [isDemoUser, showForm]);

  const loadCategories = () => {
    setLoading(true);
    setError(null);
    api.getCategories()
      .then(data => {
        setCategories(data || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Error loading categories:', err);
        setError(err.message || 'Failed to load categories');
        setCategories([]); // Ensure it's always an array
        setLoading(false);
      });
  };

  useEffect(() => {
    loadCategories();
  }, []);

  if (loading) {
    return (
      <div>
        <div className="flex justify-between items-end mb-8 animate-pulse">
          <div className="space-y-3">
            <div className="h-10 bg-slate-800 rounded-xl w-64" />
            <div className="h-4 bg-slate-800 rounded w-96" />
          </div>
          <div className="h-11 bg-slate-800 rounded-xl w-40" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3, 4, 5, 6].map(i => (
            <div key={i} className="h-64 bg-slate-800/40 rounded-2xl border border-slate-700/50 animate-pulse flex flex-col p-6 space-y-4">
              <div className="w-12 h-12 bg-slate-700/50 rounded-xl" />
              <div className="space-y-2">
                <div className="h-6 bg-slate-700/50 rounded w-2/3" />
                <div className="h-4 bg-slate-700/30 rounded w-full" />
                <div className="h-4 bg-slate-700/30 rounded w-5/6" />
              </div>
              <div className="mt-auto h-4 bg-slate-700/50 rounded w-1/3" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div>
      {error && (
        <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400">
          <p className="font-semibold">Error: {error}</p>
          <button
            onClick={loadCategories}
            className="mt-2 text-sm underline hover:text-red-300"
          >
            Retry
          </button>
        </div>
      )}
      <div className="flex justify-between items-end mb-8">
        <div>
          <h1 className="text-4xl font-bold text-white tracking-tight mb-2">Practice Topics</h1>
          <p className="text-slate-400">Select a category to explore problem patterns.</p>
        </div>
        {!isDemoUser && (
          <button
            onClick={() => {
              setEditingCategory(undefined);
              setShowForm(true);
            }}
            className="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-xl font-semibold transition-all"
          >
            <Plus size={20} />
            New Category
          </button>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {categories && categories.length > 0 ? categories.map(cat => (
          <div
            key={cat.id}
            className="group relative bg-slate-800 border border-slate-700 hover:border-indigo-500/50 hover:bg-slate-700/50 p-6 rounded-2xl transition-all duration-300 shadow-lg hover:shadow-indigo-500/10"
          >
            <div
              onClick={() => onSelectCategory(cat)}
              className="w-full text-left cursor-pointer"
            >
              <div className="flex justify-between items-start mb-4 relative">
                <div className="p-3 bg-slate-900 rounded-xl text-indigo-400 group-hover:bg-indigo-500 group-hover:text-white transition-all">
                  {/* Fallback to Code2 icon if category icon is not found in map */}
                  {ICON_MAP[cat.icon] || <Code2 size={24} />}
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-bold uppercase tracking-wider text-slate-500 bg-slate-900 px-2.5 py-1 rounded-full border border-slate-700 z-10">
                    {cat.patternCount} Patterns
                  </span>
                  {!isDemoUser && (
                    <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setEditingCategory(cat);
                          setShowForm(true);
                        }}
                        className="p-1.5 bg-slate-700 hover:bg-slate-600 rounded-lg text-slate-400 hover:text-white transition-colors"
                        title="Edit"
                      >
                        <Edit size={14} />
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setDeleteModal({ isOpen: true, category: cat });
                        }}
                        className="p-1.5 bg-red-500/20 hover:bg-red-500/30 rounded-lg text-red-400 hover:text-red-300 transition-colors"
                        title="Delete"
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  )}
                </div>
              </div>

              <h3 className="text-xl font-bold text-white mb-2 group-hover:text-indigo-400 transition-colors">
                {cat.name}
              </h3>
              <p className="text-slate-400 text-sm line-clamp-2 leading-relaxed">
                {cat.description}
              </p>

              <div className="mt-4 flex items-center text-indigo-400 font-medium text-sm opacity-0 group-hover:opacity-100 transition-opacity">
                Explore Patterns <ChevronRight size={16} className="ml-1" />
              </div>
            </div>
          </div>
        )) : (
          <div className="col-span-full py-12 flex flex-col items-center justify-center bg-slate-800/20 rounded-2xl border-2 border-dashed border-slate-700">
            <p className="text-slate-400 mb-4">No categories yet. Create your first category to get started!</p>
            {!isDemoUser && (
              <button
                onClick={() => {
                  setEditingCategory(undefined);
                  setShowForm(true);
                }}
                className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-semibold transition-all"
              >
                <Plus size={20} className="inline mr-2" />
                Create First Category
              </button>
            )}
          </div>
        )}
      </div>

      {showForm && !isDemoUser && (
        <CategoryForm
          category={editingCategory}
          onSave={async (category) => {
            if (editingCategory) {
              await api.updateCategory(editingCategory.id, category);
            } else {
              await api.createCategory(category);
            }
            loadCategories();
          }}
          onClose={() => {
            setShowForm(false);
            setEditingCategory(undefined);
          }}
        />
      )}

      <DeleteConfirmationModal
        isOpen={deleteModal.isOpen}
        onClose={() => setDeleteModal({ isOpen: false, category: null })}
        onConfirm={async () => {
          if (deleteModal.category) {
            setIsDeleting(true);
            try {
              await api.deleteCategory(deleteModal.category.id);
              loadCategories();
              setDeleteModal({ isOpen: false, category: null });
            } catch (error) {
              console.error('Error deleting category:', error);
            } finally {
              setIsDeleting(false);
            }
          }
        }}
        title="Delete Category"
        message="Are you sure you want to delete this category? All patterns and problems under this category will also be deleted."
        itemName={deleteModal.category?.name}
        isLoading={isDeleting}
      />
    </div>
  );
};

export default Dashboard;