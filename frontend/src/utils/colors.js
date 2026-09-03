export function getRiskColor(level) {
  switch (level) {
    case 'LOW':
      return 'var(--color-risk-low)';
    case 'MODERATE':
      return 'var(--color-risk-moderate)';
    case 'HIGH':
      return 'var(--color-risk-high)';
    case 'CRITICAL':
      return 'var(--color-risk-critical)';
    default:
      return 'var(--text-tertiary)';
  }
}

export function getStatusColor(status) {
  switch (status) {
    case 'healthy':
      return 'var(--color-healthy)';
    case 'degraded':
      return 'var(--color-degraded)';
    case 'unhealthy':
      return 'var(--color-unhealthy)';
    case 'critical':
      return 'var(--color-critical)';
    default:
      return 'var(--text-tertiary)';
  }
}

export function getRiskBadgeClass(level) {
  switch (level) {
    case 'LOW':
      return 'badge-low';
    case 'MODERATE':
      return 'badge-moderate';
    case 'HIGH':
      return 'badge-high';
    case 'CRITICAL':
      return 'badge-critical';
    default:
      return '';
  }
}

export function getStatusBadgeClass(status) {
  switch (status) {
    case 'healthy':
      return 'badge-healthy';
    case 'degraded':
      return 'badge-degraded';
    case 'unhealthy':
      return 'badge-unhealthy';
    case 'critical':
      return 'badge-critical';
    default:
      return '';
  }
}

export function getScoreColor(score) {
  if (score >= 80) return 'var(--color-risk-critical)';
  if (score >= 60) return 'var(--color-risk-high)';
  if (score >= 30) return 'var(--color-risk-moderate)';
  return 'var(--color-risk-low)';
}
