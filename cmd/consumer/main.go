// Package main provides the message consumer for Redis Streams in the transactional outbox pattern.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/rueidis"

	"github.com/jnst/transactional-outbox-pattern/internal/config"
	"github.com/jnst/transactional-outbox-pattern/internal/model"
)

const (
	mailProcessingDelay = 100 * time.Millisecond
	redisBlockTimeout   = 1000 // milliseconds
	errorRetryDelay     = 1 * time.Second
	signalBufferSize    = 1
)

// MessageHandler processes messages from Redis Streams.
type MessageHandler struct {
	redisClient rueidis.Client
}

// NewMessageHandler creates a new message handler instance.
func NewMessageHandler(redisClient rueidis.Client) *MessageHandler {
	return &MessageHandler{
		redisClient: redisClient,
	}
}

// HandleUserCreatedEvent processes user creation events.
func (h *MessageHandler) HandleUserCreatedEvent(ctx context.Context, event *model.UserCreatedEvent) error {
	log.Printf("Processing user_created event: ID=%d, Name=%s, Email=%s", event.UserID, event.Name, event.Email)

	// ここで外部サービス（メール送信など）を呼び出す
	// 例: ウェルカムメール送信
	if err := h.sendWelcomeEmail(ctx, event.Name, event.Email); err != nil {
		return err
	}

	log.Printf("Successfully processed user_created event for user %d", event.UserID)

	return nil
}

func (*MessageHandler) sendWelcomeEmail(_ context.Context, name, email any) error {
	// TODO: 実際のメール送信ロジックをここに実装
	// 今回はログ出力のみ
	log.Printf("Sending welcome email to %s (%s)", name, email)

	// メール送信の処理時間をシミュレート
	time.Sleep(mailProcessingDelay)

	log.Printf("Welcome email sent successfully to %s", email)

	return nil
}

func setupRedisClient(cfg *config.Config) (rueidis.Client, error) {
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{cfg.RedisAddr},
	})
	if err != nil {
		return nil, err
	}

	return redisClient, nil
}

func setupSignalHandling() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, signalBufferSize)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, stopping consumer...")
		cancel()
	}()

	return ctx, cancel
}

func createConsumerGroup(ctx context.Context, redisClient rueidis.Client, streamKey, groupName string) {
	createGroupCmd := redisClient.B().XgroupCreate().Key(streamKey).Group(groupName).Id("0").Mkstream().Build()
	if err := redisClient.Do(ctx, createGroupCmd).Error(); err != nil {
		log.Printf("Consumer group creation result (may already exist): %v", err)
	}
}

func runConsumerLoop(ctx context.Context, handler *MessageHandler, streamKey, groupName, consumerName string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopped")
			return
		default:
			if err := handler.consumeMessages(ctx, streamKey, groupName, consumerName); err != nil {
				log.Printf("Error consuming messages: %v", err)
				time.Sleep(errorRetryDelay)
			}
		}
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	redisClient, err := setupRedisClient(cfg)
	if err != nil {
		log.Fatal("failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	handler := NewMessageHandler(redisClient)
	ctx, cancel := setupSignalHandling()
	defer cancel()

	streamKey := "user:events"
	groupName := "email-service"
	consumerName := cfg.ConsumerName

	createConsumerGroup(ctx, redisClient, streamKey, groupName)

	log.Printf("Starting message consumer (stream: %s, group: %s, consumer: %s)...",
		streamKey, groupName, consumerName)

	runConsumerLoop(ctx, handler, streamKey, groupName, consumerName)
}

func (h *MessageHandler) readMessages(
	ctx context.Context,
	streamKey, groupName, consumerName string,
) (map[string][]rueidis.XRangeEntry, error) {
	readCmd := h.redisClient.B().Xreadgroup().Group(groupName, consumerName).
		Count(1).
		Block(redisBlockTimeout).
		Streams().
		Key(streamKey).
		Id(">").
		Build()

	result := h.redisClient.Do(ctx, readCmd)
	if err := result.Error(); err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, nil // タイムアウト（正常）
		}

		return nil, err
	}

	return result.AsXRead()
}

func (h *MessageHandler) acknowledgeMessage(ctx context.Context, streamKey, groupName, messageID string) {
	ackCmd := h.redisClient.B().Xack().Key(streamKey).Group(groupName).Id(messageID).Build()
	if err := h.redisClient.Do(ctx, ackCmd).Error(); err != nil {
		log.Printf("failed to ACK message %s: %v", messageID, err)
	} else {
		log.Printf("ACKed message %s", messageID)
	}
}

func (h *MessageHandler) processStreamMessages(
	ctx context.Context,
	streamKey, groupName string,
	messages []rueidis.XRangeEntry,
) {
	for _, message := range messages {
		if err := h.processMessage(ctx, streamKey, groupName, message); err != nil {
			log.Printf("failed to process message %s: %v", message.ID, err)
			continue
		}

		h.acknowledgeMessage(ctx, streamKey, groupName, message.ID)
	}
}

func (h *MessageHandler) consumeMessages(ctx context.Context, streamKey, groupName, consumerName string) error {
	streams, err := h.readMessages(ctx, streamKey, groupName, consumerName)
	if err != nil {
		return err
	}

	if streams == nil {
		return nil // タイムアウト
	}

	for streamName, messages := range streams {
		log.Printf("Processing stream: %s", streamName)
		h.processStreamMessages(ctx, streamKey, groupName, messages)
	}

	return nil
}

func (h *MessageHandler) processMessage(ctx context.Context, _, _ string, message rueidis.XRangeEntry) error {
	log.Printf("Received message %s: %v", message.ID, message.FieldValues)

	// イベントタイプを取得
	eventType, ok := message.FieldValues["event_type"]
	if !ok {
		return errors.New("missing event_type in message")
	}

	// ペイロードを取得してパース
	payloadStr, ok := message.FieldValues["payload"]
	if !ok {
		return errors.New("missing payload in message")
	}

	// イベントタイプに応じて処理
	switch model.EventAction(eventType) {
	case model.EventActionUserCreated:
		var event model.UserCreatedEvent
		if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
			return fmt.Errorf("failed to parse user_created payload: %w", err)
		}

		return h.HandleUserCreatedEvent(ctx, &event)
	default:
		log.Printf("Unknown event type: %s", eventType)
		return nil // 未知のイベントタイプは無視
	}
}
