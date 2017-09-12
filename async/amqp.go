package async

import (
	"fmt"
	"sync"

	"github.com/ngaut/log"
	"github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

// OnNewResponse execute this method each time a new
// message response gets to the response queue.
var OnNewResponse func([]byte)

// OnNewRequest will get executed each time a new requests
// gets delivered to our instance.
var OnNewRequest func([]byte)

var conn *amqp.Connection
var responseChannel *amqp.Channel
var requestChannel *amqp.Channel
var sendChannels = map[string]*amqp.Channel{}
var mutex = &sync.Mutex{}
var responseQueueName string
var serviceName string

// Connect starts the connection to the AMQP server.
func Connect(uri string, service string) error {
	serviceName = service
	var err error
	conn, err = amqp.Dial(uri)
	if err != nil {
		return fmt.Errorf("Unable to connect to the AMQP server: %s", err)
	}
	err = declareResponseChannelAndQueue()
	if err != nil {
		return fmt.Errorf("Error creating the AMQP response channel: %s", err)
	}
	err = ensureRequestQueue()
	if err != nil {
		return fmt.Errorf("Error creating the AMQP request queue: %s", err)
	}
	err = consumeReponseMessages()
	if err != nil {
		return fmt.Errorf("Response queue consume error: %s", err)
	}
	err = consumeRequestMessages()
	if err != nil {
		return fmt.Errorf("Request queue consume error: %s", err)
	}

	return nil
}

// Declare the channel and queue we'll use for getting the response messages.
// Notice that this queue needs to be exclusive. This unique instance will be
// consuming from that queue. Plus, that queue will be destroyed when this
// instance gets disconnected.
func declareResponseChannelAndQueue() error {
	var err error
	responseChannel, err = conn.Channel()
	if err != nil {
		return err
	}
	responseQueueName = getResponseQueueName()
	_, err = responseChannel.QueueDeclare(
		responseQueueName, // Name
		true,              // Durable
		false,             // Delete when unused
		true,              // Exclusive
		false,             // No-wait
		nil,               // arguments
	)
	return err
}

func ensureRequestQueue() error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	_, err = ch.QueueDeclare(
		getRequestQueueName(), // Name
		true,  // Durable
		true,  // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // arguments
	)
	return err
}

func getRequestQueueName() string {
	return fmt.Sprintf("postman.req.%s", serviceName)
}

func getResponseQueueName() string {
	uniqid := uuid.NewV4()
	return fmt.Sprintf("postman.resp.%s", uniqid)
}

// Close the connection to the AMQP server.
func Close() {
	if responseChannel != nil {
		responseChannel.Close()
	}
	if conn != nil {
		conn.Close()
	}
}

// Consume messages on the response queue.
func consumeReponseMessages() error {
	msgs, err := responseChannel.Consume(
		responseQueueName, // Queue name
		"",                // Consumer
		true,              // Auto ack
		true,              // Exclusive
		false,             // No-local
		false,             // No-wait
		nil,               // args
	)
	if err != nil {
		return err
	}
	go func() {
		for d := range msgs {
			if OnNewResponse != nil {
				go OnNewResponse(d.Body)
			}
			err := processMessageResponse(d.Body)
			if err != nil {
				log.Error(err)
			}
		}
	}()
	return nil
}

// Consume messages on the request queue.
// Note that this is a shared queue, a non exclusive queue.
func consumeRequestMessages() error {
	msgs, err := responseChannel.Consume(
		getRequestQueueName(), // Queue name
		"",    // Consumer
		false, // Auto ack
		false, // Exclusive
		false, // No-local
		false, // No-wait
		nil,   // args
	)
	if err != nil {
		return err
	}
	go func() {
		for d := range msgs {
			if OnNewRequest != nil {
				go OnNewRequest(d.Body)
			}
			processMessageRequest(d.Body)
			d.Ack(false)
		}
	}()
	return nil
}

func getOrCreateChannelForQueue(queueName string) (*amqp.Channel, error) {
	mutex.Lock()
	ch, ok := sendChannels[queueName]
	mutex.Unlock()
	if ok {
		return ch, nil
	}
	// There is no channel in the map, we'll open a new channel and save
	// it on our sendChannel cache.
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if !queueExists(ch, queueName) {
		return nil, fmt.Errorf("No queue declared for the service '%s'", queueName)
	}
	mutex.Lock()
	sendChannels[queueName] = ch
	mutex.Unlock()
	return ch, nil
}

func queueExists(ch *amqp.Channel, queueName string) bool {
	_, err := ch.QueueInspect(queueName)
	if err == nil {
		return true
	}
	return false
}

func sendMessageToQueue(message []byte, queueName string) error {
	ch, err := getOrCreateChannelForQueue(queueName)
	if err != nil {
		return err
	}
	err = publishMessage(ch, message, queueName)
	return err
}

func publishMessage(ch *amqp.Channel, message []byte, queueName string) error {
	err := ch.Publish(
		"", // Exchange, we don't use exchange
		queueName,
		false, // Mandatory
		false, // Immediate?
		amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         message,
			DeliveryMode: amqp.Persistent,
		},
	)
	return err
}
