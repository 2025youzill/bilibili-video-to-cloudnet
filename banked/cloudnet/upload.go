package cloudnet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"bvtc/client"
	"bvtc/constant"
	"bvtc/log"

	"github.com/chaunsin/netease-cloud-music/api/weapi"
	"github.com/chaunsin/netease-cloud-music/pkg/utils"
	"github.com/dhowden/tag"
)

var ctx context.Context = context.Background()

func UploadToNetCloud(filename string, splaylist bool, pid int64) error {
	// 检查文件是否存在
	ext := filepath.Ext(filename)
	bitrate := constant.BitRate

	api, err := client.GetNetcloudApi()
	if err != nil {
		log.Logger.Error("client fail to init", log.Any("err : ", err))
		return errors.New("client fail to init")
	}

	// 读取文件
	file, err := os.Open(filename)
	if err != nil {
		log.Logger.Error("fail to open file", log.Any("err : ", err))
		return errors.New("file error")
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.Logger.Error("fail to start file", log.Any("err : ", err))
		return errors.New("file error")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		log.Logger.Error("fail to read file", log.Any("err : ", err))
		return errors.New("file error")
	}

	md5, err := utils.MD5Hex(data)
	if err != nil {
		log.Logger.Error("fail to change to MD5Hex", log.Any("err : ", err))
		return errors.New("file error")
	}

	// 重新设置文件指针到开头
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Logger.Error("fail to set header", log.Any("err : ", err))
		return errors.New("file error")
	}

	// 检查此文件是否需要上传
	checkReq := weapi.CloudUploadCheckReq{
		Bitrate: bitrate,
		Ext:     ext,
		Length:  fmt.Sprintf("%d", stat.Size()),
		Md5:     md5,
		SongId:  "0",
		Version: "1",
	}
	resp, err := api.CloudUploadCheck(ctx, &checkReq)
	if err != nil {
		log.Logger.Error("fail to get token", log.Any("err : ", err), log.Any("Code : ", resp.Code))
		return errors.New("fail to get token")
	}
	if resp.Code != 200 {
		log.Logger.Error("token Code is not 200", log.Any("Code : ", resp.Code))
		return errors.New("token Code is not compare")
	}

	// 获取上传凭证
	allocReq := weapi.CloudTokenAllocReq{
		Bucket:     "", // jd-musicrep-privatecloud-audio-public
		Ext:        ext,
		Filename:   filepath.Base(filename),
		Local:      "false",
		NosProduct: "3",
		Type:       "audio",
		Md5:        md5,
	}
	allocResp, err := api.CloudTokenAlloc(ctx, &allocReq)
	if err != nil {
		log.Logger.Error("fail to get token", log.Any("err : ", err), log.Any("Code : ", resp.Code))
		return errors.New("fail to get token")
	}
	if allocResp.Code != 200 {
		log.Logger.Error("token Code is not 200", log.Any("Code : ", resp.Code))
		return errors.New("token Code is not compare")
	}

	// 上传文件
	if resp.NeedUpload {
		uploadReq := weapi.CloudUploadReq{
			Bucket:    allocResp.Bucket,
			ObjectKey: allocResp.ObjectKey,
			Token:     allocResp.Token,
			Filepath:  filename,
		}
		uploadResp, err := api.CloudUpload(ctx, &uploadReq)
		if err != nil {
			log.Logger.Error("fail to upload", log.Any("err : ", err), log.Any("Code : ", uploadResp.ErrCode))
			return errors.New("fail to upload")
		}
		if uploadResp.ErrCode != "" {
			log.Logger.Error("fail to upload", log.Any("Code : ", uploadResp.ErrCode))
			return errors.New("upload Code is not compare")
		}
	}

	// 上传歌曲相关信息
	metadata, err := tag.ReadFrom(file)
	if err != nil {
		log.Logger.Error("fail to upload", log.Any("err : ", err))
		return errors.New("fail to upload")
	}
	InfoReq := weapi.CloudInfoReq{
		Md5:        md5,
		SongId:     resp.SongId,
		Filename:   stat.Name(),
		Song:       utils.Ternary(metadata.Title() != "", metadata.Title(), filepath.Base(filename)),
		Album:      utils.Ternary(metadata.Album() != "", metadata.Album(), "未知专辑"),
		Artist:     utils.Ternary(metadata.Artist() != "", metadata.Artist(), "未知艺术家"),
		Bitrate:    bitrate,
		ResourceId: allocResp.ResourceID,
	}

	infoResp, err := api.CloudInfo(ctx, &InfoReq)
	if err != nil {
		log.Logger.Error("fail to upload music imformation", log.Any("err : ", err))
		return errors.New("fail to upload music imformation")
	}
	if infoResp.Code != 200 {
		log.Logger.Error("fail to upload music imformation", log.Any("Code : ", infoResp.Code))
		return errors.New("upload Code is not compare")
	}

	// 对上传得歌曲进行发布，和自己账户做关联,不然云盘列表看不到上传得歌曲信息
	publishReq := weapi.CloudPublishReq{
		SongId: infoResp.SongId,
	}
	publishResp, err := api.CloudPublish(ctx, &publishReq)
	if err != nil {
		log.Logger.Error("fail to publish", log.Any("err : ", err))
		return errors.New("fail to publish")
	}

	switch publishResp.Code {
	case 200:
		log.Logger.Info("success to upload", log.Any("filename : ", filename))
	case 201:
		log.Logger.Info("the music already exists", log.Any("filename : ", filename))
		return errors.New("the music already exists")
	default:
		log.Logger.Error("fail to publish", log.Any("filename : ", filename))
		return errors.New("upload Code is not compare")
	}

	// 判断是否要加入歌单还是只保存网盘
	if splaylist {
		trackId, err := strconv.ParseInt(infoResp.SongId, 10, 64)
		if err != nil {
			log.Logger.Error("转换歌曲ID失败", log.Any("err", err))
			return errors.New("fail to convert song id")
		}
		err = UploadToPlaylist(UploadToMusicReq{
			Pid:      pid,
			TrackIds: trackId,
		})
		if err != nil {
			log.Logger.Error("添加到歌单失败", log.Any("err", err))
			return errors.New("fail to add to playlist")
		}
	}
	return nil
}
