package eventstore

import (
	"context"
	"encoding/json"
	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/jiaming2012/http-cache/src/constants"
	"time"
)

const (
	CachedEntityType string = "CachedEntity"
)

// CachedEntity is a stored cached response
type CachedEntity struct {
	Key        string
	Content    []byte
	Expiration int64
	Version    uint
}

func (cachedEntity *CachedEntity) AppendToStream(db *esdb.Client) error {
	data, err := json.Marshal(cachedEntity)
	if err != nil {
		return err
	}

	eventData := esdb.EventData{
		ContentType: esdb.JsonContentType,
		EventType:   CachedEntityType,
		Data:        data,
	}

	_, err = db.AppendToStream(context.Background(), constants.HttpCacheStreamName, esdb.AppendToStreamOptions{}, eventData)
	if err != nil {
		return err
	}

	return nil
}

// Expired returns true if the item has expired.
func (cachedEntity CachedEntity) Expired() bool {
	if cachedEntity.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > cachedEntity.Expiration
}
