function resolveApiBase() {
  const envUrl = import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE;
  if (envUrl) {
    const clean = envUrl.replace(/\/+$/, '');
    return clean.endsWith('/api') ? clean : `${clean}/api`;
  }
  if (typeof window !== 'undefined' && window.location.port === '3000' && !import.meta.env.DEV) {
    return 'http://localhost:8080/api';
  }
  return '/api';
}

const API_BASE = resolveApiBase();

async function fetchJSON(url, options) {
  const res = await fetch(`${API_BASE}${url}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }

  const json = await res.json();
  if (!json.success) {
    throw new Error(json.error || 'Unknown error');
  }
  return json.data;
}

export const api = {
  // Services
  getServices: () => fetchJSON('/services'),
  getService: (id) => fetchJSON(`/services/${id}`),
  getServiceMetrics: (id, duration = 30) =>
    fetchJSON(`/services/${id}/metrics?duration=${duration}`),
  getServiceRisk: (id) => fetchJSON(`/services/${id}/risk`),
  getDependencyImpact: (id) => fetchJSON(`/services/${id}/dependency-impact`),

  // Dependencies
  getDependencies: () => fetchJSON('/dependencies'),

  // Incidents
  getIncidents: (status) =>
    fetchJSON(`/incidents${status ? `?status=${status}` : ''}`),
  getIncident: (id) => fetchJSON(`/incidents/${id}`),

  // Simulations
  getSimulations: () => fetchJSON('/simulations'),
  getScenarios: () => fetchJSON('/simulations/scenarios'),
  createSimulation: (scenario) =>
    fetchJSON('/simulations', {
      method: 'POST',
      body: JSON.stringify({ scenario }),
    }),
  stopSimulation: (id) =>
    fetchJSON(`/simulations/${id}/stop`, {
      method: 'POST',
    }),

  // System
  getHealth: () => fetchJSON('/system/health'),
};
