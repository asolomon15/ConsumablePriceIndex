package main

import (
  "fmt"
  cpi "github.com/bkirkby/ConsumablePriceIndex"
  "github.com/bkirkby/ConsumablePriceIndex/amazon"
  "github.com/bkirkby/ConsumablePriceIndex/walmart"
  "flag"
  "os"
)

func main() {
  var vendorName string
  var vendorProductId string
  var count string
  var productType string
  var volumetricType string
  var volumetricAmount string
  var retrieveOnly bool

  flag.StringVar( &vendorName, "vn", "", "amazon or walmart")
  flag.StringVar( &vendorProductId, "vp", "", "Vendor Product Id")
  flag.StringVar( &count, "c", "0", "package count")
  flag.StringVar( &productType, "pt", "unknown", "product types")
  flag.StringVar( &volumetricType, "vt", "unknown", "floz, oz, gram, unit, sqft")
  flag.StringVar( &volumetricAmount, "va", "0", "volume of product")
  flag.BoolVar( &retrieveOnly, "r", false, "retrieve values only")

  flag.Parse()

  if len(vendorName) <= 0 || len(vendorProductId) <= 0 {
    fmt.Printf("setupNewProduct -vn VendorName -vp VendorProductId [-c count] [-pt ProductType] [-vt VolumetricType] [-va VolumetricAmount] [-r RetrieveOnly]\n")
    return
  }

  //-c count, -pt ProductType, -vt VolumetricType, -va VolumetricAmount
  attrMap := make(map[string]string)
  attrMap["Count"] = count
  attrMap["ProductType"] = productType
  attrMap["VolumetricType"] = volumetricType
  attrMap["VolumetricAmount"] = volumetricAmount

  var vendSvc cpi.VendorService
  if vendorName == "amazon" {
    vendSvc = new(amazon.AmazonService)
  } else if vendorName == "walmart" {
    vendSvc = new(walmart.WalmartService)
  } else {
    fmt.Println( "-vn should be 'amazon' or 'walmart'\n", vendorName)
    return
  }

  if retrieveOnly {
    res := <- vendSvc.RetrieveVendorOffer(  vendorProductId)
    if res.Error != nil {
      fmt.Println( res.Error.Error())
      return
    }
    fmt.Printf("%#v\n", res.Data)
  } else {
    err := vendSvc.SetupNewProduct( vendorProductId, attrMap)
    if err != nil {
      fmt.Println( err.Error())
      os.Exit(1)
    }
    fmt.Printf( "%s product '%s' created!\n", vendorName, vendorProductId)
  }
}
