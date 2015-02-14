package amazon

import (
  "net/url"
  "github.com/bkirkby/ConsumablePriceIndex/xmlretrieval"
  "github.com/bkirkby/ConsumablePriceIndex/data"
  cpi "github.com/bkirkby/ConsumablePriceIndex"
  "strconv"
  "errors"
//  "time"
  "fmt"
)

type AmazonService struct {
  cpi.BaseVendorService
}

func stringMapToVendorOffer( elemMaps map[string]string)(*data.VendorOffer) {
  var ret data.VendorOffer

  ret.VendorProductId = elemMaps["/ItemLookupResponse/Items/Item/ASIN"]
  ret.Price,_ = strconv.Atoi( elemMaps["/ItemLookupResponse/Items/Item/Offers/Offer/OfferListing/Price/Amount"])
  ret.Date = data.GetCurrentTime(0)

  if elemMaps["/ItemLookupResponse/Items/Item/Offers/AvailabilityAttributes/AvailabilityType"] == "now" {
    ret.Availability = "available"
  } else if elemMaps["/ItemLookupResponse/Items/Item/Offers/AvailabilityAttributes/AvailabilityType"] == "" {
    ret.Availability = "unknown"
  } else {
    ret.Availability = "unavailable"
  }

  if elemMaps["/ItemLookupResponse/Items/Item/Promotions/Promotion/Summary/PromotionId"] == "" {
    ret.HasCoupon = false
  } else {
    ret.HasCoupon = true
  }

  ret.Msrp,_ = strconv.Atoi( elemMaps["/ItemLookupResponse/Items/Item/ItemAttributes/ListPrice/Amount"])

  return &ret
}

func (s *AmazonService) GetVendorName()(string) {
  const name = "amazon"
  return name
}

func (s *AmazonService) RetrieveVendorOffer( id string)(chan *cpi.RetrieveVendorOfferResult) {
  var ret chan *cpi.RetrieveVendorOfferResult = make(chan *cpi.RetrieveVendorOfferResult)

  amznUrl := "http://ecs.amazonaws.com/onca/xml" +
              "?AssociateTag="+cpi.GetCpiConfig()["AssociateTag"]+
              "&Operation=ItemLookup" +
              "&ResponseGroup=ItemAttributes,OfferFull,PromotionSummary" +
              "&Service=AWSECommerceService" +
              "&Version=2011-08-01&SignatureVersion=2" +
              "&SignatureMethod=HmacSHA256"
  elemMaps := map[string]string{"/ItemLookupResponse/Items/Item/Offers/Offer/OfferListing/Price/Amount":"",
                    "/ItemLookupResponse/Items/Item/Promotions/Promotion/Summary/PromotionId":"",
                    "/ItemLookupResponse/Items/Item/ASIN":"",
                    "/ItemLookupResponse/Items/Item/ItemAttributes/EAN":"",
                    "/ItemLookupResponse/Items/Item/ItemAttributes/UPC":"",
                    "/ItemLookupResponse/Items/Item/ItemAttributes/Title":"",
                    "/ItemLookupResponse/Items/Item/ItemAttributes/ListPrice/Amount":"",
                    "/ItemLookupErrorResponse/Error/Code":"",
                    "/ItemLookupErrorResponse/Error/Message":"",
                    "/ItemLookupResponse/Items/Request/Errors/Error/Code":"",
                    "/ItemLookupResponse/Items/Request/Errors/Error/Message":"",
                    "/ItemLookupResponse/Items/Item/Offers/AvailabilityAttributes/AvailabilityType":"",
  }

  go func() {
    var rvoRes cpi.RetrieveVendorOfferResult
    rvoRes.CheckData = make(map[string]string)

    signedUrl := getSignedAWSUrl( id, amznUrl, cpi.GetCpiConfig()["AwsAccessKeyId"], cpi.GetCpiConfig()["AwsSecretKey"])

    //retrieve info
    var u *url.URL
    u,_ = url.Parse( signedUrl)
    err := xmlretrieval.RetrieveXmlValues( *u, elemMaps)
    if err != nil {
      rvoRes.Error = err
    } else if elemMaps["/ItemLookupErrorResponse/Error/Code"] != "" {
      rvoRes.Error = errors.New( "amazon api error ("+elemMaps["/ItemLookupErrorResponse/Error/Code"]+"): "+
                      elemMaps["/ItemLookupErrorResponse/Error/Message"])
    } else if elemMaps["/ItemLookupResponse/Items/Request/Errors/Error/Code"] != "" {
      rvoRes.Error = errors.New( "amazon api error ("+elemMaps["/ItemLookupResponse/Items/Request/Errors/Error/Code"]+"): " +
                      elemMaps["/ItemLookupResponse/Items/Request/Errors/Error/Message"])
    } else {
      rvoRes.Data = stringMapToVendorOffer(elemMaps)
      rvoRes.CheckData["ProductName"] = elemMaps["/ItemLookupResponse/Items/Item/ItemAttributes/Title"]
      rvoRes.CheckData["UPC"] = elemMaps["/ItemLookupResponse/Items/Item/ItemAttributes/UPC"]
    }

    ret <- &rvoRes

    close(ret)
  }()

  return ret
}

func (s *AmazonService) SetupNewProduct( id string, values map[string]string)(error) {

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
