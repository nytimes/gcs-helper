package vodmodule

// Mapping represents the response expected by the vod-module.
//
// See https://github.com/kaltura/nginx-vod-module#mapping-response-format.
type Mapping struct {
	Sequences []Sequence `json:"sequences"`
}

// Sequence represents a list of media clips.
type Sequence struct {
	Clips []Clip `json:"clips"`
}

// Clip represents a single media file.
type Clip struct {
	Type string `json:"type"`
	Path string `json:"path"`
}
