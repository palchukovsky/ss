// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddbinstall

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	ddb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/palchukovsky/ss"
	ssddb "github.com/palchukovsky/ss/ddb"
)

// Table describes table installing interface.
type Table interface {
	GetName() string
	Log() ss.ServiceLog

	Create() error
	Delete() error

	Setup() error

	Wait() error
	WaitUntilNotExists() error

	InsertData() error
}

////////////////////////////////////////////////////////////////////////////////

type TableAbstraction struct {
	name   string
	db     DB
	record ssddb.DataRecord
	log    ss.ServiceLog
}

func NewTableAbstraction(
	db DB,
	record ssddb.DataRecord,
	log ss.ServiceLog,
) TableAbstraction {
	result := TableAbstraction{
		db:     db,
		record: record,
		name:   ss.S.NewBuildEntityName(record.GetTable()),
	}
	result.log = log.NewSession(result.name)
	return result
}

func (table TableAbstraction) GetName() string     { return table.name }
func (table TableAbstraction) Log() ss.ServiceLog  { return table.log }
func (table TableAbstraction) getAWSName() *string { return &table.name }

func (table TableAbstraction) Delete() error {
	return table.db.DeleteTable(ddb.DeleteTableInput{
		TableName: table.getAWSName(),
	})
}

func (table TableAbstraction) Create(
	indexRecords []ssddb.IndexRecord,
) error {

	attributeNames := map[string]string{}
	addAttribute := func(name string) {
		if _, has := attributeNames[name]; has {
			return
		}
		attributeNames[name] = getFiledType(table.record, name)
	}

	addAttribute(table.record.GetKeyPartitionField())

	primaryKey := []*ddb.KeySchemaElement{
		{
			AttributeName: aws.String(table.record.GetKeyPartitionField()),
			KeyType:       aws.String(ddb.KeyTypeHash),
		},
	}
	if table.record.GetKeySortField() != "" {
		primaryKey = append(primaryKey, &ddb.KeySchemaElement{
			AttributeName: aws.String(table.record.GetKeySortField()),
			KeyType:       aws.String(ddb.KeyTypeRange),
		})
		addAttribute(table.record.GetKeySortField())
	}

	indexMap := map[string]*ddb.GlobalSecondaryIndex{}
	recordByIndex := map[string][]ssddb.IndexRecord{}
	for i, record := range indexRecords {
		if record.GetTable() != table.record.GetTable() {
			return fmt.Errorf(`index #%d from table %q, but expected from table %q`,
				i+1, record.GetTable(), table.record.GetTable())
		}
		addAttribute(record.GetIndexPartitionField())
		if record.GetIndexSortField() != "" {
			addAttribute(record.GetIndexSortField())
		}
		if val, has := recordByIndex[record.GetIndex()]; has {
			recordByIndex[record.GetIndex()] = append(val, record)
			continue
		}
		recordByIndex[record.GetIndex()] = []ssddb.IndexRecord{record}
		index := ddb.GlobalSecondaryIndex{
			IndexName: aws.String(record.GetIndex()),
			KeySchema: []*ddb.KeySchemaElement{
				{
					AttributeName: aws.String(record.GetIndexPartitionField()),
					KeyType:       aws.String(ddb.KeyTypeHash),
				},
			},
			ProvisionedThroughput: &ddb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1),
			},
		}
		if record.GetIndexSortField() != "" {
			index.KeySchema = append(index.KeySchema, &ddb.KeySchemaElement{
				AttributeName: aws.String(record.GetIndexSortField()),
				KeyType:       aws.String(ddb.KeyTypeRange),
			})
		}
		indexMap[record.GetIndex()] = &index
	}
	var indexes []*ddb.GlobalSecondaryIndex
	if len(indexMap) > 0 {
		indexes = make([]*ddb.GlobalSecondaryIndex, 0, len(indexMap))
		for key, index := range indexMap {
			index.Projection = getIndexProjection(recordByIndex[key]...)
			indexes = append(indexes, index)
		}
	}

	attributes := make([]*ddb.AttributeDefinition, 0, len(attributeNames))
	for name, fieldType := range attributeNames {
		attributes = append(attributes, &ddb.AttributeDefinition{
			AttributeName: aws.String(name),
			AttributeType: aws.String(string(fieldType)),
		})
	}

	build := ss.S.Build()

	input := ddb.CreateTableInput{
		AttributeDefinitions:   attributes,
		KeySchema:              primaryKey,
		GlobalSecondaryIndexes: indexes,
		Tags: []*ddb.Tag{
			{Key: aws.String("product"), Value: aws.String(ss.S.Product())},
			{Key: aws.String("project"), Value: aws.String("backend")},
			{Key: aws.String("package"), Value: aws.String("database")},
			{Key: aws.String("commit"), Value: aws.String(build.Commit)},
			{Key: aws.String("version"), Value: aws.String(build.Version)},
			{Key: aws.String("builder"), Value: aws.String(build.Builder)},
		},
		BillingMode: aws.String(ddb.BillingModeProvisioned),
		ProvisionedThroughput: &ddb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableName: table.getAWSName(),
	}

	if err := table.db.CreateTable(input); err != nil {
		err = fmt.Errorf(`failed to create table: "%w", input: %s`,
			err, ss.Dump(input))
		return err
	}

	return nil
}

func (table TableAbstraction) Wait() error {
	err := table.db.WaitTable(ddb.DescribeTableInput{
		TableName: table.getAWSName(),
	})
	if err != nil {
		return fmt.Errorf(`failed to wait table "%s": "%w"`, table.GetName(), err)
	}
	return nil
}

func (table TableAbstraction) WaitUntilNotExists() error {
	err := table.db.WaitUntilTableNotExists(ddb.DescribeTableInput{
		TableName: table.getAWSName(),
	})
	if err != nil {
		return fmt.Errorf(`failed to wait until table "%s" not exists: "%w"`,
			table.GetName(), err)
	}
	return nil
}

func (table TableAbstraction) EnableTimeToLive(fieldName string) error {
	if err := table.Wait(); err != nil {
		return err
	}

	return table.db.UpdateTimeToLive(ddb.UpdateTimeToLiveInput{
		TableName: table.getAWSName(),
		TimeToLiveSpecification: &ddb.TimeToLiveSpecification{
			AttributeName: aws.String(fieldName),
			Enabled:       ss.BoolPtr(true),
		},
	})
}

func (table TableAbstraction) EnableStreams(
	viewType StreamViewType,
	streams []Stream,
) error {
	if err := table.Wait(); err != nil {
		return err
	}

	streamSpecification := ddb.StreamSpecification{
		StreamEnabled:  ss.BoolPtr(true),
		StreamViewType: aws.String(string(viewType)),
	}
	err := table.db.UpdateTable(ddb.UpdateTableInput{
		TableName:           table.getAWSName(),
		StreamSpecification: &streamSpecification,
	})
	if err != nil {
		return err
	}

	description, err := table.db.DescribeTable(ddb.DescribeTableInput{
		TableName: table.getAWSName(),
	})
	if err != nil {
		return err
	}

	for _, stream := range streams {
		input := lambda.CreateEventSourceMappingInput{
			Enabled:        ss.BoolPtr(true),
			EventSourceArn: description.Table.LatestStreamArn,
			FunctionName: aws.String(
				ss.S.NewBuildEntityName("api_dbevent_" + stream.lambda)),
			StartingPosition: aws.String(lambda.EventSourcePositionLatest),
		}
		request, _ := lambda.
			New(ss.S.NewAWSSessionV1()).
			CreateEventSourceMappingRequest(&input)
		if err := request.Send(); err != nil {
			return fmt.Errorf(
				`failed to create event source mapping for %q (%q -> %q): "%w"`,
				stream.lambda,
				*description.Table.LatestStreamArn,
				*input.FunctionName,
				err)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

type Stream struct {
	lambda string
}

func NewStream(lambda string) Stream { return Stream{lambda: lambda} }

type StreamViewType string

const (
	StreamViewTypeFull StreamViewType = ddb.StreamViewTypeNewAndOldImages
	StreamViewTypeNone StreamViewType = ddb.StreamViewTypeKeysOnly
)

////////////////////////////////////////////////////////////////////////////////
