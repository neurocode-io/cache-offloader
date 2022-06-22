package storage

import (
	"time"

	"github.com/neurocode-io/cache-offloader/pkg/model"
)

func getSize(value model.Response) float64 {
	sizeBytes := len(value.Body)
	sizeMB := float64(sizeBytes) / (1024 * 1024)

	return sizeMB
}

func getStaleStatus(timeStamp int64, staleDuration int) uint8 {
	if (time.Now().Unix() - timeStamp) >= int64(staleDuration) {
		return model.StaleValue
	}

	return model.FreshValue
}
