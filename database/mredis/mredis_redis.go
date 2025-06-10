// Copyright Maltose Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mredis

import (
	"context"

	"github.com/graingo/maltose/database/mredis/config"
	"github.com/redis/go-redis/v9"
)

// Redis is the main struct for redis operations.
type Redis struct {
	client redis.UniversalClient
	config *config.Config
}

// NewWithConfig creates and returns a new Redis client.
func NewWithConfig(config *config.Config) (*Redis, error) {
	client := config.NewClient()
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &Redis{
		client: client,
		config: config,
	}, nil
}

// Client returns the underlying universal client.
func (r *Redis) Client() redis.UniversalClient {
	return r.client
}

// Ping checks the connection to the server.
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the client, releasing any open resources.
func (r *Redis) Close() error {
	return r.client.Close()
}
