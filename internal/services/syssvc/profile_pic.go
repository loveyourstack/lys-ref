package syssvc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysformfile"
)

func (svc Service) SetUserProfilePic(ctx context.Context, uploadFile lysformfile.UploadFile, uploadsPath string) (err error) {

	defer uploadFile.File.Close()

	// get user id from context
	userId := lys.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return lyserr.User{Message: "user not authenticated", StatusCode: http.StatusForbidden}
	}

	// stream to uploads
	uploadResp, err := lysformfile.StreamToDisk([]lysformfile.UploadFile{uploadFile}, uploadsPath, userId, svc.Logger)
	if err != nil {
		return fmt.Errorf("lysformfile.StreamToDisk failed: %w", err)
	}

	if len(uploadResp.FileResults) != 1 {
		return fmt.Errorf("expected 1 file result, got %d", len(uploadResp.FileResults))
	}
	fileResult := uploadResp.FileResults[0]

	// TODO: upload to s3

	// write stored file name to user db record
	err = svc.SysUserStore.UpdatePartial(ctx, map[string]any{"profile_pic": fileResult.StoredName}, userId)
	if err != nil {
		return fmt.Errorf("svc.SysUserStore.UpdatePartial failed: %w", err)
	}

	return nil
}
