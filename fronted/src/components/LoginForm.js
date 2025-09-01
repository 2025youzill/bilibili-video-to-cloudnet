import { useNavigate } from "react-router-dom";
import { useEffect, useMemo, useRef, useState } from "react";
import { sendCaptcha, submitLogin } from "../api/login";
import { Input, Button, Card, message, Form, Typography, Row, Col, Divider } from "antd";
import { UserOutlined, LockOutlined } from "@ant-design/icons";

const { Text } = Typography;

const LoginForm = () => {
	const navigate = useNavigate();
	const [form] = Form.useForm();
	const [countdown, setCountdown] = useState(0);
	const [loading, setLoading] = useState(false);
	const [sending, setSending] = useState(false);
	const timerRef = useRef(null);

	useEffect(() => {
		return () => {
			if (timerRef.current) clearInterval(timerRef.current);
		};
	}, []);

	const phoneRules = useMemo(
		() => [
			{ required: true, message: "请输入手机号" },
			{
				pattern: /^1\d{10}$/,
				message: "请输入11位中国大陆手机号",
			},
		],
		[]
	);

	const captchaRules = useMemo(
		() => [
			{ required: true, message: "请输入验证码" },
			{ len: 4, message: "验证码为4位数字" },
		],
		[]
	);

	const handleSendCaptcha = async () => {
		try {
			const values = await form.validateFields(["phone"]);
			setSending(true);
			await sendCaptcha(values.phone);
			message.success("验证码已发送，请注意查收");
			let counter = 60;
			setCountdown(counter);
			timerRef.current = setInterval(() => {
				counter -= 1;
				setCountdown(counter);
				if (counter <= 0 && timerRef.current) {
					clearInterval(timerRef.current);
				}
			}, 1000);
		} catch (err) {
			// 校验失败或请求失败
			if (err?.errorFields) return;
			message.error("验证码发送失败");
		} finally {
			setSending(false);
		}
	};

	const onFinish = async (values) => {
		setLoading(true);
		try {
			const data = await submitLogin({ phone: values.phone, captcha: values.captcha });
			if (data.code === 200) {
				message.success("登录成功，正在跳转...");
				navigate("/bilibili");
			} else {
				message.error(data.msg || "登录失败");
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
				minHeight: "100vh",
				display: "flex",
				alignItems: "center",
				justifyContent: "center",
				background: "linear-gradient(135deg, #ffffff 0%, #f6f7ff 100%)",
				padding: 24,
			}}
		>
			<Card
				style={{
					width: 420,
					borderRadius: 14,
					boxShadow: "0 6px 24px rgba(0,0,0,0.06)",
					border: "1px solid rgba(0,0,0,0.06)",
				}}
				bodyStyle={{ padding: 26 }}
			>
				<div style={{ textAlign: "center", marginBottom: 6 }}>
					<Text strong style={{ fontSize: 18 }}>
						登录
					</Text>
				</div>
				<Divider style={{ margin: "10px 0 18px" }} />

				<Form
					form={form}
					layout="vertical"
					onFinish={onFinish}
					requiredMark={false}
					validateTrigger={["onBlur", "onSubmit"]}
				>
					<Form.Item label="手机号" name="phone" rules={phoneRules}>
						<Input size="large" addonBefore="+86" placeholder="请输入11位手机号" prefix={<UserOutlined />} />
					</Form.Item>

					<Form.Item label="验证码" name="captcha" rules={captchaRules}>
						<Row gutter={8}>
							<Col flex="auto">
								<Input size="large" placeholder="请输入4位验证码" prefix={<LockOutlined />} maxLength={4} />
							</Col>
							<Col>
								<Button
									size="large"
									type="primary"
									onClick={handleSendCaptcha}
									disabled={countdown > 0}
									loading={sending}
								>
									{countdown > 0 ? `${countdown}s` : "获取验证码"}
								</Button>
							</Col>
						</Row>
					</Form.Item>

					<Button block type="primary" size="large" htmlType="submit" loading={loading}>
						登录
					</Button>
				</Form>
			</Card>
		</div>
	);
};

export default LoginForm;
