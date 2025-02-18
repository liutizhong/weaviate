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

package schema

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	command "github.com/liutizhong/weaviate/cluster/proto/api"
	"github.com/liutizhong/weaviate/cluster/types"
	"github.com/liutizhong/weaviate/entities/models"
	entSchema "github.com/liutizhong/weaviate/entities/schema"
	"github.com/liutizhong/weaviate/entities/versioned"
	"github.com/liutizhong/weaviate/usecases/sharding"
)

var (
	ErrClassExists   = errors.New("class already exists")
	ErrClassNotFound = errors.New("class not found")
	ErrShardNotFound = errors.New("shard not found")
	ErrMTDisabled    = errors.New("multi-tenancy is not enabled")
)

type ClassInfo struct {
	Exists            bool
	MultiTenancy      models.MultiTenancyConfig
	ReplicationFactor int
	Tenants           int
	Properties        int
	ClassVersion      uint64
	ShardVersion      uint64
}

func (ci *ClassInfo) Version() uint64 {
	return max(ci.ClassVersion, ci.ShardVersion)
}

type schema struct {
	nodeID      string
	shardReader shardReader
	sync.RWMutex
	Classes map[string]*metaClass
}

func (s *schema) ClassInfo(class string) ClassInfo {
	s.RLock()
	defer s.RUnlock()
	cl, ok := s.Classes[class]
	if !ok {
		return ClassInfo{}
	}
	return cl.ClassInfo()
}

// ClassEqual returns the name of an existing class with a similar name, and "" otherwise
// strings.EqualFold is used to compare classes
func (s *schema) ClassEqual(name string) string {
	s.RLock()
	defer s.RUnlock()
	for k := range s.Classes {
		if strings.EqualFold(k, name) {
			return k
		}
	}
	return ""
}

func (s *schema) MultiTenancy(class string) models.MultiTenancyConfig {
	mtc, _ := s.metaClass(class).MultiTenancyConfig()
	return mtc
}

// Read performs a read operation `reader` on the specified class and sharding state
func (s *schema) Read(class string, reader func(*models.Class, *sharding.State) error) error {
	meta := s.metaClass(class)
	if meta == nil {
		return ErrClassNotFound
	}
	return meta.RLockGuard(reader)
}

func (s *schema) metaClass(class string) *metaClass {
	s.RLock()
	defer s.RUnlock()
	return s.Classes[class]
}

// ReadOnlyClass returns a shallow copy of a class.
// The copy is read-only and should not be modified.
func (s *schema) ReadOnlyClass(class string) (*models.Class, uint64) {
	s.RLock()
	defer s.RUnlock()
	meta := s.Classes[class]
	if meta == nil {
		return nil, 0
	}
	return meta.CloneClass(), meta.ClassVersion
}

// ReadOnlyClass returns a shallow copy of a class.
// The copy is read-only and should not be modified.
func (s *schema) ReadOnlyClasses(classes ...string) map[string]versioned.Class {
	if len(classes) == 0 {
		return nil
	}

	vclasses := make(map[string]versioned.Class, len(classes))
	s.RLock()
	defer s.RUnlock()

	for _, class := range classes {
		meta := s.Classes[class]
		if meta == nil {
			continue
		}
		vclasses[class] = versioned.Class{Class: meta.CloneClass(), Version: meta.ClassVersion}
	}

	return vclasses
}

// ReadOnlySchema returns a read only schema
// Changing the schema outside this package might lead to undefined behavior.
//
// it creates a shallow copy of existing classes
//
// This function assumes that class attributes are being overwritten.
// The properties attribute is the only one that might vary in size;
// therefore, we perform a shallow copy of the existing properties.
// This implementation assumes that individual properties are overwritten rather than partially updated
func (s *schema) ReadOnlySchema() models.Schema {
	cp := models.Schema{}
	s.RLock()
	defer s.RUnlock()
	cp.Classes = make([]*models.Class, len(s.Classes))
	i := 0
	for _, meta := range s.Classes {
		cp.Classes[i] = meta.CloneClass()
		i++
	}

	return cp
}

// ShardOwner returns the node owner of the specified shard
func (s *schema) ShardOwner(class, shard string) (string, uint64, error) {
	meta := s.metaClass(class)
	if meta == nil {
		return "", 0, ErrClassNotFound
	}

	return meta.ShardOwner(shard)
}

// ShardFromUUID returns shard name of the provided uuid
func (s *schema) ShardFromUUID(class string, uuid []byte) (string, uint64) {
	meta := s.metaClass(class)
	if meta == nil {
		return "", 0
	}
	return meta.ShardFromUUID(uuid)
}

// ShardReplicas returns the replica nodes of a shard
func (s *schema) ShardReplicas(class, shard string) ([]string, uint64, error) {
	meta := s.metaClass(class)
	if meta == nil {
		return nil, 0, ErrClassNotFound
	}
	return meta.ShardReplicas(shard)
}

// TenantsShards returns shard name for the provided tenant and its activity status
func (s *schema) TenantsShards(class string, tenants ...string) (map[string]string, uint64) {
	s.RLock()
	defer s.RUnlock()

	meta := s.Classes[class]
	if meta == nil {
		return nil, 0
	}

	return meta.TenantsShards(class, tenants...)
}

func (s *schema) CopyShardingState(class string) (*sharding.State, uint64) {
	meta := s.metaClass(class)
	if meta == nil {
		return nil, 0
	}

	return meta.CopyShardingState()
}

func (s *schema) GetShardsStatus(class, tenant string) (models.ShardStatusList, error) {
	return s.shardReader.GetShardsStatus(class, tenant)
}

type shardReader interface {
	GetShardsStatus(class, tenant string) (models.ShardStatusList, error)
}

func NewSchema(nodeID string, shardReader shardReader) *schema {
	return &schema{
		nodeID:      nodeID,
		Classes:     make(map[string]*metaClass, 128),
		shardReader: shardReader,
	}
}

func (s *schema) len() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.Classes)
}

func (s *schema) multiTenancyEnabled(class string) (bool, *metaClass, ClassInfo, error) {
	s.RLock()
	defer s.RUnlock()
	meta := s.Classes[class]
	if meta == nil {
		return false, nil, ClassInfo{}, ErrClassNotFound
	}
	info := s.Classes[class].ClassInfo()
	if !info.MultiTenancy.Enabled {
		return false, nil, ClassInfo{}, fmt.Errorf("%w for class %q", ErrMTDisabled, class)
	}
	return true, meta, info, nil
}

func (s *schema) addClass(cls *models.Class, ss *sharding.State, v uint64) error {
	s.Lock()
	defer s.Unlock()
	_, exists := s.Classes[cls.Class]
	if exists {
		return ErrClassExists
	}

	s.Classes[cls.Class] = &metaClass{
		Class: *cls, Sharding: *ss, ClassVersion: v, ShardVersion: v,
	}
	return nil
}

// updateClass modifies existing class based on the givin update function
func (s *schema) updateClass(name string, f func(*metaClass) error) error {
	s.Lock()
	defer s.Unlock()

	meta := s.Classes[name]
	if meta == nil {
		return ErrClassNotFound
	}
	return meta.LockGuard(f)
}

// replaceStatesNodeName it update the node name inside sharding states.
// WARNING: this shall be used in one node cluster environments only.
// because it will replace the shard node name if the node name got updated
// only if the replication factor is 1, otherwise it's no-op
func (s *schema) replaceStatesNodeName(new string) {
	s.Lock()
	defer s.Unlock()

	for _, meta := range s.Classes {
		meta.LockGuard(func(mc *metaClass) error {
			if meta.Class.ReplicationConfig.Factor > 1 {
				return nil
			}

			for idx := range meta.Sharding.Physical {
				cp := meta.Sharding.Physical[idx].DeepCopy()
				cp.BelongsToNodes = []string{new}
				meta.Sharding.Physical[idx] = cp
			}
			return nil
		})
	}
}

func (s *schema) deleteClass(name string) {
	s.Lock()
	defer s.Unlock()
	delete(s.Classes, name)
}

func (s *schema) addProperty(class string, v uint64, props ...*models.Property) error {
	s.Lock()
	defer s.Unlock()

	meta := s.Classes[class]
	if meta == nil {
		return ErrClassNotFound
	}
	return meta.AddProperty(v, props...)
}

func (s *schema) addTenants(class string, v uint64, req *command.AddTenantsRequest) error {
	req.Tenants = removeNilTenants(req.Tenants)

	if ok, meta, info, err := s.multiTenancyEnabled(class); !ok {
		return err
	} else {
		return meta.AddTenants(s.nodeID, req, int64(info.ReplicationFactor), v)
	}
}

func (s *schema) deleteTenants(class string, v uint64, req *command.DeleteTenantsRequest) error {
	if ok, meta, _, err := s.multiTenancyEnabled(class); !ok {
		return err
	} else {
		return meta.DeleteTenants(req, v)
	}
}

func (s *schema) updateTenants(class string, v uint64, req *command.UpdateTenantsRequest) error {
	if ok, meta, _, err := s.multiTenancyEnabled(class); !ok {
		return err
	} else {
		return meta.UpdateTenants(s.nodeID, req, v)
	}
}

func (s *schema) updateTenantsProcess(class string, v uint64, req *command.TenantProcessRequest) error {
	if ok, meta, _, err := s.multiTenancyEnabled(class); !ok {
		return err
	} else {
		return meta.UpdateTenantsProcess(s.nodeID, req, v)
	}
}

func (s *schema) getTenants(class string, tenants []string) ([]*models.TenantResponse, error) {
	ok, meta, _, err := s.multiTenancyEnabled(class)
	if !ok {
		return nil, err
	}

	// Read tenants using the meta lock guard
	var res []*models.TenantResponse
	f := func(_ *models.Class, ss *sharding.State) error {
		if len(tenants) == 0 {
			res = make([]*models.TenantResponse, len(ss.Physical))
			i := 0
			for tenant, physical := range ss.Physical {
				// Ensure we copy the belongs to nodes array to avoid it being modified
				cpy := make([]string, len(physical.BelongsToNodes))
				copy(cpy, physical.BelongsToNodes)

				res[i] = MakeTenantWithBelongsToNodes(tenant, entSchema.ActivityStatus(physical.Status), cpy)

				// Increment our result iterator
				i++
			}
		} else {
			res = make([]*models.TenantResponse, 0, len(tenants))
			for _, tenant := range tenants {
				if physical, ok := ss.Physical[tenant]; ok {
					// Ensure we copy the belongs to nodes array to avoid it being modified
					cpy := make([]string, len(physical.BelongsToNodes))
					copy(cpy, physical.BelongsToNodes)
					res = append(res, MakeTenantWithBelongsToNodes(tenant, entSchema.ActivityStatus(physical.Status), cpy))
				}
			}
		}
		return nil
	}
	return res, meta.RLockGuard(f)
}

func (s *schema) States() map[string]types.ClassState {
	s.RLock()
	defer s.RUnlock()

	cs := make(map[string]types.ClassState, len(s.Classes))
	for _, c := range s.Classes {
		cs[c.Class.Class] = types.ClassState{
			Class:  c.Class,
			Shards: c.Sharding,
		}
	}

	return cs
}

func (s *schema) MetaClasses() map[string]*metaClass {
	s.RLock()
	defer s.RUnlock()

	return s.Classes
}

// makeTenant creates a tenant with the given name and status
func makeTenant(name, status string) models.Tenant {
	return models.Tenant{
		Name:           name,
		ActivityStatus: status,
	}
}

// MakeTenantWithBelongsToNodes creates a tenant with the given name, status, and belongsToNodes
func MakeTenantWithBelongsToNodes(name, status string, belongsToNodes []string) *models.TenantResponse {
	return &models.TenantResponse{
		Tenant:         makeTenant(name, status),
		BelongsToNodes: belongsToNodes,
	}
}
