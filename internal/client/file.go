package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// UploadResult is resultObject of POST /file/uploadSingleFile.
type UploadResult struct {
	FileID           string `json:"fileId"`
	StoreType        string `json:"storeType"`
	FilePathInServer string `json:"filePathInServer"`
	FileName         string `json:"fileName"`
	FileDesc         string `json:"fileDesc,omitempty"`
	CreateDate       string `json:"createDate"`
	StatusCd         string `json:"statusCd"`
	StatusDate       string `json:"statusDate"`
	AppID            string `json:"appId"`
	FileSize         string `json:"fileSize,omitempty"`
	FileType         string `json:"fileType,omitempty"`
	IsPicture        string `json:"isPicture,omitempty"`
}

// UploadFile uploads a local file via multipart POST /file/uploadSingleFile.
// When compress is true, form field compressFile=true is sent.
func (c *Client) UploadFile(path string, compress bool) (*UploadResult, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("文件路径不能为空")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("构造 multipart 失败: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	if compress {
		if err := w.WriteField("compressFile", "true"); err != nil {
			return nil, fmt.Errorf("写入 compressFile 失败: %w", err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("关闭 multipart 失败: %w", err)
	}

	raw, err := c.doMultipartPOST("/file/uploadSingleFile", buf.Bytes(), w.FormDataContentType())
	if err != nil {
		return nil, err
	}
	var result UploadResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parsing upload response: %w", err)
	}
	if strings.TrimSpace(result.FileID) == "" {
		return nil, fmt.Errorf("上传成功但未返回 fileId")
	}
	if result.AppID != "" {
		c.lastAppID = result.AppID
	}
	return &result, nil
}

// DownloadFile downloads a file by fileId via GET /file/getFile/{fileId} and writes binary to destPath.
// Follows redirects. Fails clearly on JSON error body or HTTP 404.
func (c *Client) DownloadFile(fileID, destPath string) error {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return fmt.Errorf("fileId 不能为空")
	}
	destPath = strings.TrimSpace(destPath)
	if destPath == "" {
		return fmt.Errorf("目标路径不能为空")
	}
	path := "/file/getFile/" + url.PathEscape(fileID)
	return c.doBinaryGET(path, destPath, nil)
}
