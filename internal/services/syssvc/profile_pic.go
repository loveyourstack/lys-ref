package syssvc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/loveyourstack/lys/lysformfile"
)

func (svc Service) SetUserProfilePic(ctx context.Context, userId int64, uploadFile lysformfile.UploadFile, uploadsPath string) (storedFileName string, err error) {

	defer uploadFile.File.Close()

	// generate random 4-byte hex string for unique file naming
	rnd := make([]byte, 4)
	if _, err := rand.Read(rnd); err != nil {
		return "", fmt.Errorf("rand.Read failed: %w", err)
	}

	// determine stored file extension based on MIME type and original file name
	ext := lysformfile.ChooseStoredExtension(uploadFile.FileHeader, uploadFile.MimeType)

	// generate stored file name
	storedFileName = fmt.Sprintf("%s-u%d-%s%s", time.Now().Format("20060102"), userId, hex.EncodeToString(rnd), ext)

	// upload to s3
	err = svc.AwsApiClient.PutS3Object(ctx, svc.S3Bucket, "profiles/"+storedFileName, uploadFile.File, uploadFile.MimeType)
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
