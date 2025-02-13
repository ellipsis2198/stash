package manager

import (
	"context"
	"fmt"

	"github.com/stashapp/stash/pkg/logger"
	"github.com/stashapp/stash/pkg/models"
	"github.com/stashapp/stash/pkg/scene/generate"
)

type GenerateCoverTask struct {
	Scene        models.Scene
	ScreenshotAt *float64
	txnManager   Repository
	Overwrite    bool
}

func (t *GenerateCoverTask) GetDescription() string {
	return fmt.Sprintf("Generating cover for %s", t.Scene.GetTitle())
}

func (t *GenerateCoverTask) Start(ctx context.Context) {
	scenePath := t.Scene.Path

	var required bool
	if err := t.txnManager.WithReadTxn(ctx, func(ctx context.Context) error {
		// don't generate the screenshot if it already exists
		required = t.required(ctx)
		return t.Scene.LoadPrimaryFile(ctx, t.txnManager.File)
	}); err != nil {
		logger.Error(err)
	}

	if !required {
		return
	}

	videoFile := t.Scene.Files.Primary()
	if videoFile == nil {
		return
	}

	var at float64
	if t.ScreenshotAt == nil {
		at = float64(videoFile.Duration) * 0.2
	} else {
		at = *t.ScreenshotAt
	}

	// we'll generate the screenshot, grab the generated data and set it
	// in the database.

	logger.Debugf("Creating screenshot for %s", scenePath)

	g := generate.Generator{
		Encoder:      instance.FFMPEG,
		FFMpegConfig: instance.Config,
		LockManager:  instance.ReadLockManager,
		ScenePaths:   instance.Paths.Scene,
		Overwrite:    true,
	}

	coverImageData, err := g.Screenshot(context.TODO(), videoFile.Path, videoFile.Width, videoFile.Duration, generate.ScreenshotOptions{
		At: &at,
	})
	if err != nil {
		logger.Errorf("Error generating screenshot: %v", err)
		logErrorOutput(err)
		return
	}

	if err := t.txnManager.WithTxn(ctx, func(ctx context.Context) error {
		qb := t.txnManager.Scene
		updatedScene := models.NewScenePartial()

		// update the scene cover table
		if err := qb.UpdateCover(ctx, t.Scene.ID, coverImageData); err != nil {
			return fmt.Errorf("error setting screenshot: %v", err)
		}

		// update the scene with the update date
		_, err = qb.UpdatePartial(ctx, t.Scene.ID, updatedScene)
		if err != nil {
			return fmt.Errorf("error updating scene: %v", err)
		}

		return nil
	}); err != nil && ctx.Err() == nil {
		logger.Error(err.Error())
	}
}

// required returns true if the sprite needs to be generated
func (t GenerateCoverTask) required(ctx context.Context) bool {
	if t.Overwrite {
		return true
	}

	// if the scene has a cover, then we don't need to generate it
	hasCover, err := t.txnManager.Scene.HasCover(ctx, t.Scene.ID)
	if err != nil {
		logger.Errorf("Error getting cover: %v", err)
		return false
	}

	return !hasCover
}
