import React, { useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { usePolling } from '../hooks/usePolling';
import { getRiskBadgeClass } from '../utils/colors';
import { timeAgo, formatDuration } from '../utils/formatters';

export const Incidents = () => {
  const navigate = useNavigate();
  const [statusFilter, setStatusFilter] = useState('');

  const fetchIncidents = useCallback(() => api.getIncidents(statusFilter), [statusFilter]);
  const { data: incidents, loading } = usePolling(fetchIncidents, 4000);

  if (loading) {
    return <div className="loading-container"><div className="spinner" /></div>;
  }

  return (
    <div>
      {/* Filter tabs */}
      <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-4)' }}>
        <div className="filter-tabs">
          {[
            { label: 'All', value: '' },
            { label: 'Open', value: 'OPEN' },
            { label: 'Investigating', value: 'INVESTIGATING' },
            { label: 'Resolved', value: 'RESOLVED' },
          ].map(tab => (
            <button
              key={tab.value}
              className={`filter-tab ${statusFilter === tab.value ? 'active' : ''}`}
              onClick={() => setStatusFilter(tab.value)}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      <div className="card" style={{ padding: 0 }}>
        {!incidents || incidents.length === 0 ? (
          <div className="empty-state">
            <h3>No incidents</h3>
            <p>{statusFilter ? `No ${statusFilter.toLowerCase()} incidents recorded.` : 'No incidents recorded yet.'}</p>
          </div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Severity</th>
                  <th>Service</th>
                  <th>Incident Title</th>
                  <th>Status</th>
                  <th>Risk Score</th>
                  <th>Duration</th>
                  <th>Started</th>
                </tr>
              </thead>
              <tbody>
                {incidents.map(inc => (
                  <tr key={inc.id} className="clickable-row" onClick={() => navigate(`/incidents/${inc.id}`)}>
                    <td>
                      <span className={`badge ${getRiskBadgeClass(inc.severity)}`}>
                        {inc.severity}
                      </span>
                    </td>
                    <td style={{ fontWeight: 500 }}>{inc.service_name}</td>
                    <td
                      style={{
                        color: 'var(--text-secondary)',
                        fontSize: 'var(--font-size-sm)',
                        maxWidth: '380px',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {inc.title}
                    </td>
                    <td>
                      <span className={`badge badge-${inc.status.toLowerCase()}`}>
                        <span className="dot" />
                        {inc.status}
                      </span>
                    </td>
                    <td style={{ fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
                      {inc.risk_score}
                    </td>
                    <td style={{ color: 'var(--text-secondary)', fontSize: 'var(--font-size-sm)' }}>
                      {formatDuration(inc.started_at, inc.resolved_at)}
                    </td>
                    <td style={{ color: 'var(--text-secondary)', fontSize: 'var(--font-size-sm)' }}>
                      {timeAgo(inc.started_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
