import axiosInstance from '../axiosInstance';

// 根据BVID获取视频→转音频→上传网易云（后端需提供此接口）
export const uploadToNetCloudByBVID = async (bvid) => {
  const response = await axiosInstance.post('/bilibili/video/load', {
    bvid // 直接传递JSON对象，键名与后端VideoStreamReq的json标签一致（"bvid"）
  });
  return response.data;
};