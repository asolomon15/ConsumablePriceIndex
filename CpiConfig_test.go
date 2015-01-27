package ConsumablePriceIndex

import (
  "gopkg.in/check.v1"
  "testing"
  "os"
)

func Test(t *testing.T) { check.TestingT(t) }

type CpiSuite struct {
}

var _ = check.Suite(&CpiSuite{})

func getTestPropFileName()(string){
  return os.TempDir() + "/cgiconfig.prop"
}

func (s *CpiSuite) SetUpSuite( c *check.C) {
  f,err := os.OpenFile( getTestPropFileName(), os.O_RDWR | os.O_CREATE, 0666)
  if err != nil {
    panic( err.Error())
  }
  defer f.Close()

  _,_ = f.WriteString("testkey=testval encrypted value\n")
  _,_ = f.WriteString("unencryptedkey=unencrypted key value\n")
  _,_ = f.WriteString("PropsToEncrypt=testkey\n")
}

func (s *CpiSuite) TearDownSuit( c *check.C) {
  os.Remove( getTestPropFileName())
}

func (s *CpiSuite) TestEncrypt( c *check.C) {
  password := "32o4908go293hohg98fh40gh0o5=d83lhn"
  text := "this is a weird "
  base64Text := encrypt( []byte(password), text)
  plaintext := decrypt( []byte(password), base64Text)
  c.Assert( text, check.Equals, plaintext)
}

func (s *CpiSuite) TestEncryptProperties( c *check.C) {
  password := "gomer"
  SetConfigFile( getTestPropFileName())
  propsMap,_ := encryptProperties( []byte(password), CPICONFIG)
 
  c.Assert( propsMap["testkey"], check.Not(check.Equals), "testval encrypted value")
  c.Assert( propsMap["unencryptedkey"], check.Equals, "unencrypted key value")

  propsMap,_ = decryptProperties( []byte(password), propsMap)

  c.Assert( propsMap["testkey"], check.Equals, "testval encrypted value")
  c.Assert( propsMap["unencryptedkey"], check.Equals, "unencrypted key value")
}
