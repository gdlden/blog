package biz

import (
	"fmt"
	"net/smtp"
	"time"

	"blog/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

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

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", s.from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	s.log.Infof("Sending verification code to %s", to)
	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		s.log.Errorf("Failed to send email to %s: %v", to, err)
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	return nil
}

func generateCode() string {
	// 6 位数字验证码
	now := time.Now().UnixNano()
	code := int(now%900000 + 100000)
	return fmt.Sprintf("%d", code)
}
