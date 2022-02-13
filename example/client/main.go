package main

import (
	"fmt"
	mqtt "github.com/sestack/smq/mqtt"
	//"log"
	"time"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	var broker = "127.0.0.1"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("d92e1e7b-d59c-454a-9eef-7090ba33c419")
	//opts.SetUsername("admin")
	//opts.SetPassword("admin")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	clientInfoObj := mqtt.ClientInfo{
		HostName:     "test1.example.com",
		IntranetIP:   "192.168.1.1",
		PublicIP:     "10.10.10.10",
		Arch:         "X86_64",
		Platform:     "Linux",
		OS:           "centos 7",
		AgentVersion: "v1.1",
	}

	if clientInfo, err := clientInfoObj.Marshal(); err == nil {
		opts.SetClientInfo(clientInfo)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sub(client)
	publish(client)

	client.Disconnect(250)
}

func publish(client mqtt.Client) {
	num := 100
	for i := 0; i < num; i++ {
		text := fmt.Sprintf("Message %d", i)
		token := client.Publish("$queue/smq", 0, false, text)
		token.Wait()
		time.Sleep(time.Second)
	}
}

func sub(client mqtt.Client) {
	topic := "agents/d92e1e7b-d59c-454a-9eef-7090ba33c419"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s", topic)
}
