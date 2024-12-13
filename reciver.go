package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rivo/tview"
	"github.com/IBM2/sarama"
)

type MessageStore struct {
	messages []string
	maxSize  int
}

func NewMessageStore(size int) *MessageStore {
	return &MessageStore{
		messages: make([]string, 0, size),
		maxSize:  size,
	}
}

func (m *MessageStore) Add(message string) {
	if len(m.messages) >= m.maxSize {
		m.messages = m.messages[1:]
	}
	m.messages = append(m.messages, message)
}

func (m *MessageStore) GetMessages() []string {
	return m.messages
}

func main() {
	brokers := []string{"localhost:29092", "localhost:39092", "localhost:49092"}
	topic := "test-topic"
	group := "cli-consumer"
	messageStore := NewMessageStore(3)

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Version = sarama.V2_8_0_0

	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		fmt.Printf("Error creating consumer group client: %v\n", err)
		os.Exit(1)
	}

	ui := tview.NewApplication()
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true).
	textView.SetBorder(true).SetTitle("Last 3 Kafka Messages")

	consumer := Consumer{
		topic:        topic,
		messageStore: messageStore,
		ui:           ui,
		textView:     textView,
	}

	go func() {
		for {
			if err := client.Consume(ui, []string{topic}, &consumer); err != nil {
				fmt.Printf("Error from consumer: %v\n", err)
			}
		}
	}()

	// Handle OS signals for graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigterm
		ui.Stop()
	}()

	if err := ui.SetRoot(textView, true).Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
	}
}

type Consumer struct {
	topic        string
	messageStore *MessageStore
	ui           *tview.Application
	textView     *tview.TextView
}

func (c *Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		c.messageStore.Add(string(message.Value))
		c.ui.QueueUpdateDraw(func() {
			c.textView.SetText(strings.Join(c.messageStore.GetMessages(), "\n"))
		})
		session.MarkMessage(message, "")
	}
	return nil
}
