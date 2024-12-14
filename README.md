# Apache Kafka

Projeto feito para a matéria de Sistemas Distribuidos

Dupla: Gabriel Bortoli Buron e Juan Paulo Fricke


## Pré-requisitos

1. **Instalar o Docker**:

   - Siga o guia de instalação para o seu sistema operacional:
     - [Docker para Windows](https://docs.docker.com/desktop/install/windows-install/)
     - [Docker para macOS](https://docs.docker.com/desktop/install/mac-install/)
     - [Docker para Linux](https://docs.docker.com/engine/install/)

---

## Configurando o Cluster Kafka

### Passo 1: Iniciar o Ambiente Kafka

1. Construa e inicie o ambiente:

   ```bash
   docker compose up -d
   ```

2. Verifique se todos os containers estão em execução:
   ```bash
   docker ps
   ```
   Você deverá ver containers para `controller-1`, `controller-2`, `controller-3`, `broker-1`, `broker-2`, and `broker-3`.

### Step 2:

1. Para baixar as dependências do projeto:

```bash
    go mod tidy
```

2. Para iniciar o script do Produtor:

```bash
    go run sender.go
```

3. Para iniciar o script do Consumidor:

```bash
    go run reciver.go
```

---

## Utilizando Produtores e Consumidores

### Cenário 1: Operação Normal (Todos os Nós Ativos)

1. **Iniciar um Produtor**:

   - Abra um terminal e acesse o `broker-1`:
     ```bash
     docker exec -it broker-1 bash
     ```
   - Execute o produtor:
     ```bash
     kafka-console-producer.sh --broker-list broker-1:19092,broker-2:19092,broker-3:19092 --topic test-topic
     ```
   - Digite mensagens e pressione `Enter` para enviar.

2. **Iniciar um Consumidor**:
   - Abra outro terminal e acesse o `broker-2`:
     ```bash
     docker exec -it broker-2 bash
     ```
   - Execute o consumidor:
     ```bash
     kafka-console-consumer.sh --bootstrap-server broker-1:19092,broker-2:19092,broker-3:19092 --topic test-topic --from-beginning
     ```
   - Você verá as mensagens enviadas pelo produtor.

### Cenário 2: Um Nó Desligado

1. Desligue um broker, por exemplo `broker-2`:
   ```bash
   docker stop broker-2
   ```
2. Envie e receba mensagens utilizando os mesmos processos do produtor e consumidor.

3. Reinicie o broker quando terminar:
   ```bash
   docker start broker-2
   ```

### Cenário 3: Adicionando um Novo Nó

1. Adicione um novo broker ao arquivo `docker-compose.yml`:
   ```yaml
   broker-4:
     image: apache/kafka:latest
     container_name: broker-(NUMERO DO CONTAINER)
     ports:
       - (NUMERO CRESCENTE)9092:9092
     environment:
       KAFKA_NODE_ID: (PROXIMO ID)
       KAFKA_NUM_PARTITIONS: ${KAFKA_NUM_PARTITIONS}
       KAFKA_DEFAULT_REPLICATION_FACTOR: ${KAFKA_DEFAULT_REPLICATION_FACTOR}
       KAFKA_PROCESS_ROLES: broker
       KAFKA_LISTENERS: "PLAINTEXT://:19092,PLAINTEXT_HOST://:9092"
       KAFKA_ADVERTISED_LISTENERS: "PLAINTEXT://broker-4:19092,PLAINTEXT_HOST://localhost:59092"
       KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT
       KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
       KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT
       KAFKA_CONTROLLER_QUORUM_VOTERS: ${KAFKA_CONTROLLER_QUORUM_VOTERS}
       KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
     depends_on:
       - controller-1
       - controller-2
       - controller-3
   ```
2. Inicie o novo broker:
   ```bash
   docker compose up -d broker-4
   ```

---

## Parando o Ambiente

1. Desligue todos os containers:
   ```bash
   docker compose down
   ```

---

## Solução de Problemas

- **Logs**:
  - Verifique os logs de qualquer container:
    ```bash
    docker logs <container_name>
    ```
- **Recriar Ambiente**:
  - Remova e recrie os containers:
    ```bash
    docker compose down -v
    docker compose up -d
    ```

---
