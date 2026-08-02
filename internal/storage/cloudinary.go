package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

const folder = "hr-management"

var cld *cloudinary.Cloudinary

func Init() {
	if cld != nil {
		return
	}

	var (
		instance *cloudinary.Cloudinary
		err      error
	)

	if url := os.Getenv("CLOUDINARY_URL"); url != "" {
		instance, err = cloudinary.NewFromURL(url)
	} else {
		instance, err = cloudinary.NewFromParams(
			os.Getenv("CLOUDINARY_CLOUD_NAME"),
			os.Getenv("CLOUDINARY_API_KEY"),
			os.Getenv("CLOUDINARY_API_SECRET"),
		)
	}

	if err != nil {
		fmt.Println("Error initializing Cloudinary:", err)
		return
	}

	cld = instance
}

func Upload(ctx context.Context, r io.Reader, publicID string) (string, error) {
	if cld == nil {
		return "", fmt.Errorf("uninitialized cloudinary")
	}

	res, err := cld.Upload.Upload(ctx, r, uploader.UploadParams{
		PublicID: folder + "/" + publicID,
	})
	if err != nil {
		return "", err
	}

	return res.SecureURL, nil
}

func Destroy(ctx context.Context, secureURL string) error {
	if cld == nil || secureURL == "" {
		return nil
	}

	publicID := PublicIDFromURL(secureURL)
	if publicID == "" {
		return nil
	}

	_, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}

func PublicIDFromURL(secureURL string) string {
	marker := "/image/upload/"
	idx := strings.Index(secureURL, marker)
	if idx < 0 {
		return ""
	}

	id := secureURL[idx+len(marker):]

	if dot := strings.LastIndex(id, "."); dot > 0 {
		id = id[:dot]
	}

	return id
}
