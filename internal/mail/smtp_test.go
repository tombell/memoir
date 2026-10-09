package mail

import (
	"context"
	"fmt"
	"github.com/tombell/memoir/internal/config"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

func TestSMTPDeliveryAndTLSRequirement(t *testing.T) {
	for _, mode := range []string{"none", "starttls"} {
		t.Run(mode, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			messages := make(chan string, 1)
			done := make(chan struct{})
			go func() {
				defer close(done)
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
				proto := textproto.NewConn(conn)
				_ = proto.PrintfLine("220 localhost SMTP")
				for {
					line, err := proto.ReadLine()
					if err != nil {
						return
					}
					switch {
					case strings.HasPrefix(line, "EHLO"):
						_ = proto.PrintfLine("250 localhost")
					case strings.HasPrefix(line, "MAIL FROM:"), strings.HasPrefix(line, "RCPT TO:"):
						_ = proto.PrintfLine("250 OK")
					case line == "DATA":
						_ = proto.PrintfLine("354 send data")
						body, err := proto.ReadDotBytes()
						if err != nil {
							return
						}
						messages <- string(body)
						_ = proto.PrintfLine("250 OK")
					case line == "QUIT":
						_ = proto.PrintfLine("221 bye")
						return
					default:
						_ = proto.PrintfLine("500 unknown")
					}
				}
			}()
			sender := New(config.SMTPConfig{Address: listener.Addr().String(), From: "memoir@example.test", TLSMode: mode})
			err = sender.Send(context.Background(), "person@example.test", "Verify email", "Open this link:\nhttps://example.test/#token=secret")
			if mode == "starttls" {
				if err == nil {
					t.Fatal("unencrypted server accepted when TLS required")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				message := <-messages
				for _, expected := range []string{"To: person@example.test", "Subject: Verify email", "https://example.test/#token=secret"} {
					if !strings.Contains(message, expected) {
						t.Fatal(fmt.Sprintf("missing %q", expected))
					}
				}
			}
			<-done
		})
	}
}

func TestSMTPHeaderInjection(t *testing.T) {
	sender := New(config.SMTPConfig{Address: "127.0.0.1:1", From: "memoir@example.test", TLSMode: "none"})
	if err := sender.Send(context.Background(), "person@example.test\r\nBcc: other@example.test", "Verify", "body"); err == nil {
		t.Fatal("header injection accepted")
	}
}

func TestSMTPCancellationDuringGreeting(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetReadDeadline(time.Now().Add(time.Second))
		var buf [1]byte
		_, _ = conn.Read(buf[:])
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	sender := New(config.SMTPConfig{Address: listener.Addr().String(), From: "memoir@example.test", TLSMode: "none"})
	start := time.Now()
	if err := sender.Send(ctx, "person@example.test", "Verify", "body"); err == nil {
		t.Fatal("cancelled SMTP send succeeded")
	}
	if time.Since(start) > time.Second {
		t.Fatal("SMTP cancellation failed to interrupt greeting")
	}
	<-done
}
