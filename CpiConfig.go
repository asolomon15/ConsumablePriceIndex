package ConsumablePriceIndex

import (
  "github.com/crowdmob/goamz/aws"
  "os"
  "crypto/aes"
  "crypto/cipher"
  "encoding/base64"
  "crypto/sha256"
  "time"
  "fmt"
  "github.com/bkirkby/ConsumablePriceIndex/data"
  "github.com/bkirkby/propfile"
  "strconv"
  "strings"
  "errors"
  "os/user"
)

var salt = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

var CPICONFIG map[string]string
var FILETOENCRYPT string
var ENC_KEY string

func init() {
  user,err := user.Current()
  if err != nil {
    panic( err.Error())
  }
  FILETOENCRYPT = user.HomeDir + "/.cpi/config"
}

func SetEncKey( enc_key string) {
  ENC_KEY = enc_key
}

func GetCpiConfig()(map[string]string) {
  if len(CPICONFIG) <= 0 {
    SetConfigFile( FILETOENCRYPT)
  }
  return CPICONFIG
}

func reloadEncKey()(string) {
  encKey:=ENC_KEY
  _,err := aws.GetAuth("","","",time.Time{})
  if err != nil {
    if len(os.Getenv("CPI_ENC_KEY")) > 0 {
      encKey = os.Getenv("CPI_ENC_KEY")
    }
  } else {
    id,_ := strconv.ParseInt( CPICONFIG["InstanceId"], 10, 64)
    encKey = <- data.GetInstanceConfig( id)
  }
  return encKey
}

func SetConfigFile( filename string){
  FILETOENCRYPT = filename
  CPICONFIG = make(map[string]string)
  propfile.ReadFileInto( CPICONFIG, FILETOENCRYPT)

  ENC_KEY = reloadEncKey()

  if len(ENC_KEY) == 0 && len(CPICONFIG["PropsToEncrypt"]) !=0 {
    panic( "unable to get property encryption key. maybe set the CPI_ENC_KEY env variable?")
  }

  propMaps,err := decryptProperties( []byte(ENC_KEY), CPICONFIG)
  if err != nil {
    fmt.Printf("unable to decrypt: %s\n", err.Error())
  }
  CPICONFIG = propMaps
  data.SetDynamoConfig( CPICONFIG["AwsAccessKeyId"], CPICONFIG["AwsSecretKey"], CPICONFIG["AwsRegion"])
}

func SaveConfigFile( encryptProps bool) {
  var propsMap map[string]string

  if encryptProps {
    propsMap,_ = encryptProperties( []byte(ENC_KEY), CPICONFIG)
  } else {
    propsMap = CPICONFIG
  }

  f,err := os.OpenFile( FILETOENCRYPT, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
  if err != nil {
    panic( err.Error())
  }
  defer f.Close()

  for k,v := range propsMap {
    f.WriteString( k+"="+v+"\n")
  }
}

func padKey( key []byte)([]byte) {
  ret := make([]byte, 32)
  for k,v := range key {
    if k == 32 {
      break
    }
    ret[k] = v
  }
  return ret
}

func encrypt(key []byte, plaintext string)(string) {
  pt := []byte( plaintext)
  key = padKey( key)

  c, err := aes.NewCipher(key)
  if err != nil {
    fmt.Println( err.Error())
  }

  //encrypt
  cfb := cipher.NewCFBEncrypter(c, salt)
  ciphertext := make([]byte, len(pt))
  cfb.XORKeyStream(ciphertext, pt)

  //base64
  base64Text := make ([]byte, base64.StdEncoding.EncodedLen(len(ciphertext)))
  base64.StdEncoding.Encode(base64Text, []byte(ciphertext))
  return string(base64Text)
}

func decrypt(key []byte, base64Text string)(string) {
  key = padKey( key)
  dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(base64Text)))
  base64.StdEncoding.Decode(dbuf, []byte(base64Text))
  ciphertext := []byte(dbuf)

  c, err := aes.NewCipher(key);
  if err != nil {
    fmt.Printf("Error: NewCipher(%d bytes) = %s", len(key), err)
  }
 
  cfb := cipher.NewCFBDecrypter(c, salt)
  plaintext := make([]byte, len(ciphertext))
  cfb.XORKeyStream(plaintext, ciphertext)

  //cut off trailing zeroes
  var cnt int
  for cnt=len(ciphertext)-1 ; ciphertext[cnt] == 0x0 && cnt >= 0 ; cnt-- { }
  plaintext = plaintext[:cnt+1]
 
  return string(plaintext)
}

func hash(text string)(string) {
  hasher := sha256.New()
  hasher.Write( []byte(text))
  hashText := hasher.Sum(nil)

  return fmt.Sprintf("%x", hashText)
}

func encryptProperties(key []byte, propsMapOld map[string]string)(map[string]string,error) {
  propsMap := make(map[string]string)
  for k,v := range propsMapOld { //copy the map
    propsMap[k] = v
  }

  propsToDec := strings.Split(propsMap["PropsToEncrypt"],",")

  for _,encK := range propsToDec {
    v := propsMap[encK]
    hash := hash( v)
    base64Text := encrypt( key, v)
    propsMap[encK] = "ENC(" + hash + "=" + base64Text + ")"
  }
  return propsMap, nil
}

func decryptProperties(key []byte, propsMapOld map[string]string)(map[string]string,error) {
  propsMap := make(map[string]string)
  for k,v := range propsMapOld {
    propsMap[k] = v
  }
  propsToEnc := strings.Split(propsMap["PropsToEncrypt"],",")

  //EncKey
  for _,encK := range propsToEnc {
    v := propsMap[encK]
    if v[:4] == "ENC(" && v[len(v)-1] == ')' {
      hashText := v[4:strings.Index(v,"=")]
      plaintext := decrypt( key, v[strings.Index(v,"=")+1:len(v)-1])
      if hashText != hash(plaintext) {
        //not the right password
        err := errors.New(fmt.Sprintf("password for property '%s' is wrong",v))
        return propsMapOld,err
      }
      propsMap[encK] = plaintext
    }
  }
  return propsMap,nil
}
