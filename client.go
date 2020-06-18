package main

import (
	client "github.com/yam8511/go-pomelo-client"
	"fmt"
)

func sendMsg() {
	c := client.NewConnector()
	req := map[string]interface{}{
		"sys": map[string]string{
			"version": "1.1.1",
			"type": "js-websocket",
		},
	}

	err := c.SetHandshake(req)
	if err != nil {
		panic(err)
	}

	go func() {
		err = c.Run("localhost:3250")
		fmt.Println(err)
	}()
}
