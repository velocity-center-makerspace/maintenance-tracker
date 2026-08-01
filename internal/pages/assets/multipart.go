package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type upload struct {
	tmpPath  string
	checksum string
	ext      string
	filename string
	mimeType string
}

type fileMetadata struct {
	contentHash string
	mimeType    string
	filename    string
	path        string
}

func parseMultipart(reader *multipart.Reader, dstRoot string) ([]upload, *json.Decoder, error) {
	var uploads []upload
	var dec *json.Decoder

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		mimeType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			return nil, nil, err
		}

		switch mimeType {
		default:
			return nil, nil, fmt.Errorf(
				"%w: %q for part %s",
				ErrUnsupportedContentType,
				mimeType,
				part.FormName(),
			)
		case "application/json":
			dec = json.NewDecoder(part)
			dec.DisallowUnknownFields()
		case "application/pdf":
			fallthrough
		case "application/msword":
			fallthrough
		case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
			if part.FileName() == "" {
				continue
			}

			tmp, err := os.CreateTemp(dstRoot, "upload-*.tmp")
			if err != nil {
				return nil, nil, err
			}

			h := sha256.New()
			tee := io.TeeReader(part, h)

			if _, err := io.Copy(tmp, tee); err != nil {
				return nil, nil, err
			}

			if err := tmp.Close(); err != nil {
				return nil, nil, err
			}

			uploads = append(uploads,
				upload{
					tmpPath:  tmp.Name(),
					checksum: hex.EncodeToString(h.Sum(nil)),
					mimeType: mimeType,
					filename: part.FileName(),
					ext:      filepath.Ext(part.FileName()),
				},
			)
		}
	}

	return uploads, dec, nil
}

type processFileParams struct {
	upload  upload
	dstRoot string
	errChan chan<- error
}

func processFile(p processFileParams) *fileMetadata {
	tmp, err := os.Open(p.upload.tmpPath)
	if err != nil {
		p.errChan <- err
		return nil
	}

	defer func() {
		for attempts := range 3 {
			err := os.Remove(p.upload.tmpPath)
			if err == nil || errors.Is(err, os.ErrNotExist) {
				return
			}

			// incrementally longer retrys
			time.Sleep(10 * (1 << attempts) * time.Millisecond)
		}
		p.errChan <- fmt.Errorf("failed to remove after 3 attempts: %w", err)
	}()

	defer func() { _ = tmp.Close() }()

	dstPath := filepath.Join(
		p.dstRoot,
		p.upload.checksum[:2],
		p.upload.checksum[2:4],
		p.upload.checksum+p.upload.ext,
	)

	if _, err := os.Stat(dstPath); err == nil {
		return &fileMetadata{
			contentHash: p.upload.checksum,
			filename:    p.upload.filename,
			mimeType:    p.upload.mimeType,
			path:        dstPath,
		}
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0744); err != nil {
		p.errChan <- err
		return nil
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		p.errChan <- err
		return nil
	}

	defer func() {
		if err := dst.Close(); err != nil {
			p.errChan <- err
		}
	}()

	if _, err := io.Copy(dst, tmp); err != nil {
		p.errChan <- err
		return nil
	}

	return &fileMetadata{
		contentHash: p.upload.checksum,
		filename:    p.upload.filename,
		mimeType:    p.upload.mimeType,
		path:        dstPath,
	}
}
