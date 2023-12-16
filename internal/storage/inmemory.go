package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

type allCache struct {
	urls *cache.Cache
}

const (
	defaultExpiration = 5 * time.Minute
	purgeTime         = 10 * time.Minute
)

func NewCache() *allCache {
	return &allCache{}
}

func (c *allCache) Init() error {
	Cache := cache.New(defaultExpiration, purgeTime)
	*c = allCache{urls: Cache}
	return nil
}

func (c *allCache) AddURL(ctx context.Context, url *URL) error {
	c.urls.Set(url.ShortenPath, url, cache.DefaultExpiration)
	return nil
}

func (c *allCache) ReturnURL(ctx context.Context, short string) (*URL, error) {
	url, ok := c.urls.Get(short)
	if ok {
		return url.(*URL), nil
	}

	return nil, fmt.Errorf("not found")

}

func (c *allCache) CheckExists(domain string, path string) (*URL, error) {
	allItems := c.urls.Items()
	for _, v := range allItems {
		if v.Object.(*URL).Domain == domain && v.Object.(*URL).OriginalPath == path {
			return v.Object.(*URL), nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func (c *allCache) ClearTable() error {
	c.urls.Flush()
	return nil
}
