import axios from "axios";

// 配置 axios 实例（与后端接口地址匹配）
const axiosInstance = axios.create({
	baseURL: "http://localhost:8080", // 后端服务地址（根据实际调整）
	timeout: 300000, // 请求超时时间设置为5分钟
});

// 可选：添加响应拦截器处理错误
axiosInstance.interceptors.response.use(
	(response) => response,
	(error) => {
		console.error("请求失败:", error);
		return Promise.reject(error);
	}
);
axiosInstance.defaults.withCredentials = true;

export default axiosInstance;
