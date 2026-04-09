import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button, Divider, message, Typography } from "antd";
import { checkLoginQrcode, getLoginQrcode } from "../api/login";

const { Text } = Typography;

const QrLoginForm = () => {
	const navigate = useNavigate();
	const [qrLoading, setQrLoading] = useState(false);
	const [qrUrl, setQrUrl] = useState("");
	const objectUrlRef = useRef("");
	const lastQrUrlRef = useRef("");
	const requestInFlightRef = useRef(false);
	const fetchInProgressRef = useRef(false);

	useEffect(() => {
		return () => {
			if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current);
		};
	}, []);

	const fetchQrcode = useCallback(async () => {
		if (fetchInProgressRef.current) return;
		fetchInProgressRef.current = true;
		setQrLoading(true);
		try {
			const blob = await getLoginQrcode();
			if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current);
			const url = URL.createObjectURL(blob);
			objectUrlRef.current = url;
			setQrUrl(url);
			requestInFlightRef.current = false;
		} catch (e) {
			message.error("获取二维码失败");
		} finally {
			setQrLoading(false);
			fetchInProgressRef.current = false;
		}
	}, []);

	const pollLogin = useCallback(async () => {
		if (requestInFlightRef.current) return;
		requestInFlightRef.current = true;
		try {
			const res = await checkLoginQrcode();
			if (res?.code === 200) {
				message.success("登录成功，正在跳转...");
				navigate("/bilibili");
				return;
			}
			// 如果不是 200，说明需要重新扫码或过期，由用户点击刷新或逻辑处理
			if (res?.code !== 801 && res?.code !== 802) {
				message.error(res?.msg || "登录已失效，请刷新二维码");
			}
		} catch (e) {
			if (e?.response?.status !== 429) {
				// 避免在轮询超时或正常等待时弹出错误
				console.error("登录轮询异常", e);
			}
		} finally {
			requestInFlightRef.current = false;
		}
	}, [navigate]);

	useEffect(() => {
		// 只有在组件挂载时执行一次
		fetchQrcode();
	}, []); // 移除 fetchQrcode 依赖，确保只执行一次

	useEffect(() => {
		if (!qrUrl) return;
		if (lastQrUrlRef.current === qrUrl) return;
		lastQrUrlRef.current = qrUrl;
		pollLogin();
	}, [qrUrl, pollLogin]);

	const handleRefresh = useCallback(async () => {
		if (qrLoading || fetchInProgressRef.current) return;
		lastQrUrlRef.current = "";
		requestInFlightRef.current = false;
		setQrUrl("");
		await fetchQrcode();
	}, [qrLoading, fetchQrcode]);

	return (
		<div style={{ width: "100%", maxWidth: 420 }}>
			<div style={{ textAlign: "center", marginBottom: 20 }}>
				<Text strong style={{ fontSize: 24, color: "#1D2129" }}>
					登录
				</Text>
			</div>
			<Divider style={{ margin: "10px 0 24px", borderColor: "rgba(0,0,0,0.1)" }} />

			<div style={{ textAlign: "center" }}>
				<div
					style={{
						width: 240,
						height: 240,
						margin: "0 auto 12px",
						background: "rgba(0,0,0,0.04)",
						borderRadius: 12,
						display: "flex",
						alignItems: "center",
						justifyContent: "center",
						overflow: "hidden",
					}}
				>
					{qrUrl ? (
						<img alt="login-qrcode" src={qrUrl} style={{ width: "100%", height: "100%" }} />
					) : (
						<Text type="secondary">{qrLoading ? "二维码加载中..." : "暂无二维码"}</Text>
					)}
				</div>
				<Text type="secondary" style={{ whiteSpace: "nowrap" }}>
					打开网易云音乐 App 扫码登录
				</Text>
			</div>

			<div style={{ marginTop: 16 }}>
				<Button block size="large" onClick={handleRefresh} loading={qrLoading}>
					刷新二维码
				</Button>
			</div>
		</div>
	);
};

export default QrLoginForm;
