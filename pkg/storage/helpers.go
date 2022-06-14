package storage

import (
	"time"

	"neurocode.io/cache-offloader/pkg/model"
)

func getSize(value model.Response) float64 {
	sizeBytes := len(value.Body)
	sizeMB := float64(sizeBytes) / (1024 * 1024)

	return sizeMB
}

func getStaleStatus(timeStamp int64, staleDuration int64) uint8 {
	if (time.Now().Unix() - timeStamp) >= staleDuration {
		return model.StaleValue
	} else {
		return model.FreshValue
	}
}
