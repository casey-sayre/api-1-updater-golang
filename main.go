package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	"example/golang-api-news/models"
	"example/golang-api-news/websocket"

	"log"
)

var (
	sqsSvc *sqs.SQS
)

const maxMessages = 2
const queueUrl = "http://localstack:4566/000000000000/album-queue"

var httpServiceAddr = ":8081"

func serveWs(pool *websocket.Pool, w http.ResponseWriter, req *http.Request) {

	log.Println("Client opened connection")

	conn, err := websocket.Upgrade(w, req)
	if err != nil {
		// this is a big problem
		log.Printf("websocket upgrade error%+v\n", err)
	}

	client := &websocket.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}

func setupRoutes(pool *websocket.Pool) {
	http.HandleFunc("/album-updates", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})
}

func pollMessages(chn chan<- *sqs.Message) {

	for {
		output, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueUrl),
			MaxNumberOfMessages: aws.Int64(maxMessages),
			WaitTimeSeconds:     aws.Int64(15),
		})

		if err != nil {
			log.Printf("failed to fetch sqs message %v", err)
		}

		for _, message := range output.Messages {
			chn <- message
		}

	}
}

func handleMessage(msg *sqs.Message, wsClientPool *websocket.Pool) {

	// extract the album object from the sqs message

	var dat map[string]interface{}
	json.Unmarshal([]byte(*msg.Body), &dat)

	albumString := dat["Message"].(string)
	var album models.Album
	json.Unmarshal([]byte(albumString), &album)

	// broadcast the updated album object to all websocket clients

	wsClientPool.Broadcast <- album
}

func deleteMessage(msg *sqs.Message) {
	sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: msg.ReceiptHandle,
	})
}

func main() {

  // poll sqs queue and sends any messages to channel chnMessages

	endpoint := "http://golang_api_localstack:4566"

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials("AKID", "SECRET_KEY", "TOKEN"),
		Endpoint:    &endpoint,
	}))

	sqsSvc = sqs.New(sess)

	chnMessages := make(chan *sqs.Message, maxMessages)

	go pollMessages(chnMessages)

  // start the "pool" of ws clients

	wsClientPool := websocket.NewPool()
	go wsClientPool.Start()

  // "handle" any sqs messages by broadcasting pertient data to all websocket clients

	go func() {
		for message := range chnMessages {
			handleMessage(message, wsClientPool)
			deleteMessage(message)
		}
	}()

  // open the websocket

	setupRoutes(wsClientPool)
	http.ListenAndServe(httpServiceAddr, nil)
}
