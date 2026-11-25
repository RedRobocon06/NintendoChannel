package info

import (
	"NintendoChannel/common"
	"NintendoChannel/constants"
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/image/draw"
)

var regionToStr = map[constants.Region]string{
	constants.Japan: "JA",
	constants.PAL:   "EN",
	constants.NTSC:  "US",
}
var titleTypeToStr = map[constants.TitleType]string{
	constants.Wii:             "wii",
	constants.NintendoDS:      "ds",
	constants.NintendoThreeDS: "3ds",
	constants.ThreeDSDownload: "3ds",
}

var consoleToImageType = map[constants.TitleType]string{
	constants.Wii:             "cover",
	constants.NintendoDS:      "cover",
	constants.NintendoThreeDS: "cover",
	constants.ThreeDSDownload: "cover",
}

var consoleToTempImageType = map[constants.TitleType][]byte{
	constants.Wii:             PlaceholderWii,
	constants.NintendoDS:      PlaceholderDS,
	constants.NintendoThreeDS: Placeholder3DS,
	constants.ThreeDSDownload: Placeholder3DS,
}

//go:embed 3ds.jpg
var Placeholder3DS []byte

//go:embed ds.jpg
var PlaceholderDS []byte

//go:embed wii.jpg
var PlaceholderWii []byte

var coverClient = &http.Client{Timeout: 5 * time.Second}

func fetchCoverImage(console, primaryRegion, gameID string) image.Image {
	regions := []string{primaryRegion, "US", "EN", "JA", "AU", "DE", "FR", "ES", "IT", "NL"}
	coverTypes := []string{"cover", "coverM"}
	exts := []string{"png", "jpg"}
	if console == "ds" || console == "3ds" {
		exts = []string{"jpg", "png"}
	}

	host := "art." + "gametdb" + ".com"
	seenRegions := map[string]bool{}
	for _, r := range regions {
		if r == "" || seenRegions[r] {
			continue
		}
		seenRegions[r] = true

		for _, ct := range coverTypes {
			for _, ext := range exts {
				url := fmt.Sprintf("https://%s/%s/%s/%s/%s.%s", host, console, ct, r, gameID, ext)
				resp, err := coverClient.Get(url)
				if err != nil {
					continue
				}
				if resp.StatusCode == http.StatusOK {
					img, _, decErr := image.Decode(resp.Body)
					resp.Body.Close()
					if decErr == nil && img != nil {
						return img
					}
					continue
				}
				resp.Body.Close()
			}
		}
	}
	return nil
}

func (i *Info) WriteCoverArt(buffer *bytes.Buffer, titleType constants.TitleType, region constants.Region, gameID string) {
	startOffset := buffer.Len()
	console := titleTypeToStr[titleType]
	if console == "" {
		console = "wii"
	}
	prefRegion := regionToStr[region]
	if prefRegion == "" {
		prefRegion = "US"
	}

	path := fmt.Sprintf("%s/%s/%s/%s.jpg", common.GetConfig().ImagesPath, console, prefRegion, gameID)
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		buffer.Write(data)
		i.Header.PictureSize = uint32(buffer.Len() - startOffset)
		return
	}

	if os.Getenv("NC_NO_COVERS") != "" {
		buffer.Write(consoleToTempImageType[titleType])
		i.Header.PictureSize = uint32(buffer.Len() - startOffset)
		return
	}

	img := fetchCoverImage(console, prefRegion, gameID)
	if img == nil {
		buffer.Write(consoleToTempImageType[titleType])
		i.Header.PictureSize = uint32(buffer.Len() - startOffset)
		return
	}

	x, y := img.Bounds().Dx(), img.Bounds().Dy()

	if titleType != constants.NintendoThreeDS && titleType != constants.NintendoDS && titleType != constants.ThreeDSDownload {
		img = resize(img, int(float64(x)*(384.0/float64(y))), 384)
	} else {
		img = resize(img, 384, int(float64(y)*(384.0/float64(x))))
	}

	offsetX := (384 - img.Bounds().Dx()) / 2
	offsetY := (384 - img.Bounds().Dy()) / 2
	offset := image.Pt(offsetX, offsetY)

	newImage := image.NewRGBA(image.Rect(0, 0, 384, 384))
	draw.Draw(newImage, newImage.Bounds(), &image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(newImage, img.Bounds().Add(offset), img, image.Point{}, draw.Over)

	imgBuffer := new(bytes.Buffer)
	err := jpeg.Encode(imgBuffer, newImage, &jpeg.Options{Quality: 85})
	common.CheckError(err)

	buffer.Write(imgBuffer.Bytes())
	i.Header.PictureSize = uint32(buffer.Len() - startOffset)

	if err := os.MkdirAll(filepath.Dir(path), 0777); err == nil {
		_ = os.WriteFile(path, imgBuffer.Bytes(), 0666)
	}
}

func (i *Info) WriteRatingImage(buffer *bytes.Buffer, region constants.Region) {
	if i.RatingID < 8 {
		return
	}
	i.Header.RatingPictureOffset = i.GetCurrentSize(buffer)

	regionToRatingGroup := map[constants.Region]constants.RatingGroup{
		constants.Japan: constants.CERO,
		constants.NTSC:  constants.ESRB,
		constants.PAL:   constants.PEGI,
	}

	buffer.Write(constants.ImagesSmall[regionToRatingGroup[region]][i.RatingID-8])
	i.Header.RatingPictureSize = uint32(len(constants.ImagesSmall[regionToRatingGroup[region]][i.RatingID-8]))
}

func resize(origImage image.Image, x, y int) image.Image {
	newImage := image.NewRGBA(image.Rect(0, 0, x, y))
	draw.BiLinear.Scale(newImage, newImage.Bounds(), origImage, origImage.Bounds(), draw.Over, nil)
	return newImage
}

func (i *Info) WriteRatingDescriptor(buffer *bytes.Buffer, region constants.Region, RatingDescriptors []string) {
	maxDescriptors := min(len(RatingDescriptors), 7)
	for j := 0; j < maxDescriptors; j++ {
		s := RatingDescriptors[j]

		if s == "" {
			continue
		}

		var descriptorImage []byte
		switch region {
		case constants.Japan:
			descriptorImage = constants.CERODescriptors[s]
		case constants.PAL:
			descriptorImage = constants.PEGIDescriptors[s]
		case constants.NTSC:
			descriptorImage = constants.ESRBDescriptors[s]
		}

		i.Header.DetailedRatingPictureTable[j].PictureOffset = i.GetCurrentSize(buffer)
		buffer.Write(descriptorImage)
		i.Header.DetailedRatingPictureTable[j].PictureSize = uint32(len(descriptorImage))
	}
}
