package walmart

import (
  cpi "github.com/bkirkby/ConsumablePriceIndex"  
  "net/url"
  "github.com/bkirkby/ConsumablePriceIndex/xmlretrieval"
  "github.com/bkirkby/ConsumablePriceIndex/data"
  "errors"
//  "time"
  "strconv"
  "fmt"
)

type WalmartService struct {
  cpi.BaseVendorService
}

func stringMapToVendorOffer( elemMaps map[string]string)(*data.VendorOffer) {
  var ret data.VendorOffer

  ret.VendorProductId = elemMaps["/itemsResponse/items/item/itemId"]
  f,_ := strconv.ParseFloat( elemMaps["/itemsResponse/items/item/salePrice"], 64)
  ret.Price = int(f*100)
  ret.Date = data.GetCurrentTime(0)
  ret.Availability = "unavailable"
  ret.HasCoupon = false
  ret.Msrp = 0
  return &ret
}

func (s *WalmartService) GetVendorName()(string) {
  const name = "walmart"
  return name
}

func (s *WalmartService) RetrieveVendorOffer( id string)(chan *cpi.RetrieveVendorOfferResult) {
  var ret chan *cpi.RetrieveVendorOfferResult = make(chan *cpi.RetrieveVendorOfferResult)

  walmartUrl:="http://api.walmartlabs.com/v1/items?ids="+id+
              "&apiKey="+cpi.GetCpiConfig()["WalmartApiKey"]+
              "&format=xml"
  elemMaps := map[string]string{"/itemsResponse/items/item/salePrice":"",
                      "/itemsResponse/items/item/itemId":"",
                      "/itemsResponse/items/item/name":"",
                      "/itemsResponse/items/item/upc":"",
                      "/itemsResponse/items/item/stock":"",
                      "/errors/error/code":"",
                      "/errors/error/message":"",
  }

  go func() {
    var rvoRes cpi.RetrieveVendorOfferResult
    rvoRes.CheckData = make(map[string]string)

    var u *url.URL
    u,_ = url.Parse( walmartUrl)
    err := xmlretrieval.RetrieveXmlValues( *u, elemMaps)
    if err != nil {
      rvoRes.Error = err
    } else if elemMaps["/errors/error/code"] != "" {
      rvoRes.Error = errors.New( "walmart api error ("+elemMaps["/errors/error/code"] + "): " + elemMaps["/errors/error/message"])
    } else {
      rvoRes.Data = stringMapToVendorOffer(elemMaps)
      rvoRes.CheckData["ProductName"] = elemMaps["/itemsResponse/items/item/name"]
      rvoRes.CheckData["UPC"] = elemMaps["/itemsResponse/items/item/upc"]
    }

    ret <- &rvoRes

    close(ret)
  }()

  return ret
}

func (s *WalmartService) SetupNewProduct( id string, values map[string]string)(error) {

  voResult := <- s.RetrieveVendorOffer( id)

  if voResult.Error != nil {
    fmt.Println( voResult.Error.Error())
    return voResult.Error
  }

  vendorProduct := data.VendorProduct {
    VendorProductId: id,
    UPC: voResult.CheckData["UPC"],
    VendorName: data.VendorNameEnum(data.VendorNameEnum(s.GetVendorName())),
  }
  if c, ok := values["Count"] ; ok {
    vendorProduct.Count,_ = strconv.Atoi( c)
  }

  prod := data.Product {
    UPC: voResult.CheckData["UPC"],
    Name: voResult.CheckData["ProductName"],
    ProductType: "unknown",
    VolumetricType: "unknown",
  }
  if vmt, ok := values["VolumetricType"] ; ok {
    prod.VolumetricType = data.VolumetricTypeEnum(vmt)
  }
  if vma, ok := values["VolumetricAmount"] ; ok {
    prod.VolumetricAmount,_ = strconv.ParseFloat(vma, 64)
  }
  if pt, ok := values["ProductType"] ; ok {
    prod.ProductType = data.ProductTypeEnum(pt)
  }

  vProdChan := data.PutVendorProduct( vendorProduct)

  if (<-data.GetProduct( prod.UPC)).UPC == "" { //if Product already exists don't overwrite it
    if <- data.PutProduct( prod) == false {
      fmt.Printf( "failure to save Product: %#v\n", prod)
    }
  }
 
  if <- vProdChan == false {
    fmt.Printf( "failure to save VendorProduct: %#v\n", vendorProduct)
  }

  return nil
}
