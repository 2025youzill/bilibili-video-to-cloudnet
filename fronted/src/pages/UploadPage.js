import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axiosInstance from '../axiosInstance';
import { uploadToNetCloudByBVID } from '../api/bilibili'; // 新增API方法
import LoadingSpinner from '../components/LoadingSpinner';

const UploadPage = () => {
  const navigate = useNavigate();
  const [bvid, setBvid] = useState(''); // BVID输入状态
  const [isProcessing, setIsProcessing] = useState(false); // 处理中状态
  const [error, setError] = useState(null);

  // 提交BVID处理
  const handleSubmit = async () => {
    if (!bvid.trim()) {
      setError('请输入BVID');
      return;
    }
    setIsProcessing(true);
    setError(null);

    try {
      // 调用后端接口：根据BVID获取视频→转音频→上传网易云
      const res = await uploadToNetCloudByBVID(bvid);
      if (res.code === 200) {
        alert('视频转音频并上传云盘成功！');
        setBvid(''); // 清空输入框
      } else {
        setError(res.msg || '处理失败，请重试');
      }
    } catch (error) {
      setError('网络请求失败，请检查BVID格式');
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <div className="bvid-container">
      <h1>哔哩哔哩视频转音频上传</h1>
      <div className="input-group">
        <input
          type="text"
          placeholder="输入BVID（如 BV1xx411c7m）"
          value={bvid}
          onChange={(e) => setBvid(e.target.value)}
          disabled={isProcessing}
        />
        <button onClick={handleSubmit} disabled={isProcessing}>
          {isProcessing ? <LoadingSpinner /> : '开始处理'}
        </button>
      </div>
      {error && <div className="error-message">{error}</div>}
    </div>
  );
};

export default UploadPage;