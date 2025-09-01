import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import LoginForm from "../components/LoginForm";
import { checkLoginStatus } from "../api/login";

const LoginPage = () => {
	const navigate = useNavigate();

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
				// console.log("无有效登录状态，请登录");
			}
		};

		checkAuth();
	}, [navigate]);

	return (
		<div className="login-container">
			<LoginForm />
		</div>
	);
};

export default LoginPage;
