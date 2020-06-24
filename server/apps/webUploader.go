package apps

import (
	"errors"
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

var (
	UploadTempDir = "D:/go_project/iCloud/uploadTemp"
)

// upload file by plugin webUploader on browser
// webUploader upload one file in one post request
// big file is upload by multi block, and each block is uploaded by one post request
func FileUpload(ctx *gin.Context) {
	var (
		multiForm    *multipart.Form
		taskId       = ctx.PostForm("task_id")
		chunkId      = ctx.PostForm("chunk")
		blockNum     = ctx.PostForm("chunks")
		n, sum       int
		files        []*multipart.FileHeader
		fileName     string
		tempDir      string
		tempFileName string
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

	fmt.Printf("%s--%s--%s--%s\n", fileName, taskId, blockNum, chunkId)

	tempDir = path.Join(UploadTempDir, fileName+"--"+taskId)
	if (blockNum == "" && chunkId == "") || chunkId == "0" {
		if err = createDir(tempDir); err != nil {
			log.Logger.Errorf("create dir %s error", tempDir)
			httpStatus = 308
			goto RESPONSE
		}
	}

	if blockNum == "" && chunkId == "" {
		tempFileName = path.Join(tempDir, fileName)
		if err = ctx.SaveUploadedFile(files[0], path.Join(tempDir, fileName)); err != nil {
			log.Logger.Errorf("storage no chunked file %s error: %v", tempFileName, err)
			httpStatus = 308
			goto RESPONSE
		}
	} else {
		tempFileName = path.Join(tempDir, fileName + "--" + chunkId)
		if err = ctx.SaveUploadedFile(files[0], tempFileName); err != nil {
			log.Logger.Errorf("storage chunked file %s error: %v", tempFileName, err)
			httpStatus = 308
			goto RESPONSE
		}
		if n, err = strconv.Atoi(chunkId); err != nil {
			log.Logger.Errorf("chunkId [%s] to int error: %v", chunkId, err)
			httpStatus = 308
			goto RESPONSE
		}
		if sum, err = strconv.Atoi(blockNum); err != nil {
			log.Logger.Errorf("sum of block [%s] to int error: %v", blockNum, err)
			httpStatus = 308
			goto RESPONSE
		}

		if n + 1 == sum {
			if err = blockMerge(fileName, taskId, sum); err != nil {
				log.Logger.Errorf("merge file %s error: %v", fileName, err)
				httpStatus = 308
				goto RESPONSE
			}
		}
	}

	if err = storageFileTo(path.Join(tempDir, fileName)); err != nil {
		httpStatus = 308
		goto RESPONSE
	}

	httpStatus, rsp["ErrorCode"], rsp["Data"] = http.StatusOK, 0, struct{}{}

RESPONSE:
	ctx.JSON(httpStatus, rsp)
}

// sumBlock is number of block file
// fileName is file name of upload, and it is the name of merged file
// taskId is created by webUploader
func blockMerge(fileName, taskId string, sumBlock int) (err error) {
	var (
		finalFile, blockFile *os.File
		finalFileName = path.Join(UploadTempDir, fileName + "--" + taskId, fileName)
		blockContent []byte
	)

	if finalFile, err = os.OpenFile(finalFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm); err != nil {
		log.Logger.Errorf("create final file[%s] used to merge blocks error: %v", finalFileName, err)
		return errors.New("create final file error")
	}
	defer finalFile.Close()

	for chunkId := 0; chunkId < sumBlock; chunkId++ {
		blockFileName := finalFileName + "--" + strconv.Itoa(chunkId)
		if blockFile, err = os.OpenFile(blockFileName, os.O_RDONLY, os.ModePerm, ); err != nil {
			log.Logger.Errorf("open block file[%s] error: %v", blockFileName, err)
			return errors.New(fmt.Sprintf("open block file[%s] error", blockFileName))
		}
		if blockContent, err = ioutil.ReadAll(blockFile); err != nil {
			log.Logger.Errorf("read block file[%s] to final file error: %v", blockFileName, err)
			return errors.New(fmt.Sprintf("read block file[%s] error", blockFileName))
		}
		blockFile.Close()
		if _, err = finalFile.Write(blockContent); err != nil {
			log.Logger.Errorf("write block file[%s] to final file error: %v", blockFileName, err)
			return errors.New(fmt.Sprintf("wtite block file[%s] to final file error", blockFileName))
		}
		if err = os.Remove(blockFileName); err != nil {
			log.Logger.Errorf("remove block file[%s] error: %v", blockFileName, err)
			return errors.New(fmt.Sprintf("remove block file[%s] error", blockFileName))
		}
	}

	return nil
}

func createDir(dir string) (err error) {
	var (
		m = "apps.webUpload.createDir()"
	)
	if err = removeDir(dir); err != nil {
		return
	}
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
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
		removeDirErr = os.RemoveAll(dir)
	}

	if removeDirErr != nil {
		log.Logger.Errorf("%s error, remove old file[%s] in %s error: %v", m, dir, UploadTempDir, removeDirErr)
		return removeDirErr
	}
	return nil
}

func storageFileTo(fileName string) (err error) {
	// TODO storage file to distributed storage system
	fmt.Printf("storage file[%s] to ... ....\n", fileName)

	// TODO remove temp dir
	return
}

