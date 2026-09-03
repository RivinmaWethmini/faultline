import React, { useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { api } from '../api/client';
import { usePolling } from '../hooks/usePolling';
import { getScoreColor, getStatusBadgeClass, getRiskBadgeClass } from '../utils/colors';
import { formatTimestamp } from '../utils/formatters';

export const ServiceDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const fetchService = useCallback(() => api.getService(id), [id]);
  const fetchMetrics = useCallback(() => api.getServiceMetrics(id, 30), [id]);
  const fetchRisk = useCallback(() => api.getServiceRisk(id), [id]);
  const fetchImpact = useCallback(() => api.getDependencyImpact(id), [id]);

  const { data: service, loading: loadingSvc } = usePolling(fetchService, 5000);
  const { data: metrics } = usePolling(fetchMetrics, 5000);
  const { data: riskAssessment } = usePolling(fetchRisk, 5000);
  const { data: impact } = usePolling(fetchImpact, 10000);

  if (loadingSvc || !service) {
    return <div className="loading-container"><div className="spinner" /></div>;
  }

  const overallScore = riskAssessment?.overallRisk ?? service.risk_score?.overall_score ?? 0;
  const riskLevel = riskAssessment?.level ?? service.risk_score?.risk_level ?? 'LOW';

  const chartMetrics = [...(metrics || [])].reverse().map(m => ({
    time: formatTimestamp(m.timestamp),
    latency: m.response_latency_ms,
    errorRate: m.error_rate,
    requestRate: m.request_rate,
    cpu: m.cpu_usage,
    memory: m.memory_usage,
    depLatency: m.dep_latency_ms,
  }));

  const riskHistory = riskAssessment?.history || [];
  const riskChartData = [...riskHistory].reverse().map(r => ({
    time: formatTimestamp(r.timestamp),
    score: r.overall_score,
  }));

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-6)' }}>
        <div>
          <div className="flex items-center gap-3">
            <h1 style={{ fontSize: 'var(--font-size-xl)', fontWeight: 700 }}>{service.display_name}</h1>
            <span className={`badge ${getStatusBadgeClass(service.status)}`}>
              <span className="dot" />
              {service.status}
            </span>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: 'var(--font-size-sm)', marginTop: '4px' }}>
            {service.description}
          </p>
        </div>
        <button className="btn btn-secondary btn-sm" onClick={() => navigate('/services')}>
          ← Back to Services
        </button>
      </div>

      {/* Risk Gauge & Anomaly Factors */}
      <div className="grid grid-3" style={{ marginBottom: 'var(--space-6)' }}>
        <div className="card">
          <div className="card-header">
            <h3>Current Risk Score</h3>
          </div>
          <div className="risk-gauge">
            <div className="score" style={{ color: getScoreColor(overallScore) }}>
              {overallScore}
            </div>
            <div className="score-label" style={{ color: getScoreColor(overallScore) }}>
              <span className={`badge ${getRiskBadgeClass(riskLevel)}`}>{riskLevel}</span>
            </div>
            <div className="risk-bar" style={{ marginTop: '12px' }}>
              <div
                className="risk-fill"
                style={{
                  width: `${overallScore}%`,
                  background: getScoreColor(overallScore),
                }}
              />
            </div>
          </div>
        </div>

        <div className="card" style={{ gridColumn: 'span 2' }}>
          <div className="card-header">
            <h3>Explainable Anomaly Factors</h3>
            <span className="card-subtitle">Deterministic statistical evaluation</span>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            {riskAssessment?.factors && riskAssessment.factors.length > 0 ? (
              riskAssessment.factors.map((factor, idx) => (
                <div
                  key={idx}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    padding: 'var(--space-2) var(--space-3)',
                    background: 'var(--bg-tertiary)',
                    borderRadius: 'var(--radius-md)',
                    borderLeft: `3px solid ${getScoreColor(factor.score)}`,
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div style={{ fontWeight: 600, fontSize: 'var(--font-size-sm)' }}>
                      {factor.name.replace(/_/g, ' ').toUpperCase()}
                    </div>
                    <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--text-secondary)', marginTop: '2px' }}>
                      {factor.reason}
                    </div>
                  </div>
                  <div style={{ textAlign: 'right', marginLeft: 'var(--space-4)' }}>
                    <span
                      style={{
                        fontWeight: 700,
                        fontSize: 'var(--font-size-md)',
                        fontVariantNumeric: 'tabular-nums',
                        color: getScoreColor(factor.score),
                      }}
                    >
                      {factor.score}
                    </span>
                  </div>
                </div>
              ))
            ) : (
              <div style={{ color: 'var(--text-secondary)', fontSize: 'var(--font-size-sm)' }}>
                All operational metrics are within normal baseline tolerances.
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Dependency Impact & Root Cause Analysis */}
      {impact && (
        <div className="card" style={{ marginBottom: 'var(--space-6)' }}>
          <div className="card-header">
            <h3>Dependency & Propagation Impact</h3>
            <span className="card-subtitle">Topology-aware failure analysis</span>
          </div>
          <div className="grid grid-3">
            <div>
              <div style={{ fontSize: 'var(--font-size-xs)', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '8px' }}>
                Dependencies (Calls)
              </div>
              {impact.downstream_services && impact.downstream_services.length > 0 ? (
                <div className="flex gap-2" style={{ flexWrap: 'wrap' }}>
                  {impact.downstream_services.map((name, i) => (
                    <span key={i} className="badge badge-healthy">{name}</span>
                  ))}
                </div>
              ) : (
                <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--text-tertiary)' }}>No downstream dependencies</span>
              )}
            </div>

            <div>
              <div style={{ fontSize: 'var(--font-size-xs)', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '8px' }}>
                Upstream Callers
              </div>
              {impact.upstream_services && impact.upstream_services.length > 0 ? (
                <div className="flex gap-2" style={{ flexWrap: 'wrap' }}>
                  {impact.upstream_services.map((name, i) => (
                    <span key={i} className="badge badge-degraded">{name}</span>
                  ))}
                </div>
              ) : (
                <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--text-tertiary)' }}>Edge service (no callers)</span>
              )}
            </div>

            <div>
              <div style={{ fontSize: 'var(--font-size-xs)', fontWeight: 600, color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '8px' }}>
                Root Cause Ranking
              </div>
              {impact.possible_root_causes && impact.possible_root_causes.length > 0 ? (
                <div style={{ fontSize: 'var(--font-size-xs)' }}>
                  <strong>{impact.possible_root_causes[0].service_name}</strong>
                  <div style={{ color: 'var(--text-secondary)' }}>{impact.possible_root_causes[0].reason}</div>
                </div>
              ) : (
                <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--text-tertiary)' }}>Normal topology</span>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Real-time Metric Charts */}
      <div className="grid grid-2" style={{ marginBottom: 'var(--space-6)' }}>
        <div className="card">
          <div className="card-header">
            <h3>Response Latency</h3>
            <span className="card-subtitle">ms</span>
          </div>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={chartMetrics}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border-secondary)" />
              <XAxis dataKey="time" tick={{ fontSize: 10, fill: 'var(--text-tertiary)' }} />
              <YAxis tick={{ fontSize: 10, fill: 'var(--text-tertiary)' }} />
              <Tooltip contentStyle={{ background: 'var(--bg-tertiary)', border: '1px solid var(--border-primary)', borderRadius: 6, fontSize: 12 }} />
              <Line type="monotone" dataKey="latency" stroke="var(--color-accent)" strokeWidth={1.5} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        <div className="card">
          <div className="card-header">
            <h3>Error Rate</h3>
            <span className="card-subtitle">%</span>
          </div>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={chartMetrics}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border-secondary)" />
              <XAxis dataKey="time" tick={{ fontSize: 10, fill: 'var(--text-tertiary)' }} />
              <YAxis tick={{ fontSize: 10, fill: 'var(--text-tertiary)' }} />
              <Tooltip contentStyle={{ background: 'var(--bg-tertiary)', border: '1px solid var(--border-primary)', borderRadius: 6, fontSize: 12 }} />
              <Line type="monotone" dataKey="errorRate" stroke="var(--color-risk-critical)" strokeWidth={1.5} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        <div className="card">
          <div className="card-header">
            <h3>Request Rate</h3>
            <span className="card-subtitle">req/s</span>
          </div>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={chartMetrics}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border-secondary)" />
              <XAxis dataKey="time" tick={{ fontSize: 10, fill: 'var(--text-tertiary)' }} />
              <YAxis tick={{ fontSize: 10, fill: 'var(--text-tertiary)' }} />
              <Tooltip contentStyle={{ background: 'var(--bg-tertiary)', border: '1px solid var(--border-primary)', borderRadius: 6, fontSize: 12 }} />
              <Line type="monotone" dataKey="requestRate" stroke="var(--color-risk-moderate)" strokeWidth={1.5} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        <div className="card">
          <div className="card-header">
            <h3>Risk Score History</h3>
            <span className="card-subtitle">0-100 score</span>
          </div>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={riskChartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border-secondary)" />
              <XAxis dataKey="time" tick={{ fontSize: 10, fill: 'var(--text-tertiary)' }} />
              <YAxis domain={[0, 100]} tick={{ fontSize: 10, fill: 'var(--text-tertiary)' }} />
              <Tooltip contentStyle={{ background: 'var(--bg-tertiary)', border: '1px solid var(--border-primary)', borderRadius: 6, fontSize: 12 }} />
              <Line type="monotone" dataKey="score" stroke="var(--color-risk-high)" strokeWidth={1.5} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
};
