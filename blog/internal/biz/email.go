package biz

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"blog/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

// loginAuth implements AUTH LOGIN for SMTP (required by QQ/163 mail).
type loginAuth struct {
	username string
	password string
	step     int
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	a.step++
	if a.step == 1 {
		return []byte(a.username), nil
	}
	return []byte(a.password), nil
}

type EmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
	log      *log.Helper
}

func NewEmailSender(c *conf.SMTP, logger log.Logger) *EmailSender {
	return &EmailSender{
		host:     c.Host,
		port:     int(c.Port),
		username: c.Username,
		password: c.Password,
		from:     c.From,
		log:      log.NewHelper(logger),
	}
}

func (s *EmailSender) SendVerificationCode(to string, code string) error {
	auth := &loginAuth{username: s.username, password: s.password}
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// Fallback from to username if not set
	from := s.from
	if from == "" {
		from = s.username
	}

	s.log.Infof("sending email via %s from=%s to=%s", addr, from, to)

	// Try TLS direct (port 465) first, fall back to STARTTLS (port 587)
	if s.port == 465 {
		return s.sendTLSDirect(addr, auth, from, to, code)
	}
	return s.sendSTARTTLS(addr, auth, from, to, code)
}

func (s *EmailSender) sendTLSDirect(addr string, auth smtp.Auth, from, to string, code string) error {
	tlsConfig := &tls.Config{ServerName: s.host, InsecureSkipVerify: false}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		s.log.Errorf("tls dial failed: %v", err)
		return fmt.Errorf("连接SMTP服务器失败: %w", err)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer client.Close()

	return s.sendMail(client, auth, from, to, code)
}

func (s *EmailSender) sendSTARTTLS(addr string, auth smtp.Auth, from, to string, code string) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		s.log.Errorf("dial failed: %v", err)
		return fmt.Errorf("连接SMTP服务器失败: %w", err)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer client.Close()

	// Check if STARTTLS is supported
	ok, _ := client.Extension("STARTTLS")
	if ok {
		tlsConfig := &tls.Config{ServerName: s.host, InsecureSkipVerify: false}
		if err := client.StartTLS(tlsConfig); err != nil {
			s.log.Errorf("STARTTLS failed: %v", err)
			return fmt.Errorf("TLS握手失败: %w", err)
		}
	}

	return s.sendMail(client, auth, from, to, code)
}

func (s *EmailSender) sendMail(client *smtp.Client, auth smtp.Auth, from, to string, code string) error {
	// Authenticate
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			s.log.Errorf("auth failed: %v", err)
			return fmt.Errorf("SMTP认证失败: %w", err)
		}
	}

	// MAIL FROM
	if err := client.Mail(from); err != nil {
		s.log.Errorf("MAIL FROM failed: %v", err)
		return fmt.Errorf("发件人地址错误: %w", err)
	}

	// RCPT TO
	if err := client.Rcpt(to); err != nil {
		s.log.Errorf("RCPT TO failed: %v", err)
		return fmt.Errorf("收件人地址错误: %w", err)
	}

	// DATA
	w, err := client.Data()
	if err != nil {
		s.log.Errorf("DATA command failed: %v", err)
		return fmt.Errorf("发送数据失败: %w", err)
	}

	subject := "=?UTF-8?B?5qOA5rWL6YKu5Lu277yIcGx1Z2luLm1pYW9sYW8uY29t77yJ?="
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
  <div style="max-width: 600px; margin: 0 auto; background: #ffffff; border-radius: 8px; padding: 30px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
    <h2 style="color: #333; text-align: center;">邮箱验证码</h2>
    <p style="font-size: 16px; color: #555;">您好，</p>
    <p style="font-size: 16px; color: #555;">您正在进行邮箱验证，验证码为：</p>
    <div style="text-align: center; margin: 30px 0;">
      <span style="font-size: 36px; font-weight: bold; letter-spacing: 8px; color: #0071e3; background: #f5f5f7; padding: 12px 24px; border-radius: 8px;">%s</span>
    </div>
    <p style="font-size: 14px; color: #999;">验证码有效期为 %d 分钟，请尽快完成验证。</p>
    <p style="font-size: 14px; color: #999;">如果这不是您本人的操作，请忽略此邮件。</p>
  </div>
</body>
</html>`, code, 10)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", from, to, subject, body)

	if _, err := fmt.Fprint(w, msg); err != nil {
		w.Close()
		s.log.Errorf("write data failed: %v", err)
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	if err := w.Close(); err != nil {
		s.log.Errorf("close data failed: %v", err)
		return fmt.Errorf("发送邮件内容失败: %w", err)
	}

	client.Quit()
	s.log.Infof("email sent successfully to %s", to)
	return nil
}

func generateCode() string {
	now := time.Now().UnixNano()
	code := int(now%900000 + 100000)
	return fmt.Sprintf("%d", code)
}
