package http

import (
	"net/http"

	"github.com/AleksandrMac/fileserver/pkg/uerror/logwrapper"
	"github.com/rs/zerolog/log"
)

func (x *Handler) Track(w http.ResponseWriter, r *http.Request) {
	uc := x.trackUC

	args, err := uc.ReadRequestData(r)
	if err != nil {
		logwrapper.ZeroLog(log.Debug().Str("func", "TrackUsecase.ReadRequestData"), err)
		http.Error(w, err.Message(), err.Status())
		return
	}

	result, err := uc.Proceed(args)
	if err != nil {
		logwrapper.ZeroLog(log.Error().Str("func", "TrackUsecase.Proceed"), err)
		http.Error(w, err.Message(), err.Status())
		return
	}

	if err := uc.WriteResponse(w, result, "application/json"); err != nil {
		log.Error().Err(err).Msg("failid write response")
	}
}
