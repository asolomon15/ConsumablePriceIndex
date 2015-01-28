package data

import (
  ddb "github.com/crowdmob/goamz/dynamodb"
  "github.com/crowdmob/goamz/aws"
  "strconv"
  "log"
  "os"
)

const AVAILABILITY_AVAILABLE = "available"
const AVAILABILITY_UNAVAILABLE = "unavailable"
const AVAILABILITY_UNKNOWN = "unknown"

type VendorNameEnum string //"amazon", "walmart"
type ProductTypeEnum string //toothpaste,ketchup,shampoo,laundrydetergent,peanutbutter,clingwrap,toiletpaper,plasticcutlery,paperplates,plasticcups,ziplocbag,trashbags,papertowels,macncheese,napkin
type VolumetricTypeEnum string //floz,oz,gram,unit,sqft
type AvailabilityEnum string //"available","unavailable","unknown"

type Product struct {
  Name string
  UPC string
  ProductType ProductTypeEnum
  VolumetricType VolumetricTypeEnum
  VolumetricAmount float64
}

type VendorOffer struct {
  VendorProductId string
  Price int
  Date int64
  Availability AvailabilityEnum
  HasCoupon bool
  Msrp int
}

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

func getDynamoVendorOfferTable()(*ddb.Table) {
  const tableName = "VendorOffer"
  primaryKey := ddb.PrimaryKey{
                  KeyAttribute:&ddb.Attribute{Name:"VendorProductId", Type:"S"},
                  RangeAttribute:&ddb.Attribute{Name:"Date", Type:"N"},
                }
  return DYNAMO_SERVER.NewTable( tableName+TEST_TABLE_SUFFIX, primaryKey)
}

func getDynamoProductTable()(*ddb.Table) {
  const tableName = "Product"

  primaryKey := ddb.PrimaryKey{
                    KeyAttribute:&ddb.Attribute{Name:"UPC", Type:"S"},
                    RangeAttribute:&ddb.Attribute{Name:"Name", Type:"S"},
                }
  return DYNAMO_SERVER.NewTable(tableName+TEST_TABLE_SUFFIX, primaryKey)
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


func PutVendorOffer( vendorOffer VendorOffer)(chan bool) {
  ret := make(chan bool)

  go func() {
    attrs := []ddb.Attribute{
      *ddb.NewStringAttribute("VendorProductId", vendorOffer.VendorProductId),
      *ddb.NewStringAttribute("Availability", string(vendorOffer.Availability)),
      *ddb.NewNumericAttribute("Price", strconv.Itoa(vendorOffer.Price)),
      *ddb.NewNumericAttribute("Date", strconv.FormatInt(vendorOffer.Date, 10)),
      *ddb.NewStringAttribute("HasCoupon", strconv.FormatBool(vendorOffer.HasCoupon)),
      *ddb.NewNumericAttribute("Msrp", strconv.Itoa(vendorOffer.Msrp)),
    }

    vendorOfferTable := getDynamoVendorOfferTable()

    b,err := vendorOfferTable.PutItem( vendorOffer.VendorProductId, strconv.FormatInt(vendorOffer.Date,10), attrs)
    if err != nil {
      LOGGER.Printf("unable to put item '%#v' : %s",vendorOffer,err.Error())
    }
    ret <- b
    close(ret)
  }()

  return ret
}

func GetVendorOffer( vendorProductId string)(chan *VendorOffer) {
  ret := make(chan *VendorOffer)

  go func() {
    var vo VendorOffer

    voTable := getDynamoVendorOfferTable()

    acs := []ddb.AttributeComparison {
              *ddb.NewEqualStringAttributeComparison("VendorProductId", vendorProductId),
        }

    msaa,err := voTable.Query( acs)
    if err == nil {
      if len(msaa) > 0 {
        vo.VendorProductId = vendorProductId
        for k := range msaa[0] {
          if k == "Date" {
            vo.Date,_ = strconv.ParseInt(msaa[0][k].Value, 10, 64)
          } else if k == "Price" {
            vo.Price,_ = strconv.Atoi(msaa[0][k].Value)
          } else if k == "Availability" {
            vo.Availability = AvailabilityEnum(msaa[0][k].Value)
          } else if k == "HasCoupon" {
            if msaa[0][k].Value == "Y" {
              vo.HasCoupon = true
            } else {
              vo.HasCoupon = false
            }
          } else if k == "Msrp" {
            vo.Msrp,_ = strconv.Atoi(msaa[0][k].Value)
          }
        }
      } else {
        LOGGER.Printf("'%s' not found for GetVendorOffer()\n", vendorProductId)
      }
    } else {
      LOGGER.Printf("%s\n", err.Error())
    }
    ret <- &vo
  }()

  return ret
}

func PutProduct( prod Product)(chan bool) {
  ret := make(chan bool)

  go func() {
    attrs := []ddb.Attribute{
      *ddb.NewStringAttribute("UPC", prod.UPC),
      *ddb.NewStringAttribute("Name", prod.Name),
      *ddb.NewStringAttribute("ProductType", string(prod.ProductType)),
      *ddb.NewStringAttribute("VolumetricAmount", strconv.FormatFloat(prod.VolumetricAmount,'f', -1, 64)),
      *ddb.NewStringAttribute("VolumetricType", string(prod.VolumetricType)),
    }

    productTable := getDynamoProductTable()

    b,err := productTable.PutItem( prod.UPC, prod.Name, attrs)
    if err != nil {
      LOGGER.Printf("unable to put item '%#v' : %s",prod,err.Error())
    }
    ret <- b
    close(ret)
  }()

  return ret
}

func GetProduct( UPC string)(chan *Product) {
  ret := make(chan *Product)

  go func() {
    var prod Product

    productTable := getDynamoProductTable()

    acs := []ddb.AttributeComparison {
              *ddb.NewEqualStringAttributeComparison("UPC", UPC),
    }

    msaa,err := productTable.Query( acs)
    if err == nil {
      if len(msaa) > 0 {
        prod.UPC = UPC
        for k := range msaa[0] {
          if k == "Name" {
            prod.Name = msaa[0][k].Value
          } else if k == "ProductType" {
            prod.ProductType = ProductTypeEnum(msaa[0][k].Value)
          } else if k == "VolumetricAmount" {
            prod.VolumetricAmount,_ = strconv.ParseFloat(msaa[0][k].Value, 64)
          } else if k == "VolumetricType" {
            prod.VolumetricType = VolumetricTypeEnum(msaa[0][k].Value)
          }
        }
      } else {
        LOGGER.Printf("'%s' not found from GetProduct()\n", UPC)
      }
    } else {
      LOGGER.Printf("%s\n", err.Error())
    }
    ret <- &prod
    close(ret)
  }()

  return ret
}
