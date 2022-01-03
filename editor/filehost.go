package editor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type Filehost struct {
	Client *http.Client
	Domain string
	Pass   string
}

func (f Filehost) Upload(data interface{}) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// this can't error out
	fw, _ := writer.CreateFormFile("file", "urlib.json")
	json.NewEncoder(fw).Encode(data)

	fw, _ = writer.CreateFormField("pass")
	fmt.Fprint(fw, f.Pass)

	writer.Close()

	res, err := f.Client.Post(f.Domain+"/upload", writer.FormDataContentType(), body)
	if err != nil {
		return "", fmt.Errorf("uploading: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("uploading code: %v", res.StatusCode)
	}

	var code string
	if _, err := fmt.Fscanf(res.Body, f.Domain+"/hosted/%s", &code); err != nil {
		return "", fmt.Errorf("scanning code: %v", err)
	}

	return strings.TrimSuffix(code, ".json"), nil
}

func (f Filehost) Apply(code string) (io.ReadCloser, error) {
	res, err := f.Client.Get(f.Domain + "/hosted/" + code + ".json")
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad filehost status code %d", res.StatusCode)
	}

	return res.Body, nil
}

func (f Filehost) ApplyCode() string {
	return f.Pass
}
