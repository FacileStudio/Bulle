package suite

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	nuageFilesPath   = "/api/files"
	nuageFoldersPath = "/api/folders"
	nuageSearchPath  = "/api/search"
	nuageReadLimit   = 100 << 10
	nuageSearchHits  = 25
)

type NuageFile struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	Size      int64  `json:"size"`
	FolderID  *int64 `json:"folder_id"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type NuageFolder struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ParentID  *int64 `json:"parent_id"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type NuageListing struct {
	Folders []NuageFolder `json:"folders"`
	Files   []NuageFile   `json:"files"`
}

type NuageHit struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Path      string `json:"path"`
	MimeType  string `json:"mime_type,omitempty"`
	Size      int64  `json:"size,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type NuageRead struct {
	File      NuageFile `json:"file"`
	Content   string    `json:"content,omitempty"`
	Truncated bool      `json:"truncated,omitempty"`
	Note      string    `json:"note,omitempty"`
}

// Nuage reaches the user's personal file area. Team spaces are left out in v1.
type Nuage struct{ gw *Gateway }

func NewNuage(gw *Gateway) Nuage { return Nuage{gw: gw} }

func (n Nuage) List(ctx context.Context, userID, folder int64) (NuageListing, error) {
	files, folders := url.Values{}, url.Values{}
	if folder > 0 {
		files.Set("folder_id", strconv.FormatInt(folder, 10))
		folders.Set("parent_id", strconv.FormatInt(folder, 10))
	}
	var listing NuageListing
	if err := n.getInto(ctx, userID, nuageFilesPath, files, &listing); err != nil {
		return NuageListing{}, err
	}
	if err := n.getInto(ctx, userID, nuageFoldersPath, folders, &listing); err != nil {
		return NuageListing{}, err
	}
	return listing, nil
}

func (n Nuage) Search(ctx context.Context, userID int64, query string) ([]NuageHit, error) {
	var out struct {
		Results []NuageHit `json:"results"`
	}
	q := url.Values{"q": {query}, "limit": {strconv.Itoa(nuageSearchHits)}}
	if err := n.getInto(ctx, userID, nuageSearchPath, q, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// Read answers a file's metadata, and its text when it is a text file. The
// model cannot see an image or a PDF from here, so those come back with a
// note telling it how the user can show them.
func (n Nuage) Read(ctx context.Context, userID, id int64) (NuageRead, error) {
	var file NuageFile
	if err := n.getInto(ctx, userID, fmt.Sprintf("%s/%d", nuageFilesPath, id), nil, &file); err != nil {
		return NuageRead{}, err
	}
	out := NuageRead{File: file}
	if strings.HasPrefix(file.MimeType, "image/") || file.MimeType == "application/pdf" {
		out.Note = "You cannot see images or PDFs through this tool. Ask the user to attach the file to a message if you need to look at it."
		return out, nil
	}
	data, err := n.gw.Get(ctx, userID, AppNuage, fmt.Sprintf("%s/%d/download", nuageFilesPath, id), nil)
	if err != nil {
		return NuageRead{}, err
	}
	text, ok := asText(data, nuageReadLimit)
	if !ok {
		out.Note = "This is a binary file and cannot be read as text."
		return out, nil
	}
	out.Content, out.Truncated = text, len(data) > nuageReadLimit
	return out, nil
}

func asText(data []byte, limit int) (string, bool) {
	if len(data) > limit {
		data = data[:limit]
		for i := 0; i < utf8.UTFMax && !utf8.Valid(data); i++ {
			data = data[:len(data)-1]
		}
	}
	if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return "", false
	}
	return string(data), true
}

// Save uploads a file into the user's personal area, at the root or in a folder.
func (n Nuage) Save(ctx context.Context, userID, folder int64, file File) (NuageFile, error) {
	body, contentType, err := uploadForm(folder, file)
	if err != nil {
		return NuageFile{}, err
	}
	raw, err := n.gw.Do(ctx, userID, Request{App: AppNuage, Method: http.MethodPost, Path: nuageFilesPath, Raw: body, ContentType: contentType})
	if err != nil {
		return NuageFile{}, err
	}
	var saved NuageFile
	if err := decodeInto(raw, &saved); err != nil {
		return NuageFile{}, fmt.Errorf("nuage: reading the saved file: %w", err)
	}
	return saved, nil
}

func uploadForm(folder int64, file File) ([]byte, string, error) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, file.Name))
	header.Set("Content-Type", file.MediaType)
	part, err := form.CreatePart(header)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(file.Data); err != nil {
		return nil, "", err
	}
	_ = form.WriteField("origin_app", "bulle")
	if folder > 0 {
		_ = form.WriteField("folder_id", strconv.FormatInt(folder, 10))
	}
	if err := form.Close(); err != nil {
		return nil, "", err
	}
	return body.Bytes(), form.FormDataContentType(), nil
}

func (n Nuage) getInto(ctx context.Context, userID int64, path string, query url.Values, target any) error {
	raw, err := n.gw.Get(ctx, userID, AppNuage, path, query)
	if err != nil {
		return err
	}
	if err := decodeInto(raw, target); err != nil {
		return fmt.Errorf("nuage: reading %s: %w", path, err)
	}
	return nil
}
