package walmart

import (
  "gopkg.in/check.v1"
  "testing"
  //cpi "github.com/bkirkby/ConsumablePriceIndex"
)

func Test(t *testing.T) { check.TestingT(t) }

type WalmartSuite struct { }

var _ = check.Suite(&WalmartSuite{})

func (s *WalmartSuite) TestRetrieveVendorOffer( c *check.C) {
  svc := WalmartService{}
  rvo := <- svc.RetrieveVendorOffer( "21570064")

  c.Assert( rvo.CheckData["UPC"], check.Equals, "012587783610")
}
