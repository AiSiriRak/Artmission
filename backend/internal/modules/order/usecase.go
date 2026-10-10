package order

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	stdDraw "image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	_ "golang.org/x/image/webp"
)

type ArtistGetter interface {
	GetArtistName(ctx context.Context, id uuid.UUID) (string, error)
}

type orderUsecase struct {
	repo         OrderRepository
	storage      ObjectStorage
	artistGetter ArtistGetter
}

// NewOrderUsecase creates the order use case with its repository and storage
// dependencies. artistGetter is needed only when creating watermarked
// deliverable previews.
func NewOrderUsecase(repo OrderRepository, storage ObjectStorage, artistGetters ...ArtistGetter) OrderUsecase {
	usecase := &orderUsecase{repo: repo, storage: storage}
	if len(artistGetters) > 0 {
		usecase.artistGetter = artistGetters[0]
	}
	return usecase
}

const orderDeliverableTTL = 15 * time.Minute

// ViewOrders validates the query, retrieves its page, and resolves preview keys to URLs.
func (u *orderUsecase) ViewOrders(ctx context.Context, query ListQuery) (Page, error) {
	normalized, err := normalizeListQuery(query)
	if err != nil {
		return Page{}, err
	}

	page, err := u.repo.ListOrders(ctx, normalized)
	if err != nil {
		return Page{}, err
	}

	for i := range page.Orders {
		key := page.Orders[i].DeliverablePreviewKey
		if key == nil {
			continue
		}
		url, err := u.storage.GetPresignedURL(ctx, *key, orderDeliverableTTL)
		if err != nil {
			return Page{}, ErrFailedToPresignDeliverablePreview
		}
		page.Orders[i].DeliverablePreviewURL = &url
	}

	return page, nil
}

// normalizeListQuery validates query and defaults every optional field.
func normalizeListQuery(q ListQuery) (ListQuery, error) {
	if !q.Participant.IsValid() {
		return ListQuery{}, ErrUnsupportedParticipantRole
	}
	if q.ParticipantID == uuid.Nil {
		return ListQuery{}, ErrMissingParticipantID
	}

	statuses, err := normalizeStatuses(q.Statuses)
	if err != nil {
		return ListQuery{}, err
	}
	q.Statuses = statuses

	if q.Sort == "" {
		q.Sort = DefaultSort
	}
	if !q.Sort.IsValid() {
		return ListQuery{}, apperror.InvalidInput(fmt.Sprintf("invalid sort field %q", q.Sort), nil)
	}

	if q.Order == "" {
		q.Order = DefaultOrder
	}
	if !q.Order.IsValid() {
		return ListQuery{}, apperror.InvalidInput(fmt.Sprintf("invalid sort order %q", q.Order), nil)
	}

	if q.Limit == 0 {
		q.Limit = DefaultLimit
	}
	if q.Limit < 1 || q.Limit > MaxLimit {
		return ListQuery{}, apperror.InvalidInput(fmt.Sprintf("limit must be between 1 and %d", MaxLimit), nil)
	}

	if q.Offset < 0 {
		return ListQuery{}, apperror.InvalidInput("offset must not be negative", nil)
	}

	return q, nil
}

// normalizeStatuses validates every status, drops duplicates, and sorts
// the result into a deterministic order.
func normalizeStatuses(in []Status) ([]Status, error) {
	if len(in) == 0 {
		return nil, nil
	}

	seen := make(map[Status]bool, len(in))
	out := make([]Status, 0, len(in))
	for _, s := range in {
		if !s.IsValid() {
			return nil, apperror.InvalidInput(fmt.Sprintf("invalid status %q", s), nil)
		}
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}

	slices.Sort(out)
	return out, nil
}

// GetOrder returns an order detail if the authenticated user is a participant
// in the order. The other party is selected based on the authenticated
// participant's role.
func (u *orderUsecase) GetOrder(
	ctx context.Context,
	participant Participant,
	participantID uuid.UUID,
	orderID uuid.UUID,
) (*OrderDetail, error) {
	if !participant.IsValid() {
		return nil, ErrUnsupportedParticipantRole
	}
	if participantID == uuid.Nil {
		return nil, ErrMissingParticipantID
	}

	if orderID == uuid.Nil {
		return nil, ErrMissingOrderID
	}

	data, err := u.repo.GetOrderByID(
		ctx,
		participant,
		participantID,
		orderID,
	)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, apperror.Internal("order repository returned no order detail", nil)
	}

	detail := &OrderDetail{
		Order:           data.Order,
		ArtworkSnapshot: data.ArtworkSnapshot,
		Deliverables:    data.Deliverables,
	}

	switch participant {
	case ParticipantCustomer:
		detail.OtherParty = data.Artist
	case ParticipantArtist:
		detail.OtherParty = data.Customer
	}

	for i := range detail.Deliverables {
		d := &detail.Deliverables[i]

		if d.PreviewImageKey == "" {
			continue
		}

		url, err := u.storage.GetPresignedURL(
			ctx,
			d.PreviewImageKey,
			orderDeliverableTTL,
		)
		if err != nil {
			return nil, ErrFailedToPresignDeliverablePreview
		}

		d.PreviewImageURL = url
	}

	return detail, nil
}

// ConfirmOrder accepts or rejects an order on behalf of its artist.
// The order must belong to the artist and currently be PENDING.
// Accept = true transitions the order's status to NOT_PAID.
// Accept = false transitions the order's status to CANCEL.
func (u *orderUsecase) ConfirmOrder(
	ctx context.Context,
	artistID uuid.UUID,
	orderID uuid.UUID,
	input ConfirmOrderInput,
) (Status, error) {
	if artistID == uuid.Nil {
		return "", ErrMissingParticipantID
	}

	if orderID == uuid.Nil {
		return "", ErrMissingOrderID
	}

	var status Status
	switch input.Accept {
	case true:
		status = StatusNotPaid
	case false:
		status = StatusCancel
	}

	if err := u.repo.ConfirmOrder(ctx, artistID, orderID, status); err != nil {
		return "", err
	}

	return status, nil
}

// CreateOrder validates the artwork reference, builds a pending order, and persists it.
func (u *orderUsecase) CreateOrder(ctx context.Context, customerID uuid.UUID, input CreateInput) (*Order, error) {
	if input.ArtworkID == uuid.Nil {
		return nil, apperror.InvalidInput("artwork id must not be empty", nil)
	}

	order, err := normalizeOrder(uuid.New(), customerID, input)
	if err != nil {
		return nil, err
	}

	err = u.repo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// normalizeOrder validates customer-provided fields and builds a timestamped pending order.
func normalizeOrder(id uuid.UUID, customerID uuid.UUID, input CreateInput) (*Order, error) {
	if customerID == uuid.Nil {
		return nil, apperror.InvalidInput("customer id must not be empty", nil)
	}
	name, err := requiredText("name", input.Name)
	if err != nil {
		return nil, err
	}

	description, err := requiredText("customer description", input.CustomerDescription)
	if err != nil {
		return nil, err
	}
	if input.DeadlineAt.IsZero() {
		return nil, apperror.InvalidInput("deadline must not be empty", nil)
	}

	now := time.Now()
	return &Order{
		ID:                  id,
		CustomerID:          customerID,
		ArtworkID:           &input.ArtworkID,
		Name:                name,
		CustomerDescription: description,
		DeadlineAt:          input.DeadlineAt,
		Status:              StatusPending,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// requiredText trims surrounding whitespace and rejects values that become empty.
func requiredText(field, value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", apperror.InvalidInput(field+" must not be blank", nil)
	}
	return normalized, nil
}

// CreateDeliverable creates a new deliverable version for an order owned
// by the authenticated artist. It validates the uploaded image, generates
// a preview, uploads both images, and persists the deliverable metadata.
// If persistence fails, it attempts to remove the uploaded objects.
func (u *orderUsecase) CreateDeliverable(
	ctx context.Context,
	artistID uuid.UUID,
	input CreateDeliverableInput,
) (*Deliverable, error) {
	if artistID == uuid.Nil {
		return nil, ErrMissingParticipantID
	}

	if input.OrderID == uuid.Nil {
		return nil, ErrMissingOrderID
	}

	// Read the uploaded image, enforce the size limit, and detect its
	// supported MIME type and file extension.
	original, originalContentType, extension, err := readDeliverableImage(input.DeliverableImage)
	if err != nil {
		return nil, err
	}

	if u.artistGetter == nil {
		return nil, ErrArtistGetterUnavailable
	}
	artistName, err := u.artistGetter.GetArtistName(ctx, artistID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Generate a resized preview with a watermark.
	preview, err := createDeliverablePreview(
		original,
		artistName,
		now,
	)
	if err != nil {
		return nil, apperror.InvalidInput("failed to process deliverable image", err)
	}

	// Generate one ID to associate the uploaded objects with this
	// deliverable record.
	deliverableID := uuid.New()

	// Upload the unmodified original and the watermarked preview
	// under separate private object keys.
	originalImageKey, previewImageKey, err := u.uploadDeliverableImages(
		ctx,
		input.OrderID,
		deliverableID,
		original,
		preview,
		originalContentType,
		extension,
	)
	if err != nil {
		return nil, err
	}

	deliverable := normalizeOrderDeliverable(
		deliverableID,
		originalImageKey,
		previewImageKey,
		now,
	)

	if err := u.repo.CreateDeliverable(
		ctx,
		artistID,
		input.OrderID,
		deliverable,
	); err != nil {
		if cleanupErr := u.deleteDeliverableImages(
			ctx,
			originalImageKey,
			previewImageKey,
		); cleanupErr != nil {
			return nil, apperror.Internal(
				"failed to create deliverable and clean up uploaded images",
				errors.Join(err, cleanupErr),
			)
		}
		return nil, err
	}

	return deliverable, nil
}

// normalizeOrderDeliverable initializes a new deliverable entity with
// its ID, default decision, object keys, and timestamps.
// The repository assigns the version when persisting the entity.
func normalizeOrderDeliverable(
	id uuid.UUID,
	originalImageKey string,
	previewImageKey string,
	createdAt time.Time,
) *Deliverable {
	return &Deliverable{
		ID:               id,
		Decision:         DeliverableDecisionWait,
		OriginalImageKey: originalImageKey,
		PreviewImageKey:  previewImageKey,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
	}
}

// uploadDeliverableImages uploads the original image and its watermarked
// JPEG preview to separate object keys and returns those keys.
// If uploading the preview fails, it attempts to remove the original.
func (u *orderUsecase) uploadDeliverableImages(
	ctx context.Context,
	orderID uuid.UUID,
	deliverableID uuid.UUID,
	original []byte,
	preview []byte,
	originalContentType string,
	originalExtension string,
) (string, string, error) {
	baseKey := fmt.Sprintf(
		"orders/%s/deliverables/%s",
		orderID,
		deliverableID,
	)

	originalKey := baseKey + "/original." + originalExtension
	previewKey := baseKey + "/preview.jpg"

	if err := u.storage.Upload(
		ctx,
		originalKey,
		bytes.NewReader(original),
		originalContentType,
	); err != nil {
		return "", "", apperror.Internal(
			"failed to upload original deliverable image",
			err,
		)
	}

	if err := u.storage.Upload(
		ctx,
		previewKey,
		bytes.NewReader(preview),
		"image/jpeg",
	); err != nil {
		if cleanupErr := u.storage.Delete(ctx, originalKey); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
		return "", "", apperror.Internal(
			"failed to upload deliverable preview image",
			err,
		)
	}

	return originalKey, previewKey, nil
}

// deleteDeliverableImages attempts to delete each supplied object key.
// It continues after a failure and returns all cleanup errors combined.
func (u *orderUsecase) deleteDeliverableImages(
	ctx context.Context,
	keys ...string,
) error {
	var cleanupErr error
	for _, key := range keys {
		if err := u.storage.Delete(ctx, key); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("delete %q: %w", key, err))
		}
	}
	return cleanupErr
}

// readDeliverableImage reads an uploaded image while enforcing the
// maximum file size. It returns the bytes, MIME type, and extension
// for supported JPEG, PNG, and WebP images.
func readDeliverableImage(reader io.Reader) ([]byte, string, string, error) {
	if reader == nil {
		return nil, "", "", apperror.InvalidInput("deliverable image must not be empty", nil)
	}

	content, err := io.ReadAll(io.LimitReader(reader, MaxDeliverableImageSize+1))
	if err != nil {
		return nil, "", "", apperror.InvalidInput("failed to read deliverable image", err)
	}
	if len(content) == 0 {
		return nil, "", "", apperror.InvalidInput("deliverable image must not be empty", nil)
	}
	if len(content) > MaxDeliverableImageSize {
		return nil, "", "", ErrDeliverableImageTooLarge
	}

	switch mimetype.Detect(content).String() {
	case "image/jpeg":
		return content, "image/jpeg", "jpg", nil
	case "image/png":
		return content, "image/png", "png", nil
	case "image/webp":
		return content, "image/webp", "webp", nil
	default:
		return nil, "", "", ErrInvalidDeliverableImage
	}
}

// createDeliverablePreview decodes the original image with EXIF orientation,
// resizes it to the configured maximum dimension without upscaling, adds a
// watermark that fits the image, and encodes the result as a JPEG preview.
func createDeliverablePreview(
	original []byte,
	artistName string,
	submittedAt time.Time,
) ([]byte, error) {
	src, err := imaging.Decode(
		bytes.NewReader(original),
		imaging.AutoOrientation(true),
	)
	if err != nil {
		return nil, fmt.Errorf("decode deliverable image: %w", err)
	}

	width, height := src.Bounds().Dx(), src.Bounds().Dy()
	if width <= 0 || height <= 0 {
		return nil, errors.New("deliverable image has invalid dimensions")
	}

	// Resize the image while preserving its aspect ratio.
	if max(width, height) > deliverablePreviewSize {
		if width >= height {
			height = max(1, height*deliverablePreviewSize/width)
			width = deliverablePreviewSize
		} else {
			width = max(1, width*deliverablePreviewSize/height)
			height = deliverablePreviewSize
		}

		src = imaging.Resize(src, width, height, imaging.Lanczos)
	}

	// Create a mutable image for drawing the watermark.
	watermarkImage := imaging.Clone(src)
	bounds := watermarkImage.Bounds()
	width, height = bounds.Dx(), bounds.Dy()

	// Calculate watermark dimensions relative to the image.
	padding := max(4, width/100)
	availableWidth := max(1, width-2*padding)
	bandHeight := max(18, min(height/10, 60))
	bandHeight = min(bandHeight, height)

	// Draw a translucent background at the bottom of the image.
	band := image.Rect(0, height-bandHeight, width, height)
	stdDraw.Draw(
		watermarkImage,
		band,
		image.NewUniform(color.NRGBA{A: 180}),
		image.Point{},
		stdDraw.Over,
	)

	// Build the watermark text using the artist name and submission date.
	watermarkText := artistName + " - " +
		submittedAt.UTC().Format("2006-01-02")

	// Measure the text using the existing fixed-size font.
	face := basicfont.Face7x13
	textWidth := font.MeasureString(face, watermarkText).Ceil()
	textHeight := face.Metrics().Height.Ceil()

	// Scale the rendered text to fit both the available width and band height.
	availableTextHeight := max(1, bandHeight*2/3)
	scale := min(
		float64(availableWidth)/float64(max(1, textWidth)),
		float64(availableTextHeight)/float64(max(1, textHeight)),
	)

	scaledWidth := max(1, int(float64(textWidth)*scale))
	scaledHeight := max(1, int(float64(textHeight)*scale))

	// Render the text to a temporary image.
	textImage := image.NewRGBA(
		image.Rect(0, 0, max(1, textWidth), max(1, textHeight)),
	)

	drawer := font.Drawer{
		Dst:  textImage,
		Src:  image.NewUniform(color.White),
		Face: face,
		Dot: fixed.P(
			0,
			face.Metrics().Ascent.Ceil(),
		),
	}
	drawer.DrawString(watermarkText)

	// Resize the rendered text to fit the available space.
	scaledText := image.NewRGBA(
		image.Rect(0, 0, scaledWidth, scaledHeight),
	)
	xdraw.CatmullRom.Scale(
		scaledText,
		scaledText.Bounds(),
		textImage,
		textImage.Bounds(),
		xdraw.Src,
		nil,
	)

	// Center the text vertically within the watermark band.
	textX := padding
	textY := height - bandHeight + (bandHeight-scaledHeight)/2

	stdDraw.Draw(
		watermarkImage,
		image.Rect(
			textX,
			textY,
			textX+scaledWidth,
			textY+scaledHeight,
		),
		scaledText,
		image.Point{},
		stdDraw.Over,
	)

	// Encode the watermarked preview as JPEG.
	var output bytes.Buffer
	if err := imaging.Encode(
		&output,
		watermarkImage,
		imaging.JPEG,
		imaging.JPEGQuality(88),
	); err != nil {
		return nil, fmt.Errorf("encode deliverable preview: %w", err)
	}

	return output.Bytes(), nil
}
