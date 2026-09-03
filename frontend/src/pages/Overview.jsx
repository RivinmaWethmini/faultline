import React, { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { usePolling } from '../hooks/usePolling';
import { getRiskBadgeClass, getStatusBadgeClass, getScoreColor } from '../utils/colors';
import { timeAgo } from '../utils/formatters';

export const Overview = () => {
  const navigate = useNavigate();

  const fetchServices = useCallback(() => api.getServices(), []);
  const fetchIncidents = useCallback(() => api.getIncidents(), []);

  const { data: services, loading: loadingServices } = usePolling(fetchServices, 5000);
  const { data: incidents, loading: loadingIncidents } = usePolling(fetchIncidents, 5000);

  if (loadingServices || loadingIncidents) {
    return <div className="loading-container"><div className="spinner" /></div>;
  }

  const activeIncidents = incidents?.filter(i => i.status !== 'RESOLVED' && i.status !== 'resolved') || [];
  const recentIncidents = incidents?.slice(0, 5) || [];
  const highestRisk = services?.reduce((max, s) =>
    (s.risk_score?.overall_score ?? 0) > (max?.risk_score?.overall_score ?? 0) ? s : max
  , services[0]);

  const healthyCount = services?.filter(s => s.status === 'healthy').length ?? 0;
  const totalCount = services?.length ?? 0;
  const systemStatus = activeIncidents.length > 0
    ? (activeIncidents.some(i => i.severity === 'CRITICAL') ? 'critical' : 'degraded')
    : 'healthy';

  return (
    <div>
      {/* System health summary */}
      <div className="grid grid-4" style={{ marginBottom: 'var(--space-6)' }}>
        <div className="metric-card">
          <div className="metric-label">System Status</div>
          <div className="flex items-center gap-2">
            <span className={`badge ${systemStatus === 'healthy' ? 'badge-healthy' : systemStatus === 'critical' ? 'badge-critical' : 'badge-degraded'}`}>
              <span className="dot" />
              {systemStatus.toUpperCase()}
            </span>
          </div>
        </div>

        <div className="metric-card">
          <div className="metric-label">Active Incidents</div>
          <div className="metric-value" style={{ color: activeIncidents.length > 0 ? 'var(--color-risk-critical)' : 'var(--color-risk-low)' }}>
            {activeIncidents.length}
          </div>
        </div>

        <div className="metric-card">
          <div className="metric-label">Services Healthy</div>
          <div className="metric-value">
            {healthyCount}<span className="metric-unit">/ {totalCount}</span>
          </div>
        </div>

        <div className="metric-card">
          <div className="metric-label">Highest Risk</div>
          {highestRisk?.risk_score ? (
            <div>
              <div className="metric-value" style={{ color: getScoreColor(highestRisk.risk_score.overall_score) }}>
                {highestRisk.risk_score.overall_score}
              </div>
              <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--text-secondary)', marginTop: '2px' }}>
                {highestRisk.display_name}
              </div>
            </div>
          ) : (
            <div className="metric-value" style={{ color: 'var(--color-risk-low)' }}>—</div>
          )}
        </div>
      </div>

      {/* Services table */}
      <div className="section">
        <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-4)' }}>
          <h3 className="section-title" style={{ marginBottom: 0 }}>Service Health</h3>
        </div>
        <div className="card" style={{ padding: 0 }}>
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Service</th>
                  <th>Status</th>
                  <th>Risk Score</th>
                  <th>Risk Level</th>
                  <th>Last Updated</th>
                </tr>
              </thead>
              <tbody>
                {services?.map(svc => (
                  <tr key={svc.id} className="clickable-row" onClick={() => navigate(`/services/${svc.id}`)}>
                    <td>
                      <div style={{ fontWeight: 500 }}>{svc.display_name}</div>
                      <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--text-tertiary)' }}>{svc.name}</div>
                    </td>
                    <td>
                      <span className={`badge ${getStatusBadgeClass(svc.status)}`}>
                        <span className="dot" />
                        {svc.status}
                      </span>
                    </td>
                    <td>
                      <span style={{ fontWeight: 600, fontVariantNumeric: 'tabular-nums', color: getScoreColor(svc.risk_score?.overall_score ?? 0) }}>
                        {svc.risk_score?.overall_score ?? '—'}
                      </span>
                    </td>
                    <td>
                      {svc.risk_score && (
                        <span className={`badge ${getRiskBadgeClass(svc.risk_score.risk_level)}`}>
                          {svc.risk_score.risk_level}
                        </span>
                      )}
                    </td>
                    <td style={{ color: 'var(--text-secondary)', fontSize: 'var(--font-size-sm)' }}>
                      {svc.risk_score ? timeAgo(svc.risk_score.timestamp) : '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Recent incidents */}
      <div className="section">
        <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-4)' }}>
          <h3 className="section-title" style={{ marginBottom: 0 }}>Recent Incidents</h3>
          <button className="btn btn-secondary btn-sm" onClick={() => navigate('/incidents')}>View all</button>
        </div>
        <div className="card" style={{ padding: 0 }}>
          {recentIncidents.length === 0 ? (
            <div className="empty-state">
              <h3>No incidents</h3>
              <p>All services are operating normally.</p>
            </div>
          ) : (
            <div className="table-container">
              <table>
                <thead>
                  <tr>
                    <th>Severity</th>
                    <th>Service</th>
                    <th>Title</th>
                    <th>Status</th>
                    <th>Started</th>
                  </tr>
                </thead>
                <tbody>
                  {recentIncidents.map(inc => (
                    <tr key={inc.id} className="clickable-row" onClick={() => navigate(`/incidents/${inc.id}`)}>
                      <td><span className={`badge ${getRiskBadgeClass(inc.severity)}`}>{inc.severity}</span></td>
                      <td style={{ fontWeight: 500 }}>{inc.service_name}</td>
                      <td style={{ color: 'var(--text-secondary)', fontSize: 'var(--font-size-sm)', maxWidth: '400px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{inc.title}</td>
                      <td><span className={`badge badge-${inc.status.toLowerCase()}`}><span className="dot" />{inc.status}</span></td>
                      <td style={{ color: 'var(--text-secondary)', fontSize: 'var(--font-size-sm)' }}>{timeAgo(inc.started_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
