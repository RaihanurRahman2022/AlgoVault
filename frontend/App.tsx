
import React, { useState, useEffect } from 'react';
import { api } from './services/apiService';
import { Category, Pattern, Problem, User, ViewState, LearningTopic } from './types';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import CategoryDetail from './pages/CategoryDetail';
import PatternDetail from './pages/PatternDetail';
import ProblemDetail from './pages/ProblemDetail';
import LearningTopicDetail from './pages/LearningTopicDetail';
import Sidebar from './components/Sidebar';
import { ChevronLeft, LogOut, Code2, Bell, Search, User as UserIcon } from 'lucide-react';

const App: React.FC = () => {
  const [user, setUser] = useState<User | null>(() => {
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
  const [selectedLearningTopic, setSelectedLearningTopic] = useState<LearningTopic | null>(null);

  const handleLogin = (u: User, t: string) => {
    setUser(u);
    setToken(t);
    localStorage.setItem('token', t);
    localStorage.setItem('user', JSON.stringify(u));
  };

  const handleLogout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
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
    } else if (viewState === ViewState.LEARNING_TOPIC) {
      setViewState(ViewState.CATEGORIES);
      setSelectedLearningTopic(null);
    }
  };

  const handleSelectPractice = () => {
    setViewState(ViewState.CATEGORIES);
    setSelectedLearningTopic(null);
    setSelectedCategory(null);
    setSelectedPattern(null);
    setSelectedProblem(null);
  };

  const handleSelectTopic = (topic: LearningTopic) => {
    setSelectedLearningTopic(topic);
    setViewState(ViewState.LEARNING_TOPIC);
    // Reset practice states
    setSelectedCategory(null);
    setSelectedPattern(null);
    setSelectedProblem(null);
  };

  if (!user) {
    return <Login onLogin={handleLogin} />;
  }

  return (
    <div className="flex h-screen bg-slate-900 text-slate-100 overflow-hidden">
      {/* Sidebar */}
      <Sidebar
        currentView={viewState}
        onSelectPractice={handleSelectPractice}
        onSelectTopic={handleSelectTopic}
        selectedTopicId={selectedLearningTopic?.id}
      />

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Header */}
        <header className="h-16 bg-slate-800/50 backdrop-blur-md border-b border-slate-700/50 flex items-center justify-between px-8 shrink-0">
          <div className="flex items-center gap-4">
            {viewState !== ViewState.CATEGORIES && (
              <button
                onClick={navigateBack}
                className="p-2 hover:bg-slate-700 rounded-lg transition-colors text-slate-400 hover:text-white"
              >
                <ChevronLeft size={20} />
              </button>
            )}
            <h2 className="text-sm font-medium text-slate-400">
              {viewState === ViewState.CATEGORIES && "Dashboard"}
              {viewState === ViewState.PATTERNS && `Practice / ${selectedCategory?.name}`}
              {viewState === ViewState.PROBLEMS && `Practice / ${selectedCategory?.name} / ${selectedPattern?.name}`}
              {viewState === ViewState.PROBLEM_DETAIL && `Practice / ${selectedProblem?.title}`}
              {viewState === ViewState.LEARNING_TOPIC && `Learning / ${selectedLearningTopic?.name}`}
            </h2>
          </div>

          <div className="flex items-center gap-6">
            <div className="hidden md:flex items-center gap-2 px-3 py-1.5 bg-slate-900/50 border border-slate-700/50 rounded-lg text-slate-500">
              <Search size={16} />
              <span className="text-xs">Search...</span>
            </div>

            <div className="flex items-center gap-4">
              <button className="text-slate-400 hover:text-white transition-colors">
                <Bell size={20} />
              </button>
              <div className="h-8 w-px bg-slate-700/50"></div>
              <div className="flex items-center gap-3">
                <div className="text-right hidden sm:block">
                  <p className="text-xs font-medium text-slate-200">{user.name}</p>
                  <p className="text-[10px] text-slate-500 capitalize">{user.role || 'User'}</p>
                </div>
                <button
                  onClick={handleLogout}
                  className="p-2 bg-slate-700/50 hover:bg-red-500/10 hover:text-red-400 rounded-lg transition-all"
                  title="Logout"
                >
                  <LogOut size={18} />
                </button>
              </div>
            </div>
          </div>
        </header>

        {/* Scrollable Content Area */}
        <main className="flex-1 overflow-y-auto p-8 custom-scrollbar">
          <div className="max-w-6xl mx-auto">
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
                onPatternUpdated={() => { }}
                isDemoUser={user?.role === 'demo'}
              />
            )}

            {viewState === ViewState.PROBLEM_DETAIL && selectedProblem && (
              <ProblemDetail
                problem={selectedProblem}
                onUpdate={(updated) => setSelectedProblem(updated)}
                onDelete={() => {
                  setViewState(ViewState.PROBLEMS);
                  setSelectedProblem(null);
                }}
                isDemoUser={user?.role === 'demo'}
              />
            )}

            {viewState === ViewState.LEARNING_TOPIC && selectedLearningTopic && (
              <LearningTopicDetail topic={selectedLearningTopic} />
            )}
          </div>
        </main>
      </div>
    </div>
  );
};

export default App;

