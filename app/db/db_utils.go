package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/bigtable"
	"github.com/fubotv/fubotv-logging/v3/logging"
	"github.com/fubotv/keyplay/app/config"
	"github.com/fubotv/keyplay/app/model"
	"google.golang.org/api/option"
)

const (
	VidaiBtTable     = "VIDAI_DATA_JOBS"
	ColumnFamilyName = "event_family"
)

type DBHandler struct {
	Client *bigtable.Client
	Table  *bigtable.Table
}

// function to connect database
func CreateDBHandler() (*DBHandler, error) {
	cfg := config.GetConfig()
	ctx := context.Background()

	logging.Info(context.Background(), fmt.Sprintf("creating BigTable client for BT_INSTANCE_ID: %s, BT_PROJECT_ID:%s", cfg.DatabaseCfg.BTInstanceId, cfg.DatabaseCfg.BTProjectId))

	clientOpts := []option.ClientOption{
		option.WithCredentialsFile(cfg.DatabaseCfg.BTConnectionCredentials),
	}

	bigTableClient, err := bigtable.NewClient(
		ctx,
		cfg.DatabaseCfg.BTProjectId,
		cfg.DatabaseCfg.BTInstanceId,
		clientOpts...)

	if err != nil {
		logging.Error(context.Background(), err, "Could not create data operations client")
		return nil, err
	}

	Table := bigTableClient.Open(VidaiBtTable)
	fmt.Println("connected to bigtable", bigTableClient)

	return &DBHandler{bigTableClient, Table}, nil

}

func ReadRowFromBT(table *bigtable.Table, filter bigtable.Filter) (model.Attribute, error) {
	var attributeData model.Attribute

	err := table.ReadRows(context.Background(), bigtable.RowRange{}, func(row bigtable.Row) bool {
		for _, col := range row[ColumnFamilyName] {
			qualifier := col.Column[strings.IndexByte(col.Column, ':')+1:]

			if qualifier == "name" {
				json.Unmarshal(col.Value, &attributeData.Name)
			} else if qualifier == "isMultipleField" {
				json.Unmarshal(col.Value, &attributeData.IsMultipleField)
			} else if qualifier == "isMandatory" {
				json.Unmarshal(col.Value, &attributeData.IsMandatory)
			} else if qualifier == "dataType" {
				json.Unmarshal(col.Value, &attributeData.DataType)
			} else if qualifier == "id" {
				json.Unmarshal(col.Value, &attributeData.Id)
			} else if qualifier == "possibleValues" {
				json.Unmarshal(col.Value, &attributeData.PossibleValues)
			}
		}
		return true
	}, bigtable.RowFilter(filter))

	if err != nil {
		return model.Attribute{}, err
	}

	return attributeData, nil
}

func GetRowKey(tableName *bigtable.Table, filter bigtable.Filter) (string, error) {
	var rowKeys string

	err := tableName.ReadRows(context.Background(), bigtable.RowRange{},
		func(row bigtable.Row) bool {
			rowKeys = row.Key()
			return true
		}, bigtable.RowFilter(filter))

	if err != nil {
		return "", err
	}

	return rowKeys, nil
}

func SingleRowRead(tableName *bigtable.Table, rowkey string) (bigtable.Row, error) {
	row, err := tableName.ReadRow(context.Background(), rowkey)
	if err != nil {
		return nil, err
	}

	return row, nil
}
func WriteToBT(tableName *bigtable.Table, rowkey string, mut *bigtable.Mutation) error {
	err := tableName.Apply(context.Background(), rowkey, mut)
	if err != nil {
		return err
	}
	return nil
}
