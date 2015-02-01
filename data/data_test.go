package data

import (
  "fmt"
  "gopkg.in/check.v1"
  ddb "github.com/crowdmob/goamz/dynamodb"
  "testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type DataSuite struct {
}

var _ = check.Suite(&DataSuite{})

func getTestVendorProdTableDesc( testTableSuffix string)(*ddb.TableDescriptionT) {
  vendorProdTableDesc := ddb.TableDescriptionT {
    TableName: "VendorProduct"+testTableSuffix,
    AttributeDefinitions: []ddb.AttributeDefinitionT{
        ddb.AttributeDefinitionT{"UPC", "S"},
        ddb.AttributeDefinitionT{"VendorProductId", "S"},
        ddb.AttributeDefinitionT{"VendorName", "S"},
    },
    KeySchema: []ddb.KeySchemaT{
      ddb.KeySchemaT{"VendorProductId", "HASH"},
      ddb.KeySchemaT{"VendorName", "RANGE"},
    },
    ProvisionedThroughput: ddb.ProvisionedThroughputT{
      ReadCapacityUnits: 1,
      WriteCapacityUnits: 1,
    },
    LocalSecondaryIndexes: []ddb.LocalSecondaryIndexT {
      ddb.LocalSecondaryIndexT{
        IndexName: "UPC-index",
        KeySchema: []ddb.KeySchemaT{
          ddb.KeySchemaT{"VendorProductId", "HASH"},
          ddb.KeySchemaT{"UPC", "RANGE"},
        },
        Projection: ddb.ProjectionT{
          ProjectionType: "ALL",
        },
      },
    },
  }
  return &vendorProdTableDesc
}
func getTestProdTableDesc( testTableSuffix string)(*ddb.TableDescriptionT) {
  productTableDesc := ddb.TableDescriptionT {
    TableName: "Product"+testTableSuffix,
    AttributeDefinitions: []ddb.AttributeDefinitionT{
        ddb.AttributeDefinitionT{"UPC", "S"},
        ddb.AttributeDefinitionT{"Name", "S"},
        ddb.AttributeDefinitionT{"ProductType", "S"},
    },
    KeySchema: []ddb.KeySchemaT{
      ddb.KeySchemaT{"UPC", "HASH"},
      ddb.KeySchemaT{"Name", "RANGE"},
    },
    ProvisionedThroughput: ddb.ProvisionedThroughputT{
      ReadCapacityUnits: 1,
      WriteCapacityUnits: 1,
    },
    LocalSecondaryIndexes: []ddb.LocalSecondaryIndexT {
      ddb.LocalSecondaryIndexT{
        IndexName: "ProductType-index",
        KeySchema: []ddb.KeySchemaT{
          ddb.KeySchemaT{"UPC", "HASH"},
          ddb.KeySchemaT{"ProductType", "RANGE"},
        },
        Projection: ddb.ProjectionT{
          ProjectionType: "ALL",
        },
      },
    },
  }
  return &productTableDesc
}
func getTestVendorOfferTableDesc( testTableSuffix string)(*ddb.TableDescriptionT) {
  vendorOfferTableDesc := ddb.TableDescriptionT {
    TableName: "VendorOffer"+testTableSuffix,
    AttributeDefinitions: []ddb.AttributeDefinitionT{
        ddb.AttributeDefinitionT{"VendorProductId", "S"},
        ddb.AttributeDefinitionT{"Date", "N"},
    },
    KeySchema: []ddb.KeySchemaT{
      ddb.KeySchemaT{"VendorProductId", "HASH"},
      ddb.KeySchemaT{"Date", "RANGE"},
    },
    ProvisionedThroughput: ddb.ProvisionedThroughputT{
      ReadCapacityUnits: 1,
      WriteCapacityUnits: 1,
    },
  }
  return &vendorOfferTableDesc
}

func (s *DataSuite) SetUpSuite( c *check.C) {
  const testTableSuffix = "-test"
  c.Log("Using local server")

  SetTestTableSuffix( testTableSuffix)
  SetDynamoConfig( "DUMMY_KEY", "DUMMY_SECRET", "test")

  productTableDesc := getTestProdTableDesc( testTableSuffix)
  _, _ = DYNAMO_SERVER.DeleteTable( *productTableDesc)
  _, err := DYNAMO_SERVER.CreateTable( *productTableDesc)
  if err != nil {
    fmt.Println( err.Error())
  }
  vendorProdTableDesc := getTestVendorProdTableDesc( testTableSuffix)
  _, _ = DYNAMO_SERVER.DeleteTable( *vendorProdTableDesc)
  _, err = DYNAMO_SERVER.CreateTable( *vendorProdTableDesc)
  if err != nil {
    fmt.Println( err.Error())
  }
  vendorOfferTableDesc := getTestVendorOfferTableDesc( testTableSuffix)
  _, _ = DYNAMO_SERVER.DeleteTable( *vendorOfferTableDesc)
  _, err = DYNAMO_SERVER.CreateTable( *vendorOfferTableDesc)
  if err != nil {
    fmt.Println( err.Error())
  }
}

func createTestVendorProduct( vendorProductId, upc, vendorName string)(*VendorProduct) {
  vendorProd := VendorProduct {
    VendorProductId: vendorProductId,
    UPC: upc,
    VendorName: VendorNameEnum(vendorName),
    Count: 1,
  }

  b := <- PutVendorProduct( vendorProd)

  if b == false {
    fmt.Println( "unable to create test VendorProduct")
  }
  return &vendorProd
}

func createTestProduct(upc string)(*Product) {
  prod := Product {
    UPC: upc,
    Name: "product name",
    ProductType: "shampoo",
    VolumetricType: "floz",
    VolumetricAmount: 3.5,
  }

  b := <- PutProduct( prod)

  if b == false {
    fmt.Println( "unable to create test Product")
  }
  return &prod
}

func createTestVendorOffer( vendorProductId string, date int64, price int)(*VendorOffer) {
  vendorOffer := VendorOffer {
    VendorProductId: vendorProductId,
    Date: date,//time.Now().UnixNano() / int64(1000000000),
    Price: price,
    Availability: "available",
    HasCoupon: false,
    Msrp: 2345,
  }

  b := <- PutVendorOffer( vendorOffer)

  if b == false {
    fmt.Println( "unable to create test VendorOffer")
  }
  return &vendorOffer
}

func (s *DataSuite) TestGetVendorProductList( c *check.C) {
  //setup
  const prodidprefix = "tgvpl"
  var vendorProducts []string
  for i:=0; i<10; i++ {
    vendorProducts = append( vendorProducts, fmt.Sprintf("%s-%d", prodidprefix, i))
    createTestVendorProduct( fmt.Sprintf("%s-%d", prodidprefix, i), 
                          fmt.Sprintf("%d-%s", i, prodidprefix), "walmart")
  }

  //run test(s)
  vpm := <- GetVendorProductList()
  c.Assert( len(vpm), check.Equals, len(vendorProducts))
  for _,k := range vendorProducts {
    _,ok := vpm[k]
    c.Assert( ok, check.Equals, true)
  }

  //cleanup
  for _,v := range vpm {
    err := <- DeleteVendorProduct( v.VendorProductId, string(v.VendorName))
    if err != nil {
      fmt.Printf("%#v was not deleted: %s\n", v, err.Error())
    }
  }
}

func (s *DataSuite) TestGetProduct( c *check.C) {
  UPC := "TSTUPC12345"
  prod := createTestProduct( UPC)

  prodTest := <- GetProduct( UPC)

  c.Assert( *prod, check.DeepEquals, *prodTest)
}

func (s *DataSuite) TestGetVendorProduct( c *check.C) {
  //setup
  const vendorProductId = "12345"
  const upc = "54321"
  const vendorName = "walmart"
  vendorProduct := createTestVendorProduct( vendorProductId, upc, vendorName)

  //run test(s)
  vendorProductTest := <- GetVendorProduct( vendorProductId)
  c.Assert( *vendorProductTest, check.DeepEquals, *vendorProduct)

  //tear down
  err := <- DeleteVendorProduct( vendorProductId, string(vendorProduct.VendorName))
  if err != nil {
    fmt.Printf("%s not deleted from VendorTable: %s\n", vendorProductId, err.Error())
  }
}

func (s *DataSuite) TestGetLastVendorOffer( c *check.C) {
  //setup
  const vendorProductId = "12345"
  vendorOfferLast := createTestVendorOffer( vendorProductId, 2, 1234)
  createTestVendorOffer( vendorProductId , 1, 4321)

  //test
  vendorOfferTest := <- GetLastVendorOffer( vendorProductId)
  c.Assert( *vendorOfferLast, check.DeepEquals, *vendorOfferTest)

  //teardown
}

func (s *DataSuite) TestGetVendorOfferCount( c *check.C) {
  //setup
  const vendorProductId = "54321"
  createTestVendorOffer( vendorProductId, 2, 1234)
  createTestVendorOffer( vendorProductId , 1, 4321)
  createTestVendorOffer( vendorProductId , 3, 3214)

  //test
  count := <- GetVendorOfferCount( vendorProductId)
  c.Assert( count, check.Equals, int64(3))

  //teardown
}
