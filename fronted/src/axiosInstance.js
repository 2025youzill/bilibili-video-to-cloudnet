import axios from "axios";

// 配置 axios 实例
const axiosInstance = axios.create({
	baseURL: "/api",
	timeout: 6000, // 请求超时时间设置为1分钟
	withCredentials: true, // 允许跨域携带cookie
	headers: {
		"Content-Type": "application/json",
	},
});

// 添加请求拦截器
axiosInstance.interceptors.request.use(
	(config) => {
		// 在发送请求之前做些什么
		return config;
	},
	(error) => {
		// 对请求错误做些什么
		console.error("请求错误:", error);
		return Promise.reject(error);
	}
);

// 添加响应拦截器处理错误
axiosInstance.interceptors.response.use(
	(response) => response,
	(error) => {
		console.error("请求失败:", error);
		if (error.code === "ERR_NETWORK") {
			console.error("网络连接失败，请检查网络连接或后端服务是否启动");
		} else if (error.code === "ECONNABORTED") {
			console.error("请求超时，请检查网络连接");
		}
		return Promise.reject(error);
	}
);

export default axiosInstance;
