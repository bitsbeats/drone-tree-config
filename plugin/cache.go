package plugin

import (
	"fmt"
	"sync"
	"time"

	"github.com/drone/drone-go/drone"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	msgCacheHit    = "%s config-cache found entry for %s"
	msgCacheExpire = "config-cache expired entry for %s"
	msgCacheAdd    = "%s config-cache added entry for %s"
)

// configCache is used to cache config responses on a per request basis
type configCache struct {
	syncMap sync.Map
}

// cacheKey holds the unique key details which are associated with the config request
type cacheKey struct {
	slug    string
	ref     string
	before  string
	after   string
	event   string
	trigger string
	author  string
}

// cacheEntry holds the response and ttl for a config request
type cacheEntry struct {
	config string
	error  error
	ttl    *time.Timer
}

// newCacheEntry creates a new cacheEntry using the config string and error provided. The returned struct will have a
// nil ttl value -- it will be established when a entry is added to the cache via the add function.
func newCacheEntry(config string, error error) *cacheEntry {
	entry := &cacheEntry{
		config: config,
		error:  error,
	}
	return entry
}

// newCacheKey creates a new cacheKey for the provided request.
func newCacheKey(req *request) cacheKey {
	ck := cacheKey{
		slug:    req.Repo.Slug,
		ref:     req.Build.Ref,
		before:  req.Build.Before,
		after:   req.Build.After,
		event:   req.Build.Event,
		author:  req.Build.Author,
		trigger: req.Build.Trigger,
	}
	return ck
}

// add an entry to the cache
func (c *configCache) add(uuid uuid.UUID, key cacheKey, entry *cacheEntry, ttl time.Duration) {
	logrus.Infof(msgCacheAdd, uuid, fmt.Sprintf("%+v", key))

	entry.ttl = time.AfterFunc(ttl, func() {
		c.expire(key)
	})

	c.syncMap.Store(key, entry)
}

// expire is typically called internally via a time.Afterfunc
func (c *configCache) expire(key cacheKey) {
	logrus.Infof(msgCacheExpire, fmt.Sprintf("%+v", key))

	if entry, _ := c.syncMap.Load(key); entry != nil {
		entry.(*cacheEntry).ttl.Stop()
		c.syncMap.Delete(key)
	}
}

// retrieve an entry from the cache, if it exists
func (c *configCache) retrieve(uuid uuid.UUID, key cacheKey) (*cacheEntry, bool) {
	entry, exists := c.syncMap.Load(key)
	if exists {
		logrus.Infof(msgCacheHit, uuid, fmt.Sprintf("%+v", key))
		return entry.(*cacheEntry), true
	}

	return nil, false
}

// cacheAndReturn caches the result (if enabled) and returns the (drone.Config, error) that should be
// returned to the Find request.
func (p *Plugin) cacheAndReturn(uuid uuid.UUID, key cacheKey, entry *cacheEntry) (*drone.Config, error) {
	var config *drone.Config
	if entry.config != "" {
		config = &drone.Config{Data: entry.config}
	}

	// cache the config before we return it, if enabled
	if p.cacheTTL > 0 {
		p.cache.add(uuid, key, entry, p.cacheTTL)
	}

	return config, entry.error
}
