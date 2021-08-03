package mongo

import (
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
	"net/http"
	"strings"
)

const deviceLocalIdFieldName = "Device.LocalId"
const deviceUserIdFieldName = "UserId"
const deviceHiddenFieldName = "Hidden"

var deviceLocalIdKey string
var deviceUserIdKey string
var deviceHiddenKey string

func init() {
	var err error
	deviceLocalIdKey, err = getBsonFieldName(model.Device{}, deviceLocalIdFieldName)
	if err != nil {
		log.Fatal(err)
	}
	deviceUserIdKey, err = getBsonFieldName(model.Device{}, deviceUserIdFieldName)
	if err != nil {
		log.Fatal(err)
	}
	deviceHiddenKey, err = getBsonFieldName(model.Device{}, deviceHiddenFieldName)
	if err != nil {
		log.Fatal(err)
	}

	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceCollection)
		err = db.ensureIndex(collection, "devicelocalidindex", deviceLocalIdKey, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceCollection)
}

func (this *Mongo) ListDevices(userId string, limit int, offset int, sort string, showHidden bool) (result []model.Device, total int64, err error, errCode int) {
	opt := options.Find()
	opt.SetLimit(int64(limit))
	opt.SetSkip(int64(offset))

	parts := strings.Split(sort, ".")
	sortby := deviceLocalIdKey
	switch parts[0] {
	case "local_id":
		sortby = deviceLocalIdKey
		//TODO: other fields
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bsonx.Doc{{sortby, bsonx.Int32(direction)}})

	filter := bson.M{deviceUserIdKey: userId}
	if !showHidden {
		filter[deviceHiddenKey] = false
	}

	ctx, _ := getTimeoutContext()
	collection := this.deviceCollection()

	total, err = collection.CountDocuments(ctx, filter)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	cursor, err := collection.Find(ctx, filter, opt)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	for cursor.Next(ctx) {
		element := model.Device{}
		err = cursor.Decode(&element)
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
		result = append(result, element)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ReadDevice(localId string) (result model.Device, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	temp := this.deviceCollection().FindOne(
		ctx,
		bson.M{
			deviceLocalIdKey: localId,
		})
	err = temp.Err()
	if err == mongo.ErrNoDocuments {
		return result, err, http.StatusNotFound
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	err = temp.Decode(&result)
	if err == mongo.ErrNoDocuments {
		return result, err, http.StatusNotFound
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}

func (this *Mongo) SetDevice(device model.Device) (error, int) {
	ctx, _ := getTimeoutContext()
	_, err := this.deviceCollection().ReplaceOne(
		ctx,
		bson.M{
			deviceLocalIdKey: device.LocalId,
		},
		device,
		options.Replace().SetUpsert(true))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Mongo) RemoveDevice(localId string) (error, int) {
	ctx, _ := getTimeoutContext()
	_, err := this.deviceCollection().DeleteMany(
		ctx,
		bson.M{
			deviceLocalIdKey: localId,
		})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
