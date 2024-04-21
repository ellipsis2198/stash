package manager

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path"
	"strconv"

	"github.com/nfnt/resize"
	"github.com/stashapp/stash/pkg/fsutil"
	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/models"
)

type GenerateHspScreenshotTask struct {
	repository models.Repository
	Scene      models.Scene
	Folder     string
	Overwrite  bool
}

func (t *GenerateHspScreenshotTask) GetDescription() string {
	return fmt.Sprintf("Generating hsp screenshot for %s", t.Scene.Path)
}

const maxRes = 360

func (t *GenerateHspScreenshotTask) Start(ctx context.Context) {
	if !t.required() {
		return
	}

	if err := fsutil.EnsureDir(path.Join(t.Folder, "hsp_screenshot")); err != nil {
		logger.Error(err.Error())
		return
	}

	r := t.repository
	var cover []byte
	var err error

	// Get cover
	if err := r.WithTxn(ctx, func(ctx context.Context) error {
		cover, err = t.repository.Scene.GetCover(ctx, t.Scene.ID)
		return err
	}); err != nil {
		logger.Error(err.Error())
		return
	}

	// Stop if none
	if cover == nil {
		return
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(cover))
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// Get the dimensions of the original image
	originalWidth := img.Bounds().Max.X
	originalHeight := img.Bounds().Max.Y

	// Calculate the scaling factor to fit within 360 pixels
	var scaleFactor float64
	if originalWidth > originalHeight {
		scaleFactor = maxRes / float64(originalWidth)
	} else {
		scaleFactor = maxRes / float64(originalHeight)
	}

	// Resize the image
	newWidth := uint(float64(originalWidth) * scaleFactor)
	newHeight := uint(float64(originalHeight) * scaleFactor)
	resizedImg := resize.Resize(newWidth, newHeight, img, resize.NearestNeighbor)

	// Encode the resized image to JPEG format (change to PNG if needed)
	var resizedImageBuf bytes.Buffer
	if err = jpeg.Encode(&resizedImageBuf, resizedImg, nil); err != nil {
		logger.Error(err.Error())
		return
	}

	// Set output
	cover = resizedImageBuf.Bytes()

	// Write cache
	pth := t.getScreenshotPath()
	if err = os.WriteFile(pth, cover, 0600); err != nil {
		logger.Error(err.Error())
		return
	}
}

func (t *GenerateHspScreenshotTask) required() bool {
	if t.Overwrite {
		return true
	}

	pth := t.getScreenshotPath()
	_, err := os.Stat(pth)
	return os.IsNotExist(err)
}

func (t *GenerateHspScreenshotTask) getScreenshotPath() string {
	return path.Join(t.Folder, "hsp_screenshot", strconv.Itoa(t.Scene.ID))
}
