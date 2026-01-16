
import React, { useState, useEffect } from 'react';
import { api } from './services/apiService';
import { Category, Pattern, Problem, User, ViewState } from './types';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import CategoryDetail from './pages/CategoryDetail';
import PatternDetail from './pages/PatternDetail';
import ProblemDetail from './pages/ProblemDetail';
import { ChevronLeft, LogOut, Code2 } from 'lucide-react';

const App: React.FC = () => {
  const [user, setUser] = useState<User | null>(() => {
    // Load user from localStorage on initial load
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      try {
        return JSON.parse(storedUser);
      } catch {
        return null;
      }
    }
    return null;
  });
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
  const [viewState, setViewState] = useState<ViewState>(ViewState.CATEGORIES);
  
  const [selectedCategory, setSelectedCategory] = useState<Category | null>(null);
  const [selectedPattern, setSelectedPattern] = useState<Pattern | null>(null);
  const [selectedProblem, setSelectedProblem] = useState<Problem | null>(null);

  const handleLogin = (u: User, t: string) => {
    setUser(u);
    setToken(t);
    localStorage.setItem('token', t);
    localStorage.setItem('user', JSON.stringify(u)); // Store user with role
  };

  const handleLogout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user'); // Remove user from localStorage
  };

  const navigateBack = () => {
    if (viewState === ViewState.PROBLEM_DETAIL) {
      setViewState(ViewState.PROBLEMS);
      setSelectedProblem(null);
    } else if (viewState === ViewState.PROBLEMS) {
      setViewState(ViewState.PATTERNS);
      setSelectedPattern(null);
    } else if (viewState === ViewState.PATTERNS) {
      setViewState(ViewState.CATEGORIES);
      setSelectedCategory(null);
    }
  };

  if (!user) {
    return <Login onLogin={handleLogin} />;
  }

  return (
    <div className="flex flex-col min-h-screen bg-slate-900 text-slate-100">
      {/* Navbar */}
      <header className="sticky top-0 z-50 bg-slate-800/80 backdrop-blur-md border-b border-slate-700 px-6 py-4">
        <div className="flex items-center justify-between max-w-7xl mx-auto">
          <div className="flex items-center gap-3">
            {viewState !== ViewState.CATEGORIES && (
              <button 
                onClick={navigateBack}
                className="p-1 hover:bg-slate-700 rounded-lg transition-colors"
              >
                <ChevronLeft size={24} />
              </button>
            )}
            <div className="flex items-center gap-2 text-indigo-400 font-bold text-xl">
              <Code2 size={28} />
              <span>AlgoVault</span>
            </div>
          </div>
          
          <div className="flex items-center gap-4">
            <span className="text-sm text-slate-400">Welcome, {user.name}</span>
            <button 
              onClick={handleLogout}
              className="flex items-center gap-2 px-3 py-1.5 bg-slate-700 hover:bg-red-600/20 hover:text-red-400 rounded-lg text-sm font-medium transition-all"
            >
              <LogOut size={16} />
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 w-full max-w-7xl mx-auto p-6 md:p-8">
        {viewState === ViewState.CATEGORIES && (
          <Dashboard 
            onSelectCategory={(cat) => {
              setSelectedCategory(cat);
              setViewState(ViewState.PATTERNS);
            }}
            isDemoUser={user?.role === 'demo'}
          />
        )}

        {viewState === ViewState.PATTERNS && selectedCategory && (
          <CategoryDetail 
            category={selectedCategory}
            onSelectPattern={(pat) => {
              setSelectedPattern(pat);
              setViewState(ViewState.PROBLEMS);
            }}
            isDemoUser={user?.role === 'demo'}
          />
        )}

        {viewState === ViewState.PROBLEMS && selectedPattern && selectedCategory && (
          <PatternDetail 
            pattern={selectedPattern}
            categoryName={selectedCategory.name}
            onSelectProblem={(prob) => {
              setSelectedProblem(prob);
              setViewState(ViewState.PROBLEM_DETAIL);
            }}
            onPatternUpdated={() => {
              // Refresh will happen automatically when navigating back
            }}
            isDemoUser={user?.role === 'demo'}
          />
        )}

        {viewState === ViewState.PROBLEM_DETAIL && selectedProblem && (
          <ProblemDetail 
            problem={selectedProblem}
            onUpdate={(updated) => {
              setSelectedProblem(updated);
              // If we have a pattern selected, we could refresh the problems list
              // but since we're in detail view, just update the current problem
            }}
            onDelete={() => {
              // Navigate back to problems list after deletion
              setViewState(ViewState.PROBLEMS);
              setSelectedProblem(null);
            }}
            isDemoUser={user?.role === 'demo'}
          />
        )}
      </main>

      {/* Footer */}
      <footer className="bg-slate-950 border-t border-slate-800 py-6 text-center text-slate-500 text-sm">
        <p>&copy; {new Date().getFullYear()} AlgoVault. All Practice Problems in One Place.</p>
        <p className="mt-1">Powered by Gemini AI and Golang Backend</p>
      </footer>
    </div>
  );
};

export default App;
