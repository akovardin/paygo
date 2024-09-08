package mail

import (
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"gopkg.in/gomail.v2"
)

// mailer config
// пароль аккаунта: robot@getapp.store t62fuwetff
// пароль приложения:  adxyubcssoyzhvcb
type Config struct {
	Password string
	Out      string
	In       string
	Port     int
	Username string
}

type Mailer struct {
	config Config
}

func New(config Config) *Mailer {
	return &Mailer{
		config: config,
	}
}

type Message struct {
	From    string `json:"from_email"`
	Name    string `json:"from_name"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Html    string `json:"html"`
}

func (m *Mailer) Send(message Message) error {
	// get settings from database?

	msg := gomail.NewMessage()
	msg.SetHeader("From", message.From)
	msg.SetHeader("To", message.To)
	msg.SetHeader("Subject", message.Subject)
	msg.SetBody("text/html", message.Html)
	//msg.Attach("/home/User/cat.jpg")

	n := gomail.NewDialer(m.config.Out, m.config.Port, m.config.Username, m.config.Password)

	if err := n.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}

func (m *Mailer) Read() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(m.config.In+":993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(m.config.Username, m.config.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Println("Last 4 messages:")
	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
