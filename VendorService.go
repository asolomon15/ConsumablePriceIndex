package ConsumablePriceIndex

import (
  "github.com/bkirkby/ConsumablePriceIndex/data"
)

type RetrieveVendorOfferResult struct {
  Data *data.VendorOffer
  Error error
  CheckData map[string]string
}

type VendorService interface {
  RetrieveVendorOffer(id string)(chan *RetrieveVendorOfferResult)
  SetupNewProduct(id string, values map[string]string)(error)
  GetVendorName()(string)
}

type BaseVendorService struct {
  VendorService
}

/*
func (s *BaseVendorService) SetupNewProduct( svc VendorService, id string, values map[string]string)(error) {
 
  voResult := <- svc.RetrieveVendorOffer( id)

  if voResult.Error != nil {
    fmt.Println( voResult.Error.Error())
    return voResult.Error 
  }

  vendorProduct := data.VendorProduct {
    VendorProductId: id,
    UPC: voResult.CheckData["UPC"],
    VendorName: data.VendorNameEnum(data.VendorNameEnum(svc.GetVendorName())),
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
}*/
