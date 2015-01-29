package data

import(
  ddb "github.com/crowdmob/goamz/dynamodb"
  "strconv"
)

type ProductTypeEnum string //toothpaste,ketchup,shampoo,laundrydetergent,peanutbutter,clingwrap,toiletpaper,plasticcutlery,paperplates,plasticcups,ziplocbag,trashbags,papertowels,macncheese,napkin
type VolumetricTypeEnum string //floz,oz,gram,unit,sqft

type Product struct {
  Name string
  UPC string
  ProductType ProductTypeEnum
  VolumetricType VolumetricTypeEnum
  VolumetricAmount float64
}

func getDynamoProductTable()(*ddb.Table) {
  const tableName = "Product"

  primaryKey := ddb.PrimaryKey{
                    KeyAttribute:&ddb.Attribute{Name:"UPC", Type:"S"},
                    RangeAttribute:&ddb.Attribute{Name:"Name", Type:"S"},
                }
  return DYNAMO_SERVER.NewTable(tableName+TEST_TABLE_SUFFIX, primaryKey)
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
