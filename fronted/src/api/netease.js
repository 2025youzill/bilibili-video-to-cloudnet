import axiosInstance from "../axiosInstance";

// 获取歌单列表
export const getPlaylists = async () => {
	try {
		const response = await axiosInstance.get("/netcloud/playlist");
		return response.data;
	} catch (error) {
		// console.error("获取歌单列表失败:", error);
		throw error;
	}
};
