package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

type Cache struct {
	client *goredis.Client
}

func NewCache(addrOrURL string) (*Cache, error) {
	if addrOrURL == "" {
		return nil, fmt.Errorf("redis address or URL cannot be empty")
	}

	var opts *goredis.Options
	if strings.HasPrefix(addrOrURL, "redis://") || strings.HasPrefix(addrOrURL, "rediss://") {
		parsed, err := goredis.ParseURL(addrOrURL)
		if err != nil {
			return nil, fmt.Errorf("parsing redis URL: %w", err)
		}
		opts = parsed
	} else {
		opts = &goredis.Options{
			Addr: addrOrURL,
		}
	}

	client := goredis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return &Cache{client: client}, nil
}

func (c *Cache) Ping(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

// SetLatestMetric caches the most recent metric for a service (5 minute TTL).
func (c *Cache) SetLatestMetric(ctx context.Context, serviceID uuid.UUID, m *domain.Metric) error {
	if c.client == nil {
		return nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshalling metric: %w", err)
	}
	key := fmt.Sprintf("metric:latest:%s", serviceID)
	return c.client.Set(ctx, key, data, 5*time.Minute).Err()
}

// GetLatestMetric retrieves the cached latest metric for a service.
func (c *Cache) GetLatestMetric(ctx context.Context, serviceID uuid.UUID) (*domain.Metric, error) {
	if c.client == nil {
		return nil, nil
	}
	key := fmt.Sprintf("metric:latest:%s", serviceID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting cached metric: %w", err)
	}

	var m domain.Metric
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshalling metric: %w", err)
	}
	return &m, nil
}

// SetRiskScore caches the latest risk score for a service (2 minute TTL).
func (c *Cache) SetRiskScore(ctx context.Context, serviceID uuid.UUID, rs *domain.RiskScore) error {
	if c.client == nil {
		return nil
	}
	data, err := json.Marshal(rs)
	if err != nil {
		return fmt.Errorf("marshalling risk score: %w", err)
	}
	key := fmt.Sprintf("risk:latest:%s", serviceID)
	return c.client.Set(ctx, key, data, 2*time.Minute).Err()
}

// GetRiskScore retrieves the cached risk score for a service.
func (c *Cache) GetRiskScore(ctx context.Context, serviceID uuid.UUID) (*domain.RiskScore, error) {
	if c.client == nil {
		return nil, nil
	}
	key := fmt.Sprintf("risk:latest:%s", serviceID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting cached risk score: %w", err)
	}

	var rs domain.RiskScore
	if err := json.Unmarshal(data, &rs); err != nil {
		return nil, fmt.Errorf("unmarshalling risk score: %w", err)
	}
	return &rs, nil
}

// SetSimulationState stores active simulation modifiers in Redis with a TTL.
func (c *Cache) SetSimulationState(ctx context.Context, serviceName string, modifiers map[string]float64, ttl time.Duration) error {
	if c.client == nil {
		return nil
	}
	data, err := json.Marshal(modifiers)
	if err != nil {
		return fmt.Errorf("marshalling simulation state: %w", err)
	}
	key := fmt.Sprintf("sim:modifiers:%s", serviceName)
	return c.client.Set(ctx, key, data, ttl).Err()
}

// ClearSimulationState removes active simulation modifiers for a service immediately.
func (c *Cache) ClearSimulationState(ctx context.Context, serviceName string) error {
	if c.client == nil {
		return nil
	}
	key := fmt.Sprintf("sim:modifiers:%s", serviceName)
	return c.client.Del(ctx, key).Err()
}

// GetSimulationState retrieves active simulation modifiers for a service.
func (c *Cache) GetSimulationState(ctx context.Context, serviceName string) (map[string]float64, error) {
	if c.client == nil {
		return nil, nil
	}
	key := fmt.Sprintf("sim:modifiers:%s", serviceName)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting simulation state: %w", err)
	}

	var modifiers map[string]float64
	if err := json.Unmarshal(data, &modifiers); err != nil {
		return nil, fmt.Errorf("unmarshalling simulation state: %w", err)
	}
	return modifiers, nil
}

// SetServiceStatus caches service status.
func (c *Cache) SetServiceStatus(ctx context.Context, serviceName string, status string) error {
	if c.client == nil {
		return nil
	}
	key := fmt.Sprintf("service:status:%s", serviceName)
	return c.client.Set(ctx, key, status, 2*time.Minute).Err()
}

// GetServiceStatus retrieves cached service status.
func (c *Cache) GetServiceStatus(ctx context.Context, serviceName string) (string, error) {
	if c.client == nil {
		return "", nil
	}
	key := fmt.Sprintf("service:status:%s", serviceName)
	val, err := c.client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", nil
	}
	return val, err
}
