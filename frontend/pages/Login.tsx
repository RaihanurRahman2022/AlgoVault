
import React, { useState } from 'react';
import { api } from '../services/apiService';
import { User } from '../types';
import { Code2, Mail, Lock, ArrowRight, Loader2 } from 'lucide-react';

interface LoginProps {
  onLogin: (user: User, token: string) => void;
}

const Login: React.FC<LoginProps> = ({ onLogin }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [isRegister, setIsRegister] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    
    try {
      if (isRegister) {
        // Register new user
        const result = await api.register(email, password, name);
        onLogin(result.user, result.token);
      } else {
        // Login existing user
        const result = await api.login(email, password);
        onLogin(result.user, result.token);
      }
    } catch (err: any) {
      setError(err.message || (isRegister ? 'Failed to create account. Please try again.' : 'Invalid credentials. Please try again.'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-md bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl p-8">
        <div className="flex flex-col items-center mb-8">
          <div className="bg-indigo-600/20 p-4 rounded-xl mb-4">
            <Code2 size={40} className="text-indigo-500" />
          </div>
          <h1 className="text-3xl font-bold text-white mb-2">AlgoVault</h1>
          <p className="text-slate-400 text-center">Your companion for technical interview mastery.</p>
        </div>

        <div className="flex gap-2 mb-6">
          <button
            type="button"
            onClick={() => {
              setIsRegister(false);
              setError('');
            }}
            className={`flex-1 py-2 px-4 rounded-lg font-medium transition-all ${
              !isRegister
                ? 'bg-indigo-600 text-white'
                : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
            }`}
          >
            Sign In
          </button>
          <button
            type="button"
            onClick={() => {
              setIsRegister(true);
              setError('');
            }}
            className={`flex-1 py-2 px-4 rounded-lg font-medium transition-all ${
              isRegister
                ? 'bg-indigo-600 text-white'
                : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
            }`}
          >
            Create Account
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          {isRegister && (
            <div className="space-y-2">
              <label className="text-sm font-medium text-slate-300">Full Name</label>
              <div className="relative">
                <input 
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required={isRegister}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg pl-4 pr-4 py-2.5 focus:outline-none focus:ring-2 focus:ring-indigo-500 text-white transition-all"
                  placeholder="John Doe"
                />
              </div>
            </div>
          )}

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">Email Address</label>
            <div className="relative">
              <Mail className="absolute left-3 top-3 text-slate-500" size={20} />
              <input 
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                className="w-full bg-slate-800 border border-slate-700 rounded-lg pl-10 pr-4 py-2.5 focus:outline-none focus:ring-2 focus:ring-indigo-500 text-white transition-all"
                placeholder="you@example.com"
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-slate-300">Password</label>
            <div className="relative">
              <Lock className="absolute left-3 top-3 text-slate-500" size={20} />
              <input 
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="w-full bg-slate-800 border border-slate-700 rounded-lg pl-10 pr-4 py-2.5 focus:outline-none focus:ring-2 focus:ring-indigo-500 text-white transition-all"
                placeholder="••••••••"
              />
            </div>
          </div>

          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
              {error}
            </div>
          )}

          <button 
            type="submit" 
            disabled={loading}
            className="w-full bg-indigo-600 hover:bg-indigo-500 text-white font-semibold py-3 rounded-lg flex items-center justify-center gap-2 transition-all active:scale-95 disabled:opacity-50"
          >
            {loading ? (
              <Loader2 size={20} className="animate-spin" />
            ) : (
              <>
                {isRegister ? 'Create Account' : 'Sign In'}
                <ArrowRight size={20} />
              </>
            )}
          </button>
        </form>

        <div className="mt-6 p-4 bg-indigo-500/10 border border-indigo-500/20 rounded-lg">
          <p className="text-xs font-semibold text-indigo-400 uppercase tracking-wider mb-2">Demo Account</p>
          <p className="text-slate-400 text-sm mb-1">
            <span className="font-mono text-indigo-300">Email:</span> demo@algovault.com
          </p>
          <p className="text-slate-400 text-sm">
            <span className="font-mono text-indigo-300">Password:</span> demo123
          </p>
          <p className="text-slate-500 text-xs mt-2 italic">Demo users have read-only access</p>
        </div>
      </div>
      
      <div className="mt-8 text-slate-600 text-sm flex gap-4">
        <span>Go (Golang) API</span>
        <span>JWT Secured</span>
        <span>SQLite DB</span>
      </div>
    </div>
  );
};

export default Login;
