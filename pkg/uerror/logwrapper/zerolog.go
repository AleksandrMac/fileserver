package logwrapper

import (
	"github.com/AleksandrMac/fileserver/pkg/uerror"
	"github.com/rs/zerolog"
)

func ZeroLog(l *zerolog.Event, err uerror.UError) {
	l = l.Int("status", err.Status()).Err(err)

	for k, v := range err.Payload() {
		l = l.Any(k, v)
	}

	l.Msg(err.Message())
}
