package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type SMTPMailer struct {
	host     string
	port     int
	user     string
	pass     string
	fromName string
	fromMail string
	appName  string
}

func NewSMTPMailer(host, port, user, pass, fromName, fromMail, appName string) (*SMTPMailer, error) {
	p, err := strconv.Atoi(port)
	if err != nil || p <= 0 {
		return nil, fmt.Errorf("invalid SMTP port")
	}
	if host == "" || user == "" || pass == "" || fromMail == "" {
		return nil, fmt.Errorf("smtp credentials incomplete")
	}
	if fromName == "" {
		fromName = appName
	}
	if appName == "" {
		appName = "Pistoni"
	}
	return &SMTPMailer{host: host, port: p, user: user, pass: pass, fromName: fromName, fromMail: fromMail, appName: appName}, nil
}

func (m *SMTPMailer) SendCodeEmail(email, code, subject, intro string) error {
	html := renderCodeEmail(m.appName, subject, intro, code)
	plain := fmt.Sprintf("%s\n\n%s\n\nCode: %s\n\nThis code expires in 15 minutes.", m.appName, intro, code)
	return m.send(email, subject, plain, html)
}

func (m *SMTPMailer) send(to, subject, plain, html string) error {
	fromHeader := m.fromMail
	if m.fromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", m.fromName, m.fromMail)
	}

	boundary := "pistoni-boundary-7d9f9a"
	msg := strings.Builder{}
	msg.WriteString("From: " + fromHeader + "\r\n")
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n\r\n")
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	msg.WriteString(plain + "\r\n\r\n")
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	msg.WriteString(html + "\r\n\r\n")
	msg.WriteString("--" + boundary + "--")

	addr := net.JoinHostPort(m.host, strconv.Itoa(m.port))
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	log.Printf("[mailer] smtp send attempt to=%s host=%s port=%d subject=%q", to, m.host, m.port, subject)

	timeout := 10 * time.Second
	var conn net.Conn
	var err error
	if m.port == 465 {
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", addr, &tls.Config{ServerName: m.host})
	} else {
		conn, err = net.DialTimeout("tcp", addr, timeout)
	}
	if err != nil {
		log.Printf("[mailer] smtp dial failed to=%s err=%v", to, err)
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, m.host)
	if err != nil {
		log.Printf("[mailer] smtp client init failed to=%s err=%v", to, err)
		return err
	}
	defer c.Quit()

	if ok, _ := c.Extension("STARTTLS"); ok && m.port != 465 {
		if err := c.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
			log.Printf("[mailer] smtp starttls failed to=%s err=%v", to, err)
			return err
		}
	}
	if ok, _ := c.Extension("AUTH"); ok {
		if err := c.Auth(auth); err != nil {
			log.Printf("[mailer] smtp auth failed to=%s err=%v", to, err)
			return err
		}
	}
	if err := c.Mail(m.fromMail); err != nil {
		log.Printf("[mailer] smtp MAIL FROM failed to=%s err=%v", to, err)
		return err
	}
	if err := c.Rcpt(to); err != nil {
		log.Printf("[mailer] smtp RCPT TO failed to=%s err=%v", to, err)
		return err
	}
	wc, err := c.Data()
	if err != nil {
		log.Printf("[mailer] smtp DATA failed to=%s err=%v", to, err)
		return err
	}
	if _, err := wc.Write([]byte(msg.String())); err != nil {
		_ = wc.Close()
		log.Printf("[mailer] smtp write body failed to=%s err=%v", to, err)
		return err
	}
	if err := wc.Close(); err != nil {
		log.Printf("[mailer] smtp finalize failed to=%s err=%v", to, err)
		return err
	}

	log.Printf("[mailer] smtp send ok to=%s", to)
	return nil
}

func renderCodeEmail(appName, title, message, code string) string {
	if appName == "" {
		appName = "Pistoni"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="color-scheme" content="light">
  <meta name="forced-color-adjust" content="none">
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background-color:#f5f5f5 !important;color-scheme:light;forced-color-adjust:none;font-family:'SF Pro Display',-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <div style="max-width:800px;margin:40px auto;background:#ffffff !important;border-radius:4px;overflow:hidden;box-shadow:0 0.1px rgba(0,0,0,0.1);">
    <div style="padding:48px 48px 32px;">
      <div style="padding-bottom:32px;text-align:start;">
        <div style="font-size:28px;font-weight:700;letter-spacing:0.5px;color:#111111 !important;">%s</div>
      </div>

      <p style="margin:24px 0;font-size:16px;color:#1a1a1a !important;">Hey there,</p>

      <h1 style="margin:0 0 16px;font-size:30px;color:#1a1a1a !important;font-weight:600;text-align:start;">
        Verify your email address
      </h1>

      <p style="margin:0 0 32px;font-size:16px;color:#666666 !important;line-height:1.6;text-align:start;">
        %s
      </p>

      <div style="background:rgb(241,242,255) !important;border-radius:9px;padding:16px;margin:32px 0;text-align:center;">
        <h2 style="margin:0;font-size:24px;color:#000000 !important;font-weight:600;letter-spacing:12px;">%s</h2>
      </div>

      <p style="margin:32px 0;font-size:14px;color:#666666 !important;text-align:start;line-height:1.6;">
        This code will expire in 15 minutes.<br>
        If you didn't request this code, you can safely ignore this email.
      </p>

      <div style="height:1px;background:#e5e5e5 !important;margin:32px 0;"></div>

      <div style="text-align:start;">
        <p style="margin:0 0 16px;font-size:14px;color:#666666 !important;">
          Need help? <a href="mailto:support@pistoni.com" style="color:#000000 !important;text-decoration:none;">Contact Support</a>
        </p>
        <p style="margin:0;font-size:12px;color:#999999 !important;line-height:1.6;">
          © 2026 %s. All rights reserved.
        </p>
      </div>
    </div>
  </div>
</body>
</html>`, title, appName, message, code, appName)
}
