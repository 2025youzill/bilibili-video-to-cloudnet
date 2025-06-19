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
			onProxyReq: function (proxyReq, req, res) {
				console.log("代理请求:", req.method, req.url, "->", proxyReq.path);
			},
			onError: function (err, req, res) {
				console.error("代理错误:", err);
			},
		})
	);
};
