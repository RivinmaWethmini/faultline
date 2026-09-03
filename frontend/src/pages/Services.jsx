import React, { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { usePolling } from '../hooks/usePolling';
import { getRiskBadgeClass, getStatusBadgeClass, getScoreColor } from '../utils/colors';

export const Services = () => {
  const navigate = useNavigate();
  const fetchServices = useCallback(() => api.getServices(), []);
  const { data: services, loading } = usePolling(fetchServices, 5000);

  if (loading) {
    return <div className="loading-container"><div className="spinner" /></div>;
  }

  return (
    <div>
      <div className="card" style={{ padding: 0 }}>
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>Service</th>
                <th>Status</th>
                <th>Risk Score</th>
                <th>Risk Level</th>
                <th>Latency Anom.</th>
                <th>Error Anom.</th>
                <th>Dep. Anom.</th>
              </tr>
            </thead>
            <tbody>
              {services?.map(svc => (
                <tr key={svc.id} className="clickable-row" onClick={() => navigate(`/services/${svc.id}`)}>
                  <td>
                    <div style={{ fontWeight: 500 }}>{svc.display_name}</div>
                    <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--text-tertiary)' }}>{svc.description}</div>
                  </td>
                  <td>
                    <span className={`badge ${getStatusBadgeClass(svc.status)}`}>
                      <span className="dot" />
                      {svc.status}
                    </span>
                  </td>
                  <td>
                    <span style={{ fontWeight: 700, fontSize: 'var(--font-size-md)', fontVariantNumeric: 'tabular-nums', color: getScoreColor(svc.risk_score?.overall_score ?? 0) }}>
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
                  <td style={{ fontVariantNumeric: 'tabular-nums', color: getScoreColor(svc.risk_score?.latency_anomaly ?? 0) }}>
                    {svc.risk_score?.latency_anomaly ?? '—'}
                  </td>
                  <td style={{ fontVariantNumeric: 'tabular-nums', color: getScoreColor(svc.risk_score?.error_anomaly ?? 0) }}>
                    {svc.risk_score?.error_anomaly ?? '—'}
                  </td>
                  <td style={{ fontVariantNumeric: 'tabular-nums', color: getScoreColor(svc.risk_score?.dependency_anomaly ?? 0) }}>
                    {svc.risk_score?.dependency_anomaly ?? '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
