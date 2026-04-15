import axiosInstance from "../axiosInstance";

export const sendCaptcha = async (phone) => {
	try {
		const response = await axiosInstance.post("/netcloud/login", {
			phone,
		});
		return response.data;
	} catch (error) {
		throw error;
	}
};

export const submitLogin = async ({ phone, captcha }) => {
	try {
		const response = await axiosInstance.post("/netcloud/login/verify", {
			phone,
			captcha,
		});
		return response.data;
	} catch (error) {
		throw error;
	}
};

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

export const createLoginQrcodeSocket = () => {
	const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
	if (process.env.NODE_ENV === "development") {
		return new WebSocket(`${protocol}//localhost:8081/bvtc/api/netcloud/login/verify`);
	}

	const basePath = axiosInstance.defaults.baseURL || "";
	return new WebSocket(`${protocol}//${window.location.host}${basePath}/netcloud/login/verify`);
};

export const checkLoginStatus = async () => {
	try {
		const response = await axiosInstance.get("/netcloud/login/check");
		return response.data;
	} catch (error) {
		throw error;
	}
};
