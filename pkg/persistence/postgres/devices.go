/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/model"
	"github.com/SENERGY-Platform/device-waiting-room/pkg/persistence/options"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	CreateTables = append(CreateTables, CreateDevicesTable)
}

func CreateDevicesTable(db *Postgres) error {
	ctx := db.getTimeoutContext()
	_, err := db.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS devices (
    	local_id TEXT PRIMARY KEY,
    	id TEXT, 
    	name TEXT,
    	device_type_id TEXT,
    	attributes JSON,
    	user_id TEXT, 
    	hidden BOOL, 
    	created_at timestamptz,
		updated_at timestamptz);
`)
	if err != nil {
		log.Println("ERROR: unable to create table:", err)
		return err
	}
	return nil
}

func (this *Postgres) ListDevices(userId string, options options.List) (result []model.Device, total int64, err error, errCode int) {
	timeout := this.getTimeoutContext()
	parts := strings.Split(options.Sort, ".")
	sortby := "name"
	switch parts[0] {
	case "local_id", "name", "created_at", "updated_at":
		sortby = parts[0]
	default:
		return result, total, errors.New("unknown sort field"), http.StatusBadRequest
	}
	direction := "ASC"
	if len(parts) > 1 && parts[1] == "desc" {
		direction = "DESC"
	}

	where, args := this.getDeviceWhere(userId, options)

	query := fmt.Sprintf(`SELECT 
		local_id, 
		id, 
		name, 
		device_type_id, 
		attributes, 
		user_id, 
		hidden, 
		created_at, 
		updated_at 
	FROM devices WHERE %v ORDER BY %v %v LIMIT %v OFFSET %v`, where, sortby, direction, options.Limit, options.Offset)

	rows, err := this.db.QueryContext(timeout, query, args...)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	for rows.Next() {
		element := model.Device{}
		attrBuf := []byte{}
		err = rows.Scan(&element.LocalId, &element.Id, &element.Name, &element.DeviceTypeId, &attrBuf, &element.UserId, &element.Hidden, &element.CreatedAt, &element.LastUpdate)
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
		err = json.Unmarshal(attrBuf, &element.Attributes)
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
		result = append(result, element)
	}
	total, err, errCode = this.listDevicesTotal(where, args)
	if err != nil {
		return result, total, err, errCode
	}
	return result, total, nil, http.StatusOK
}

func (this *Postgres) listDevicesTotal(where string, args []any) (total int64, err error, errCode int) {
	timeout := this.getTimeoutContext()
	query := fmt.Sprintf(`SELECT COUNT(local_id) FROM devices WHERE %v`, where)
	row := this.db.QueryRowContext(timeout, query, args...)
	if err = row.Err(); err != nil {
		return total, err, http.StatusInternalServerError
	}
	err = row.Scan(&total)
	if err != nil {
		return total, err, http.StatusInternalServerError
	}
	if err = row.Err(); err != nil {
		return total, err, http.StatusInternalServerError
	}
	return total, nil, http.StatusOK
}

func (this *Postgres) getDeviceWhere(userId string, options options.List) (where string, args []any) {
	and := []string{"user_id = $1"}
	args = []any{userId}
	if !options.ShowHidden {
		args = append(args, false)
		and = append(and, "hidden = $"+strconv.Itoa(len(args)))
	}
	if options.Search != "" {
		//TODO: search
	}
	return strings.Join(and, " AND "), args
}

func (this *Postgres) ReadDevice(localId string) (result model.Device, err error, errCode int) {
	timeout := this.getTimeoutContext()
	query := `SELECT 
		local_id, 
		id, 
		name, 
		device_type_id, 
		attributes, 
		user_id, 
		hidden, 
		created_at, 
		updated_at 
	FROM devices WHERE local_id = $1 LIMIT 1`

	rows := this.db.QueryRowContext(timeout, query, localId)
	if err = rows.Err(); err != nil {
		return result, err, getErrCode(err)
	}
	attrBuf := []byte{}
	err = rows.Scan(&result.LocalId, &result.Id, &result.Name, &result.DeviceTypeId, &attrBuf, &result.UserId, &result.Hidden, &result.CreatedAt, &result.LastUpdate)
	if err != nil {
		return result, err, getErrCode(err)
	}
	if err = rows.Err(); err != nil {
		return result, err, getErrCode(err)
	}
	err = json.Unmarshal(attrBuf, &result.Attributes)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}

func getErrCode(err error) int {
	switch err {
	case nil:
		return http.StatusOK
	case sql.ErrNoRows:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (this *Postgres) SetDevice(device model.Device) (error, int) {
	query := `INSERT INTO devices(local_id, 
			id, 
			name, 
			device_type_id, 
			attributes, 
			user_id, 
			hidden, 
			created_at, 
			updated_at) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (local_id) DO UPDATE SET
		  id = EXCLUDED.id,
		  name = EXCLUDED.name,
		  device_type_id = EXCLUDED.device_type_id,
		  attributes = EXCLUDED.attributes,
		  user_id = EXCLUDED.user_id, 
		  hidden = EXCLUDED.hidden,
		  created_at = EXCLUDED.created_at,
		  updated_at = EXCLUDED.updated_at;`

	if device.Attributes == nil {
		device.Attributes = []models.Attribute{}
	}
	attrBuf, err := json.Marshal(device.Attributes)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	timeout := this.getTimeoutContext()

	_, err = this.db.ExecContext(timeout, query,
		device.LocalId,      // $1
		device.Id,           // $2
		device.Name,         // $3
		device.DeviceTypeId, // $4
		attrBuf,             // $5
		device.UserId,       // $6
		device.Hidden,       // $7
		device.CreatedAt,    // $8
		device.LastUpdate,   // $9
	)

	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Postgres) RemoveDevice(localId string) (error, int) {
	query := "DELETE FROM devices WHERE local_id = $1"
	timeout := this.getTimeoutContext()
	_, err := this.db.ExecContext(timeout, query, localId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, http.StatusOK
		}
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
