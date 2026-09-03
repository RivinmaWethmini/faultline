import React, { useCallback, useState } from 'react';
import { api } from '../api/client';
import { usePolling } from '../hooks/usePolling';
import { timeAgo } from '../utils/formatters';

export const Simulator = () => {
  const [selectedScenario, setSelectedScenario] = useState('');
  const [triggering, setTriggering] = useState(false);
  const [stoppingId, setStoppingId] = useState(null);

  const fetchScenarios = useCallback(() => api.getScenarios(), []);
  const fetchSimulations = useCallback(() => api.getSimulations(), []);

  const { data: scenarios } = usePolling(fetchScenarios, 30000);
  const { data: simulations, refresh: refreshSims } = usePolling(fetchSimulations, 4000);

  const handleTrigger = async () => {
    if (!selectedScenario) return;
    setTriggering(true);
    try {
      await api.createSimulation(selectedScenario);
      refreshSims();
    } catch (err) {
      console.error('Failed to trigger simulation:', err);
    } finally {
      setTriggering(false);
    }
  };

  const handleStop = async (id) => {
    setStoppingId(id);
    try {
      await api.stopSimulation(id);
      refreshSims();
    } catch (err) {
      console.error('Failed to stop simulation:', err);
    } finally {
      setStoppingId(null);
    }
  };

  const runningSimulations = simulations?.filter(s => s.status === 'running') || [];

  return (
    <div>
      {/* Running simulations banner */}
      {runningSimulations.length > 0 && (
        <div
          style={{
            background: 'rgba(248, 81, 73, 0.1)',
            border: '1px solid rgba(248, 81, 73, 0.3)',
            borderRadius: 'var(--radius-lg)',
            padding: 'var(--space-4) var(--space-5)',
            marginBottom: 'var(--space-6)',
          }}
        >
          <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-3)' }}>
            <div className="flex items-center gap-2">
              <span className="badge badge-active pulse"><span className="dot" />ACTIVE</span>
              <span style={{ fontSize: 'var(--font-size-sm)', fontWeight: 600 }}>
                {runningSimulations.length} Failure Simulation(s) Active
              </span>
            </div>
            <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--text-tertiary)' }}>
              Degradation is actively propagating through the dependency graph
            </span>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
            {runningSimulations.map(sim => (
              <div
                key={sim.id}
                className="flex items-center justify-between"
                style={{
                  padding: 'var(--space-2) var(--space-3)',
                  background: 'var(--bg-secondary)',
                  borderRadius: 'var(--radius-md)',
                  border: '1px solid var(--border-primary)',
                }}
              >
                <div>
                  <strong style={{ color: 'var(--text-primary)' }}>{sim.scenario.replace(/_/g, ' ').toUpperCase()}</strong>
                  <span style={{ color: 'var(--text-secondary)', marginLeft: '8px', fontSize: 'var(--font-size-xs)' }}>
                    Target: {sim.target_service} (started {sim.started_at ? timeAgo(sim.started_at) : 'just now'})
                  </span>
                </div>
                <button
                  className="btn btn-secondary btn-sm"
                  onClick={() => handleStop(sim.id)}
                  disabled={stoppingId === sim.id}
                  style={{ color: 'var(--color-risk-critical)' }}
                >
                  {stoppingId === sim.id ? 'Reverting...' : '⏹ Stop & Revert'}
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Scenarios */}
      <div className="section">
        <h3 className="section-title">Predefined Failure Scenarios</h3>
        <p style={{ fontSize: 'var(--font-size-sm)', color: 'var(--text-secondary)', marginBottom: 'var(--space-4)' }}>
          Inject realistic failure patterns into simulated microservices to demonstrate real-time risk score spikes and multi-hop propagation.
        </p>

        <div className="grid grid-2" style={{ marginBottom: 'var(--space-5)' }}>
          {scenarios?.map(scenario => (
            <div
              key={scenario.name}
              className={`scenario-card ${selectedScenario === scenario.name ? 'active' : ''}`}
              onClick={() => setSelectedScenario(scenario.name)}
            >
              <h4>{scenario.display_name}</h4>
              <p>{scenario.description}</p>
              <div className="scenario-meta">
                <span>Target: <strong>{scenario.target_service}</strong></span>
                <span>Duration: <strong>{scenario.duration_seconds}s</strong></span>
              </div>
            </div>
          ))}
        </div>

        <button
          className="btn btn-danger"
          onClick={handleTrigger}
          disabled={!selectedScenario || triggering}
          style={{ minWidth: '180px' }}
        >
          {triggering ? 'Injecting Failure...' : '⚡ Trigger Scenario'}
        </button>
      </div>

      {/* Simulation history */}
      <div className="section">
        <h3 className="section-title">Simulation History</h3>
        <div className="card" style={{ padding: 0 }}>
          {!simulations || simulations.length === 0 ? (
            <div className="empty-state">
              <h3>No simulations yet</h3>
              <p>Run a scenario above to test fault injection and failure propagation.</p>
            </div>
          ) : (
            <div className="table-container">
              <table>
                <thead>
                  <tr>
                    <th>Scenario</th>
                    <th>Target Service</th>
                    <th>Status</th>
                    <th>Started</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {simulations.map(sim => (
                    <tr key={sim.id}>
                      <td style={{ fontWeight: 500 }}>{sim.scenario.replace(/_/g, ' ')}</td>
                      <td style={{ color: 'var(--text-secondary)' }}>{sim.target_service}</td>
                      <td>
                        <span className={`badge ${sim.status === 'running' ? 'badge-active' : 'badge-resolved'}`}>
                          <span className="dot" />
                          {sim.status.toUpperCase()}
                        </span>
                      </td>
                      <td style={{ color: 'var(--text-secondary)', fontSize: 'var(--font-size-sm)' }}>
                        {sim.started_at ? timeAgo(sim.started_at) : '—'}
                      </td>
                      <td>
                        {sim.status === 'running' ? (
                          <button
                            className="btn btn-secondary btn-sm"
                            onClick={() => handleStop(sim.id)}
                            disabled={stoppingId === sim.id}
                            style={{ padding: '2px 8px', fontSize: 'var(--font-size-xs)' }}
                          >
                            Stop
                          </button>
                        ) : (
                          <span style={{ color: 'var(--text-tertiary)', fontSize: 'var(--font-size-xs)' }}>Completed</span>
                        )}
                      </td>
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
