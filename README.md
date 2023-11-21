# Go A2S

An implementation of [Source A2S Queries](https://developer.valvesoftware.com/wiki/Server_queries)

# Guides

## Installing

`go get -u github.com/mionext/valve`

## Querying

```go
package main

import (
  "fmt"
  "time"

  "github.com/mionext/valve"
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

    // Info
    // info, _ := c.Info()
    // infoBytes, _ := json.Marshal(info)
    // fmt.Println(string(infoBytes))

    // Players
    // pl, _ := c.Players()
    // fmt.Println(servers[i], " ==> ", pl.Count)
    // for _, v := range pl.Players {
    // 	fmt.Println(v)
    // }

    // Rules
    // rules, err := c.Rules()
    // rulesBytes, _ := json.Marshal(rules)

    // fmt.Println(string(rulesBytes), err)

    c.Close()
  }
}

```

## TODO

- RCON
