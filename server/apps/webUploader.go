package apps

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"iCloud/log"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
)

const (
	UPLOAD_TEMP_DIR = "D:/go_project/iCloud/uploadTemp"
)

// upload file by plugin webUploader on browser
// webUploader upload one file in one post request
// big file is upload by multi block, and each block is uploaded by one post request
func FileUpload(ctx *gin.Context) {
	var (
		multiForm    *multipart.Form
		taskId       = ctx.PostForm("task_id")
		chunkId      = ctx.PostForm("chunk")
		files        []*multipart.FileHeader
		fileName     string
		tempFilename string
		rsp          = make(gin.H)
		err          error
		exist        bool
		m            = "apps.webUploader.FileUpload()"
		httpStatus   int
	)
	if multiForm, err = ctx.MultipartForm(); err != nil {
		log.Logger.Errorf("%s error, get file in post request error: %v", m, err)
		// 一些 Web 应用可能会将 308 Permanent Redirect 以一种非标准的方式使用以及用作其他用途。
		// 例如，Google Drive 会使用 308 Resume Incomplete 状态码来告知客户端文件上传终止且不完整
		httpStatus = 308
		goto RESPONSE
	}

	if files, exist = multiForm.File["file"]; !exist {
		log.Logger.Errorf("%s error, get file from form error, file dose not exist", m)
		httpStatus = 308
		goto RESPONSE
	}

	if len(files) != 1 {
		log.Logger.Errorf("%s, error: number of file which upload by webUploader is %d, but it should be 1", m, len(files))
		httpStatus = 308
		goto RESPONSE
	}

	fileName = files[0].Filename

	if err = createDir(path.Join(UPLOAD_TEMP_DIR, taskId)); err != nil {
		httpStatus = 308
		goto RESPONSE
	}

	tempFilename = path.Join(UPLOAD_TEMP_DIR, taskId, fmt.Sprintf("%s-%s", fileName, chunkId))

	if err = ctx.SaveUploadedFile(files[0], tempFilename); err != nil {
		log.Logger.Errorf("%s, error: save file[%s] which upload by webUploader error: %v", m, fileName, err)
		httpStatus = 308
		goto RESPONSE
	}

	httpStatus, rsp["ErrorCode"], rsp["Data"] = http.StatusOK, 0, struct{}{}

RESPONSE:
	ctx.JSON(httpStatus, rsp)
}

// blockNum is number of block file
// fileName is file name of upload, and it is the name of merged file
// taskId is created by webUploader
func BlockFileMerge(ctx *gin.Context) {
	var (
		taskId          = ctx.PostForm("task_id")
		fileName        = ctx.PostForm("fileName")
		file, blockFile *os.File
		files           []os.FileInfo
		m               = "apps.webUpload.BlockFileMerge()"
		blockContent    = new([]byte)
		err             error
		rsp             = make(gin.H)
	)
	finalFileName := path.Join(UPLOAD_TEMP_DIR, taskId, fileName)
	fmt.Println(finalFileName)
	if files, err = ioutil.ReadDir(path.Join(UPLOAD_TEMP_DIR, taskId)); err != nil {
		log.Logger.Errorf("%s error, create file which is used to merge upload block file error: %v", m, err)
		rsp["ErrorCode"], rsp["Data"] = 1, "block file merge error"
	}
	switch {
	case len(files) == 1:
		if err = os.Rename(path.Join(UPLOAD_TEMP_DIR, taskId, files[0].Name()), finalFileName); err != nil {
			log.Logger.Errorf("%s error, rename file[%s] error: %v", m, finalFileName, err)
			rsp["ErrorCode"], rsp["Data"] = 1, "block file merge error"
			goto RESPONSE
		}
	case len(files) > 1:
		if file, err = os.OpenFile(finalFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm); err != nil {
			log.Logger.Errorf("%s error, create final file which is used to merge upload block file error: %v", m, err)
			rsp["ErrorCode"], rsp["Data"] = 1, "block file merge error"
			goto RESPONSE
		}
		defer file.Close()

		for i := 0; i < len(files); i++ {
			blockName := path.Join(UPLOAD_TEMP_DIR, taskId, fmt.Sprintf("%s-%s", fileName, strconv.Itoa(i)))
			if blockFile, err = os.OpenFile(blockName, os.O_RDONLY, os.ModePerm, ); err != nil {
				log.Logger.Errorf("%s error, open block file[%s] to merge error: %v", m, blockName, err)
				rsp["ErrorCode"], rsp["Data"] = 1, "block file merge error"
				goto RESPONSE
			}
			if *blockContent, err = ioutil.ReadAll(blockFile); err != nil {
				log.Logger.Errorf("%s error, read block file[%s] to final file error: %v", m, blockName, err)
				rsp["ErrorCode"], rsp["Data"] = 1, "block file merge error"
				return
			}
			blockFile.Close()
			if _, err = file.Write(*blockContent); err != nil {
				log.Logger.Errorf("%s error, write block file[%s] to final file error: %v", m, blockName, err)
				rsp["ErrorCode"], rsp["Data"] = 1, "block file merge error"
				goto RESPONSE
			}
			if err = os.Remove(blockName); err != nil {
				log.Logger.Errorf("%s error, remove block[%s] error: %v", m, blockName, err)
				rsp["ErrorCode"], rsp["Data"] = 1, "block file merge error"
				goto RESPONSE
			}
		}
	default:
		log.Logger.Errorf("%s error, no block file in dir[%s]", m, path.Join(UPLOAD_TEMP_DIR, taskId))
		rsp["ErrorCode"], rsp["Data"] = 1, "block file merge error"
		goto RESPONSE
	}

RESPONSE:
	ctx.JSON(http.StatusOK, rsp)
}

func createDir(dir string) (err error) {
	var (
		m = "apps.webUpload.createDir()"
	)
	if err = removeDir(dir); err != nil {
		return
	}
	if err = os.Mkdir(dir, os.ModePerm); err != nil {
		log.Logger.Errorf("%s error, mkdir %s error: %v", m, dir, err)
		return
	}

	return nil
}

func removeDir(dir string) (err error) {
	var (
		m            = "apps.webUploader.removeDir()"
		removeDirErr error
	)
	if _, err = os.Stat(dir); err != nil {
		if os.IsExist(err) {
			removeDirErr = os.Remove(dir)
		}
	} else {
		removeDirErr = os.Remove(dir)
	}

	if removeDirErr != nil {
		log.Logger.Errorf("%s error, remove old file[%s] in %s error: %v", m, dir, UPLOAD_TEMP_DIR, removeDirErr)
		return removeDirErr
	}
	return nil
}

func storageFileTo(fileName string) (err error) {
	// TODO storage file to distributed storage system

	err = removeDir(fileName)
	return
}
