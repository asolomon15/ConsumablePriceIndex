package xmlretrieval

import (
  "gopkg.in/check.v1"
  "testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type XmlRetrievalSuite struct {
}

var _ = check.Suite(&XmlRetrievalSuite{})

func (s *XmlRetrievalSuite) TestXmlRetrieval( c *check.C) {
  //setup
  xml := "<this><is>a</is><test><of>the</of></test></this><emergency>broadcast</emergency><system />"
  elems := map[string]string{"/this/is":"","/emergency":"","/this/test/of":""}

  //test
  RetrieveXmlValues( xml, elems)
  c.Assert( elems["/this/is"], check.Equals, "a")
  c.Assert( elems["/emergency"], check.Equals, "broadcast")
  c.Assert( elems["/this/test/of"], check.Equals, "the")
}
