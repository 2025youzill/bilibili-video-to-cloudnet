import axios from 'axios';

const baseURL = 'http://localhost:8080';

// 新增：发送验证码接口
export const sendCaptcha = async (phone) => {
  try {
    const response = await axios.get(`${baseURL}/netcloud/login?phone=${phone}`);
    return response.data;
  } catch (error) {
    console.error('发送验证码失败', error);
    throw error;
  }
};

// 验证验证码（原路径"/login/verify"改为"/netcloud/login/verify"）
export const submitLogin = async ({ phone, captcha }) => {
  try {
    const response = await axios.post(`${baseURL}/netcloud/login/verify`, {
      phone: phone,
      captcha: captcha
    });
    return response.data;
  } catch (error) {
    console.error('验证验证码失败', error);
    throw error;
  }
};

// 检查登录状态（新增）
export const checkLoginStatus = async () => {
  try {
    const response = await axios.get(`${baseURL}/netcloud/login/check`);
    return response.data;
  } catch (error) {
    console.error('检查登录状态失败', error);
    throw error;
  }
};