package s3

import (
	"context"
	"io"
	"net/textproto"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
)

type Client interface {
	CheckFile(ctx context.Context, key string) (bool, error)
	Upload(ctx context.Context, key string, body io.ReadSeeker, extraHeaders textproto.MIMEHeader) error
	UploadFolder(ctx context.Context, key string) error
	GetObjectSize(ctx context.Context, key string) (int64, error)
	GetHead(ctx context.Context, key string) (ObjectHead, error)
	Delete(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (Object, error)
}

type Object struct {
	ContentType        *string
	ContentLength      *int64
	ETag               *string
	ContentDisposition *string

	Body io.ReadCloser
}

type ObjectHead struct {
	ContentLength *int64
	ContentType   *string
}

type Config struct {
	Endpoint string `mapstructure:"endpoint"`
	Region   string `mapstructure:"region"`
	Bucket   string `mapstructure:"bucket"`

	// How to form path to bucket
	// true -> <domain>/bucket, false -> <bucket>.<domain>
	S3ForcePathStyle bool `mapstructure:"force_path_style"`
	DisableSSL       bool `mapstructure:"disable_ssl"`

	Key    string `mapstructure:"key"`
	Secret string `mapstructure:"secret"`

	Permissions string `mapstructure:"permissions"`
}

type client struct {
	Config Config
	Logger log.Logger
	Stat   Stat
	s3     *s3.S3
}

type Stat struct {
	RequestDuration stat.TimerCtor `labels:"status,method"`
}

var ErrNotFoundObject = errors.New("object not found")

func NewClient(config Config, logger log.Logger, registry stat.Registry) Client {
	awsConfig := aws.Config{
		Endpoint:         &config.Endpoint,
		Region:           &config.Region,
		Credentials:      credentials.NewStaticCredentials(config.Key, config.Secret, ""),
		DisableSSL:       aws.Bool(config.DisableSSL),
		S3ForcePathStyle: aws.Bool(config.S3ForcePathStyle),
	}

	session := session.Must(session.NewSession(&awsConfig))

	s3Client := s3.New(session)

	ret := client{
		Config: config,
		Logger: logger,
		s3:     s3Client,
	}

	stat.NewRegistrar(registry.ForSubsystem("s3")).MustRegister(&ret.Stat)

	return ret
}

func (c client) Upload(
	ctx context.Context,
	key string,
	body io.ReadSeeker,
	extraHeaders textproto.MIMEHeader,
) (err error) {
	timer := c.Stat.RequestDuration.Timer(ctx).Start()
	defer func() {
		timer.WithLabels(stat.Labels{
			"status": stat.TypedErrorLabel(ctx, err),
			"method": "Upload",
		}).Stop()
	}()

	putObjectInput := &s3.PutObjectInput{
		ACL:    aws.String(c.Config.Permissions),
		Body:   body,
		Key:    aws.String(key),
		Bucket: aws.String(c.Config.Bucket),
	}

	if header := extraHeaders.Get(HeaderContentDisposition); header != "" {
		putObjectInput.SetContentDisposition(header)
	}
	if header := extraHeaders.Get(HeaderContentType); header != "" {
		putObjectInput.SetContentType(header)
	}

	_, err = c.s3.PutObject(putObjectInput)
	return errors.Wrap(c.handleAWSError(err), "failed to put object")
}

func (c client) UploadFolder(ctx context.Context, key string) (err error) {
	timer := c.Stat.RequestDuration.Timer(ctx).Start()
	defer func() {
		timer.WithLabels(stat.Labels{
			"status": stat.TypedErrorLabel(ctx, err),
			"method": "UploadFolder",
		}).Stop()
	}()

	_, err = c.s3.PutObject(&s3.PutObjectInput{
		ACL:    aws.String(c.Config.Permissions),
		Key:    aws.String(key + "/"),
		Bucket: aws.String(c.Config.Bucket),
	})

	return errors.Wrap(c.handleAWSError(err), "failed to put empty object")
}

func (c client) CheckFile(ctx context.Context, key string) (bool, error) {
	_, err := c.GetHead(ctx, key)

	if errors.Is(err, ErrNotFoundObject) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "failed to check file")
	}

	return true, nil
}

func (c client) GetObjectSize(ctx context.Context, key string) (size int64, err error) {
	timer := c.Stat.RequestDuration.Timer(ctx).Start()
	defer func() {
		timer.WithLabels(stat.Labels{
			"status": stat.TypedErrorLabel(ctx, err),
			"method": "GetObjectSize",
		}).Stop()
	}()

	list, err := c.s3.ListObjects(
		&s3.ListObjectsInput{
			Bucket: aws.String(c.Config.Bucket),
			Prefix: aws.String(key),
		})

	if err != nil {
		return 0, errors.Wrap(c.handleAWSError(err), "failed to get object list")
	}

	totalSize := int64(0)

	for _, value := range list.Contents {
		totalSize += *value.Size
	}

	return totalSize, nil
}

func (c client) GetHead(ctx context.Context, key string) (resp ObjectHead, err error) {
	timer := c.Stat.RequestDuration.Timer(ctx).Start()
	defer func() {
		timer.WithLabels(stat.Labels{
			"status": stat.TypedErrorLabel(ctx, err),
			"method": "GetHead",
		}).Stop()
	}()

	objectHead, err := c.s3.HeadObject(
		&s3.HeadObjectInput{
			Key:    aws.String(key),
			Bucket: aws.String(c.Config.Bucket),
		})
	if err != nil {
		return ObjectHead{}, errors.Wrap(c.handleAWSError(err), "failed to head object")
	}

	return ObjectHead{
		ContentLength: objectHead.ContentLength,
		ContentType:   objectHead.ContentType,
	}, nil
}

func (c client) Delete(ctx context.Context, key string) (err error) {
	timer := c.Stat.RequestDuration.Timer(ctx).Start()
	defer func() {
		timer.WithLabels(stat.Labels{
			"status": stat.TypedErrorLabel(ctx, err),
			"method": "Delete",
		}).Stop()
	}()

	_, err = c.s3.DeleteObject(
		&s3.DeleteObjectInput{
			Key:    aws.String(key),
			Bucket: aws.String(c.Config.Bucket),
		})

	return errors.Wrap(c.handleAWSError(err), "failed to delete object")
}

func (c client) Get(ctx context.Context, key string) (resp Object, err error) {
	timer := c.Stat.RequestDuration.Timer(ctx).Start()
	defer func() {
		timer.WithLabels(stat.Labels{
			"status": stat.TypedErrorLabel(ctx, err),
			"method": "Get",
		}).Stop()
	}()

	object, err := c.s3.GetObject(
		&s3.GetObjectInput{
			Key:    aws.String(key),
			Bucket: aws.String(c.Config.Bucket),
		})
	if err != nil {
		return Object{}, errors.Wrap(c.handleAWSError(err), "failed to get object")
	}

	return Object{
		ContentType:        object.ContentType,
		ContentLength:      object.ContentLength,
		ETag:               object.ETag,
		ContentDisposition: object.ContentDisposition,
		Body:               object.Body,
	}, nil
}

func (c client) handleAWSError(err error) error {
	if err == nil {
		return nil
	}

	var aerr awserr.Error
	if errors.As(err, &aerr) {
		switch aerr.Code() {
		case "NoSuchKey", "NotFound":
			return errors.Wrap(ErrNotFoundObject, err.Error())
		}
	}

	return err
}
