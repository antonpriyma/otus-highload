package smtp

import (
	"net/textproto"

	"github.com/antonpriyma/otus-highload/pkg/errors"
)

var (
	ErrMessageCantBeSent = errors.Typed("smtp_message_cant_be_sent", "message can't be sent")
	ErrBadAddress        = errors.Typed("smtp_bad_address", "bad address")
)

var smtpCodeToError = map[int]error{
	550: ErrMessageCantBeSent,
}

func typedSMTPError(err error) error {
	var smtpError *textproto.Error
	if errors.As(err, &smtpError) {
		typedErr, ok := smtpCodeToError[smtpError.Code]
		if ok {
			return errors.Transform(err, typedErr)
		}
	}

	return err
}
