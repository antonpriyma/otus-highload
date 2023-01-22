package email

import (
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/emersion/go-message"
)

func Multipart(header Header, parts ...Part) Part {
	return multipart{
		parts:  parts,
		header: header,
	}
}

type multipart struct {
	parts  []Part
	header Header
}

func (m multipart) MarshalEML(writer *message.Writer) error {
	for i, part := range m.parts {
		if err := m.marshalEML(writer, part); err != nil {
			return errors.Wrapf(err, "failed to marshal %d part of multipart", i)
		}
	}

	return nil
}

func (m multipart) marshalEML(writer *message.Writer, part Part) error {
	if writer == nil {
		return errors.New("writer is nil")
	}

	partHeaders := part.Header()
	err := ValidateHeaders(partHeaders)
	if err != nil {
		return errors.Wrap(err, "failed to validate part headers")
	}

	newPart, err := writer.CreatePart(partHeaders.MessageHeaders())
	if err != nil {
		return errors.Wrap(err, "failed to create part in message")
	}

	if err := part.MarshalEML(newPart); err != nil {
		return errors.Wrapf(err, "failed to marshal EML")
	}

	if err := newPart.Close(); err != nil {
		return errors.Wrap(err, "failed to close part writer")
	}

	return nil
}

func (m multipart) Header() Header {
	return m.header
}

type simplePart struct {
	content []byte
	header  Header
}

func SimplePart(
	header Header,
	content []byte,
) Part {
	if header.Get(HeaderContentTransferEncoding) == "" {
		// lib auto encoding content, basing on header
		header.Add(HeaderContentTransferEncoding, "base64")
	}

	return simplePart{
		content: content,
		header:  header,
	}
}

func (p simplePart) MarshalEML(writer *message.Writer) error {
	if writer == nil {
		return errors.New("writer is nil")
	}

	_, err := writer.Write(p.content)
	return errors.Wrap(err, "failed to write part content to message")
}

func (p simplePart) Header() Header {
	return p.header
}
