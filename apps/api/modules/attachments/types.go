package attachments

type AttachmentResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	Size      int64  `json:"size"`
}

type UploadRequest struct {
	File string `json:"file"`
}
