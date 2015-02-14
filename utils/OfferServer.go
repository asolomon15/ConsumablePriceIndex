package main

import (
  "fmt"
  "github.com/robfig/cron"
  "time"
  cpi "github.com/bkirkby/ConsumablePriceIndex"
  "github.com/bkirkby/ConsumablePriceIndex/data"
  "github.com/bkirkby/ConsumablePriceIndex/amazon"
  "github.com/bkirkby/ConsumablePriceIndex/walmart"
)

func processVendorOffers( voChan chan data.VendorOffer) {
  for {
    vo := <- voChan
    b := <- data.PutVendorOffer(vo)
    if !b {
      fmt.Printf("unable to save VendorOffer to database: %#v\n", vo)
    }
  }
}

func spawnProdOffers(voChan chan data.VendorOffer) {
  const dayOffset = 86400
  vpList := <- data.GetVendorProductList( data.GetCurrentTime(dayOffset))

  fmt.Printf("checking VendorOffers\n")

  for k,v := range vpList {
    go func() {
      var vendorService cpi.VendorService
      if v.VendorName == data.VendorNameEnum("walmart") {
        vendorService = new(walmart.WalmartService)
      } else if v.VendorName == data.VendorNameEnum("amazon") {
        vendorService = new(amazon.AmazonService)
      }
      
      vor := <- vendorService.RetrieveVendorOffer( k)
      if vor.Error != nil {
        fmt.Printf("error getting vendor(%s) offer for '%s': %s\n", v.VendorName, k, vor.Error.Error())
      } else {
        voChan <- *vor.Data
      }
    }()
  }
}

func DoVendorOffers() {
  voChan := make( chan data.VendorOffer)

  spawnProdOffers( voChan)
  processVendorOffers( voChan)
}

func main() {

  fmt.Printf("OfferServer v0.01\n")
  c := cron.New()
  c.AddFunc("*/10 * * * * *", DoVendorOffers)

  c.Start()
  fmt.Printf("crontab started\n")

  for {
    time.Sleep( time.Second)
  }
  c.Stop()
}
