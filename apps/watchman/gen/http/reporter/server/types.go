// Code generated by goa v3.4.3, DO NOT EDIT.
//
// reporter HTTP server types
//
// Command:
// $ goa gen github.com/lbryio/lbrytv/apps/watchman/design -o apps/watchman

package server

import (
	"unicode/utf8"

	reporter "github.com/lbryio/lbrytv/apps/watchman/gen/reporter"
	goa "goa.design/goa/v3/pkg"
)

// AddRequestBody is the type of the "reporter" service "add" endpoint HTTP
// request body.
type AddRequestBody struct {
	// LBRY URL
	URL *string `form:"url,omitempty" json:"url,omitempty" xml:"url,omitempty"`
	// Event duration, ms
	Duration *int32 `form:"duration,omitempty" json:"duration,omitempty" xml:"duration,omitempty"`
	// Current playback report stream position, ms
	Position *int32 `form:"position,omitempty" json:"position,omitempty" xml:"position,omitempty"`
	// Relative stream position, pct, 0—100
	RelPosition *int32 `form:"rel_position,omitempty" json:"rel_position,omitempty" xml:"rel_position,omitempty"`
	// Rebuffering events count
	RebufCount *int32 `form:"rebuf_count,omitempty" json:"rebuf_count,omitempty" xml:"rebuf_count,omitempty"`
	// Rebuffering events total duration, ms
	RebufDuration *int32 `form:"rebuf_duration,omitempty" json:"rebuf_duration,omitempty" xml:"rebuf_duration,omitempty"`
	// Video format, stb (binary stream) or HLS
	Format *string `form:"format,omitempty" json:"format,omitempty" xml:"format,omitempty"`
	// Cache status of video
	Cache *string `form:"cache,omitempty" json:"cache,omitempty" xml:"cache,omitempty"`
	// Player server name
	Player *string `form:"player,omitempty" json:"player,omitempty" xml:"player,omitempty"`
	// User ID
	UserID *int32 `form:"user_id,omitempty" json:"user_id,omitempty" xml:"user_id,omitempty"`
	// Client download rate, bit/s
	Rate *int32 `form:"rate,omitempty" json:"rate,omitempty" xml:"rate,omitempty"`
	// Client device
	Device *string `form:"device,omitempty" json:"device,omitempty" xml:"device,omitempty"`
}

// NewAddPlaybackReport builds a reporter service add endpoint payload.
func NewAddPlaybackReport(body *AddRequestBody) *reporter.PlaybackReport {
	v := &reporter.PlaybackReport{
		URL:           *body.URL,
		Duration:      *body.Duration,
		Position:      *body.Position,
		RelPosition:   *body.RelPosition,
		RebufCount:    *body.RebufCount,
		RebufDuration: *body.RebufDuration,
		Format:        *body.Format,
		Cache:         body.Cache,
		Player:        *body.Player,
		UserID:        *body.UserID,
		Rate:          body.Rate,
		Device:        *body.Device,
	}

	return v
}

// ValidateAddRequestBody runs the validations defined on AddRequestBody
func ValidateAddRequestBody(body *AddRequestBody) (err error) {
	if body.URL == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("url", "body"))
	}
	if body.Duration == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("duration", "body"))
	}
	if body.Position == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("position", "body"))
	}
	if body.RelPosition == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("rel_position", "body"))
	}
	if body.RebufCount == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("rebuf_count", "body"))
	}
	if body.RebufDuration == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("rebuf_duration", "body"))
	}
	if body.Format == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("format", "body"))
	}
	if body.Player == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("player", "body"))
	}
	if body.UserID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("user_id", "body"))
	}
	if body.Device == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("device", "body"))
	}
	if body.URL != nil {
		if utf8.RuneCountInString(*body.URL) > 512 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("body.url", *body.URL, utf8.RuneCountInString(*body.URL), 512, false))
		}
	}
	if body.Duration != nil {
		if *body.Duration < 0 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.duration", *body.Duration, 0, true))
		}
	}
	if body.Duration != nil {
		if *body.Duration > 60000 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.duration", *body.Duration, 60000, false))
		}
	}
	if body.Position != nil {
		if *body.Position < 0 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.position", *body.Position, 0, true))
		}
	}
	if body.RelPosition != nil {
		if *body.RelPosition < 0 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.rel_position", *body.RelPosition, 0, true))
		}
	}
	if body.RelPosition != nil {
		if *body.RelPosition > 100 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.rel_position", *body.RelPosition, 100, false))
		}
	}
	if body.RebufCount != nil {
		if *body.RebufCount < 0 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.rebuf_count", *body.RebufCount, 0, true))
		}
	}
	if body.RebufDuration != nil {
		if *body.RebufDuration < 0 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.rebuf_duration", *body.RebufDuration, 0, true))
		}
	}
	if body.RebufDuration != nil {
		if *body.RebufDuration > 60000 {
			err = goa.MergeErrors(err, goa.InvalidRangeError("body.rebuf_duration", *body.RebufDuration, 60000, false))
		}
	}
	if body.Format != nil {
		if !(*body.Format == "stb" || *body.Format == "hls") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("body.format", *body.Format, []interface{}{"stb", "hls"}))
		}
	}
	if body.Cache != nil {
		if !(*body.Cache == "local" || *body.Cache == "player" || *body.Cache == "miss") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("body.cache", *body.Cache, []interface{}{"local", "player", "miss"}))
		}
	}
	if body.Player != nil {
		if utf8.RuneCountInString(*body.Player) > 64 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("body.player", *body.Player, utf8.RuneCountInString(*body.Player), 64, false))
		}
	}
	if body.Device != nil {
		if !(*body.Device == "ios" || *body.Device == "adr" || *body.Device == "web") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("body.device", *body.Device, []interface{}{"ios", "adr", "web"}))
		}
	}
	return
}
