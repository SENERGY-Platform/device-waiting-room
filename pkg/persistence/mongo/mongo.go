package mongo

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/configuration"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Mongo struct {
	config configuration.Config
	db     *mongo.Client
}

var CreateCollections = []func(db *Mongo) error{}

func New(ctx context.Context, wg *sync.WaitGroup, conf configuration.Config) (*Mongo, error) {
	timeout, _ := getTimeoutContext()
	db, err := mongo.Connect(timeout, options.Client().ApplyURI(conf.MongoUrl))
	if err != nil {
		return nil, err
	}
	client := &Mongo{config: conf, db: db}
	for _, creators := range CreateCollections {
		err = creators(client)
		if err != nil {
			client.disconnect()
			return nil, err
		}
	}
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		<-ctx.Done()
		client.disconnect()
		if wg != nil {
			wg.Done()
		}
	}()
	return client, nil
}

func (this *Mongo) ensureIndex(collection *mongo.Collection, indexname string, indexKey string, asc bool, unique bool) error {
	ctx, _ := getTimeoutContext()
	var direction int32 = -1
	if asc {
		direction = 1
	}
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{indexKey, direction}},
		Options: options.Index().SetName(indexname).SetUnique(unique),
	})
	return err
}

func (this *Mongo) ensureCompoundIndex(collection *mongo.Collection, indexname string, asc bool, unique bool, indexKeys ...string) error {
	ctx, _ := getTimeoutContext()
	var direction int32 = -1
	if asc {
		direction = 1
	}
	keys := []bson.E{}
	for _, key := range indexKeys {
		keys = append(keys, bson.E{Key: key, Value: direction})
	}
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D(keys),
		Options: options.Index().SetName(indexname).SetUnique(unique),
	})
	return err
}

func (this *Mongo) ensureTextIndex(collection *mongo.Collection, indexname string, indexKeys ...string) error {
	if len(indexKeys) == 0 {
		return errors.New("expect at least one key")
	}
	keys := bson.D{}
	for _, key := range indexKeys {
		keys = append(keys, bson.E{Key: key, Value: "text"})
	}
	ctx, _ := getTimeoutContext()
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetName(indexname),
	})
	return err
}

func (this *Mongo) disconnect() {
	timeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
	log.Println("disconnect mongo:", this.db.Disconnect(timeout))
}

func (this *Mongo) getSearchTokens(device model.Device) string {
	return strings.Join([]string{
		device.LocalId,
		device.Name,
		strings.NewReplacer(
			"_", " ",
			"-", " ",
			".", " ",
			":", " ",
			"/", " ").
			Replace(device.Name),
	}, " ")
}

func getBsonFieldName(obj interface{}, fieldName string) (bsonName string, err error) {
	field, found := reflect.TypeOf(obj).FieldByName(fieldName)
	if !found {
		return "", errors.New("field '" + fieldName + "' not found")
	}
	tags, err := bsoncodec.DefaultStructTagParser.ParseStructTags(field)
	return tags.Name, err
}

func getBsonFieldPath(obj interface{}, path string) (bsonPath string, err error) {
	t := reflect.TypeOf(obj)
	pathParts := strings.Split(path, ".")
	bsonPathParts := []string{}
	for _, name := range pathParts {
		field, found := t.FieldByName(name)
		if !found {
			return "", errors.New("field path '" + path + "' not found at '" + name + "'")
		}
		tags, err := bsoncodec.DefaultStructTagParser.ParseStructTags(field)
		if err != nil {
			return bsonPath, err
		}
		bsonPathParts = append(bsonPathParts, tags.Name)
		t = field.Type
	}
	bsonPath = strings.Join(bsonPathParts, ".")
	return
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
