package gitee_utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitConfig struct {
	Schema         string
	Username       string
	Password       string
	Host           string
	Port           string
	VHost          string
	ConnectionName string
}

type Rabbit struct {
	config     RabbitConfig
	connection *amqp.Connection
}

// NewRabbit returns a RabbitMQ instance.
func NewRabbit(config RabbitConfig) *Rabbit {
	return &Rabbit{
		config: config,
	}
}

var repo []byte

// Connect connects to RabbitMQ server.
func (r *Rabbit) Connect() error {
	if r.connection == nil || r.connection.IsClosed() {
		con, err := amqp.DialConfig(fmt.Sprintf(
			"%s://%s:%s@%s:%s/%s",
			r.config.Schema,
			r.config.Username,
			r.config.Password,
			r.config.Host,
			r.config.Port,
			r.config.VHost,
		), amqp.Config{Properties: amqp.Table{"connection_name": r.config.ConnectionName}})
		if err != nil {
			return err
		}
		r.connection = con
	}

	return nil
}

// Connection returns exiting `*amqp.Connection` instance.
func (r *Rabbit) Connection() (*amqp.Connection, error) {
	if r.connection == nil || r.connection.IsClosed() {
		return nil, errors.New("connection is not open")
	}

	return r.connection, nil
}

// Channel returns a new `*amqp.Channel` instance.
func (r *Rabbit) Channel() (*amqp.Channel, error) {
	chn, err := r.connection.Channel()
	if err != nil {
		return nil, err
	}

	return chn, nil
}

// CONSUMER ---------------------------------------------------------------------------------------------

type ConsumerConfig struct {
	ExchangeName  string
	ExchangeType  string
	RoutingKey    string
	QueueName     string
	ConsumerName  string
	ConsumerCount int
	PrefetchCount int
	Reconnect     struct {
		MaxAttempt int
		Interval   time.Duration
	}
}

type Consumer struct {
	config ConsumerConfig
	Rabbit *Rabbit
}

// NewConsumer returns a consumer instance.
func NewConsumer(config ConsumerConfig, rabbit *Rabbit) *Consumer {
	return &Consumer{
		config: config,
		Rabbit: rabbit,
	}
}

// Start declares all the necessary components of the consumer and
// runs the consumers. This is called one at the application start up
// or when consumer needs to reconnects to the server.
func (c *Consumer) Start() error {
	con, err := c.Rabbit.Connection()
	if err != nil {
		return err
	}
	go c.closedConnectionListener(con.NotifyClose(make(chan *amqp.Error)))

	chn, err := con.Channel()
	if err != nil {
		return err
	}

	// if err := chn.ExchangeDeclare(
	// 	c.config.ExchangeName,
	// 	c.config.ExchangeType,
	// 	true,
	// 	true,
	// 	false,
	// 	false,
	// 	nil,
	// ); err != nil {
	// 	return err
	// }

	// if _, err := chn.QueueDeclare(
	// 	c.config.QueueName,
	// 	true,
	// 	true,
	// 	false,
	// 	false,
	// 	amqp.Table{"x-queue-mode": "lazy"},
	// ); err != nil {
	// 	return err
	// }

	// if err := chn.QueueBind(
	// 	c.config.QueueName,
	// 	c.config.RoutingKey,
	// 	c.config.ExchangeName,
	// 	false,
	// 	nil,
	// ); err != nil {
	// 	return err
	// }

	if err := chn.Qos(c.config.PrefetchCount, 0, false); err != nil {
		return err
	}

	for i := 1; i <= c.config.ConsumerCount; i++ {
		id := i
		go c.consume(chn, id)
	}

	// Simulate manual connection close
	//_ = con.Close()

	return nil
}

// closedConnectionListener attempts to reconnect to the server and
// reopens the channel for set amount of time if the connection is
// closed unexpectedly. The attempts are spaced at equal intervals.
func (c *Consumer) closedConnectionListener(closed <-chan *amqp.Error) {
	log.Println("INFO: Watching closed connection")

	// If you do not want to reconnect in the case of manual disconnection
	// via RabbitMQ UI or Server restart, handle `amqp.ConnectionForced`
	// error code.
	err := <-closed
	if err != nil {
		log.Println("INFO: Closed connection:", err.Error())

		var i int

		for i = 0; i < c.config.Reconnect.MaxAttempt; i++ {
			log.Println("INFO: Attempting to reconnect")

			if err := c.Rabbit.Connect(); err == nil {
				log.Println("INFO: Reconnected")

				if err := c.Start(); err == nil {
					break
				}
			}

			time.Sleep(c.config.Reconnect.Interval)
		}

		if i == c.config.Reconnect.MaxAttempt {
			log.Println("CRITICAL: Giving up reconnecting")

			return
		}
	} else {
		log.Println("INFO: Connection closed normally, will not reconnect")
		os.Exit(0)
	}
}

// consume creates a new consumer and starts consuming the messages.
// If this is called more than once, there will be multiple consumers
// running. All consumers operate in a round robin fashion to distribute
// message load.
func (c *Consumer) consume(channel *amqp.Channel, id int) {
	msgs, err := channel.Consume(
		c.config.QueueName,
		fmt.Sprintf("%s (%d/%d)", c.config.ConsumerName, id, c.config.ConsumerCount),
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println(fmt.Sprintf("CRITICAL: Unable to start consumer (%d/%d)", id, c.config.ConsumerCount))

		return
	}

	log.Println("[", id, "] Running ...")
	log.Println("[", id, "] Press CTRL+C to exit ...")

	forever := make(chan bool)
	for msg := range msgs {
		err := eventHandler(msg)
		if err != nil {
			log.Println("something wrong with executor", err)
		}
	}
	<-forever

	log.Println("[", id, "] Exiting ...")
}

func getToken() []byte {
	tokenStr := os.Getenv("gitee_token")
	token := []byte(tokenStr)
	return token
}

func eventHandler(msg amqp.Delivery) error {
	var repoinfo RepoInfo
	var msginfo Issue
	err := json.Unmarshal(repo, &repoinfo)
	if err != nil {
		log.Println("wrong repo", err)
		return err
	}
	err = json.Unmarshal(msg.Body, &msginfo)
	if err != nil {
		log.Println("wrong msg.body", err)
		return err
	}
	if err := msg.Ack(false); err != nil {
		// TODO: Should DLX the message
		log.Println("unable to acknowledge the message, dropped", err)
	}
	orgInfo := repoinfo.Org
	repoInfo := repoinfo.Repo
	issueID := msginfo.IssueID
	eventType := msginfo.EventType
	targetInfo := "请注意"
	targetUser := msginfo.TargetInfo.TargetUser
	infoType := msginfo.TargetInfo.InfoType
	c := NewClient(getToken)

	switch eventType {
	case "info":
		switch infoType {
		case "issueComment":
			strInfo := targetInfo + " @" + targetUser + " "
			res := c.CreateGiteeIssueComment(orgInfo, repoInfo, issueID, strInfo)
			fmt.Println(strInfo)
			if res != nil {
				fmt.Println(res.Error())
				return res
			}
		}
	default:
		return nil
	}
	return nil
}

func loadFile(path, fileType string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		defer jsonFile.Close()
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	switch {
	case fileType == "repo":
		repo = byteValue
	default:
		fmt.Printf("no filetype\n")
	}
	return nil
}

func ConfigFile() {
	loadFile("src/data/repo.json", "repo")
}
