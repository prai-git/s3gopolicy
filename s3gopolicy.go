package s3gopolicy

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AWSCredentials Amazon Credentials
type AWSCredentials struct {
	AWSSecretKeyID string
	AWSAccessKeyID string
}

// UploadConfig generate policies from config
type UploadConfig struct {
	BucketName  string
	ObjectKey   string
	ContentType string
	FileSize    int64
}

// UploadPolicies Amazon s3 upload policies
type UploadPolicies struct {
	URL  string
	Form map[string]string
}

// PolicyJSON is policy rule
type PolicyJSON struct {
	Expiration string        `json:"expiration"`
	Conditions []interface{} `json:"conditions"`
}

const expirationTimeFormat = "2006-01-02T15:04:05ZZ07:00"
const expirationHour = 1 * time.Hour
const uploadURLFormat = "http://%s.s3.amazonaws.com/" // <bucketName>

// NowTime mockable time.Now()
var NowTime = func() time.Time {
	return time.Now()
}

// CreatePolicies create amazon s3 to upload policies return
func CreatePolicies(awsCredentials AWSCredentials, fileInfo UploadConfig) (UploadPolicies, error) {
	data, err := json.Marshal(&PolicyJSON{
		Expiration: NowTime().Add(expirationHour).Format(expirationTimeFormat),
		Conditions: []interface{}{
			map[string]string{"bucket": fileInfo.BucketName},
			map[string]string{"key": fileInfo.ObjectKey},
			map[string]string{"Content-Type": fileInfo.ContentType},
			[]interface{}{"content-length-range", fileInfo.FileSize, fileInfo.FileSize},
		},
	})
	if err != nil {
		return UploadPolicies{}, err
	}

	policy := strings.Replace(base64.StdEncoding.EncodeToString(data), "\n", "", -1)
	mac := hmac.New(sha1.New, []byte(awsCredentials.AWSSecretKeyID))
	mac.Write([]byte(policy))
	expectedMAC := mac.Sum(nil)
	signature := strings.Replace(base64.StdEncoding.EncodeToString(expectedMAC), "\n", "", -1)

	uploadURL := fmt.Sprintf(uploadURLFormat, fileInfo.BucketName)
	return UploadPolicies{
		URL: uploadURL,
		Form: map[string]string{
			"AWSAccessKeyId": awsCredentials.AWSAccessKeyID,
			"key":            fileInfo.ObjectKey,
			"Content-Type":   fileInfo.ContentType,
			"signature":      signature,
			"policy":         policy,
		},
	}, nil
}