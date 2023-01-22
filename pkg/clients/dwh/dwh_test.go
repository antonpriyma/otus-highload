package dwh

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestEventMessageToMap(t *testing.T) {
	tests := []struct {
		name         string
		eventMessage EventMessage
		want         map[string]interface{}
	}{
		{
			name: "with extra fields",
			eventMessage: EventMessage{
				TimestampNano:                 123,
				Type:                          1,
				Product:                       "some product",
				UserID:                        "some id",
				UserAgent:                     "some agent",
				IP:                            "some ip",
				AppleIdentifierForAdvertisers: "some idfa",
				GoogleAdvertisingID:           "some gaid",
				VID:                           "some vid",
				UTM:                           "some utm",
				Platform:                      "some platform",
				OSVersion:                     "some os version",
				AppVersion:                    "some app version",
				ExtraFields: map[string]interface{}{
					"some extra 1": "some extra 1",
					"some":         "body",
				},
			},
			want: map[string]interface{}{
				"event_date":    int64(123),
				"event_type_id": TypeID(1),
				"product":       "some product",
				"user_id":       "some id",
				"useragent":     "some agent",
				"ip":            "some ip",
				"idfa":          "some idfa",
				"gaid":          "some gaid",
				"vid":           "some vid",
				"utm":           "some utm",
				"platform":      "some platform",
				"osversion":     "some os version",
				"appversion":    "some app version",
				"some extra 1":  "some extra 1",
				"some":          "body",
			},
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(tt *testing.T) {
			got := testCase.eventMessage.toMap()
			require.Equal(tt, testCase.want, got)
		})
	}
}

type ImportPublish struct {
	User         string `dwh:"0"`
	ICSSender    string `dwh:"1"`
	ICSOrganizer string `dwh:"2"`
	EventUID     string `dwh:"3"`
	Dtstart      string `dwh:"4"`
}

func (i ImportPublish) MessageType() string {
	return "add_event_from_ics"
}

type ReminderForPublish struct {
	ImportPublish ImportPublish
	NotifyType    string `dwh:"5"`
}

func (r ReminderForPublish) MessageType() string {
	return "remind_event_from_ics"
}

type DuplicatedTags struct {
	User         string `dwh:"0"`
	ICSSender    string `dwh:"1"`
	ICSOrganizer string `dwh:"1"`
}

func (i DuplicatedTags) MessageType() string {
	return "duplicated_tags"
}

type OutOfRangeTags struct {
	User         string `dwh:"0"`
	ICSSender    string `dwh:"1"`
	ICSOrganizer string `dwh:"3"`
}

func (i OutOfRangeTags) MessageType() string {
	return "out_of_range_tags"
}

type WrongTagType struct {
	User         string `dwh:"wrong"`
	ICSSender    string `dwh:"1"`
	ICSOrganizer string `dwh:"2"`
}

func (i WrongTagType) MessageType() string {
	return "wrong_tag_type"
}

func TestMarshalCompactMessage(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)

	Now = func() time.Time {
		return utils.MustParseTimeNano("2020-08-01T14:30:00+03:00")
	}

	tests := []struct {
		name         string
		eventMessage CompactMessage
		want         string
	}{
		{
			name: "import publish",
			eventMessage: ImportPublish{
				"recipient@mail.ru",
				"sender@mail.ru",
				"organizer@mail.ru",
				"uid",
				"some_dtstart",
			},
			want: hostname + " [DWH]add_event_from_ics||" + fmt.Sprint(Now().Unix()) + "||recipient@mail.ru||sender@mail.ru||organizer@mail.ru||uid||some_dtstart[/DWH]",
		},
		{
			name: "reminder for publish",
			eventMessage: ReminderForPublish{
				ImportPublish{
					"recipient@mail.ru",
					"sender@mail.ru",
					"organizer@mail.ru",
					"uid",
					"some_dtstart",
				},
				"email",
			},
			want: hostname + " [DWH]remind_event_from_ics||" + fmt.Sprint(Now().Unix()) + "||recipient@mail.ru||sender@mail.ru||organizer@mail.ru||uid||some_dtstart||email[/DWH]",
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(tt *testing.T) {
			got, err := MarshalCompactMessage(testCase.eventMessage, hostname)

			require.NoError(tt, err)
			require.Equal(tt, testCase.want, got)
		})
	}
}

func TestMarshalCompactMessageError(t *testing.T) {
	Now = func() time.Time {
		return utils.MustParseTimeNano("2020-08-01T14:30:00+03:00")
	}

	tests := []struct {
		name         string
		eventMessage CompactMessage
		errMsg       string
	}{
		{
			name: "duplicated tags",
			eventMessage: DuplicatedTags{
				"recipient@mail.ru",
				"sender@mail.ru",
				"organizer@mail.ru",
			},
			errMsg: "duplicated tag",
		},
		{
			name: "out of range",
			eventMessage: OutOfRangeTags{
				"recipient@mail.ru",
				"sender@mail.ru",
				"organizer@mail.ru",
			},
			errMsg: "is not in permissible range",
		}, {
			name: "wrong tag type",
			eventMessage: WrongTagType{
				"recipient@mail.ru",
				"sender@mail.ru",
				"organizer@mail.ru",
			},
			errMsg: "failed to convert convert tag",
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(tt *testing.T) {
			_, err := MarshalCompactMessage(testCase.eventMessage, "")

			require.Error(tt, err)
			require.Contains(tt, err.Error(), testCase.errMsg)
		})
	}
}
