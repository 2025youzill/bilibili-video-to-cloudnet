import { useNavigate } from "react-router-dom";
import { useState } from "react";
import axiosInstance from "../axiosInstance";
import { sendCaptcha } from "../api/login";
import { Input, Button, Card, message } from "antd";
import { UserOutlined, LockOutlined } from "@ant-design/icons";

const LoginForm = () => {
	const navigate = useNavigate();
	const [phone, setPhone] = useState("");
	const [captcha, setCaptcha] = useState("");
	const [countdown, setCountdown] = useState(0);
	const [loading, setLoading] = useState(false);

	const handleSendCaptcha = async () => {
		if (!phone.trim()) {
			message.warning("请先输入手机号");
			return;
		}
		try {
			await sendCaptcha(phone);
			message.success("验证码已发送，请注意查收");
			let timer = 60;
			setCountdown(timer);
			const interval = setInterval(() => {
				timer--;
				setCountdown(timer);
				if (timer <= 0) clearInterval(interval);
			}, 1000);
		} catch (error) {
			message.error("验证码发送失败");
		}
	};

	const handleLogin = async () => {
		if (!phone || !captcha) {
			message.warning("请填写手机号和验证码");
			return;
		}
		setLoading(true);
		try {
			const response = await axiosInstance.post(
				"/netcloud/login/verify",
				{
					phone: phone,
					captcha: captcha,
				},
				{
					headers: {
						"Content-Type": "application/json",
					},
				}
			);
			if (response.data.code === 200) {
				message.success("登录成功");
				navigate("/bilibili");
			} else {
				message.error(response.data.msg || "登录失败");
			}
		} catch (error) {
			message.error("登录请求失败");
		} finally {
			setLoading(false);
		}
	};

	return (
		<div
			style={{
				display: "flex",
				justifyContent: "center",
				alignItems: "center",
				minHeight: "100vh",
				background: "#f0f2f5",
			}}
		>
			<Card
				title="网易云音乐登录"
				style={{
					width: 400,
					boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
				}}
				headStyle={{
					textAlign: "center",
					fontSize: "20px",
					fontWeight: "bold",
				}}
			>
				<form
					onSubmit={(e) => {
						e.preventDefault();
						handleLogin();
					}}
					style={{ display: "flex", flexDirection: "column", gap: "16px" }}
				>
					<Input
						size="large"
						prefix={<UserOutlined />}
						placeholder="请输入手机号"
						value={phone}
						onChange={(e) => setPhone(e.target.value)}
					/>
					<div style={{ display: "flex", gap: "8px" }}>
						<Input
							size="large"
							prefix={<LockOutlined />}
							placeholder="请输入验证码"
							value={captcha}
							onChange={(e) => setCaptcha(e.target.value)}
						/>
						<Button
							type="primary"
							size="large"
							onClick={handleSendCaptcha}
							disabled={countdown > 0}
							style={{ width: "120px" }}
						>
							{countdown > 0 ? `${countdown}秒后重试` : "发送验证码"}
						</Button>
					</div>
					<Button type="primary" size="large" htmlType="submit" loading={loading} style={{ marginTop: "8px" }}>
						登录
					</Button>
				</form>
			</Card>
		</div>
	);
};

export default LoginForm;
