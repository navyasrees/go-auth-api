package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

var Writer *kafka.Writer

func InitProducer() {
    Writer = kafka.NewWriter(kafka.WriterConfig{
        Brokers:  []string{os.Getenv("KAFKA_BROKER")},
        Topic:    "send-email",
        Balancer: &kafka.LeastBytes{},
    })
    log.Println("📬 Kafka producer initialized")
}

func SendOTPEmail(toEmail, otp string) error {
    emailPayload := EmailPayload{
        To:      toEmail,
        Subject: "Your OTP Code",
        Body:    "Your OTP code is: " + otp,
    }

    jsonData, err := json.Marshal(emailPayload)
    if err != nil {
        return err
    }

    msg := kafka.Message{
        Key:   []byte(toEmail),
        Value: jsonData,
    }

    return Writer.WriteMessages(context.Background(), msg)
}

func SendPasswordResetOTPEmail(toEmail, otp string) error {
    emailPayload := EmailPayload{
        To:      toEmail,
        Subject: "Password Reset OTP",
        Body:    "Your password reset OTP code is: " + otp + "\n\nThis OTP will expire in 15 minutes.",
    }

    jsonData, err := json.Marshal(emailPayload)
    if err != nil {
        return err
    }

    msg := kafka.Message{
        Key:   []byte(toEmail),
        Value: jsonData,
    }

    return Writer.WriteMessages(context.Background(), msg)
}
