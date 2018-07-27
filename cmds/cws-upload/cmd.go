package cws_upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	extensionIDKey   = "extension.id"
	clientIDKey      = "google.client.id"
	clientSecretKey  = "google.client.secret"
	refreshTokenKey  = "google.refresh.token"
	zipPathKey       = "zipPath"
	publishKey       = "publish"
	publishTargetKey = "target"

	rootURI         = "https://www.googleapis.com"
	refreshTokenURI = "https://www.googleapis.com/oauth2/v4/token"
)

var (
	log             *zap.Logger
	targetAllowed   = []string{"default", "trustedTesters"}
	requiredFields  = []string{extensionIDKey, clientIDKey, clientSecretKey, refreshTokenKey, zipPathKey}
	allowedFileType = []string{"application/zip"}
)

type AccessToken struct {
	Token string `json:"access_token"`
	Type  string `json:"token_type"`
}

type ItemResource struct {
	ID          string
	Kind        string
	UploadState string      `json:"uploadState"`
	ItemError   []ErrorItem `json:"itemError"`
}

type ErrorItem struct {
	Code   string `json:"error_code"`
	Detail string `json:"error_detail"`
}

func (i *ItemResource) LogFields() (field []zapcore.Field) {
	field = []zapcore.Field{
		zap.String("ID", i.ID),
		zap.String("Kind", i.Kind),
		zap.String("UploadState", i.UploadState),
	}

	for i, err := range i.ItemError {
		field = append(field, zap.String(fmt.Sprintf("error_%d_code", i), err.Code))
		field = append(field, zap.String(fmt.Sprintf("error_%d_detail", i), err.Detail))
	}

	return
}

func InitCommand(zapLogger *zap.Logger) *cobra.Command {
	log = zapLogger
	CobraCommand := &cobra.Command{
		Use:   "upload",
		Short: "Upload a zip file into CWS",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.Wrap(proceed(), "proceed")
		},
	}

	// Define flag for zip path
	CobraCommand.PersistentFlags().StringP(zipPathKey, "z", "", "CWS zip file path")
	CobraCommand.MarkFlagRequired(zipPathKey)
	viper.BindPFlag(zipPathKey, CobraCommand.PersistentFlags().Lookup(zipPathKey))

	// Define flag for publish optional
	CobraCommand.PersistentFlags().BoolP(publishKey, "p", false, "Publish CWS item immediately after zip file uploaded")
	CobraCommand.MarkFlagRequired(publishKey)
	viper.BindPFlag(publishKey, CobraCommand.PersistentFlags().Lookup(publishKey))

	// Define flag for publish target
	CobraCommand.PersistentFlags().StringP(publishTargetKey, "t", targetAllowed[0], "Publish target (trustedTesters/default)")
	CobraCommand.MarkFlagRequired(publishTargetKey)
	viper.BindPFlag(publishTargetKey, CobraCommand.PersistentFlags().Lookup(publishTargetKey))

	return CobraCommand
}

func proceed() error {
	for _, field := range requiredFields {
		if len(viper.GetString(field)) == 0 {
			return errors.New(field + " is required")
		}
	}

	log.Debug("Validate zip file should matched with allowed mime type")
	file, err := os.Open(viper.GetString(zipPathKey))
	if err != nil {
		return errors.Wrap(err, "os.Open")
	}
	defer file.Close()

	if err := validateFile(file); err != nil {
		return errors.Wrap(err, "validateFile")
	}

	log.Debug("Uploading the file ...")
	if err := upload(file); err != nil {
		return errors.Wrap(err, "upload")
	}

	if viper.GetBool(publishKey) {
		log.Debug("Going to publish item immediately ...")
		return errors.Wrap(publish(), "publish")
	}

	return nil
}

func validateFile(file *os.File) error {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return errors.Wrap(err, "file.Read")
	}

	// Reset the read pointer if necessary.
	file.Seek(0, 0)

	contentType := http.DetectContentType(buffer)
	log.Debug("Got file with type: " + contentType)

	for _, allowed := range allowedFileType {
		if allowed == contentType {
			return nil
		}
	}

	return errors.New("Zip file is invalid")
}

func getHeader() (header http.Header, err error) {
	header = make(http.Header)
	accessToken, err := fetchToken()
	if err != nil {
		err = errors.Wrap(err, "fetchToken")
		return
	}

	header.Add("Authorization", accessToken.Type+" "+accessToken.Token)
	return
}

// publish method used to publish immediately
func publish() error {
	header, err := getHeader()
	if err != nil {
		return errors.Wrap(err, "getHeader")
	}

	var b bytes.Buffer
	extensionId := viper.GetString(extensionIDKey)
	targetPublish := viper.GetString(publishTargetKey)
	targetVerified := false
	for _, v := range targetAllowed {
		if v == targetPublish {
			targetVerified = true
			break
		}
	}

	if !targetVerified {
		return errors.New("Publish target is invalid")
	}

	publishUri := strings.Join([]string{rootURI, "/chromewebstore/v1.1/items/", extensionId, "/publish?publishTarget=", targetPublish}, "")
	req, err := http.NewRequest("POST", publishUri, &b)
	if err != nil {
		return errors.Wrap(err, "http.NewRequest")
	}

	req.Header = header

	return errors.Wrap(fetchResponseFromCws(req), "fetchResponseFromCws")
}

// upload method used to upload file reader into CWS
func upload(file *os.File) error {
	header, err := getHeader()
	if err != nil {
		return errors.Wrap(err, "getHeader")
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", file.Name())
	if err != nil {
		return errors.Wrap(err, "w.CreateFormFile")
	}

	if _, err = io.Copy(fw, file); err != nil {
		return errors.Wrap(err, "io.Copy")
	}
	w.Close()

	extensionId := viper.GetString(extensionIDKey)
	req, err := http.NewRequest("PUT", strings.Join([]string{rootURI, "/upload/chromewebstore/v1.1/items/", extensionId}, ""), &b)
	if err != nil {
		return errors.Wrap(err, "http.NewRequest")
	}

	req.Header = header
	req.Header.Add("Content-Type", w.FormDataContentType())

	return errors.Wrap(fetchResponseFromCws(req), "fetchResponseFromCws")
}

func fetchResponseFromCws(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http.DefaultClient.Do")
	}

	if resp.StatusCode != 200 {
		return errors.New("Failed to request. Status code: " + string(resp.StatusCode))
	}

	decoder := json.NewDecoder(resp.Body)
	var itemResource ItemResource
	if err := decoder.Decode(&itemResource); err != nil {
		return errors.Wrap(err, "decoder.Decode")
	}

	log.With(itemResource.LogFields()...).Debug("Request completed")

	return nil
}

func fetchToken() (*AccessToken, error) {
	log.Debug("Going to refresh token")
	data := url.Values{}
	data.Add("client_id", viper.GetString(clientIDKey))
	data.Add("client_secret", viper.GetString(clientSecretKey))
	data.Add("refresh_token", viper.GetString(refreshTokenKey))
	data.Add("grant_type", "refresh_token")

	resp, err := http.PostForm(refreshTokenURI, data)
	if err != nil {
		return nil, errors.Wrap(err, "http.PostForm")
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("Failed to refresh token")
	}

	decoder := json.NewDecoder(resp.Body)
	var accessToken AccessToken
	if err := decoder.Decode(&accessToken); err != nil {
		return nil, errors.Wrap(err, "decoder.Decode")
	}

	return &accessToken, nil
}
