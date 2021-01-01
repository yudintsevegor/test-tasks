package store

import (
	"context"
	"errors"
	"fmt"
	"log"

	"not-for-work/aviasales_test/config"
	"not-for-work/aviasales_test/internal/uniq"

	"go.mongodb.org/mongo-driver/x/bsonx"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	EmptyResultErr   = errors.New("store: empty result")
	EmptyDatabaseErr = errors.New("store: empty database")
	NotUniqWordsErr  = errors.New("store: not uniq words")
)

type Store struct {
	client   *mongo.Client
	database *mongo.Database

	collectionName string
}

func New(cfg config.Mongo, dropData bool) (*Store, error) {
	DSN := fmt.Sprintf("mongodb://%s:%s/%s", cfg.Host, cfg.Port, cfg.Name)

	client, err := mongo.NewClient(options.Client().ApplyURI(DSN))
	if err != nil {
		return nil, fmt.Errorf("create new mongo client by DSN %q: %w", DSN, err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("connect to mongo client by DSN %q: %w", DSN, err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("ping to mongo client by DSN %q: %w", DSN, err)
	}

	index := mongo.IndexModel{
		Keys: bsonx.Doc{bsonx.Elem{
			Key:   "key",
			Value: bsonx.Int32(1),
		}},
		Options: options.Index().SetUnique(true),
	}

	database := client.Database(cfg.Name)
	_, err = database.Collection(cfg.CollectionName).Indexes().CreateOne(context.Background(), index)
	if err != nil {
		return nil, fmt.Errorf("create uniq field in collection %q: %w", cfg.CollectionName, err)
	}

	if dropData {
		err = database.Collection(cfg.CollectionName).Drop(context.Background())
		if err != nil {
			return nil, fmt.Errorf("drop collection %q: %w", cfg.CollectionName, err)
		}

		log.Println("STORE DATA DROPPED")
	}

	s := &Store{
		client:         client,
		database:       database,
		collectionName: cfg.CollectionName,
	}

	return s, nil
}

func (s *Store) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

func (s *Store) Get(ctx context.Context, key string, record *Object) error {
	if s.database == nil {
		return EmptyDatabaseErr
	}

	filter := bson.M{"key": key}
	collection := s.database.Collection(s.collectionName)
	err := collection.FindOne(ctx, filter).Decode(record)
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		return EmptyResultErr
	case err != nil:
		return fmt.Errorf("mongo: finding one: %w", err)
	}

	return nil
}

func (s *Store) Put(ctx context.Context, record Object) (int, error) {
	if s.database == nil {
		return 0, EmptyDatabaseErr
	}

	var wordsCount int
	fn := func() error {
		object := new(Object)
		err := s.Get(ctx, record.Key, object)
		switch {
		case errors.Is(err, EmptyResultErr):
			collection := s.database.Collection(s.collectionName)
			_, err = collection.InsertOne(ctx, record)
			if err != nil {
				return fmt.Errorf("mongo: insert one: %w", err)
			}
			wordsCount = len(record.Words)
			return nil
		case err != nil:
			return fmt.Errorf("mongo: dinding one: %w", err)
		}

		words, ok := uniq.Find(object.Words, record.Words)
		if !ok {
			return NotUniqWordsErr
		}
		wordsCount = len(words) + len(object.Words)

		err = s.Update(ctx, record.Key, words)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}

		return nil
	}

	err := s.MakeTransaction(fn)
	if err != nil {
		return 0, fmt.Errorf("make transaction: %w", err)
	}

	return wordsCount, nil
}

func (s *Store) Update(ctx context.Context, key string, values []string) error {
	if s.database == nil {
		return EmptyDatabaseErr
	}

	var (
		filter = bson.M{"key": key}
		update = bson.M{"$push": bson.M{"words": bson.M{"$each": values}}}
	)

	collection := s.database.Collection(s.collectionName)
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("update one in collection %q: %w", s.collectionName, err)
	}

	return nil
}

func (s *Store) MakeTransaction(fn func() error) error {
	if s.database == nil {
		return EmptyDatabaseErr
	}

	session, err := s.client.StartSession()
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	defer session.EndSession(context.Background())

	if err := session.StartTransaction(); err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	sessionFunc := func(ctx mongo.SessionContext) error {
		if err := fn(); err != nil {
			return fmt.Errorf("general function execution: %w", err)
		}

		if err := session.CommitTransaction(ctx); err != nil {
			return fmt.Errorf("commiting transaction: %w", err)
		}

		return nil
	}

	err = mongo.WithSession(context.Background(), session, sessionFunc)
	if err != nil {
		return fmt.Errorf("with session: %w", err)
	}

	return nil
}
