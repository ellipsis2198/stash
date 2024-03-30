package heresphere

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stashapp/stash/internal/manager/config"
	"github.com/stashapp/stash/pkg/models"
)

/*
 * Finds the selected VR Tag string
 */
func getVrTag() (varTag string, err error) {
	// Find setting
	varTag = config.GetInstance().GetUIVRTag()
	if len(varTag) == 0 {
		err = fmt.Errorf("zero length vr tag")
	}
	return
}

/*
 * Finds the selected minimum play percentage value
 */
func getMinPlayPercent() (per int, err error) {
	per = config.GetInstance().GetUIMinPlayPercent()
	if per < 0 {
		err = fmt.Errorf("unset minimum play percent")
	}
	return
}

/*
 * Returns the hsp filepath and an os.Stat for checking os.IsNotExist
 */
func getHspFile(primaryFile *models.VideoFile) (string, error) {
	path := primaryFile.Base().Path
	//fileBaseNameWithoutExt := fmt.Sprintf("%s.%d.hsp", strings.TrimSuffix(path, filepath.Ext(path)), version)
	fileBaseNameWithoutExt := fmt.Sprintf("%s.hsp", strings.TrimSuffix(path, filepath.Ext(path)))

	_, err := os.Stat(fileBaseNameWithoutExt)
	return fileBaseNameWithoutExt, err
}
