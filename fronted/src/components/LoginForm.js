import { useNavigate } from "react-router-dom";
import { useState } from "react";
import axiosInstance from "../axiosInstance";
import { sendCaptcha } from "../api/login"; // 新增：引入验证码发送接口

const LoginForm = () => {
  const navigate = useNavigate();
  const [phone, setPhone] = useState('');
  const [captcha, setCaptcha] = useState('');
  const [countdown, setCountdown] = useState(0); // 新增：倒计时状态

  // 新增：发送验证码逻辑
  const handleSendCaptcha = async () => {
    if (!phone.trim()) {
      alert("请先输入手机号");
      return;
    }
    try {
      await sendCaptcha(phone);
      alert("验证码已发送，请注意查收");
      // 启动60秒倒计时
      let timer = 60;
      setCountdown(timer);
      const interval = setInterval(() => {
        timer--;
        setCountdown(timer);
        if (timer <= 0) clearInterval(interval);
      }, 1000);
    } catch (error) {
      alert("验证码发送失败");
    }
  };

  const handleLogin = async () => {
    if (!phone || !captcha) {
      alert("请填写手机号和验证码");
      return;
    }
    try {
      // 原路径"/login/verify"改为"/netcloud/login/verify"
      const response = await axiosInstance.post("/netcloud/login/verify", {
        phone: phone,
        captcha: captcha
      });
      if (response.data.code === 200) {
        navigate("/upload");
      } else {
        alert(response.data.msg || "登录失败");
      }
    } catch (error) {
      alert("登录请求失败");
    }
  };

  return (
    <div className="login-form">
      <form onSubmit={(e) => {
        e.preventDefault();
        handleLogin();
      }}>
        <input 
          type="text" 
          placeholder="请输入手机号" 
          value={phone} 
          onChange={(e) => setPhone(e.target.value)} 
        />
        <div style={{ display: "flex", gap: "8px" }}>
          <input 
            type="text" 
            placeholder="请输入验证码" 
            value={captcha} 
            onChange={(e) => setCaptcha(e.target.value)} 
          />
          <button 
            type="button" 
            onClick={handleSendCaptcha}
            disabled={countdown > 0} // 倒计时期间禁用
          >
            {countdown > 0 ? `${countdown}秒后重试` : "发送"}
          </button>
        </div>
        <button type="submit" className="btn btn-primary w-full mt-4">
          登录
        </button>
      </form>
    </div>
  );
};

export default LoginForm;