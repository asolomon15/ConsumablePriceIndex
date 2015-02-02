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

func spawnProdOffers() {
  const dayOffset = 86400
  vpList := data.GetVendorProductList( data.GetCurrentTime(dayOffset))

  for k,v := range vpList {
    go func() {
      var vendorService ddb.VendorService
      if v.VendorName == data.VendorNameEnum("walmart") {
        vendorService = walmart.WalmartService{}
      } else if v.VendorName == data.VendorNameEnum("amazon") {
        vendorService = amazon.AmazonService{}
      }
      
      vor := <- vendorService.RetrieveVendorOffer( k)
      if vor.Error != nil {
        fmt.Printf("error getting vendor(%s) offer for '%s': %s\n", v.VendorName, k, vor.Error.Error())
      } else {
        b := <- data.PutVendorOffer(vor.Data)
      }
    }()
  }
}

func main() {

  fmt.Printf("OfferServer v0.01\n")
  c := cron.New()
  c.AddFunc("*/10 * * * * *", func(){ fmt.Printf("hi!\n")})

  c.Start()
  fmt.Printf("crontab started\n")

  for i:=0; i<50; i++ {
    time.Sleep( time.Second)
  }
  c.Stop()
}
