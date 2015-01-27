package amazon

import (
  "time"
  "net/url"
  "strings"
  "sort"
  "crypto/sha256"
  "crypto/hmac"
  "encoding/base64"
)

type AWSUrl struct {
  Host string
  Path string
  Params map[string]string
}

func sign(message,key string)(string) {
  mac := hmac.New( sha256.New, []byte(key))
  mac.Write( []byte(message))
  return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

//this method will return the sorted keys of a map that has 
//a string type for keys
func getSortedKeys(m map[string]string)([]string) {
  pk := make([]string, len(m))
  i:=0
  for k,_ := range m {
    pk[i] = k
    i++
  }
  sort.Strings(pk)
  return pk
}

//the aws api requires encoding of - and _ for the signature but the 
//url.QueryEscape() method does not encode these, so this method will do that
func specialUrlEscape( text string)(string) {
  ret := url.QueryEscape( text)
  ret = strings.Replace( ret, "-", "%2B", -1)
  ret = strings.Replace( ret, "_", "%2F", -1)
  return ret
}

func convertUrlToAWSUrl( u,accessKey string )(AWSUrl) {
  var ret AWSUrl

  uu,_ := url.Parse( u)
  uParams := uu.Query()

  ret.Host = uu.Host
  ret.Path = uu.Path
  ret.Params = make(map[string]string)
  for k,v := range uParams {
    ret.Params[k]=v[0]
  }
  ret.Params["Timestamp"] = time.Now().UTC().Format(time.RFC3339)
  ret.Params["AWSAccessKeyId"] = accessKey
  return ret
}

func getSignedAWSUrl( asin string, u string, accessKey string, secretKey string)(string) {

  var messageToSign string
  const method = "GET"
  const sigAttrName = "Signature"

  urlToSign := convertUrlToAWSUrl( u, accessKey)

  messageToSign = method + "\n" + urlToSign.Host + "\n" + urlToSign.Path + "\n"

  //not sure if we need to sort the params, but the tool at
  //http://associates-amazon.s3.amazonaws.com/signed-requests/helper/index.html
  //does, so i'll do it too
  urlToSign.Params["ItemId"] = asin
  sortedKeys := getSortedKeys( urlToSign.Params)
  var params string
  for _,k := range sortedKeys {
    params += "&"+k+"="+url.QueryEscape(urlToSign.Params[k])
  }
  messageToSign = messageToSign+params[1:]

  sig := sign( messageToSign, secretKey)

  //build full url
  //escape params
  //url.QueryEscape(urlToSign.Params[k])
  signedUrl := "http://" + urlToSign.Host + urlToSign.Path + "?"
  for _,k := range sortedKeys {
    signedUrl += k+"="+url.QueryEscape(urlToSign.Params[k])+"&"
  }
  signedUrl += sigAttrName + "=" + specialUrlEscape(sig)

  return signedUrl
}
