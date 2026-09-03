import React from 'react';
import { NavLink } from 'react-router-dom';

const navItems = [
  { to: '/', label: 'Overview', icon: '◉' },
  { to: '/services', label: 'Services', icon: '⬡' },
  { to: '/dependencies', label: 'Dependencies', icon: '⤮' },
  { to: '/incidents', label: 'Incidents', icon: '⚠' },
  { to: '/simulator', label: 'Simulator', icon: '⚡' },
];

export const Sidebar = () => {
  return (
    <aside className="app-sidebar">
      <div className="sidebar-logo">
        <div className="logo-icon">
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
            <path d="M1 8L6 1L11 8" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M3 11L6 6L9 11" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" opacity="0.6"/>
          </svg>
        </div>
        <h1>Faultline</h1>
      </div>
      <nav className="sidebar-nav">
        <div className="sidebar-section">Platform</div>
        {navItems.map(item => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) => (isActive ? 'active' : '')}
          >
            <span style={{ fontSize: '14px', width: '18px', textAlign: 'center' }}>{item.icon}</span>
            {item.label}
          </NavLink>
        ))}
      </nav>
      <div style={{ padding: '12px 20px', borderTop: '1px solid var(--border-primary)', fontSize: 'var(--font-size-xs)', color: 'var(--text-tertiary)' }}>
        Faultline v1.0.0
      </div>
    </aside>
  );
};
