package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/leahgarrett/image-management-system/services/backend/internal/db"
	"github.com/sqlc-dev/pqtype"
)

type imageMetadata struct {
	CaptureDate string `json:"captureDate"`
	CameraMake  string `json:"cameraMake"`
	CameraModel string `json:"cameraModel"`
}

type registerImageRequest struct {
	ImageID          string        `json:"imageId"`
	OriginalFilename string        `json:"originalFilename"`
	ThumbnailKey     string        `json:"thumbnailKey"`
	WebKey           string        `json:"webKey"`
	OriginalKey      string        `json:"originalKey"`
	ThumbnailSize    int64         `json:"thumbnailSize"`
	WebSize          int64         `json:"webSize"`
	OriginalSize     int64         `json:"originalSize"`
	Width            int32         `json:"width"`
	Height           int32         `json:"height"`
	Metadata         imageMetadata `json:"metadata"`
}

type imageResponse struct {
	ID               string        `json:"id"`
	ImageID          string        `json:"imageId"`
	OriginalFilename string        `json:"originalFilename"`
	ThumbnailKey     string        `json:"thumbnailKey"`
	WebKey           string        `json:"webKey"`
	OriginalKey      string        `json:"originalKey"`
	ThumbnailSize    int64         `json:"thumbnailSize"`
	WebSize          int64         `json:"webSize"`
	OriginalSize     int64         `json:"originalSize"`
	Width            int32         `json:"width"`
	Height           int32         `json:"height"`
	UploadedAt       time.Time     `json:"uploadedAt"`
	Published        bool          `json:"published"`
	People           []string      `json:"people"`
	Tags             []string      `json:"tags"`
	DateType         string        `json:"dateType,omitempty"`
	ExactDate        string        `json:"exactDate,omitempty"`
	OccasionCategory string        `json:"occasionCategory,omitempty"`
	OccasionName     string        `json:"occasionName,omitempty"`
	Metadata         imageMetadata `json:"metadata"`
}

func imageToResponse(img db.Image, people []db.ImagePerson, tags []db.Tag) imageResponse {
	resp := imageResponse{
		ID:               img.ID.String(),
		ImageID:          img.ImageID,
		OriginalFilename: img.OriginalFilename.String,
		ThumbnailKey:     img.ThumbnailKey.String,
		WebKey:           img.WebKey.String,
		OriginalKey:      img.OriginalKey.String,
		ThumbnailSize:    img.ThumbnailSize.Int64,
		WebSize:          img.WebSize.Int64,
		OriginalSize:     img.OriginalSize.Int64,
		Width:            img.Width.Int32,
		Height:           img.Height.Int32,
		UploadedAt:       img.UploadedAt,
		Published:        img.Published,
		DateType:         img.DateType.String,
		OccasionCategory: img.OccasionCategory.String,
		OccasionName:     img.OccasionName.String,
		People:           []string{},
		Tags:             []string{},
	}
	if img.ExactDate.Valid {
		resp.ExactDate = img.ExactDate.Time.Format("2006-01-02")
	}
	for _, p := range people {
		resp.People = append(resp.People, p.Name)
	}
	for _, tg := range tags {
		resp.Tags = append(resp.Tags, tg.Name)
	}
	return resp
}

// RegisterImage handles POST /api/v1/images.
func (h *Handlers) RegisterImage(w http.ResponseWriter, r *http.Request) {
	var req registerImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, errValidation("invalid request body"))
		return
	}
	if req.ImageID == "" {
		writeError(w, r, errValidation("imageId is required"))
		return
	}

	userIDStr, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, r, errUnauthorized("missing user context"))
		return
	}

	parsedUID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, r, errInternal("invalid user ID"))
		return
	}

	metaJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		writeError(w, r, errInternal("failed to marshal metadata"))
		return
	}

	params := db.CreateImageParams{
		ImageID:          req.ImageID,
		OriginalFilename: sql.NullString{String: req.OriginalFilename, Valid: req.OriginalFilename != ""},
		ThumbnailKey:     sql.NullString{String: req.ThumbnailKey, Valid: req.ThumbnailKey != ""},
		WebKey:           sql.NullString{String: req.WebKey, Valid: req.WebKey != ""},
		OriginalKey:      sql.NullString{String: req.OriginalKey, Valid: req.OriginalKey != ""},
		ThumbnailSize:    sql.NullInt64{Int64: req.ThumbnailSize, Valid: req.ThumbnailSize > 0},
		WebSize:          sql.NullInt64{Int64: req.WebSize, Valid: req.WebSize > 0},
		OriginalSize:     sql.NullInt64{Int64: req.OriginalSize, Valid: req.OriginalSize > 0},
		Width:            sql.NullInt32{Int32: req.Width, Valid: req.Width > 0},
		Height:           sql.NullInt32{Int32: req.Height, Valid: req.Height > 0},
		UploadedBy:       uuid.NullUUID{UUID: parsedUID, Valid: true},
		Exif:             pqtype.NullRawMessage{RawMessage: metaJSON, Valid: true},
	}

	img, err := h.q.CreateImage(r.Context(), params)
	if err != nil {
		writeError(w, r, errInternal("failed to create image"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(imageToResponse(img, nil, nil))
}
