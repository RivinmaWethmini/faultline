import React, { useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { api } from '../api/client';
import { usePolling } from '../hooks/usePolling';
import { getRiskBadgeClass, getScoreColor } from '../utils/colors';
import { formatDateTime, formatDuration } from '../utils/formatters';

export const IncidentDetail = () => {
  const { id } = useParams();

  const fetchIncident = useCallback(() => api.getIncident(id), [id]);
  const { data: incident, loading } = usePolling(fetchIncident, 5000);

  if (loading || !incident) {
    return <div className="loading-container"><div className="spinner" /></div>;
  }

  const anomalies = incident.anomalies
    ? [
        { label: 'Latency', value: incident.anomalies.latency_anomaly },
        { label: 'Error Rate', value: incident.anomalies.error_anomaly },
        { label: 'Timeout', value: incident.anomalies.timeout_anomaly },
        { label: 'Traffic', value: incident.anomalies.traffic_anomaly },
        { label: 'Dependency', value: incident.anomalies.dependency_anomaly },
      ]
    : [];

  return (
    <div>
      {/* Header */}
      <div style={{ marginBottom: 'var(--space-6)' }}>
        <div className="flex items-center gap-3" style={{ marginBottom: 'var(--space-2)' }}>
          <span className={`badge ${getRiskBadgeClass(incident.severity)}`}>{incident.severity}</span>
          <span className={`badge badge-${incident.status.toLowerCase()}`}><span className="dot" />{incident.status}</span>
        </div>
        <h1 style={{ fontSize: 'var(--font-size-xl)', fontWeight: 700, marginBottom: 'var(--space-2)' }}>
          {incident.title}
        </h1>
        <div className="flex items-center gap-4" style={{ fontSize: 'var(--font-size-sm)', color: 'var(--text-secondary)' }}>
          <span>Service: <strong style={{ color: 'var(--text-primary)' }}>{incident.service_name}</strong></span>
          <span>Started: {formatDateTime(incident.started_at)}</span>
          <span>Duration: {formatDuration(incident.started_at, incident.resolved_at)}</span>
        </div>
      </div>

      <div className="grid grid-2" style={{ marginBottom: 'var(--space-6)' }}>
        {/* Risk + Anomaly breakdown */}
        <div className="card">
          <div className="card-header"><h3>Risk Assessment</h3></div>
          <div className="risk-gauge" style={{ marginBottom: 'var(--space-5)' }}>
            <div className="score" style={{ color: getScoreColor(incident.risk_score) }}>
              {incident.risk_score}
            </div>
            <div className="risk-bar">
              <div className="risk-fill" style={{
                width: `${incident.risk_score}%`,
                background: getScoreColor(incident.risk_score),
              }} />
            </div>
          </div>
          <div className="anomaly-breakdown">
            {anomalies.map(a => (
              <div key={a.label} className="anomaly-row">
                <span className="anomaly-label">{a.label}</span>
                <div className="anomaly-bar">
                  <div className="anomaly-fill" style={{
                    width: `${a.value}%`,
                    background: getScoreColor(a.value),
                  }} />
                </div>
                <span className="anomaly-value" style={{ color: getScoreColor(a.value) }}>{a.value}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Root cause & propagation */}
        <div className="card">
          <div className="card-header"><h3>Analysis</h3></div>

          <div style={{ marginBottom: 'var(--space-5)' }}>
            <div style={{ fontSize: 'var(--font-size-xs)', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.04em', marginBottom: 'var(--space-2)' }}>
              Root Cause Analysis
            </div>
            <p style={{ fontSize: 'var(--font-size-sm)', color: 'var(--text-secondary)', lineHeight: 1.6 }}>
              {incident.root_cause}
            </p>
          </div>

          {incident.propagation_path && incident.propagation_path.length > 0 && (
            <div style={{ marginBottom: 'var(--space-5)' }}>
              <div style={{ fontSize: 'var(--font-size-xs)', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.04em', marginBottom: 'var(--space-2)' }}>
                Propagation Path
              </div>
              <div className="propagation-chain">
                {incident.propagation_path.map((node, i) => (
                  <React.Fragment key={i}>
                    {i > 0 && <span className="chain-arrow">→</span>}
                    <span className="chain-node">{node}</span>
                  </React.Fragment>
                ))}
              </div>
            </div>
          )}

          {incident.impacted_services && incident.impacted_services.length > 0 && (
            <div>
              <div style={{ fontSize: 'var(--font-size-xs)', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.04em', marginBottom: 'var(--space-2)' }}>
                Impacted Services
              </div>
              <div className="flex gap-2" style={{ flexWrap: 'wrap' }}>
                {incident.impacted_services.map((svc, i) => (
                  <span key={i} className="badge badge-degraded">{svc}</span>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Timeline */}
      <div className="card">
        <div className="card-header"><h3>Timeline</h3></div>
        {incident.events && incident.events.length > 0 ? (
          <div className="timeline">
            {incident.events.map(evt => (
              <div key={evt.id} className="timeline-item">
                <div className={`timeline-dot ${evt.event_type}`} />
                <div className="timeline-content">
                  <div className="timeline-time">{formatDateTime(evt.created_at)}</div>
                  <div className="timeline-message">{evt.message}</div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <p>No timeline events.</p>
          </div>
        )}
      </div>
    </div>
  );
};
