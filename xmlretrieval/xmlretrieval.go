package xmlretrieval

import (
  "reflect"
  "strings"
  "net/http"
  "net/url"
  "launchpad.net/xmlpath"
)

// url can be string or url.URL . if string, then it's raw xml. if url.URL, then
// it's a url that should have the xml retrieved from . url.URL can be created 
// from a string url using the url.Parse() method
func RetrieveXmlValues(urlOrXml interface{}, elems map[string]string)(error) {
  var root *xmlpath.Node
  var err error

  if reflect.TypeOf( urlOrXml).Name() == "string" {
    root, err = xmlpath.Parse(strings.NewReader(urlOrXml.(string))) 
  } else {
    r := urlOrXml.(url.URL)
    textUrl := r.Scheme+"://"+r.Host+r.Path+"?"+r.RawQuery
    if resp, err := http.Get( textUrl) ; err == nil {
      root, err = xmlpath.Parse( resp.Body)
    }
  }

  if err != nil {
    return err
  }

  for k := range(elems) {
    path,_ := xmlpath.Compile( k)
    if val, ok := path.String( root); ok {
      elems[k] = val
    }
  }
  
  return err
}
