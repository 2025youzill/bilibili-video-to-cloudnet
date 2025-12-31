import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import LoginForm from "../components/LoginForm";
import { checkLoginStatus } from "../api/login";
import { Button, GithubOutlined, EyeOutlined, EyeInvisibleOutlined } from "antd";
import { GithubOutlined as GithubIcon, EyeOutlined as EyeIcon, EyeInvisibleOutlined as EyeInvisibleIcon } from "@ant-design/icons";

const LoginPage = () => {
	const navigate = useNavigate();
	const [backgroundVisible, setBackgroundVisible] = useState(true);
	const [topbarVisible, setTopbarVisible] = useState(false);

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
		<div className="login-page-with-bg">
			{/* 隐形触发区域 - 鼠标移到顶部时显示框 */}
			<div
				style={{
					position: "fixed",
					top: 0,
					left: 0,
					right: 0,
					height: "30px",
					zIndex: 999,
				}}
				onMouseEnter={() => setTopbarVisible(true)}
			/>

			{/* 顶部导航栏 - 自动隐藏 */}
			<div
				style={{
					position: "fixed",
					top: 0,
					left: 0,
					right: 0,
					height: "50px",
					background: "rgba(255, 255, 255, 0.1)",
					backdropFilter: "blur(10px)",
					WebkitBackdropFilter: "blur(10px)",
					borderBottom: "1px solid rgba(255, 255, 255, 0.2)",
					display: "flex",
					alignItems: "center",
					justifyContent: "flex-end",
					padding: "0 30px",
					gap: "20px",
					zIndex: 1000,
					transform: topbarVisible ? "translateY(0)" : "translateY(-100%)",
					transition: "transform 0.3s ease-in-out",
				}}
				onMouseEnter={() => setTopbarVisible(true)}
				onMouseLeave={() => setTopbarVisible(false)}
			>
				{/* 背景切换按钮 */}
				<Button
					type="text"
					icon={backgroundVisible ? <EyeIcon /> : <EyeInvisibleIcon />}
					onClick={() => setBackgroundVisible(!backgroundVisible)}
					style={{
						color: "#333",
						fontWeight: "600",
						display: "flex",
						alignItems: "center",
						gap: "6px",
					}}
				>
					{backgroundVisible ? "隐藏背景" : "显示背景"}
				</Button>

				{/* GitHub 图标 */}
				<a
					href="https://github.com/2025youzill/bilibili-video-to-cloudnet"
					target="_blank"
					rel="noopener noreferrer"
					style={{
						display: "flex",
						alignItems: "center",
						color: "#333",
						fontSize: "24px",
						transition: "all 0.3s ease",
					}}
					onMouseEnter={(e) => {
						e.currentTarget.style.color = "#000";
						e.currentTarget.style.transform = "scale(1.1)";
					}}
					onMouseLeave={(e) => {
						e.currentTarget.style.color = "#333";
						e.currentTarget.style.transform = "scale(1)";
					}}
				>
					<GithubIcon />
				</a>
			</div>

		{/* 登录表单 */}
		<div className="login-form-glass">
			<LoginForm />
		</div>
		</div>
	);
};

export default LoginPage;
