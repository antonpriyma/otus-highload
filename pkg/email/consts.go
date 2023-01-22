package email

const (
	ContentTypeAlternative = "multipart/alternative"
	ContentTypeMixed       = "multipart/mixed"
	ContentTypeHTML        = "text/html"
	ContentTypePlain       = "text/plain"
	ContentTypeAMP         = "text/x-amp-html"
	ContentTypeICS         = "application/ics"
	ContentTypeCalendar    = "text/calendar"
)

const (
	HeaderContentType             = "Content-Type"
	HeaderContentDisposition      = "Content-Disposition"
	HeaderContentTransferEncoding = "Content-Transfer-Encoding"
)
