package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/jiaming2012/http-cache/src/cache/memory"
	"github.com/jiaming2012/http-cache/src/constants"
	log "github.com/jiaming2012/http-cache/src/logger"
	"time"
)

type Storage struct {
	data *memory.Storage
	db   *esdb.Client
}

func (s Storage) Get(key string) []byte {
	return s.data.Get(key)
}

func (s Storage) Set(key string, content []byte, duration time.Duration) {
	cachedEntity := CachedEntity{
		Key:        key,
		Content:    content,
		Expiration: int64(duration),
		Version:    constants.HttpCacheStreamVersion,
	}

	err := cachedEntity.AppendToStream(s.db)
	if err != nil {
		log.Logger.Errorf("failed to cache %v: %v", key, err)
	}
}

func handleEvent(ev *esdb.SubscriptionEvent, data *memory.Storage) error {
	event := ev.EventAppeared.Event

	switch event.EventType {
	case CachedEntityType:
		// todo: proper event source build
		var entity CachedEntity
		jsonErr := json.Unmarshal(event.Data, &entity)
		if jsonErr != nil {
			return fmt.Errorf("failed to restore CachedEntity for EventID, %w", event.EventID)
		}

		// build in memory cache
		if entity.Version >= constants.HttpCacheStreamVersion {
			log.Logger.Debugf("Adding EventNumber %v to cached", event.EventNumber)
			data.Set(entity.Key, entity.Content, time.Duration(entity.Expiration))
		}
	default:
		log.Logger.Fatalf("Unknown EventType: %v", event.EventType)
	}
	return nil
}

func BuildCacheFromCachedEntityEventStream(ctx context.Context, db *esdb.Client, data *memory.Storage) {
	stream, dbErr := db.SubscribeToStream(ctx, constants.HttpCacheStreamName, esdb.SubscribeToStreamOptions{
		From: esdb.Revision(0),
	})

	if dbErr != nil {
		panic(dbErr)
	}

	defer stream.Close()

	for {
		select {
		case <-ctx.Done():
			break
		default:
			ev := stream.Recv()

			if ev.SubscriptionDropped != nil {
				log.Logger.Warnf("Subscription dropped for %v", constants.HttpCacheStreamName)
				break
			}

			if ev.EventAppeared != nil {
				err := handleEvent(ev, data)
				if err != nil {
					log.Logger.Error(err)
				}
			}
		}
	}

	log.Logger.Infof("Exiting BuildCacheFromCachedEntityEventStream ...")
}

func NewStorage() (*Storage, error) {
	settings, err := esdb.ParseConnectionString(constants.EsdbURL)
	if err != nil {
		return nil, fmt.Errorf("esdb.ParseConnectionString: %w", err)
	}

	db, err := esdb.NewClient(settings)
	if err != nil {
		return nil, fmt.Errorf("esdb.NewClient: %w", err)
	}

	data := memory.NewStorage()

	// todo: send context.Done() when web server is shut down
	go BuildCacheFromCachedEntityEventStream(context.Background(), db, data)

	return &Storage{
		db:   db,
		data: data,
	}, nil
}
