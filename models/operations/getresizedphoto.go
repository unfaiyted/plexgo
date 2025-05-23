// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package operations

import (
	"encoding/json"
	"fmt"
	"github.com/unfaiyted/plexgo/internal/utils"
	"net/http"
)

// MinSize - images are always scaled proportionally. A value of '1' in minSize will make the smaller native dimension the dimension resized against.
type MinSize int64

const (
	MinSizeZero MinSize = 0
	MinSizeOne  MinSize = 1
)

func (e MinSize) ToPointer() *MinSize {
	return &e
}
func (e *MinSize) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case 0:
		fallthrough
	case 1:
		*e = MinSize(v)
		return nil
	default:
		return fmt.Errorf("invalid value for MinSize: %v", v)
	}
}

// Upscale - allow images to be resized beyond native dimensions.
type Upscale int64

const (
	UpscaleZero Upscale = 0
	UpscaleOne  Upscale = 1
)

func (e Upscale) ToPointer() *Upscale {
	return &e
}
func (e *Upscale) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case 0:
		fallthrough
	case 1:
		*e = Upscale(v)
		return nil
	default:
		return fmt.Errorf("invalid value for Upscale: %v", v)
	}
}

type GetResizedPhotoRequest struct {
	// The width for the resized photo
	Width float64 `queryParam:"style=form,explode=true,name=width"`
	// The height for the resized photo
	Height float64 `queryParam:"style=form,explode=true,name=height"`
	// The opacity for the resized photo
	Opacity int64 `default:"100" queryParam:"style=form,explode=true,name=opacity"`
	// The width for the resized photo
	Blur float64 `queryParam:"style=form,explode=true,name=blur"`
	// images are always scaled proportionally. A value of '1' in minSize will make the smaller native dimension the dimension resized against.
	MinSize MinSize `queryParam:"style=form,explode=true,name=minSize"`
	// allow images to be resized beyond native dimensions.
	Upscale Upscale `queryParam:"style=form,explode=true,name=upscale"`
	// path to image within Plex
	URL string `queryParam:"style=form,explode=true,name=url"`
}

func (g GetResizedPhotoRequest) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(g, "", false)
}

func (g *GetResizedPhotoRequest) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &g, "", false, false); err != nil {
		return err
	}
	return nil
}

func (o *GetResizedPhotoRequest) GetWidth() float64 {
	if o == nil {
		return 0.0
	}
	return o.Width
}

func (o *GetResizedPhotoRequest) GetHeight() float64 {
	if o == nil {
		return 0.0
	}
	return o.Height
}

func (o *GetResizedPhotoRequest) GetOpacity() int64 {
	if o == nil {
		return 0
	}
	return o.Opacity
}

func (o *GetResizedPhotoRequest) GetBlur() float64 {
	if o == nil {
		return 0.0
	}
	return o.Blur
}

func (o *GetResizedPhotoRequest) GetMinSize() MinSize {
	if o == nil {
		return MinSize(0)
	}
	return o.MinSize
}

func (o *GetResizedPhotoRequest) GetUpscale() Upscale {
	if o == nil {
		return Upscale(0)
	}
	return o.Upscale
}

func (o *GetResizedPhotoRequest) GetURL() string {
	if o == nil {
		return ""
	}
	return o.URL
}

type GetResizedPhotoResponse struct {
	// HTTP response content type for this operation
	ContentType string
	// HTTP response status code for this operation
	StatusCode int
	// Raw HTTP response; suitable for custom response parsing
	RawResponse *http.Response
}

func (o *GetResizedPhotoResponse) GetContentType() string {
	if o == nil {
		return ""
	}
	return o.ContentType
}

func (o *GetResizedPhotoResponse) GetStatusCode() int {
	if o == nil {
		return 0
	}
	return o.StatusCode
}

func (o *GetResizedPhotoResponse) GetRawResponse() *http.Response {
	if o == nil {
		return nil
	}
	return o.RawResponse
}
