package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
	"gopkg.in/gomail.v2"
)

type EmailPayload struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

func StartEmailConsumer() {
    r := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{os.Getenv("KAFKA_BROKER")},
        Topic:   "send-email",
        GroupID: "auth-service-email-consumer",
    })

    go func() {
        log.Println("📬 Email consumer started...")
        for {
            m, err := r.ReadMessage(context.Background())
            if err != nil {
                log.Println("Kafka read error:", err)
                continue
            }

            var email EmailPayload
            if err := json.Unmarshal(m.Value, &email); err != nil {
                log.Println("Invalid email payload:", err)
                continue
            }

            if err := sendEmail(email); err != nil {
                log.Println("❌ Failed to send email:", err)
            } else {
                log.Println("✅ Email sent to", email.To)
            }
        }
    }()
}

func sendEmail(payload EmailPayload) error {
    // Use Mailpit SMTP settings
    msg := gomail.NewMessage()
    msg.SetHeader("From", "auth-service@example.com")
    msg.SetHeader("To", payload.To)
    msg.SetHeader("Subject", payload.Subject)
    msg.SetBody("text/plain", payload.Body)

    // Mailpit SMTP configuration
    dialer := gomail.NewDialer(
        "localhost", // Mailpit host
        1025,        // Mailpit SMTP port
        "",          // No username required for Mailpit
        "",          // No password required for Mailpit
    )

    return dialer.DialAndSend(msg)
}
