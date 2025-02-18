//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package state

import (
	"context"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/liutizhong/weaviate/adapters/handlers/graphql"
	"github.com/liutizhong/weaviate/adapters/handlers/rest/tenantactivity"
	"github.com/liutizhong/weaviate/adapters/repos/classifications"
	"github.com/liutizhong/weaviate/adapters/repos/db"
	rCluster "github.com/liutizhong/weaviate/cluster"
	"github.com/liutizhong/weaviate/usecases/auth/authentication/anonymous"
	"github.com/liutizhong/weaviate/usecases/auth/authentication/apikey"
	"github.com/liutizhong/weaviate/usecases/auth/authentication/oidc"
	"github.com/liutizhong/weaviate/usecases/auth/authorization"
	"github.com/liutizhong/weaviate/usecases/backup"
	"github.com/liutizhong/weaviate/usecases/cluster"
	"github.com/liutizhong/weaviate/usecases/config"
	"github.com/liutizhong/weaviate/usecases/locks"
	"github.com/liutizhong/weaviate/usecases/memwatch"
	"github.com/liutizhong/weaviate/usecases/modules"
	"github.com/liutizhong/weaviate/usecases/monitoring"
	"github.com/liutizhong/weaviate/usecases/objects"
	"github.com/liutizhong/weaviate/usecases/replica"
	"github.com/liutizhong/weaviate/usecases/scaler"
	"github.com/liutizhong/weaviate/usecases/schema"
	"github.com/liutizhong/weaviate/usecases/sharding"
	"github.com/liutizhong/weaviate/usecases/traverser"
)

// State is the only source of application-wide state
// NOTE: This is not true yet, see gh-723
// TODO: remove dependencies to anything that's not an ent or uc
type State struct {
	OIDC            *oidc.Client
	AnonymousAccess *anonymous.Client
	APIKey          *apikey.Client
	Authorizer      authorization.Authorizer
	AuthzController authorization.Controller

	ServerConfig          *config.WeaviateConfig
	Locks                 locks.ConnectorSchemaLock
	Logger                *logrus.Logger
	gqlMutex              sync.Mutex
	GraphQL               graphql.GraphQL
	Modules               *modules.Provider
	SchemaManager         *schema.Manager
	Scaler                *scaler.Scaler
	Cluster               *cluster.State
	RemoteIndexIncoming   *sharding.RemoteIndexIncoming
	RemoteNodeIncoming    *sharding.RemoteNodeIncoming
	RemoteReplicaIncoming *replica.RemoteReplicaIncoming
	Traverser             *traverser.Traverser

	ClassificationRepo *classifications.DistributedRepo
	Metrics            *monitoring.PrometheusMetrics
	ServerMetrics      *monitoring.ServerMetrics
	BackupManager      *backup.Handler
	DB                 *db.DB
	BatchManager       *objects.BatchManager
	ClusterHttpClient  *http.Client
	ReindexCtxCancel   context.CancelFunc
	MemWatch           *memwatch.Monitor

	ClusterService *rCluster.Service
	TenantActivity *tenantactivity.Handler

	Migrator *db.Migrator
}

// GetGraphQL is the safe way to retrieve GraphQL from the state as it can be
// replaced at runtime. Instead of passing appState.GraphQL to your adapters,
// pass appState itself which you can abstract with a local interface such as:
//
// type gqlProvider interface { GetGraphQL graphql.GraphQL }
func (s *State) GetGraphQL() graphql.GraphQL {
	s.gqlMutex.Lock()
	gql := s.GraphQL
	s.gqlMutex.Unlock()
	return gql
}

func (s *State) SetGraphQL(gql graphql.GraphQL) {
	s.gqlMutex.Lock()
	s.GraphQL = gql
	s.gqlMutex.Unlock()
}
