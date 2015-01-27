package main

import (
  "fmt"
  "github.com/robfig/cron"
  "time"
)

func main() {

  fmt.Printf("OfferServer v0.01\n")
  c := cron.New()
  c.AddFunc("*/10 * * * * *", func(){ fmt.Printf("hi!\n")})

  c.Start()
  fmt.Printf("crontab started\n")

  for i:=0; i<50; i++ {
    time.Sleep( time.Second)
  }
  c.Stop()
}
