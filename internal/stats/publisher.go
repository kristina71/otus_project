package stats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kristina71/otus_project/internal/config"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const PublisherMsgKey = "amqp.rotation.service.key"

type Message struct {
	BannerID  string    `json:"bannerId"`
	SlotID    string    `json:"slotId"`
	GroupID   string    `json:"groupId"`
	Type      string    `json:"-"`
	Timestamp time.Time `json:"-"`
}

type Publisher struct {
	conn         *amqp.Connection
	ch           *amqp.Channel
	exchangeName string
	queueName    string
}

func NewPublisher(cnf config.PublisherConfig) (*Publisher, error) {
	connection, err := amqp.Dial(cnf.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbit server: %w", err)
	}

	return &Publisher{
		conn:         connection,
		queueName:    cnf.QueueName,
		exchangeName: cnf.ExchangeName,
	}, nil
}

func (p *Publisher) Start() error {
	channel, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to get channel for rabbit connection: %w", err)
	}
	p.ch = channel

	if err := channel.ExchangeDeclare(
		p.exchangeName,
		"direct",
		false,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to create exchange: %w", err)
	}

	if _, err := channel.QueueDeclare(
		p.queueName,
		false,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf(" failed to create queue: %w", err)
	}

	if err := channel.QueueBind(
		p.queueName,
		PublisherMsgKey,
		p.exchangeName,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	zap.L().Info("rotation service stats publisher successfully started")
	return nil
}

func (p *Publisher) Stop() error {
	return p.conn.Close()
}

func (p *Publisher) Publish(msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error during marhalling message data: %w", err)
	}
	if err := p.ch.Publish(
		p.exchangeName,
		PublisherMsgKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Timestamp:   msg.Timestamp,
			Type:        msg.Type,
			AppId:       "banner-rotation",
			Body:        data,
		},
	); err != nil {
		return fmt.Errorf("error during publishing data to rabbit queue: %w", err)
	}
	return nil
}
