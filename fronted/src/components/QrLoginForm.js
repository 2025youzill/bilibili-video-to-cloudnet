import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button, Divider, message, Typography } from "antd";
import { createLoginQrcodeSocket, getLoginQrcode } from "../api/login";

const { Text } = Typography;
const COOKIE_SETTLE_DELAY_MS = 200;
const SOCKET_RETRY_LIMIT = 3;
const SOCKET_RETRY_DELAY_MS = 600;

const QrLoginForm = ({ autoFetch = true }) => {
	const navigate = useNavigate();
	const [qrLoading, setQrLoading] = useState(false);
	const [qrUrl, setQrUrl] = useState("");
	const [qrStatus, setQrStatus] = useState("idle"); // idle, scanning, success, expired, timeout
	const [isMobileLike, setIsMobileLike] = useState(false);
	const objectUrlRef = useRef("");
	const fetchInProgressRef = useRef(false);
	const mountedRef = useRef(true);
	const socketRef = useRef(null);
	const navigateTimerRef = useRef(null);
	const fetchSeqRef = useRef(0);
	const suppressSocketErrorRef = useRef(false);
	const reconnectTimerRef = useRef(null);
	const reconnectAttemptsRef = useRef(0);
	const qrUrlRef = useRef("");
	const qrStatusRef = useRef("idle");

	useEffect(() => {
		qrUrlRef.current = qrUrl;
	}, [qrUrl]);

	useEffect(() => {
		qrStatusRef.current = qrStatus;
	}, [qrStatus]);

	const clearReconnectTimer = useCallback(() => {
		if (reconnectTimerRef.current) {
			window.clearTimeout(reconnectTimerRef.current);
			reconnectTimerRef.current = null;
		}
	}, []);

	const closeSocket = useCallback(() => {
		clearReconnectTimer();
		if (socketRef.current) {
			suppressSocketErrorRef.current = true;
			socketRef.current.close();
			socketRef.current = null;
		}
	}, [clearReconnectTimer]);

	const connectLoginSocket = useCallback(() => {
		closeSocket();

		const socket = createLoginQrcodeSocket();
		socketRef.current = socket;
		suppressSocketErrorRef.current = false;

		const scheduleReconnect = () => {
			if (!mountedRef.current || reconnectTimerRef.current || fetchInProgressRef.current) return;
			if (!qrUrlRef.current) return;
			if (["success", "expired", "timeout"].includes(qrStatusRef.current)) return;
			if (reconnectAttemptsRef.current >= SOCKET_RETRY_LIMIT) {
				message.error("登录连接异常，请刷新二维码重试");
				return;
			}

			reconnectAttemptsRef.current += 1;
			reconnectTimerRef.current = window.setTimeout(() => {
				reconnectTimerRef.current = null;
				if (!mountedRef.current || fetchInProgressRef.current) return;
				connectLoginSocket();
			}, SOCKET_RETRY_DELAY_MS);
		};

		socket.onopen = () => {
			reconnectAttemptsRef.current = 0;
			clearReconnectTimer();
		};

		socket.onmessage = (event) => {
			if (!mountedRef.current) return;

			try {
				const res = JSON.parse(event.data);

				switch (res?.code) {
					case 200:
						setQrStatus("success");
						closeSocket();
						if (navigateTimerRef.current) {
							window.clearTimeout(navigateTimerRef.current);
						}
						navigateTimerRef.current = window.setTimeout(() => {
							navigate("/bilibili");
						}, 1500);
						return;
					case 800:
						setQrStatus("expired");
						message.error("二维码已过期，请刷新");
						closeSocket();
						return;
					case 801:
						setQrStatus("idle");
						return;
					case 802:
						setQrStatus("scanning");
						return;
					case 408:
						setQrStatus("timeout");
						message.warning("登录超时，请重新获取二维码");
						closeSocket();
						return;
					default:
						if (res?.msg) {
							message.error(res.msg);
						}
						closeSocket();
					}
				} catch (error) {
					console.error("二维码登录消息解析失败", error);
				}
			};

		socket.onerror = () => {
			if (!mountedRef.current) return;
			if (suppressSocketErrorRef.current) return;
			scheduleReconnect();
		};

		socket.onclose = () => {
			const shouldSuppressError = suppressSocketErrorRef.current;
			if (socketRef.current === socket) {
				socketRef.current = null;
			}
			if (shouldSuppressError) {
				suppressSocketErrorRef.current = false;
				return;
			}
			scheduleReconnect();
		};
	}, [clearReconnectTimer, closeSocket, navigate]);

	useEffect(() => {
		const mediaQuery = window.matchMedia("(max-width: 768px), (pointer: coarse)");
		const updateMobileState = (event) => {
			setIsMobileLike(event.matches);
		};

		setIsMobileLike(mediaQuery.matches);
		if (typeof mediaQuery.addEventListener === "function") {
			mediaQuery.addEventListener("change", updateMobileState);
			return () => mediaQuery.removeEventListener("change", updateMobileState);
		}

		mediaQuery.addListener(updateMobileState);
		return () => mediaQuery.removeListener(updateMobileState);
	}, []);

	useEffect(() => {
		mountedRef.current = true;
		return () => {
			mountedRef.current = false;
			closeSocket();
			clearReconnectTimer();
			if (navigateTimerRef.current) {
				window.clearTimeout(navigateTimerRef.current);
			}
			if (objectUrlRef.current) {
				URL.revokeObjectURL(objectUrlRef.current);
			}
		};
	}, [closeSocket]);

	const fetchQrcode = useCallback(async () => {
		if (fetchInProgressRef.current) return;
		const fetchSeq = ++fetchSeqRef.current;
		fetchInProgressRef.current = true;
		reconnectAttemptsRef.current = 0;
		closeSocket();
		setQrLoading(true);
		setQrStatus("idle");

		try {
			const blob = await getLoginQrcode();
			if (!mountedRef.current) return;
			if (objectUrlRef.current) {
				URL.revokeObjectURL(objectUrlRef.current);
			}
			const url = URL.createObjectURL(blob);
			objectUrlRef.current = url;
			setQrUrl(url);
			await new Promise((resolve) => window.setTimeout(resolve, COOKIE_SETTLE_DELAY_MS));
			if (!mountedRef.current || fetchSeq !== fetchSeqRef.current) return;
			connectLoginSocket();
		} catch (error) {
			message.error("获取二维码失败");
		} finally {
			if (mountedRef.current) {
				setQrLoading(false);
			}
			fetchInProgressRef.current = false;
		}
	}, [closeSocket, connectLoginSocket]);

	useEffect(() => {
		if (!autoFetch) return;
		fetchQrcode();
	}, [autoFetch, fetchQrcode]);

	const handleRefresh = useCallback(async () => {
		if (qrLoading || fetchInProgressRef.current) return;
		setQrUrl("");
		await fetchQrcode();
	}, [fetchQrcode, qrLoading]);

	const handleSaveQrcode = useCallback(() => {
		if (!qrUrl) {
			message.warning("请先获取二维码");
			return;
		}

		const link = document.createElement("a");
		link.href = qrUrl;
		link.download = "netcloud-login-qrcode.png";
		link.rel = "noopener";
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
	}, [qrUrl]);

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
						position: "relative",
					}}
				>
					{qrUrl ? (
						<>
							<img
								alt="login-qrcode"
								src={qrUrl}
								style={{
									width: "100%",
									height: "100%",
									filter: qrStatus === "expired" || qrStatus === "timeout" ? "blur(4px) grayscale(100%)" : "none",
									opacity: qrStatus === "scanning" ? 0.3 : 1,
									transition: "all 0.3s ease",
								}}
							/>
							{qrStatus === "scanning" && (
								<div style={{ position: "absolute", textAlign: "center" }}>
									<Text strong style={{ color: "#1890ff" }}>
										已扫码，请在手机上确认
									</Text>
								</div>
							)}
							{qrStatus === "success" && (
								<div
									style={{
										position: "absolute",
										width: "100%",
										height: "100%",
										display: "flex",
										flexDirection: "column",
										alignItems: "center",
										justifyContent: "center",
										background: "rgba(255, 255, 255, 0.72)",
										textAlign: "center",
										padding: "0 24px",
									}}
								>
									<Text strong style={{ color: "#52c41a", fontSize: 18 }}>
										登录成功
									</Text>
									<Text type="secondary" style={{ marginTop: 8 }}>
										正在跳转...
									</Text>
								</div>
							)}
							{(qrStatus === "expired" || qrStatus === "timeout") && (
								<div
									style={{
										position: "absolute",
										width: "100%",
										height: "100%",
										display: "flex",
										flexDirection: "column",
										alignItems: "center",
										justifyContent: "center",
										background: "rgba(255, 255, 255, 0.6)",
										cursor: "pointer",
									}}
									onClick={handleRefresh}
								>
									<Text strong>{qrStatus === "expired" ? "二维码已失效" : "登录超时"}</Text>
									<Button type="link">点击刷新</Button>
								</div>
							)}
						</>
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
			{isMobileLike && qrUrl && (
				<div style={{ marginTop: 12 }}>
					<Button block size="large" onClick={handleSaveQrcode}>
						保存二维码到相册
					</Button>
				</div>
			)}
		</div>
	);
};

export default QrLoginForm;
