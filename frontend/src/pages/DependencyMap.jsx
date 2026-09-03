import React, { useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { usePolling } from '../hooks/usePolling';
import { getScoreColor } from '../utils/colors';

export const DependencyMap = () => {
  const navigate = useNavigate();

  const fetchServices = useCallback(() => api.getServices(), []);
  const fetchDeps = useCallback(() => api.getDependencies(), []);

  const { data: services, loading: loadingSvc } = usePolling(fetchServices, 5000);
  const { data: deps, loading: loadingDeps } = usePolling(fetchDeps, 10000);

  // Layout nodes in a layered graph
  const layout = useMemo(() => {
    if (!services || !deps) return { nodes: [], edges: [] };

    const layers = {
      'api-gateway': 0,
      'auth-service': 1,
      'order-service': 1,
      'inventory-service': 2,
      'payment-service': 2,
    };

    // Group services by layer
    const layerGroups = [[], [], []];
    services.forEach(s => {
      const layer = layers[s.name] ?? 0;
      layerGroups[layer]?.push(s);
    });

    const width = 800;
    const nodeWidth = 140;
    const nodeHeight = 56;

    const nodes = [];
    layerGroups.forEach((group, layerIdx) => {
      const y = 40 + layerIdx * 140;
      const totalWidth = group.length * nodeWidth + (group.length - 1) * 60;
      const startX = (width - totalWidth) / 2;

      group.forEach((svc, i) => {
        nodes.push({
          x: startX + i * (nodeWidth + 60),
          y,
          service: svc,
        });
      });
    });

    // Build edges
    const nodeMap = new Map(nodes.map(n => [n.service.id, n]));
    const edges = deps
      .map(d => ({
        from: nodeMap.get(d.source_id),
        to: nodeMap.get(d.target_id),
      }))
      .filter(e => e.from && e.to);

    return { nodes, edges };
  }, [services, deps]);

  if (loadingSvc || loadingDeps) {
    return <div className="loading-container"><div className="spinner" /></div>;
  }

  const nodeWidth = 140;
  const nodeHeight = 56;

  return (
    <div>
      <div className="dep-graph">
        <svg width="100%" viewBox="0 0 800 440" style={{ overflow: 'visible' }}>
          <defs>
            <marker id="arrowhead" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto">
              <polygon points="0 0, 8 3, 0 6" fill="var(--text-tertiary)" />
            </marker>
          </defs>

          {/* Edges */}
          {layout.edges.map((edge, i) => {
            const fromX = edge.from.x + nodeWidth / 2;
            const fromY = edge.from.y + nodeHeight;
            const toX = edge.to.x + nodeWidth / 2;
            const toY = edge.to.y;

            return (
              <line
                key={i}
                x1={fromX}
                y1={fromY}
                x2={toX}
                y2={toY - 4}
                stroke="var(--border-primary)"
                strokeWidth="1.5"
                markerEnd="url(#arrowhead)"
                className="dep-edge"
              />
            );
          })}

          {/* Nodes */}
          {layout.nodes.map(node => {
            const score = node.service.risk_score?.overall_score ?? 0;
            const borderColor = score >= 30 ? getScoreColor(score) : 'var(--border-primary)';

            return (
              <g
                key={node.service.id}
                className="dep-node"
                onClick={() => navigate(`/services/${node.service.id}`)}
              >
                <rect
                  x={node.x}
                  y={node.y}
                  width={nodeWidth}
                  height={nodeHeight}
                  rx="8"
                  ry="8"
                  fill="var(--bg-card)"
                  stroke={borderColor}
                  strokeWidth="1.5"
                />
                {/* Status dot */}
                <circle
                  cx={node.x + 14}
                  cy={node.y + nodeHeight / 2}
                  r="4"
                  fill={getScoreColor(score)}
                />
                <text
                  x={node.x + 24}
                  y={node.y + 22}
                  fill="var(--text-primary)"
                  fontSize="11"
                  fontWeight="600"
                  fontFamily="var(--font-family)"
                >
                  {node.service.display_name}
                </text>
                <text
                  x={node.x + 24}
                  y={node.y + 38}
                  fill="var(--text-secondary)"
                  fontSize="10"
                  fontFamily="var(--font-family)"
                >
                  Risk: {score}
                </text>
                {/* Risk score indicator bar */}
                <rect
                  x={node.x}
                  y={node.y + nodeHeight - 3}
                  width={nodeWidth * (score / 100)}
                  height="3"
                  rx="0"
                  ry="0"
                  fill={getScoreColor(score)}
                  opacity="0.8"
                />
              </g>
            );
          })}
        </svg>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-4" style={{ marginTop: 'var(--space-4)', justifyContent: 'center' }}>
        {[
          { label: 'Low (0-29)', color: 'var(--color-risk-low)' },
          { label: 'Moderate (30-59)', color: 'var(--color-risk-moderate)' },
          { label: 'High (60-79)', color: 'var(--color-risk-high)' },
          { label: 'Critical (80-100)', color: 'var(--color-risk-critical)' },
        ].map(item => (
          <div key={item.label} className="flex items-center gap-2" style={{ fontSize: 'var(--font-size-xs)', color: 'var(--text-secondary)' }}>
            <div style={{ width: 10, height: 10, borderRadius: '50%', background: item.color }} />
            {item.label}
          </div>
        ))}
      </div>
    </div>
  );
};
