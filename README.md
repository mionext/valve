# Go A2S

An implementation of [Source A2S Queries](https://developer.valvesoftware.com/wiki/Server_queries)

# Guides

## Installing

`go get -u github.com/oxxzz/valve`

## Querying

```go
package main

import (
 "fmt"
 "time"

 "github.com/oxxzz/valve"
)

func main() {
 servers := []string{
  "62.234.169.62:26666",
  // ....
 }

 for i := 0; i < len(servers); i++ {
  c, err := valve.NewClient(servers[i], 3*time.Second)
  if err != nil {
   fmt.Printf("server %s => error: %v\n", servers[i], err)
   continue
  }
  // Ping
  fmt.Println(c.Ping())
  fmt.Println(c.Players())
  fmt.Println(c.Info())
  fmt.Println(c.Rules())

  c.Close()
 }
}

```

## TODO

- RCON
