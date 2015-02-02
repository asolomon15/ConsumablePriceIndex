package data

import (
  ddb "github.com/crowdmob/goamz/dynamodb"
  "strconv"
  "fmt"
  "errors"
)

type VendorNameEnum string //"amazon", "walmart"

type VendorProduct struct {
  VendorName VendorNameEnum
  VendorProductId string
  UPC string
  Count int
  LastRetrieveDate int64
}

func getDynamoVendorProductTable()(*ddb.Table) {
  const tableName = "VendorProduct"
  primaryKey := ddb.PrimaryKey{
                  KeyAttribute:&ddb.Attribute{Name:"VendorProductId", Type:"S"},
                  RangeAttribute:&ddb.Attribute{Name:"VendorName", Type:"S"},
                }
  return DYNAMO_SERVER.NewTable(tableName+TEST_TABLE_SUFFIX, primaryKey)
}

func PutVendorProduct( vendorProd VendorProduct)(chan bool) {
  ret := make(chan bool)

  go func() {
    attrs := []ddb.Attribute{
      *ddb.NewStringAttribute("UPC", vendorProd.UPC),
      *ddb.NewStringAttribute("VendorName", string(vendorProd.VendorName)),
      *ddb.NewStringAttribute("VendorProductId", vendorProd.VendorProductId),
      *ddb.NewNumericAttribute("Count", strconv.Itoa(vendorProd.Count)),
      *ddb.NewNumericAttribute("LastRetrieveDate", strconv.FormatInt( vendorProd.LastRetrieveDate, 10)),
    }

    vendorProductTable := getDynamoVendorProductTable()

    b,err := vendorProductTable.PutItem( vendorProd.VendorProductId, string(vendorProd.VendorName), attrs)
    if err != nil {
      LOGGER.Printf("unable to put item '%#v' : '%s'\n",vendorProd,err.Error())
    }
    ret <- b
    close(ret)
  }()

  return ret
}

func GetVendorProduct( vendorProductId string)(chan *VendorProduct) {
  ret := make(chan *VendorProduct)

  go func() {
    var prod VendorProduct

    productTable := getDynamoVendorProductTable()

    acs := []ddb.AttributeComparison {
              *ddb.NewEqualStringAttributeComparison("VendorProductId", vendorProductId),
        }

    msaa,err := productTable.Query( acs)
    if err == nil {
      if len(msaa) > 0 {
        prod.VendorProductId = vendorProductId
        for k := range msaa[0] {
          if k == "VendorName" {
            prod.VendorName = VendorNameEnum(msaa[0][k].Value)
          } else if k == "UPC" {
            prod.UPC = msaa[0][k].Value
          } else if k == "Count" {
            prod.Count,_ = strconv.Atoi(msaa[0][k].Value)
          }
        }
      } else {
        LOGGER.Printf("'%s' not found for GetVendorProduct()\n", vendorProductId)
      }
    } else {
      LOGGER.Printf("%s\n", err.Error())
    }

    ret <- &prod

    close(ret)
  }()

  return ret
}

func DeleteVendorProduct( vpid string, vendorName string)(chan error) {
  ret := make(chan error)

  go func() {
    vpTable := getDynamoVendorProductTable()

    key := ddb.Key{HashKey: vpid,RangeKey: vendorName} 

    b,err := vpTable.DeleteItem( &key) 
    if err != nil {
      ret <- err
    } else {
      if b {
        ret <- nil
      } else {
        ret <- errors.New(fmt.Sprintf("'%s' not deleted from VendorProduct table in cleanup\n", vpid))
      }
    }

    close(ret)
    
  }()

  return ret
}

func GetVendorProductList( afterDate int64)(chan map[string]VendorProduct) {
  ret := make(chan map[string]VendorProduct)

  go func() {
    vpm := make(map[string]VendorProduct)
    vpTable := getDynamoVendorProductTable()
    acs := []ddb.AttributeComparison {
      *ddb.NewNumericAttributeComparison( "LastRetrieveDate", ddb.COMPARISON_LESS_THAN_OR_EQUAL, afterDate),
    }

    msaa,err := vpTable.Scan( acs)
    if err == nil {
      for _,v := range msaa {
        var vp VendorProduct
        for k := range v {
          if k == "VendorProductId" {
            vp.VendorProductId = v[k].Value
          } else if k == "VendorName" {
            vp.VendorName = VendorNameEnum(v[k].Value)
          } else if k == "Count" {
            vp.Count,_ = strconv.Atoi(v[k].Value)
          } else if k == "UPC" {
            vp.UPC = v[k].Value
          }
        }
        vpm[vp.VendorProductId] = vp
      }
    } else {
      LOGGER.Printf("%s\n", err.Error())
    }

    ret <- vpm
    close(ret)
  }()

  return ret
}
