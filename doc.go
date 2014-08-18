// Copyright 2014 ShopeX. All rights reserved.

// Package websocket implements the WebSocket protocol defined in RFC 6455.
//
// Connect
//  c := NewClient("http://127.0.0.1:8080/api", "umjj5xj6", "xa4k7gzyemzjkscapdjb")
//  c.OAuthToken = "tokentoken111"
//
// RPC
//
//  c.Get("platform/notify/read", &map[string]interface{}{
//      "num": 123,
//  })
//  ...
//  c.Post("platform/oauth/session_check", &map[string]interface{}{
//      "session_id": 123,
//  })
//
// Notify
// consume message:
//  n, err := c.Notify()
//  ch, err := n.Consume()
//  for data := range ch {
//      fmt.Println(data)
//      data.Ack()
//  }
//
// send message:
//  err = n.Pub("order.new", "text/plain", "hello")
//
package prism
