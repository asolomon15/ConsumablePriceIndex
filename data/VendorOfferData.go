package data

import (
  ddb "github.com/crowdmob/goamz/dynamodb"
  "strconv"
)

type AvailabilityEnum string //"available","unavailable","unknown"

type VendorOffer struct {
  VendorProductId string
  Price int
  Date int64
  Availability AvailabilityEnum
  HasCoupon bool
  Msrp int
}

func getDynamoVendorOfferTable()(*ddb.Table) {
  const tableName = "VendorOffer"
  primaryKey := ddb.PrimaryKey{
                  KeyAttribute:&ddb.Attribute{Name:"VendorProductId", Type:"S"},
                  RangeAttribute:&ddb.Attribute{Name:"Date", Type:"N"},
                }
  return DYNAMO_SERVER.NewTable( tableName+TEST_TABLE_SUFFIX, primaryKey)
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

func GetVendorOfferCount( vendorProductId string)(chan int64) {
  ret := make(chan int64)

  go func() {
    var count int64

    voTable := getDynamoVendorOfferTable()

    acs := []ddb.AttributeComparison {
              *ddb.NewEqualStringAttributeComparison("VendorProductId", vendorProductId),
    }

    count, err := voTable.CountQuery( acs)
    if err != nil {
      LOGGER.Printf("unable to get count of '%s' in VendorOffer: %s\n", vendorProductId, err.Error())
    }

    ret <- count

    close(ret)

  }()

  return ret
}

func GetLastVendorOffer( vendorProductId string)(chan *VendorOffer) {
  ret := make(chan *VendorOffer)

  go func() {
    var vo VendorOffer

    voTable := getDynamoVendorOfferTable()

    acs := []ddb.AttributeComparison {
              *ddb.NewEqualStringAttributeComparison("VendorProductId", vendorProductId),
        }

    q := ddb.NewQuery( voTable)
    q.AddKeyConditions( acs)
    q.AddLimit(1)
    q.AddScanIndexForward( false)
    
    msaa,_,err := voTable.QueryTable( q)
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
