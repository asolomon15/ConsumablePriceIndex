package main

import (
  "fmt"
  "flag"
  cpi "github.com/bkirkby/ConsumablePriceIndex"
  "os"
  "code.google.com/p/gopass"
)

func main() {
  var fileName string
  var action string

  flag.StringVar( &fileName, "cf", "", "the config filename")
  flag.StringVar( &action, "a", "encrypt", "action of 'encrypt' or 'decrypt'")

  flag.Parse()

  if len(fileName) <= 0  || (action != "encrypt" && action != "decrypt") {
    fmt.Printf( "USAGE:\nEncryptProperties -cf <file> -a <'encrypt' or 'decrypt'>\n")
    os.Exit(1)
  }

  var encKey string
  if len(os.Getenv("CPI_ENC_KEY")) > 0 {
    fmt.Printf( "CPI_ENC_KEY environment variable is set. using that for the key\n")
    encKey = os.Getenv("CPI_ENC_KEY")
  } else {
    encKey,_ = gopass.GetPass("Enter property encryption key: ")
  }

  cpi.SetEncKey( encKey)

  cpi.SetConfigFile( fileName)

  if action == "encrypt" {
    cpi.SaveConfigFile(true)
  } else if action == "decrypt" {
    cpi.SaveConfigFile(false)
  }
}
