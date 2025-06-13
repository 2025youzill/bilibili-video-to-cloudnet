import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import LoginForm from "../components/LoginForm";
import axiosInstance from "../axiosInstance";
import { checkLoginStatus } from "../api/login";

const LoginPage = () => {
	const navigate = useNavigate();
	const [error, setError] = useState(null);

	// 合并为单个 useEffect 检查登录状态
	useEffect(() => {
		const checkAuth = async () => {
			try {
				// 统一使用 checkLoginStatus 接口（已封装 axios 请求）
				const res = await checkLoginStatus();
				if (res.code === 200) {
					// 后端 CheckCookie 接口返回 code=200 表示已登录（根据后端代码 netclogin.CheckCookie 实现）
					navigate("/bilibili");
				}
			} catch (error) {
				// 接口调用失败或未登录，保持当前页面
				console.log("无有效登录状态，请登录");
			}
		};

		checkAuth();
	}, [navigate]);

	const handleLogin = async (phone, captcha) => {
		try {
			const verifyResponse = await axiosInstance.post("/login/verify", {
				phone: phone,
				captcha: captcha,
			});

			if (verifyResponse.data.code === 200) {
				navigate("/bilibili");
			} else {
				setError("验证码错误，请重试");
			}
		} catch (error) {
			setError("登录失败，请重试");
		}
	};

	return (
		<div className="login-container">
			<h1>网易云音乐登录</h1>
			<LoginForm onLogin={handleLogin} />
			{error && <div className="error-message">{error}</div>}
		</div>
	);
};

export default LoginPage;
