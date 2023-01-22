package smtp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/antonpriyma/otus-highload/pkg/email"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/stretchr/testify/require"
)

const host = `127.0.0.1`
const port = 1026

func smtpStub(data []string, wg *sync.WaitGroup, clientMessages *bufio.Writer, l net.Listener, errorsChan chan error) {
	wg.Add(1)
	defer wg.Done()

	conn, err := l.Accept()
	if err != nil {
		errorsChan <- errors.Errorf("Accept error: %v", err)
		return
	}
	defer conn.Close()

	tc := textproto.NewConn(conn)
	for i := 0; i < len(data) && data[i] != ""; i++ {
		_ = tc.PrintfLine(data[i])
		for len(data[i]) >= 4 && data[i][3] == '-' {
			i++
			_ = tc.PrintfLine(data[i])
		}
		if data[i] == "221 Goodbye" {
			return
		}
		read := false
		for !read || data[i] == "354 Go ahead" {
			msg, err := tc.ReadLine()
			_, _ = clientMessages.Write([]byte(msg + "\r\n"))
			read = true
			if err != nil {
				errorsChan <- errors.Errorf("Read error: %v", err)
				return
			}
			if data[i] == "354 Go ahead" && msg == "." {
				break
			}
		}
	}
}

func TestSendMailWithTLS(t *testing.T) {
	hostname, _ := os.Hostname()

	for _, c := range []struct {
		name                   string
		serverMessages         string
		expectedClientMessages string
		expectedError          error
		expectSMTPError        bool
		sender                 emailSender
	}{
		{
			name:                   "simple",
			serverMessages:         sendMailServer,
			expectedClientMessages: sendMailClient,
			sender: emailSender{
				SkipTLS: true,
			},
		}, {
			name:                   "with TLS",
			serverMessages:         sendMailServerWithTLS,
			expectedClientMessages: sendMailClientWithTLS,
			expectedError:          email.ErrFailedStartTLS,
			expectSMTPError:        true,
			sender:                 emailSender{},
		}, {
			name:                   "tls unsupported",
			serverMessages:         sendMailServer,
			expectedClientMessages: sendMailClientUnsupportedTLS,
			expectedError:          email.ErrTLSUnsupported,
			expectSMTPError:        true,
			sender:                 emailSender{},
		}, {
			name:                   "with auth",
			serverMessages:         sendMailServerWithAUTH,
			expectedClientMessages: sendMailClientWithAUTH,
			sender: emailSender{
				Auth:    smtp.PlainAuth("", "username", "password", host),
				SkipTLS: true,
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			c.sender.Host = host
			c.sender.Port = port
			c.sender.Logger = log.Null

			c.expectedClientMessages = strings.Replace(c.expectedClientMessages, "localhost", hostname, 1)
			c.serverMessages = strings.Replace(c.serverMessages, "localhost", hostname, 1)

			var clientMessages bytes.Buffer
			clientMessagesBuf := bufio.NewWriter(&clientMessages)

			wg := sync.WaitGroup{}

			server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
			if err != nil {
				require.NoError(t, err)
			}

			defer server.Close()

			smtpError := make(chan error, 1)

			go smtpStub(strings.Split(c.serverMessages, "\n"), &wg, clientMessagesBuf, server, smtpError)

			opts := email.SendOpts{
				EmailFrom: "test@example.com",
			}

			err = c.sender.sendEmail(
				context.Background(),
				[]string{"other@example.com"},
				[]byte(strings.Replace(testMessage, "\n", "\r\n", -1)),
				opts,
			)
			require.True(t, errors.Is(err, c.expectedError))

			wg.Wait()

			if !c.expectSMTPError {
				select {
				case err = <-smtpError:
					require.NoError(t, err)
				default:
				}
			}

			clientMessagesBuf.Flush()

			c.expectedClientMessages = strings.Join(strings.Split(c.expectedClientMessages, "\n"), "\r\n")
			if c.expectedClientMessages != clientMessages.String() {
				t.Errorf("Got:\n%s\nExpected:\n%s", clientMessages.String(), c.expectedClientMessages)
			}
		})

	}
}

var sendMailServer = `220 hello world
250 mx.google.com
250 Sender ok
250 Receiver ok
354 Go ahead
250 Data ok
221 Goodbye
`

var sendMailClient = `EHLO localhost
MAIL FROM:<test@example.com>
RCPT TO:<other@example.com>
DATA
From: test@example.com
To: other@example.com
Subject: SendMail test

SendMail is working for me.
.
QUIT
`

var sendMailServerWithTLS = `220 hello world
250-mx.google.com
250 STARTTLS
250 Sender ok
250 Receiver ok
354 Go ahead
250 Data ok
221 Goodbye
`

// Breaking with error
var sendMailClientWithTLS = `EHLO localhost
STARTTLS

`

var sendMailClientUnsupportedTLS = `EHLO localhost

`

var sendMailServerWithAUTH = `220 hello world
250-mx.google.com
250 AUTH
235 Auth ok
250 Sender ok
250 Receiver ok
354 Go ahead
250 Data ok
221 Goodbye
`

var sendMailClientWithAUTH = `EHLO localhost
AUTH PLAIN AHVzZXJuYW1lAHBhc3N3b3Jk
MAIL FROM:<test@example.com>
RCPT TO:<other@example.com>
DATA
From: test@example.com
To: other@example.com
Subject: SendMail test

SendMail is working for me.
.
QUIT
`

var testMessage = `From: test@example.com
To: other@example.com
Subject: SendMail test

SendMail is working for me.
`
