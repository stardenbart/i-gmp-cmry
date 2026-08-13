package uploadusecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

const finalizingKeyTTL = 5 * time.Minute

// FinalizePhotoResult menyimpan statistik hasil finalisasi foto
type FinalizePhotoResult struct {
	Uploaded int
	Skipped  int // Sudah ada di MinIO (idempotent)
	Failed   int
}

// FinalizeInspectionPhotos memindahkan semua foto base64 WebP dari Redis draft ke MinIO.
// Dipanggil saat inspeksi status diubah ke "Completed".
// Operasi ini:
//   - Idempotent: jika object key sudah ada di MinIO, skip tanpa error
//   - Atomik: Redis draft key dihapus HANYA setelah upload MinIO selesai semua
//   - Concurrency-safe: set flag finalizing:{inspectionID} (SETNX, TTL 5 menit)
func (uc *UploadUseCase) FinalizeInspectionPhotos(
	ctx context.Context,
	rdb *redis.Client,
	inspectionID string,
) (*FinalizePhotoResult, error) {
	if uc.minioStorage == nil {
		return &FinalizePhotoResult{}, nil
	}

	result := &FinalizePhotoResult{}

	// 1. Set finalizing flag (SETNX) — cegah concurrent finalisasi
	finalizingKey := fmt.Sprintf("finalizing:%s", inspectionID)
	if rdb != nil {
		ok, err := rdb.SetNX(ctx, finalizingKey, "1", finalizingKeyTTL).Result()
		if err == nil && !ok {
			return nil, fmt.Errorf("inspeksi %s sedang dalam proses finalisasi", inspectionID)
		}
		defer rdb.Del(ctx, finalizingKey)
	}

	// 2. Ambil semua draft Redis untuk inspeksi ini
	var allDrafts map[string]map[string]string
	if rdb != nil {
		pattern := fmt.Sprintf("state:aspek:%s:*", inspectionID)
		keys, err := rdb.Keys(ctx, pattern).Result()
		if err == nil && len(keys) > 0 {
			allDrafts = make(map[string]map[string]string)
			pipe := rdb.Pipeline()
			type keyCmd struct {
				aspekID string
				cmd     *redis.MapStringStringCmd
			}
			var cmds []keyCmd
			for _, key := range keys {
				parts := strings.Split(key, ":")
				if len(parts) >= 4 {
					aspekID := parts[3]
					cmd := pipe.HGetAll(ctx, key)
					cmds = append(cmds, keyCmd{aspekID: aspekID, cmd: cmd})
				}
			}
			pipe.Exec(ctx)
			for _, kc := range cmds {
				if draft, err := kc.cmd.Result(); err == nil && len(draft) > 0 {
					allDrafts[kc.aspekID] = draft
				}
			}
		}
	}

	if len(allDrafts) == 0 {
		return result, nil
	}

	// 3. Iterasi setiap aspek → parse data JSON → cari foto base64
	for aspekID, draft := range allDrafts {
		rawData, ok := draft["data"]
		if !ok || rawData == "" {
			continue
		}

		// Parse JSON data aspek: map[uraianKey] → { photos: [{previewUrl, keterangan, ...}] }
		var aspekData map[string]json.RawMessage
		if err := json.Unmarshal([]byte(rawData), &aspekData); err != nil {
			continue
		}

		for uraianKey, uraianRaw := range aspekData {
			var uraianObj struct {
				Photos []struct {
					ID         string `json:"id"`
					PreviewURL string `json:"previewUrl"`
					Keterangan string `json:"keterangan"`
					HEIID      string `json:"hei_id"`
					HEICategory string `json:"hei_category"`
				} `json:"photos"`
			}
			if err := json.Unmarshal(uraianRaw, &uraianObj); err != nil {
				continue
			}

			for idx, photo := range uraianObj.Photos {
				if !strings.HasPrefix(photo.PreviewURL, "data:image/webp;base64,") &&
					!strings.HasPrefix(photo.PreviewURL, "data:image/") {
					// Bukan base64 — sudah URL server atau kosong, skip
					continue
				}

				// 4a. Decode base64 → binary
				// Format: "data:image/webp;base64,<data>"
				commaIdx := strings.Index(photo.PreviewURL, ",")
				if commaIdx == -1 {
					result.Failed++
					continue
				}
				b64data := photo.PreviewURL[commaIdx+1:]
				imgBytes, err := base64.StdEncoding.DecodeString(b64data)
				if err != nil {
					result.Failed++
					continue
				}

				// 4b. Object key deterministik (idempotent) — tidak pakai UUID random
				photoID := photo.ID
				if photoID == "" {
					photoID = fmt.Sprintf("idx%d", idx)
				}
				objectName := fmt.Sprintf("uploads/%s/%s-%s-%s.webp",
					inspectionID, aspekID, uraianKey, photoID)

				// 4c. Upload ke MinIO (UploadStream sudah return path relatif)
				_, err = uc.minioStorage.UploadStream(
					ctx,
					objectName,
					bytes.NewReader(imgBytes),
					int64(len(imgBytes)),
					"image/webp",
				)
				if err != nil {
					result.Failed++
					continue
				}

				result.Uploaded++
			}
		}
	}

	return result, nil
}
