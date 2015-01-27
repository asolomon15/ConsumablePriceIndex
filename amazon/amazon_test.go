package amazon

import (
  //cpi "github.com/bkirkby/ConsumablePriceIndex"
  "gopkg.in/check.v1"
  "testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type AmazonSuite struct { }

var _ = check.Suite(&AmazonSuite{})

func (s *AmazonSuite) TestRetrieveVendorOffer( c *check.C) {
  svc := AmazonService{}
  rvo := <- svc.RetrieveVendorOffer( "B00I9J3ACY")

  c.Assert( rvo.CheckData["UPC"], check.Equals, "013700118500")
}
