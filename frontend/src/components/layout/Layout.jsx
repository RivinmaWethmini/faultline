import React from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { Sidebar } from './Sidebar';

const pageTitles = {
  '/': 'Overview',
  '/services': 'Services',
  '/dependencies': 'Dependency Map',
  '/incidents': 'Incidents',
  '/simulator': 'Failure Simulator',
};

export const Layout = () => {
  const location = useLocation();

  // Resolve title — handle dynamic routes
  let title = pageTitles[location.pathname] || '';
  if (location.pathname.startsWith('/services/')) title = 'Service Detail';
  if (location.pathname.startsWith('/incidents/') && location.pathname !== '/incidents') title = 'Incident Detail';

  return (
    <div className="app-layout">
      <Sidebar />
      <main className="app-main">
        <header className="app-header">
          <h2>{title}</h2>
        </header>
        <div className="app-content">
          <Outlet />
        </div>
      </main>
    </div>
  );
};
