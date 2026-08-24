package syssvc

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"time"

	"github.com/loveyourstack/lys/lysformfile"
)

// SetUserProfilePic stores a user's profile picture as a JPEG in S3 and updates the user's record in the database with the stored file name.
// uploadFile must be a 400 px by 400 px gif, jpg, or png image file with a maximum size of 1 MB.
func (svc Service) SetUserProfilePic(ctx context.Context, userId int64, uploadFile lysformfile.UploadFile) (storedFileName string, err error) {

	defer uploadFile.File.Close()

	var fileReader io.Reader = uploadFile.File

	// convert to jpeg with fixed quality to reduce file size and ensure consistent format
	fileReader, err = toJpeg(uploadFile.File, uploadFile.MimeType)
	if err != nil {
		return "", fmt.Errorf("toJpeg failed: %w", err)
	}

	// generate random 4-byte hex string for unique file naming
	rnd := make([]byte, 4)
	if _, err := rand.Read(rnd); err != nil {
		return "", fmt.Errorf("rand.Read failed: %w", err)
	}

	// generate stored file name
	storedFileName = fmt.Sprintf("%s-u%d-%s.jpg", time.Now().Format("20060102"), userId, hex.EncodeToString(rnd))

	// upload to s3
	err = svc.AwsApiClient.PutS3Object(ctx, svc.S3Bucket, "profiles/"+storedFileName, fileReader, "image/jpeg")
	if err != nil {
		return "", fmt.Errorf("svc.AwsApiClient.PutS3Object failed: %w", err)
	}

	// write stored file name to user db record
	err = svc.SysUserStore.UpdatePartial(ctx, map[string]any{"profile_pic": storedFileName}, userId)
	if err != nil {
		return "", fmt.Errorf("svc.SysUserStore.UpdatePartial failed: %w", err)
	}

	return storedFileName, nil
}

// toJpeg converts an image to JPEG format with 85% quality.
// NB: only jpg, gif and png are supported: see std library imports above.
// Animated gifs will be converted to a single frame jpeg.
func toJpeg(r io.Reader, mimeType string) (io.Reader, error) {

	// decode the image
	src, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("image.Decode failed: %w", err)
	}

	// if png, draw white background to remove transparency
	if mimeType == "image/png" {
		bg := image.NewRGBA(src.Bounds())
		draw.Draw(bg, bg.Bounds(), image.White, image.Point{}, draw.Src)
		draw.Draw(bg, bg.Bounds(), src, src.Bounds().Min, draw.Over)
		src = bg
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("jpeg.Encode failed: %w", err)
	}

	return &buf, nil
}
