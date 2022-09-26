package handler

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/bigtable"
	"github.com/fubotv/keyplay/app/db"
	"github.com/fubotv/keyplay/app/model"
	"github.com/fubotv/keyplay/app/util"
	"goji.io/pat"
)

type ServiceHandler struct {
	DatabaseHandler *db.DBHandler
}

const (
	RowKeyDelimiter = "#"
)

func (s ServiceHandler) GetKeyplayAttribute(w http.ResponseWriter, r *http.Request) {

	key := pat.Param(r, "key")
	fmt.Println(key, "key data")

	filter := bigtable.RowKeyFilter("^" + "attribute" + RowKeyDelimiter + key + "$")
	fmt.Println(filter, "filter in places")

	keyplay, err := db.ReadRowFromBT(s.DatabaseHandler.Table, filter)
	fmt.Println(keyplay, "keyplay in place")

	if err != nil {
		util.JsonError(context.Background(), w, http.StatusNotFound, err)
	}

	util.Json(context.Background(), w, http.StatusOK, keyplay)

}

func (s ServiceHandler) CreateAttributeKeyplay(w http.ResponseWriter, r *http.Request) {
	var newAttribute model.Attribute

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	if len(reqBody) > 0 {
		json.Unmarshal(reqBody, &newAttribute)

		rowKey := generateRowkey(newAttribute.Name)

		fmt.Println(rowKey, "rowkey in place")

		mut := bigtable.NewMutation()
		binary.Write(new(bytes.Buffer), binary.BigEndian, int64(1))

		multipleFieldMarshalled, _ := json.Marshal(newAttribute.IsMultipleField)
		mandatoryMarshalled, _ := json.Marshal(newAttribute.IsMandatory)
		nameMarshalled, _ := json.Marshal(newAttribute.Name)
		datatypeMarshalled, _ := json.Marshal(newAttribute.DataType)

		mut.Set(db.ColumnFamilyName, "name", bigtable.Now(), nameMarshalled)
		mut.Set(db.ColumnFamilyName, "isMultipleField", bigtable.Now(), multipleFieldMarshalled)
		mut.Set(db.ColumnFamilyName, "isMandatory", bigtable.Now(), mandatoryMarshalled)
		mut.Set(db.ColumnFamilyName, "dataType", bigtable.Now(), datatypeMarshalled)

		err := db.WriteToBT(s.DatabaseHandler.Table, rowKey, mut)
		if err != nil {
			util.JsonError(context.Background(), w, http.StatusNotFound, err)
		}
	} else {
		util.JsonError(context.Background(), w, http.StatusMethodNotAllowed, errors.New("please enter all the attribute details"))
	}

}

func (s ServiceHandler) UpdateAttributeKeyplay(w http.ResponseWriter, r *http.Request) {
	var updatedAttribute model.Attribute

	key := pat.Param(r, "key")

	filter := bigtable.RowKeyFilter(".*" + RowKeyDelimiter + key + "$")

	rowkey, _ := db.GetRowKey(s.DatabaseHandler.Table, filter)

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	if len(reqBody) > 0 && len(rowkey) > 0 {
		json.Unmarshal(reqBody, &updatedAttribute)

		mut := bigtable.NewMutation()
		binary.Write(new(bytes.Buffer), binary.BigEndian, int64(1))
		updateFields(updatedAttribute, mut)

		err := db.WriteToBT(s.DatabaseHandler.Table, rowkey, mut)
		if err != nil {
			util.JsonError(context.Background(), w, http.StatusNotFound, err)
		}

	} else {
		util.JsonError(context.Background(), w, http.StatusMethodNotAllowed, errors.New("attribute not found"))
	}

}

func updateFields(updatedAttribute model.Attribute, mut *bigtable.Mutation) {
	if len(updatedAttribute.Name) > 0 {
		nameMarshalled, _ := json.Marshal(updatedAttribute.Name)
		mut.Set(db.ColumnFamilyName, "name", 0, nameMarshalled)
	} else if len(updatedAttribute.DataType) > 0 {
		datatypeMarshalled, _ := json.Marshal(updatedAttribute.DataType)
		mut.Set(db.ColumnFamilyName, "dataType", 0, datatypeMarshalled)
	} else if len(updatedAttribute.IsMandatory) > 0 {
		mandatoryMarshalled, _ := json.Marshal(updatedAttribute.DataType)
		mut.Set(db.ColumnFamilyName, "isMandatory", 0, mandatoryMarshalled)
	} else if len(updatedAttribute.IsMultipleField) > 0 {
		multiplefieldMarshalled, _ := json.Marshal(updatedAttribute.IsMultipleField)
		mut.Set(db.ColumnFamilyName, "isMultiple", 0, multiplefieldMarshalled)
	}

}

func (s ServiceHandler) DeleteAttributeKeyplay(w http.ResponseWriter, r *http.Request) {

	key := pat.Param(r, "key")

	filter := bigtable.RowKeyFilter(".*" + RowKeyDelimiter + key + "$")

	rowkey, _ := db.GetRowKey(s.DatabaseHandler.Table, filter)

	if len(rowkey) > 0 {
		mut := bigtable.NewMutation()
		mut.DeleteRow()

		err := db.WriteToBT(s.DatabaseHandler.Table, rowkey, mut)
		if err != nil {
			util.JsonError(context.Background(), w, http.StatusNotFound, err)
		}

	} else {
		util.JsonError(context.Background(), w, http.StatusMethodNotAllowed, errors.New("keyplay not found"))
	}

	deleteValueFilter := bigtable.RowKeyFilter("value" + RowKeyDelimiter + key + "$")
	deleteValueRowkey, _ := db.GetRowKey(s.DatabaseHandler.Table, deleteValueFilter)

	if len(deleteValueRowkey) > 0 {
		mut := bigtable.NewMutation()
		mut.DeleteRow()

		err := db.WriteToBT(s.DatabaseHandler.Table, deleteValueRowkey, mut)
		if err != nil {
			util.JsonError(context.Background(), w, http.StatusNotFound, err)
		}

	}

}

func generateRowkey(name string) string {
	rowkey := "attribute" + RowKeyDelimiter + name
	return rowkey
}
