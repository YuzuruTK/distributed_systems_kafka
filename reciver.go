package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/olekukonko/tablewriter"
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

type Consumer struct {
	topic        string
	messageStore *MessageStore
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Adiciona a mensagem ao store
		c.messageStore.Add(string(message.Value))

		// Exibe as mensagens em formato de tabela, incluindo o IP do broker
		printMessages(c.messageStore.GetMessages(), message.Headers) // Passa as mensagens e cabeçalhos (que contêm o IP)

		// Salva as mensagens no arquivo
		saveMessagesToFile(c.messageStore.GetMessages()) // Salva mensagens no arquivo
		session.MarkMessage(message, "")
	}
	return nil
}

// Função para exibir mensagens em formato de tabela
func printMessages(messages []string, headers []*sarama.RecordHeader) {
	// Limpa a tela antes de imprimir
	fmt.Print("\033[H\033[2J")

	// Extraímos o IP do broker da primeira header (se disponível)
	var brokerIP string
	for _, header := range headers {
		if string(header.Key) == "broker-ip" { // A chave 'broker-ip' pode ser ajustada conforme necessário
			brokerIP = string(header.Value)
			break
		}
	}

	// Criação da tabela
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Broker IP", "Message 1", "Message 2", "Message 3"}) // Exemplo para 3 mensagens

	// Preenche as mensagens nas colunas
	row := []string{brokerIP}
	for _, message := range messages {
		row = append(row, message)
	}
	// Adiciona a linha na tabela
	table.Append(row)

	table.Render()
}

// Função para salvar mensagens em um arquivo
func saveMessagesToFile(messages []string) {
	// Cria ou abre o arquivo de mensagens
	file, err := os.OpenFile("messages.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Escreve as mensagens no arquivo
	for _, message := range messages {
		_, err := file.WriteString(message + "\n")
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			return
		}
	}
}

func main() {
	brokers := []string{"localhost:29092", "localhost:39092", "localhost:49092"}
	topic := "random-jokes"
	group := "cli-consumer"
	messageStore := NewMessageStore(3)

	// Configuração do Sarama Consumer
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Version = sarama.V2_8_0_0

	// Criação do consumidor Kafka
	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		fmt.Printf("Error creating consumer group client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Configuração do consumidor para o Kafka
	consumer := &Consumer{
		topic:        topic,
		messageStore: messageStore,
	}

	// Contexto com cancelamento
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Consumindo mensagens do Kafka em uma goroutine
	go func() {
		for {
			if err := client.Consume(ctx, []string{topic}, consumer); err != nil {
				fmt.Printf("Error from consumer: %v\n", err)
			}

			// Saída do loop se o contexto for cancelado
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// Tratamento de sinal para encerramento gracioso
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
	cancel()
}
