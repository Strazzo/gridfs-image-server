package paint

import (
	"fmt"
	"image"
	"log"
	"sync"

	"github.com/disintegration/imaging"
)

//ResizeType defines which image manipulation will or can be used
//each ResizeType has its corresponding Resizer
type ResizeType string

//AvailableResizeTypes contains a map with both key and value
//of all available resize types
//simply check with _, found := AvaiableResizeTypes[ResizeType]
var defaultAvailableResizeTypes = map[ResizeType]ResizeType{
	TypeResize: TypeResize,
	TypeCrop:   TypeCrop,
	TypeFit:    TypeFit,
}

var extraAllowedTypes = map[ResizeType]Resizer{}
var extraResizerLock = sync.Mutex{}

const (
	//TypeResize will either force the given sizes, or resize via original ratio when either height or width is not specified
	TypeResize ResizeType = "resize"
	//TypeCrop will generate an image with exact sizes, but only a part of the image is visible
	TypeCrop ResizeType = "crop"
	//TypeFit will resize the image according to the original ratio, but will not exceed the given bounds
	TypeFit ResizeType = "fit"
)

//AddResizer allows a custom resizer to use
func AddResizer(resizeType ResizeType, resizer Resizer) {
	extraResizerLock.Lock()
	defer extraResizerLock.Unlock()
	log.Printf("Registering additional resizer %s\n", resizeType)
	extraAllowedTypes[resizeType] = resizer
}

//GetAvailableTypes returns all available types
func GetAvailableTypes() map[ResizeType]ResizeType {
	extraResizerLock.Lock()
	defer extraResizerLock.Unlock()
	result := defaultAvailableResizeTypes
	for rtype := range extraAllowedTypes {
		result[rtype] = rtype
	}

	return result
}

//GetCustomResizers returns a read only list of custom resizers
func GetCustomResizers() map[ResizeType]Resizer {
	extraResizerLock.Lock()
	defer extraResizerLock.Unlock()

	return extraAllowedTypes
}

//Resizer can resize an image
//dstWidth and dstHeight are the desired output values
//but it is not promised that the output image has exactly those bounds
type Resizer interface {
	Resize(input image.Image, dstWidth, dstHeight int) (image.Image, error)
}

//newResizerByType returns a resizer for the given
//type. If an invalid type was given
//a plainResizer will be created
func newResizerByType(resizeType ResizeType, customResizer map[ResizeType]Resizer) Resizer {
	resizers := map[ResizeType]Resizer{
		TypeResize: plainResizer{},
		TypeFit:    fitResizer{},
		TypeCrop:   cropResizer{},
	}
	log.Println("%s %#v.", "resizers", customResizer)

	for rtype, resizer := range customResizer {
		resizers[rtype] = resizer
		log.Println("%s %#v.", rtype, resizer)
	}

	resizer, found := resizers[resizeType]

	if !found {
		// an error here would be a regression
		// so for now we use a fallback behaviour
		// in the future we can refactor it to
		// only support registered resizers
		resizer = resizers[TypeResize]
	}

	return resizer
}

type plainResizer struct {
}

func (p plainResizer) Resize(input image.Image, dstWidth, dstHeight int) (image.Image, error) {
	if dstWidth < 0 && dstHeight < 0 {
		return nil, fmt.Errorf("Either width or height must be greater zero to keep the existing ratio")
	}

	//since we use -1 as optional and imaging uses zero as optional
	//we change -1 to 0 to keep the aspect ratio
	if dstWidth < 0 {
		dstWidth = 0
	}

	if dstHeight < 0 {
		dstHeight = 0
	}

	return imaging.Resize(input, dstWidth, dstHeight, imaging.Lanczos), nil
}

type fitResizer struct {
}

func (f fitResizer) Resize(input image.Image, dstWidth, dstHeight int) (image.Image, error) {
	if dstWidth < 0 || dstHeight < 0 {
		return nil, fmt.Errorf("Please specify both width and height for your target image")
	}

	originalBounds := input.Bounds()
	originalRatio := float64(originalBounds.Dx()) / float64(originalBounds.Dy())

	targetRatio := float64(dstWidth) / float64(dstHeight)

	if targetRatio < originalRatio {
		dstHeight = int(float64(dstWidth) / originalRatio)
	} else {
		dstWidth = int(float64(dstHeight) * originalRatio)
	}

	return imaging.Resize(input, int(dstWidth), int(dstHeight), imaging.Lanczos), nil
}

type cropResizer struct {
}

func (c cropResizer) Resize(input image.Image, dstWidth, dstHeight int) (image.Image, error) {
	if dstWidth < 0 && dstHeight < 0 {
		return nil, fmt.Errorf("Either width or height must be greater zero to keep the existing ratio")
	}

	originalBounds := input.Bounds()
	originalRatio := float64(originalBounds.Dx()) / float64(originalBounds.Dy())

	if dstWidth < 0 {
		dstWidth = int(float64(dstHeight) * originalRatio)
	}

	if dstHeight < 0 {
		dstHeight = int(float64(dstWidth) / originalRatio)
	}

	return imaging.Thumbnail(input, dstWidth, dstHeight, imaging.Lanczos), nil
}
