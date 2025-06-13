import axiosInstance from "../axiosInstance";

// 根据BVID获取视频→转音频→上传网易云（后端需提供此接口）
export const uploadToNetCloudByBVID = async (bvidList, splaylist = false, pid = "") => {
	const response = await axiosInstance.post("/bilibili/load", {
		bvid: bvidList, // 直接传递bvid数组
		splaylist,
		pid: pid ? parseInt(pid) : undefined,
	});
	return response.data;
};
