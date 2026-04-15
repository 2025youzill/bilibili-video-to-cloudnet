package socket

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"bvtc/log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func resetTestManager() {
	defaultManager = newManager()
	startOnce = sync.Once{}
}

func TestUpgrade_EstablishesSocketConnection(t *testing.T) {
	// 1. 重置全局状态
	resetTestManager()
	t.Cleanup(resetTestManager) // 测试结束后自动重置

	// 禁用日志
	log.Logger = zap.NewNop()
	gin.SetMode(gin.TestMode)

	// 2. 搭建测试路由
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		client, err := Upgrade(c, "test-session")
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}

		go func() {
			_ = client.SendJSON(gin.H{
				"code": 200,
				"msg":  "connected",
			})
		}()
	})

	// 3. 启动测试服务器
	server := httptest.NewServer(router)
	defer server.Close()

	// 4. 客户端拨号 WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket failed: %v", err)
	}
	defer resp.Body.Close()
	defer conn.Close()

	// 5. 设置读超时
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	// 6. 读取服务端消息
	var payload map[string]any
	if err := conn.ReadJSON(&payload); err != nil {
		t.Fatalf("read websocket message failed: %v", err)
	}

	// 7. 断言消息内容
	wantMsg := "connected"
	if got := payload["msg"]; got != wantMsg {
		raw, _ := json.Marshal(payload)
		t.Fatalf("msg mismatch:\nwant: %q\ngot : %q\nfull: %s", wantMsg, got, raw)
	}

	// 8. 优雅关闭管理器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	// 9. 验证连接已关闭
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected connection closed after shutdown, but no error")
	}
}