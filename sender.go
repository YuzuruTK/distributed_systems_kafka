package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
)

type Joke struct {
	ID   string `json:"id"`
	Joke string `json:"value"`
}

func generateRandomJoke() (*Joke, error) {
	// Defina o URL da API que retorna piadas aleatórias
	apiURL := "https://api.chucknorris.io/jokes/random" // API de piadas aleatórias do Chuck Norris

	// Criar um cliente HTTP customizado que ignora erros de certificado
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Requisição para a API
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Ler o corpo da resposta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON da resposta da API
	var apiResponse Joke
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	// Criar o objeto Joke com os dados da API
	joke := &Joke{
		ID:   apiResponse.ID,
		Joke: apiResponse.Joke,
	}

	return joke, nil
}

func createTopicIfNotExists(brokers []string, topic string) error {
	// Conecta-se ao broker Kafka para criar o tópico
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("erro ao conectar ao broker Kafka: %v", err)
	}
	defer conn.Close()

	// Tenta criar o tópico
	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		return fmt.Errorf("erro ao criar o tópico Kafka: %v", err)
	}

	fmt.Printf("Tópico '%s' criado com sucesso!\n", topic)
	return nil
}

func main() {
	brokers := []string{"localhost:29092", "localhost:39092", "localhost:49092"} // Brokers Kafka
	topic := "random-jokes"

	// Cria o tópico se não existir
	err := createTopicIfNotExists(brokers, topic)
	if err != nil {
		log.Fatalf("Erro ao criar o tópico: %v", err)
	}

	// Inicializa o writer Kafka
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})

	// Teste de conectividade
	testMessage := kafka.Message{
		Value: []byte("Test message"),
	}

	err = writer.WriteMessages(context.Background(), testMessage)
	if err != nil {
		log.Fatalf("Erro ao conectar ou enviar mensagem para o Kafka: %v", err)
	}

	// Verifique a conectividade com uma mensagem de teste
	fmt.Println("Conectado ao Kafka com sucesso!")

	// Envio contínuo de piadas para o Kafka
	for {
		joke, err := generateRandomJoke()
		if err != nil {
			log.Fatalf("Erro ao gerar piada: %v", err)
		}

		// Converte a piada para JSON
		jokeBytes, err := json.Marshal(joke)
		if err != nil {
			log.Fatalf("Erro ao converter piada para JSON: %v", err)
		}

		// Cria a mensagem Kafka
		kafkaMessage := kafka.Message{
			Value: jokeBytes,
		}

		// Envia a mensagem para o Kafka
		err = writer.WriteMessages(context.Background(), kafkaMessage)
		if err != nil {
			log.Fatalf("Erro ao enviar mensagem para o Kafka: %v", err)
		}

		// Log da mensagem enviada
		fmt.Printf("Piada enviada para o Kafka: %s\n", joke.Joke)

		// Atraso antes de enviar a próxima mensagem
		time.Sleep(5 * time.Second)
	}
}
