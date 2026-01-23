package usecase

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	. "github.com/AleksandrMac/fileserver/internal/domain"
	"github.com/AleksandrMac/fileserver/internal/interfaces"
	"github.com/AleksandrMac/fileserver/pkg/uerror"
)

type TrackUC struct {
	docServerUrl         string
	docServerUrlInternal string
	fileRepo             interfaces.FileRepo
}

func NewTrackUC(
	fileRepo interfaces.FileRepo,
	docServerUrl,
	docServerUrlInternal string,
) interfaces.TrackUsecase {
	return &TrackUC{
		docServerUrl:         docServerUrl,
		docServerUrlInternal: docServerUrlInternal,
		fileRepo:             fileRepo,
	}
}

func (x *TrackUC) ReadRequestData(r *http.Request) (args *TrackRequest, err uerror.UError) {
	args = new(TrackRequest)
	if err := json.NewDecoder(r.Body).Decode(args); err != nil {
		return nil, uerror.NewUError(
			http.StatusBadRequest, "failed parse payload", err, nil,
		)
	}

	if err := args.Valid(); err != nil {
		return nil, uerror.NewUError(
			http.StatusBadRequest, "failed validate payload", err, nil,
		)
	}

	if args.Url == "" && (args.Status == 2 || args.Status == 3) {
		return nil, uerror.NewUError(
			http.StatusBadRequest, "failed validate payload", errors.New("missing url tag"), nil,
		)
	}

	return
}

func (x *TrackUC) Proceed(data *TrackRequest) (_ *TrackResponse, err uerror.UError) {
	// 4. Обрабатываем только статусы 2 (редактирование) и 3 (нужно сохранить)
	if data.Status == 2 || data.Status == 3 {
		filename, err := base64.URLEncoding.DecodeString(data.Key)
		if err != nil {
			return nil, uerror.NewUError(http.StatusBadRequest,
				"failed decode key", err, map[string]any{
					"key": data.Key,
				},
			)
		}

		fullFilename, err := x.fileRepo.GetFullPath(string(filename))
		if err != nil {
			return nil, uerror.NewUError(http.StatusInternalServerError,
				"failed get full path", err, map[string]any{
					"path": filename,
				},
			)
		}

		// 6. Скачиваем обновлённый документ от Document Server
		uri := x.updateUri(data.Url)

		resp, err := http.Get(uri)
		if err != nil {
			return nil, uerror.NewUError(
				http.StatusInternalServerError, "failed download updated doocument", err, map[string]any{
					"url": uri,
				},
			)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, uerror.NewUError(http.StatusInternalServerError,
				"download failed", errors.New("download failed"), map[string]any{
					"url":    uri,
					"status": resp.Status,
				},
			)
		}

		// 7. Сохраняем поверх существующего файла
		if err := x.fileRepo.SaveFile(fullFilename, resp.Body); err != nil {
			return nil, uerror.NewUError(http.StatusInternalServerError,
				"failed to write document", err, map[string]any{
					"fullfilename": fullFilename,
				},
			)
		}
	}

	return
}

func (x *TrackUC) WriteResponse(w http.ResponseWriter, data *TrackResponse, format string) error {
	w.Header().Set("Content-Type", format)
	if format == "application/json" {
		return json.NewEncoder(w).Encode(map[string]int{"error": 0})
	}

	_, err := w.Write([]byte("unsupported format"))
	return err
}

func (x *TrackUC) updateUri(uri string) string {
	if x.docServerUrl == "" {
		return uri
	}

	u, _ := url.Parse(uri)
	u.Scheme = "http"
	u.Host = x.docServerUrlInternal

	return u.String()
}
