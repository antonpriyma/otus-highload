package dwh

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/fatih/structs"
)

//go:generate mockgen -destination=./mock/mock_generated.go -package=mock github.com/antonpriyma/otus-highload/pkg/clients/dwh Client

const (
	dwhTag                  = "dwh"
	compactMessageSeparator = "||"
	prefixLength            = 2
)

var Now = time.Now

type Client interface {
	SendMessage(ctx context.Context, msg Message) error
	SendCompactMessage(ctx context.Context, msg CompactMessage) error
}

type Message interface {
	MarshalDWH() ([]byte, error)
}

type CompactMessage interface {
	MessageType() string
}

type TypeID int

type EventMessage struct {
	TimestampNano int64 `structs:"event_date"`

	Type    TypeID `structs:"event_type_id"`
	Product string `structs:"product"`

	Email     string `structs:"email,omitempty"`
	UserID    string `structs:"user_id,omitempty"`
	UserAgent string `structs:"useragent,omitempty"`
	IP        string `structs:"ip"`

	AppleIdentifierForAdvertisers string `structs:"idfa,omitempty"`
	GoogleAdvertisingID           string `structs:"gaid,omitempty"`
	VID                           string `structs:"vid,omitempty"` // value until first ":"
	UTM                           string `structs:"utm,omitempty"`

	Platform   string `structs:"platform,omitempty"`
	OSVersion  string `structs:"osversion,omitempty"`
	AppVersion string `structs:"appversion,omitempty"`

	ExtraFields map[string]interface{} `structs:"-"`
}

func (m EventMessage) toMap() map[string]interface{} {
	s := structs.New(m)
	res := s.Map()

	for k, v := range m.ExtraFields {
		res[k] = v
	}

	return res
}

func (m EventMessage) MarshalDWH() ([]byte, error) {
	return json.Marshal(m.toMap())
}

// hostname [DWH]тип_события||дата_события(timestamp)||поле1||поле2||...[/DWH]
func MarshalCompactMessage(m CompactMessage, hostname string) (string, error) {
	content, err := extractFields(m)
	if err != nil {
		return "", errors.Wrap(err, "failed to get dwh message content")
	}

	return fmt.Sprintf("%s [DWH]%s[/DWH]", hostname, strings.Join(content, compactMessageSeparator)), nil
}

func extractFields(m CompactMessage) (res []string, err error) {
	fields := expandNestedFields(structs.Fields(m))
	prefix := prefixCompactMessage(m)
	tags := newTagsParser(len(fields))

	res = make([]string, len(fields)+len(prefix))
	copy(res, prefix)

	for _, field := range fields {
		order, err := tags.ParseAndCheckOrder(field.Tag(dwhTag))
		if err != nil {
			return nil, errors.Wrapf(err, "wrong order %v in field %v", field.Tag(dwhTag), field)
		}

		res[order+prefixLength] = fmt.Sprintf("%v", field.Value())
	}

	return res, nil
}

func expandNestedFields(fields []*structs.Field) (expandedFields []*structs.Field) {
	for _, field := range fields {
		switch field.Kind() {
		case reflect.Struct:
			expandedFields = append(expandedFields, expandNestedFields(field.Fields())...)
		default:
			expandedFields = append(expandedFields, field)
		}
	}

	return expandedFields
}

type tagsParser struct {
	exists map[int]bool
	max    int
}

func newTagsParser(max int) tagsParser {
	return tagsParser{
		exists: make(map[int]bool, max),
		max:    max,
	}
}

func prefixCompactMessage(m CompactMessage) []string {
	return []string{
		m.MessageType(),
		fmt.Sprint(Now().Unix()),
	}
}

func (t *tagsParser) ParseAndCheckOrder(tag string) (order int, err error) {
	order, err = strconv.Atoi(tag)
	if err != nil {
		return 0, errors.Errorf("failed to convert convert tag %v to number order", tag)
	}

	if order >= t.max || order < 0 {
		return 0, errors.Errorf("order from tag %v is not in permissible range 0 - %d", tag, t.max)
	}

	if t.exists[order] {
		return 0, errors.Errorf("duplicated tag %v in dwh message", tag)
	}

	t.exists[order] = true
	return order, nil

}
