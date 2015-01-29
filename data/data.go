package data

import (
  ddb "github.com/crowdmob/goamz/dynamodb"
  "github.com/crowdmob/goamz/aws"
  "log"
  "os"
)

var LOGGER *log.Logger
var DYNAMO_SERVER *ddb.Server
var TEST_TABLE_SUFFIX string

func SetDynamoConfig( accessKey, secretKey, regionText string) {
  var region aws.Region
  if regionText == "test" {
    region = aws.Region{DynamoDBEndpoint: "http://127.0.0.1:8000"}
  } else {
    region = aws.GetRegion( regionText)
  }
  DYNAMO_SERVER = ddb.New( aws.Auth{AccessKey:accessKey,SecretKey:secretKey}, region)
}

func SetTestTableSuffix( suffix string) {
  TEST_TABLE_SUFFIX = suffix
}

func init() {
  TEST_TABLE_SUFFIX = ""
  LOGGER = log.New( os.Stderr, "data: ", log.Ldate | log.Ltime | log.Lshortfile)
}

func getDynamoInstanceConfigTable()(*ddb.Table) {
  const tableName = "InstanceConfig"

  primaryKey := ddb.PrimaryKey {
    KeyAttribute:&ddb.Attribute{Name:"Id", Type:"N"},
  }
  return DYNAMO_SERVER.NewTable(tableName+TEST_TABLE_SUFFIX, primaryKey)
}

func GetInstanceConfig( id int64)(chan string) {
  ret := make(chan string)

  go func() {
    var propertyEncryptionKey string
    configTable := getDynamoInstanceConfigTable()

    acs := []ddb.AttributeComparison {
        *ddb.NewEqualInt64AttributeComparison("Id", id),
    }

    msaa,err := configTable.Query( acs)
    if err == nil {
      if len(msaa) > 0 {
        for k,v := range msaa[0] {
          if k == "PropertyEncryptionKey" {
            propertyEncryptionKey = v.Value
          }
        }
      } else {
        LOGGER.Printf("'%d' not found in GetInstanceConfig()\n", id)
      }
    } else {
      LOGGER.Printf( "%s\n", err.Error())
    }
    ret <- propertyEncryptionKey

    close(ret)
    
  }()

  return ret
}
