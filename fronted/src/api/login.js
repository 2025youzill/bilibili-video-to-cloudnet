import axiosInstance from "../axiosInstance";

// 新增：发送验证码接口
export const sendCaptcha = async (phone) => {
	try {
		const response = await axiosInstance.post("/netcloud/login", {
			phone: phone,
		});
		return response.data;
	} catch (error) {
		// console.error("发送验证码失败", error);
		throw error;
	}
};

// 验证验证码（原路径"/login/verify"改为"/netcloud/login/verify"）
export const submitLogin = async ({ phone, captcha }) => {
	try {
		const response = await axiosInstance.post("/netcloud/login/verify", {
			phone: phone,
			captcha: captcha,
		});
		return response.data;
	} catch (error) {
		// console.error("验证验证码失败", error);
		throw error;
	}
};

// 获取网易云登录二维码（返回 image/png）
export const getLoginQrcode = async () => {
	try {
		const response = await axiosInstance.get("/netcloud/login", {
			responseType: "blob",
		});
		return response.data;
	} catch (error) {
		throw error;
	}
};

// 检查二维码登录状态（后端为长轮询：可能阻塞一段时间）
export const checkLoginQrcode = async () => {
	try {
		const response = await axiosInstance.get("/netcloud/login/verify", {
			timeout: 200000,
		});
		return response.data;
	} catch (error) {
		throw error;
	}
};

// 检查登录状态（新增）
export const checkLoginStatus = async () => {
	try {
		const response = await axiosInstance.get("/netcloud/login/check");
		return response.data;
	} catch (error) {
		// console.error("检查登录状态失败", error);
		throw error;
	}
};
