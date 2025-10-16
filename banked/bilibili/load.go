package bilibili

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bvtc/client"
	"bvtc/config"
	"bvtc/constant"
	"bvtc/log"
	"bvtc/response"
	redis_pool "bvtc/tool/pool"
	"bvtc/tool/randomstring"

	"github.com/CuteReimu/bilibili/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/semaphore"
)

type VideoStreamReq struct {
	Bvid          []string          `json:"bvid"`                    // 稿件 bvid
	Splaylist     bool              `json:"splaylist"`               // 是否上传到歌单
	Pid           int64             `json:"pid,omitempty"`           // 歌单 id
	TitleOverride map[string]string `json:"titleOverride,omitempty"` // 可选：自定义标题，key 为 bvid
}

// 任务结构体
type LoadMP4Task struct {
	ID        string         `json:"id"`         // 任务ID
	Status    string         `json:"status"`     // 任务状态
	Progress  int            `json:"progress"`   // 进度百分比 (0-100)
	Total     int            `json:"total"`      // 总文件数
	Success   []string       `json:"success"`    // 成功处理的视频标题
	Failed    []failed       `json:"failed"`     // 失败处理的视频
	Error     string         `json:"error"`      // Status为failed时，错误信息
	CreatedAt time.Time      `json:"created_at"` // 创建时间
	UpdatedAt time.Time      `json:"updated_at"` // 更新时间
	Request   VideoStreamReq `json:"request"`    // 原始请求
}

// 失败处理的视频
type failed struct {
	Title string `json:"title,omitempty"` // 视频标题
	Error string `json:"error,omitempty"` // 错误信息
}

// 任务管理器
type TaskManager struct {
	tasks map[string]*LoadMP4Task
	mutex sync.RWMutex // 互斥锁
}

var taskManager = &TaskManager{
	tasks: make(map[string]*LoadMP4Task),
}

type result struct {
	Title string
	Err   error
}

// CreateLoadMP4Task 创建上传任务
func CreateLoadMP4Task(ctx *gin.Context) {
	var req VideoStreamReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Logger.Error("bind json fail", log.Any("err", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("invalid request format"))
		return
	}

	sid, err := ctx.Cookie("SessionId")
	if err != nil {
		log.Logger.Error("fail to get sessionId", log.Any("err : ", err))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("fail to get sessionId"))
		return
	}
	rdb := redis_pool.GetRdb()
	rtcx := redis_pool.GetRctx()
	key := "session:" + sid
	cookieFile, rerr := rdb.HGet(rtcx, key, "cookieFile").Result()
	if rerr != nil || cookieFile == "" {
		log.Logger.Error("session not found or expired", log.Any("err : ", rerr))
		ctx.JSON(http.StatusBadRequest, response.FailMsg("session not found or expired"))
		return
	}

	if len(req.Bvid) == 0 || req.Bvid[0] == "" {
		log.Logger.Error("bvid is empty")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("bvid is empty"))
		return
	}

	if req.Splaylist && req.Pid == 0 {
		log.Logger.Error("pid is required when splaylist is true")
		ctx.JSON(http.StatusBadRequest, response.FailMsg("pid is required when splaylist is true"))
		return
	}

	// 创建任务
	task := taskManager.createTask(req)

	// 启动异步处理
	go LoadMP4Async(task.ID, cookieFile)

	// 返回任务ID
	ctx.JSON(http.StatusOK, response.SuccessMsg(map[string]string{"task_id": task.ID}))
}

// CheckLoadMP4Task 查询任务状态
func CheckLoadMP4Task(ctx *gin.Context) {
	taskID := ctx.Param("taskId")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, response.FailMsg("task_id is required"))
		return
	}

	task, exists := taskManager.getTask(taskID)
	if !exists {
		ctx.JSON(http.StatusNotFound, response.FailMsg("task not found"))
		return
	}

	if task.Status == constant.TaskStatusPending {
		ctx.JSON(http.StatusNotAcceptable, response.FailMsg("task is not running"))
		return
	}

	if task.Status == constant.TaskStatusCompleted {
		taskManager.cleanTask(taskID)
		ctx.JSON(http.StatusOK, response.SuccessMsg(task))
		return
	}

	if task.Status == constant.TaskStatusRunning {
		ctx.JSON(http.StatusAccepted, response.SuccessMsg(task))
		return
	}

	if task.Status == constant.TaskStatusFailed {
		taskManager.cleanTask(taskID)
		ctx.JSON(http.StatusInternalServerError, response.FailMsg(task.Error))
		return
	}

	if task.Status == constant.TaskStatusOuttime {
		taskManager.cleanTask(taskID)
		ctx.JSON(http.StatusInternalServerError, response.FailMsg("task is outtime"))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessMsg(task))
}

// processLoadMP4Task 异步处理任务
func LoadMP4Async(taskID string, cookiefile string) {
	task, _ := taskManager.getTask(taskID)
	// 更新状态为运行中
	taskManager.updateTask(taskID, constant.TaskStatusRunning, 0, "")

	cli, err := client.GetBiliClient()
	if err != nil {
		log.Logger.Error("client init fail", log.Any("err", err))
		taskManager.updateTask(taskID, constant.TaskStatusFailed, 0, err.Error())
		return
	}

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(config.GetConfig().Music.Concurrency)
	resultChan := make(chan result, len(task.Request.Bvid))

	// 启动所有处理协程
	for i, bvid := range task.Request.Bvid {
		wg.Add(1)
		if err := sem.Acquire(context.Background(), 1); err != nil {
			log.Logger.Error("获取信号量失败", log.Any("err", err))
			continue
		}

		go func(index int, bvid string) {
			defer wg.Done()
			defer sem.Release(1)

			videoinfo, err := cli.GetVideoInfo(bilibili.VideoParam{Bvid: bvid})
			if err != nil {
				// cannot reference videoinfo when err != nil; use bvid as title fallback
				resultChan <- result{Title: bvid, Err: fmt.Errorf("get video info fail: %v", err)}
				return
			}
			cid := videoinfo.Cid

			stream, err := cli.GetVideoStream(bilibili.GetVideoStreamParam{Bvid: bvid, Cid: cid})
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("get video stream fail: %v", err)}
				return
			}

			// 应用可选的标题覆盖
			title := videoinfo.Title
			if task != nil && task.Request.TitleOverride != nil {
				if t, ok := task.Request.TitleOverride[bvid]; ok {
					t = strings.TrimSpace(t)
					if t != "" {
						title = t
					}
				}
			}
			title = sanitizeFilename(title)
			url := stream.Durl[0].Url
			filename := filepath.Join(constant.Filepath, fmt.Sprintf("%s.mp4", title))
			defer os.Remove(filename)

			err = os.MkdirAll(constant.Filepath, 0o755)
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("创建输出目录失败: %v", err)}
				return
			}

			referer := cli.Resty().Header.Get("Referer")
			useragent := cli.Resty().Header.Get("User-Agent")
			resp, err := resty.New().R().
				SetHeader("User-Agent", useragent).
				SetHeader("Referer", referer).
				SetOutput(filename).
				Get(url)
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("下载失败: %v", err)}
				return
			}
			if resp.StatusCode() != 200 {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("请求失败: status code %d", resp.StatusCode())}
				return
			}

			var audioreq AudioReq
			audioreq.Filename = filename
			audioreq.Artist = videoinfo.Owner.Name
			audioreq.Title = title

			mid := videoinfo.Owner.Mid
			artistinfo, err := cli.GetUserCard(bilibili.GetUserCardParam{Mid: mid})
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("获取用户空间详情失败: %v", err)}
				return
			}
			coverurl := artistinfo.Card.Face
			coverfilename := filepath.Join(constant.Filepath, fmt.Sprintf("%s.jpeg", randomstring.GenerateRandomString(16)))
			defer os.Remove(coverfilename)
			coverresp, err := resty.New().R().
				SetOutput(coverfilename).
				Get(coverurl)
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("下载封面失败: %v", err)}
				return
			}
			if coverresp.StatusCode() != 200 {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("请求封面失败: status code %d", coverresp.StatusCode())}
				return
			}
			audioreq.CoverArt = coverfilename

			err = TranslateVideoToAudio(audioreq, task.Request.Splaylist, task.Request.Pid, cookiefile)
			if err != nil {
				resultChan <- result{Title: videoinfo.Title, Err: fmt.Errorf("上传失败: %v", err)}
				return
			}

			resultChan <- result{Title: title, Err: nil}
		}(i, bvid)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集处理结果
	for result := range resultChan {
		if result.Err != nil {
			taskManager.addFailed(taskID, failed{
				Title: result.Title,
				Error: result.Err.Error(),
			})
		} else {
			taskManager.addSuccess(taskID, result.Title)
		}
	}

	// 更新最终状态
	taskManager.updateTask(taskID, constant.TaskStatusCompleted, 100, "")
}

// Task控制函数

// 创建新任务
func (tm *TaskManager) createTask(req VideoStreamReq) *LoadMP4Task {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	task := &LoadMP4Task{
		ID:        randomstring.GenerateRandomString(16),
		Status:    constant.TaskStatusPending,
		Progress:  0,
		Total:     len(req.Bvid),
		Success:   make([]string, 0),
		Failed:    make([]failed, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Request:   req,
	}

	tm.tasks[task.ID] = task
	return task
}

// 获取任务
func (tm *TaskManager) getTask(taskID string) (*LoadMP4Task, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	task, exists := tm.tasks[taskID]
	return task, exists
}

// 更新任务状态
func (tm *TaskManager) updateTask(taskID string, status string, progress int, error string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Status = status
		task.Progress = progress
		task.UpdatedAt = time.Now()
		if error != "" {
			task.Error = error
		}
	}
}

// 添加成功结果
func (tm *TaskManager) addSuccess(taskID string, title string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Success = append(task.Success, title)
		task.Progress = (len(task.Success) + len(task.Failed)) * 100 / task.Total
		task.UpdatedAt = time.Now()
	}
}

// 添加失败结果
func (tm *TaskManager) addFailed(taskID string, failedItem failed) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Failed = append(task.Failed, failedItem)
		task.Progress = (len(task.Success) + len(task.Failed)) * 100 / task.Total
		task.UpdatedAt = time.Now()
	}
}

func (tm *TaskManager) cleanTask(taskID string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	delete(tm.tasks, taskID)
}

func sanitizeFilename(filename string) string {
	// 替换所有可能造成问题的字符
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(filename)
}
