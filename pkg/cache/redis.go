package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	Address  string
	Username string
	Password string
	Database int
}

type CacheClient interface {
	Client() *redis.Client
	Close() error
	AttemptLeaderElection(ctx context.Context, instanceUuid string, group string) (bool, error)
	SetLeader(leader bool)
	IsLeader() bool
}

type cacheClient struct {
	client *redis.Client
	leader bool
}

// NewCachingClient creates a new caching client.
func NewCacheClient(instanceUuid string, opts Options) CacheClient {
	client := redis.NewClient(&redis.Options{
		Addr:     opts.Address,
		Username: opts.Username,
		Password: opts.Password,
		DB:       opts.Database,
	})

	return &cacheClient{
		client: client,
	}
}

// Client returns the underlying redis client.
func (c *cacheClient) Client() *redis.Client {
	return c.client
}

// Close closes the caching client.
func (c *cacheClient) Close() error {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return fmt.Errorf("failed to close caching client: %w", err)
		}
	}
	return nil
}

// AttemptLeaderElection attempts to elect the current instance as the leader for the given group.
func (c *cacheClient) AttemptLeaderElection(ctx context.Context, instanceUuid string, group string) (bool, error) {
	return c.client.SetNX(ctx, group, instanceUuid, 5*time.Second).Result()
}

// SetLeader sets the leader status of the current instance.
func (c *cacheClient) SetLeader(leader bool) {
	c.leader = leader
}

// IsLeader returns whether the current instance is the leader.
func (c *cacheClient) IsLeader() bool {
	return c.leader
}
