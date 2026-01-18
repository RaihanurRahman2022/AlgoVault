
import React, { useEffect, useState } from 'react';
import {
    Code2,
    Layout,
    Server,
    Box,
    Cloud,
    Code,
    Users,
    Terminal,
    BookOpen,
    CheckCircle2,
    ChevronRight,
    Menu,
    X
} from 'lucide-react';
import { api } from '../services/apiService';
import { LearningTopic, ViewState } from '../types';

interface SidebarProps {
    currentView: ViewState;
    onSelectPractice: () => void;
    onSelectTopic: (topic: LearningTopic) => void;
    selectedTopicId?: string;
}

const Sidebar: React.FC<SidebarProps> = ({
    currentView,
    onSelectPractice,
    onSelectTopic,
    selectedTopicId
}) => {
    const [topics, setTopics] = useState<LearningTopic[]>([]);
    const [isOpen, setIsOpen] = useState(true);

    useEffect(() => {
        const fetchTopics = async () => {
            try {
                const data = await api.getLearningTopics();
                setTopics(data);
            } catch (error) {
                console.error('Failed to fetch topics:', error);
            }
        };
        fetchTopics();
    }, []);

    const getIcon = (iconName: string) => {
        switch (iconName) {
            case 'Layout': return <Layout size={18} />;
            case 'Server': return <Server size={18} />;
            case 'Box': return <Box size={18} />;
            case 'Cloud': return <Cloud size={18} />;
            case 'Code': return <Code size={18} />;
            case 'Users': return <Users size={18} />;
            case 'Terminal': return <Terminal size={18} />;
            default: return <BookOpen size={18} />;
        }
    };

    const isPracticeActive = [
        ViewState.CATEGORIES,
        ViewState.PATTERNS,
        ViewState.PROBLEMS,
        ViewState.PROBLEM_DETAIL
    ].includes(currentView) && !selectedTopicId;

    return (
        <>
            {/* Mobile Toggle */}
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="lg:hidden fixed bottom-6 right-6 z-50 p-3 bg-indigo-600 text-white rounded-full shadow-lg hover:bg-indigo-700 transition-all"
            >
                {isOpen ? <X size={24} /> : <Menu size={24} />}
            </button>

            <aside className={`
        fixed inset-y-0 left-0 z-40 w-64 bg-slate-800 border-r border-slate-700 transform transition-transform duration-300 ease-in-out
        ${isOpen ? 'translate-x-0' : '-translate-x-full'}
        lg:translate-x-0 lg:static lg:block
      `}>
                <div className="flex flex-col h-full">
                    {/* Logo */}
                    <div className="p-6 border-b border-slate-700">
                        <div className="flex items-center gap-2 text-indigo-400 font-bold text-xl">
                            <Code2 size={28} />
                            <span>AlgoVault</span>
                        </div>
                    </div>

                    {/* Navigation */}
                    <div className="flex-1 overflow-y-auto p-4 space-y-8">
                        {/* Practice Section */}
                        <div>
                            <h3 className="px-3 text-xs font-semibold text-slate-500 uppercase tracking-wider mb-4">
                                Practice
                            </h3>
                            <button
                                onClick={onSelectPractice}
                                className={`
                  w-full flex items-center justify-between px-3 py-2 rounded-lg transition-all group
                  ${isPracticeActive
                                        ? 'bg-indigo-600/20 text-indigo-400 font-medium'
                                        : 'text-slate-400 hover:bg-slate-700 hover:text-slate-200'}
                `}
                            >
                                <div className="flex items-center gap-3">
                                    <Code2 size={18} className={isPracticeActive ? 'text-indigo-400' : 'text-slate-500 group-hover:text-slate-300'} />
                                    <span>Problem Tracker</span>
                                </div>
                                {isPracticeActive && <div className="w-1.5 h-1.5 rounded-full bg-indigo-400" />}
                            </button>
                        </div>

                        {/* Learning Section */}
                        <div>
                            <h3 className="px-3 text-xs font-semibold text-slate-500 uppercase tracking-wider mb-4">
                                Learning Resources
                            </h3>
                            <div className="space-y-1">
                                {topics.map((topic) => {
                                    const isActive = selectedTopicId === topic.id;
                                    return (
                                        <button
                                            key={topic.id}
                                            onClick={() => onSelectTopic(topic)}
                                            className={`
                        w-full flex items-center justify-between px-3 py-2.5 rounded-lg transition-all group
                        ${isActive
                                                    ? 'bg-indigo-600/20 text-indigo-400 font-medium'
                                                    : 'text-slate-400 hover:bg-slate-700 hover:text-slate-200'}
                      `}
                                        >
                                            <div className="flex items-center gap-3">
                                                <span className={isActive ? 'text-indigo-400' : 'text-slate-500 group-hover:text-slate-300'}>
                                                    {getIcon(topic.icon)}
                                                </span>
                                                <span className="text-sm">{topic.name}</span>
                                            </div>
                                            {isActive ? (
                                                <div className="w-1.5 h-1.5 rounded-full bg-indigo-400" />
                                            ) : (
                                                <ChevronRight size={14} className="opacity-0 group-hover:opacity-100 text-slate-600 transition-opacity" />
                                            )}
                                        </button>
                                    );
                                })}
                            </div>
                        </div>

                        {/* Roadmap Section (Coming Soon) */}
                        <div className="pt-4 border-t border-slate-700/50">
                            <div className="px-3 py-2 bg-slate-900/50 rounded-lg border border-slate-700/50">
                                <div className="flex items-center gap-2 text-xs font-medium text-slate-500 mb-1">
                                    <CheckCircle2 size={12} />
                                    <span>YOUR PROGRESS</span>
                                </div>
                                <div className="w-full bg-slate-800 rounded-full h-1.5 mb-1">
                                    <div className="bg-indigo-500 h-1.5 rounded-full w-1/3"></div>
                                </div>
                                <span className="text-[10px] text-slate-600">12 of 45 topics completed</span>
                            </div>
                        </div>
                    </div>

                    {/* User Info (Mobile) */}
                    <div className="lg:hidden p-4 border-t border-slate-700 bg-slate-800/50">
                        <div className="flex items-center gap-3 px-3 py-2">
                            <div className="w-8 h-8 rounded-full bg-indigo-600 flex items-center justify-center text-xs font-bold">
                                U
                            </div>
                            <div className="flex-1 min-w-0">
                                <p className="text-sm font-medium text-slate-200 truncate">User</p>
                                <p className="text-xs text-slate-500 truncate">user@example.com</p>
                            </div>
                        </div>
                    </div>
                </div>
            </aside>
        </>
    );
};

export default Sidebar;
