
import React, { useEffect, useState } from 'react';
import { api } from '../services/apiService';
import { LearningTopic, LearningResource, RoadmapItem } from '../types';
import {
    BookOpen,
    PlayCircle,
    Link as LinkIcon,
    CheckCircle2,
    Circle,
    Clock,
    ExternalLink,
    ChevronRight,
    ArrowRight
} from 'lucide-react';
import MarkdownRenderer from '../components/MarkdownRenderer';

interface LearningTopicDetailProps {
    topic: LearningTopic;
}

const LearningTopicDetail: React.FC<LearningTopicDetailProps> = ({ topic }) => {
    const [resources, setResources] = useState<LearningResource[]>([]);
    const [roadmap, setRoadmap] = useState<RoadmapItem[]>([]);
    const [activeTab, setActiveTab] = useState<'resources' | 'roadmap'>('resources');
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            setLoading(true);
            try {
                const [resData, roadmapData] = await Promise.all([
                    api.getLearningResources(topic.id),
                    api.getRoadmap(topic.id)
                ]);
                setResources(resData);
                setRoadmap(roadmapData);
            } catch (error) {
                console.error('Failed to fetch topic data:', error);
            } finally {
                setLoading(false);
            }
        };
        fetchData();
    }, [topic.id]);

    const getResourceIcon = (type: string) => {
        switch (type) {
            case 'video': return <PlayCircle className="text-red-400" size={20} />;
            case 'link': return <LinkIcon className="text-blue-400" size={20} />;
            default: return <BookOpen className="text-emerald-400" size={20} />;
        }
    };

    const getStatusIcon = (status: string) => {
        switch (status) {
            case 'completed': return <CheckCircle2 className="text-emerald-500" size={20} />;
            case 'in-progress': return <Clock className="text-amber-500" size={20} />;
            default: return <Circle className="text-slate-600" size={20} />;
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
            </div>
        );
    }

    return (
        <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Header */}
            <div className="bg-slate-800/50 rounded-2xl p-8 border border-slate-700/50 backdrop-blur-sm">
                <div className="flex flex-col md:flex-row md:items-center gap-6">
                    <div className="w-16 h-16 bg-indigo-600/20 rounded-2xl flex items-center justify-center text-indigo-400">
                        <BookOpen size={32} />
                    </div>
                    <div className="flex-1">
                        <h1 className="text-3xl font-bold text-white mb-2">{topic.name}</h1>
                        <p className="text-slate-400 max-w-2xl">{topic.description}</p>
                    </div>
                </div>
            </div>

            {/* Tabs */}
            <div className="flex gap-1 p-1 bg-slate-800/50 rounded-xl border border-slate-700/50 w-fit">
                <button
                    onClick={() => setActiveTab('resources')}
                    className={`px-6 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === 'resources'
                            ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/20'
                            : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700/50'
                        }`}
                >
                    Resources
                </button>
                <button
                    onClick={() => setActiveTab('roadmap')}
                    className={`px-6 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === 'roadmap'
                            ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/20'
                            : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700/50'
                        }`}
                >
                    Roadmap
                </button>
            </div>

            {/* Content */}
            <div className="grid grid-cols-1 gap-6">
                {activeTab === 'resources' ? (
                    resources.length > 0 ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {resources.map((res) => (
                                <div key={res.id} className="group bg-slate-800/40 hover:bg-slate-800/60 rounded-xl p-6 border border-slate-700/50 transition-all hover:border-indigo-500/50">
                                    <div className="flex items-start justify-between mb-4">
                                        <div className="p-2 bg-slate-900/50 rounded-lg">
                                            {getResourceIcon(res.type)}
                                        </div>
                                        {res.url && (
                                            <a
                                                href={res.url}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="text-slate-500 hover:text-indigo-400 transition-colors"
                                            >
                                                <ExternalLink size={18} />
                                            </a>
                                        )}
                                    </div>
                                    <h3 className="text-lg font-semibold text-slate-100 mb-2 group-hover:text-indigo-400 transition-colors">
                                        {res.title}
                                    </h3>
                                    <div className="text-sm text-slate-400 line-clamp-3 mb-4">
                                        <MarkdownRenderer content={res.content} />
                                    </div>
                                    <button className="flex items-center gap-2 text-xs font-medium text-indigo-400 hover:text-indigo-300 transition-colors">
                                        Read More <ArrowRight size={14} />
                                    </button>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="text-center py-20 bg-slate-800/20 rounded-2xl border border-dashed border-slate-700">
                            <BookOpen size={48} className="mx-auto text-slate-700 mb-4" />
                            <h3 className="text-lg font-medium text-slate-400">No resources yet</h3>
                            <p className="text-sm text-slate-500">We're still gathering the best materials for this topic.</p>
                        </div>
                    )
                ) : (
                    <div className="max-w-3xl mx-auto w-full space-y-4">
                        {roadmap.length > 0 ? (
                            roadmap.map((item, index) => (
                                <div key={item.id} className="relative pl-12 pb-8 last:pb-0">
                                    {/* Line */}
                                    {index !== roadmap.length - 1 && (
                                        <div className="absolute left-[19px] top-8 bottom-0 w-0.5 bg-slate-700"></div>
                                    )}

                                    {/* Icon */}
                                    <div className="absolute left-0 top-0 z-10 bg-slate-900 p-1 rounded-full">
                                        {getStatusIcon(item.status)}
                                    </div>

                                    <div className="bg-slate-800/40 rounded-xl p-6 border border-slate-700/50 hover:border-slate-600 transition-colors">
                                        <h3 className="text-lg font-semibold text-slate-100 mb-1">{item.title}</h3>
                                        <p className="text-sm text-slate-400">{item.description}</p>
                                    </div>
                                </div>
                            ))
                        ) : (
                            <div className="text-center py-20 bg-slate-800/20 rounded-2xl border border-dashed border-slate-700">
                                <Clock size={48} className="mx-auto text-slate-700 mb-4" />
                                <h3 className="text-lg font-medium text-slate-400">Roadmap coming soon</h3>
                                <p className="text-sm text-slate-500">We're designing the perfect learning path for you.</p>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};

export default LearningTopicDetail;
