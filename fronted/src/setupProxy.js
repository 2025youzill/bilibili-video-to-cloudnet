const { createProxyMiddleware } = require("http-proxy-middleware");

module.exports = function (app) {
	app.use(
		"/api",
		createProxyMiddleware({
			target: "http://localhost:8080",
			changeOrigin: true,
			secure: false,
			logLevel: "debug",
			pathRewrite: {
				"^/api": "/api", // 保持 /api 前缀
			},
			// 关键：避免代理侧超时断开 SSE
			timeout: 0,
			proxyTimeout: 0,
			onProxyRes: function (proxyRes, req, res) {
				// 明确保持长连接，防止中间层缓冲
				proxyRes.headers["Cache-Control"] = "no-cache";
				proxyRes.headers["Connection"] = "keep-alive";
				proxyRes.headers["X-Accel-Buffering"] = "no";
			},
			onProxyReq: function (proxyReq, req, res) {
				// 可选：标记期望 SSE，部分代理有帮助
				proxyReq.setHeader("Accept", "text/event-stream");
			},
			onError: function (err, req, res) {
				// console.error("代理错误:", err);
			},
		})
	);
};
