package interfaces

import (
	"net/http"

	. "github.com/AleksandrMac/fileserver/internal/domain"
	"github.com/AleksandrMac/fileserver/pkg/uerror"
)

type TrackUsecase UseCaseI[*TrackRequest, *TrackResponse]

type UseCaseI[Input, Output any] interface {
	ReadRequestData(r *http.Request) (args Input, err uerror.UError)
	Proceed(data Input) (result Output, err uerror.UError)
	WriteResponse(w http.ResponseWriter, result Output, format string) error
}
