import React, { useState, useEffect } from "react";
import { Input, Button, List, message, Modal, Checkbox, Space, Pagination, App, Progress, Spin } from "antd";
import { UserOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import axiosInstance from "../axiosInstance";
import { checkLoginStatus } from "../api/login";

const BilibiliPage = () => {
	const [videoId, setVideoId] = useState("");
	const [videoInfo, setVideoInfo] = useState(null);
	const [playlists, setPlaylists] = useState([]);
	const [loading, setLoading] = useState(false);
	const [isModalVisible, setIsModalVisible] = useState(false);
	const [isLoginModalVisible, setIsLoginModalVisible] = useState(false);
	const [selectedVideos, setSelectedVideos] = useState([]);
	const [uploading, setUploading] = useState(false);
	const [currentPage, setCurrentPage] = useState(1);
	const [pageSize] = useState(10);
	const navigate = useNavigate();
	const { message: messageApi } = App.useApp();

	const [taskStatus, setTaskStatus] = useState(null);
	const [progress, setProgress] = useState(0);
	const [progressModalVisible, setProgressModalVisible] = useState(false);
	const [progressError, setProgressError] = useState("");
	const [progressTimer, setProgressTimer] = useState(null);
	const [uploadResult, setUploadResult] = useState(null);
	const [avatarUrl, setAvatarUrl] = useState("");
	const [isHovering, setIsHovering] = useState(false);
	const [confirmModalVisible, setConfirmModalVisible] = useState(false);
	const [selectedPlaylist, setSelectedPlaylist] = useState(null);
	const [searchKeyword, setSearchKeyword] = useState("");
	// 根据 bvid 获取原标题的工具函数
	const getOriginalTitleByBvid = (bvid) => {
		if (!videoInfo?.video_list) return "";
		const found = videoInfo.video_list.find((v) => v.bvid === bvid);
		return found?.title || "";
	};

	// 调用后端 AI 建议接口（SSE流式），批量预填建议标题
	const prepareTitleSuggestions = async () => {
		if (!selectedVideos || selectedVideos.length === 0) return;
		setTitleSuggesting(true);
		return new Promise((resolve, reject) => {
			let receivedAny = false;
			try {
				const base = axiosInstance.defaults.baseURL || "/api";
				const url = `${base}/bilibili/suggest-title-batch/stream?bvids=${encodeURIComponent(selectedVideos.join(","))}`;
				const es = new EventSource(url);

				es.addEventListener("open", () => {
					// 可在这里提示已开始
				});

				const handleProgress = (ev) => {
					try {
						const data = JSON.parse(ev.data || "{}");
						const { bvid, suggestedTitle, error } = data || {};
						if (!bvid && Array.isArray(data?.results)) {
							// 兼容聚合结构（防旧格式）
							const map = {};
							data.results.forEach((r) => {
								if (r.bvid && r.suggestedTitle) map[r.bvid] = r.suggestedTitle;
							});
							if (Object.keys(map).length) {
								setTitleOverride((prev) => ({ ...prev, ...map }));
							}
							return;
						}
						if (bvid && suggestedTitle) {
							setTitleOverride((prev) => ({ ...prev, [bvid]: suggestedTitle }));
						}
						if (error) {
							// 单项失败仅提示，不中断整体
							// message.warn?.(`AI 建议失败(${bvid}): ${error}`);
						}
						// 第一条进度就关闭转圈，边流边填
						if (!receivedAny) {
							receivedAny = true;
							// 保持轻遮罩直到 done，避免“立刻消失”的突兀
						}
					} catch {
						/* ignore */
					}
				};
				es.addEventListener("progress", handleProgress);
				// 某些代理会把未命名事件当默认 message 发
				es.onmessage = handleProgress;

				es.addEventListener("error", (ev) => {
					// 不主动关闭，交给 EventSource 自动重连，避免只收到首条后断流
					// 可以在这里添加轻量提示或日志
					// console.warn("SSE error, waiting for reconnect", ev);
				});

				es.addEventListener("done", () => {
					es.close();
					setTitleSuggesting(false);
					resolve();
				});
			} catch (e) {
				setTitleSuggesting(false);
				reject(e);
			}
		});
	};

	// 单个条目重新生成建议（SSE单条）
	const regenerateSuggestion = async (bvid) => {
		if (!bvid) return;
		setTitleSuggesting(true);
		return new Promise((resolve, reject) => {
			let receivedAny = false;
			try {
				const base = axiosInstance.defaults.baseURL || "/api";
				const url = `${base}/bilibili/suggest-title-batch/stream?bvids=${encodeURIComponent(bvid)}`;
				const es = new EventSource(url);

				const handleSingleProgress = (ev) => {
					try {
						const data = JSON.parse(ev.data || "{}");
						if (data?.bvid === bvid && data?.suggestedTitle) {
							setTitleOverride((prev) => ({ ...prev, [bvid]: data.suggestedTitle }));
						}
						if (!receivedAny) {
							receivedAny = true;
							// 保持轻遮罩直到 done
						}
					} catch {
						/* ignore */
					}
				};
				es.addEventListener("progress", handleSingleProgress);
				es.onmessage = handleSingleProgress;

				es.addEventListener("error", () => {
					// 不主动关闭，交给 EventSource 自动重连
					// console.warn("SSE error (single), waiting for reconnect");
				});

				es.addEventListener("done", () => {
					es.close();
					setTitleSuggesting(false);
					resolve();
				});
			} catch (e) {
				setTitleSuggesting(false);
				reject(e);
			}
		});
	};

	// 新增：标题选择与编辑相关状态
	const [keepTitleModalVisible, setKeepTitleModalVisible] = useState(false);
	const [titleEditModalVisible, setTitleEditModalVisible] = useState(false);
	const [useOriginalTitle, setUseOriginalTitle] = useState(true);
	const [titleOverride, setTitleOverride] = useState({}); // { bvid: title }
	const [titleSuggesting, setTitleSuggesting] = useState(false);

	const showError = (msg) => {
		messageApi.error(msg);
	};

	const handleSearch = async () => {
		if (!videoId) {
			showError("请输入视频ID");
			return;
		}
		setLoading(true);
		try {
			let url = "bilibili/list";
			if (videoId.toLowerCase().startsWith("av")) {
				const avid = parseInt(videoId.toLowerCase().replace("av", ""));
				url += `?avid=${avid}`;
			} else if (videoId.toLowerCase().startsWith("bv")) {
				url += `?bvid=${encodeURIComponent(videoId)}`;
			} else {
				showError("请输入正确的bvid号(´～`)");
				setLoading(false);
				return;
			}

			// console.log("请求URL:", url);
			const response = await axiosInstance.get(url);
			// console.log("响应数据:", response.data);
			setVideoInfo(response.data.data);
			// 重置选中的视频
			setSelectedVideos([]);
		} catch (error) {
			// console.error("请求错误:", error);
			// 打印完整的错误对象，帮助调试
			// console.log("错误详情:", {
			// 	response: error.response,
			// 	status: error.response?.status,
			// 	data: error.response?.data,
			// });

			// 检查错误响应
			if (error.response) {
				// console.log("错误状态码:", error.response.status);
				// console.log("错误信息:", error.response.data);
				if (error.response.status === 400) {
					showError("请输入正确的bvid号(´～`)");
				} else {
					showError(error.response.data.msg || "获取视频信息失败");
				}
			} else {
				showError("网络请求失败，请检查网络连接");
			}
		} finally {
			setLoading(false);
		}
	};

	const handleSave = async () => {
		if (selectedVideos.length === 0) {
			message.warning("请至少选择一个视频");
			return;
		}

		try {
			const cloudCheck = await axiosInstance.get("netcloud/login/check");
			if (!cloudCheck.data.data) {
				setIsLoginModalVisible(true);
				return;
			}

			const response = await axiosInstance.get("netcloud/playlist");
			if (response.data.code === 200) {
				setPlaylists(response.data.data);
				setIsModalVisible(true);
			} else {
				message.error("获取歌单列表失败");
			}
		} catch (error) {
			if (error.response && error.response.status === 400) {
				setIsLoginModalVisible(true);
			} else {
				message.error("获取歌单列表失败");
			}
		}
	};

	const handleVideoSelect = (bvid, checked) => {
		if (checked) {
			setSelectedVideos([...selectedVideos, bvid]);
		} else {
			setSelectedVideos(selectedVideos.filter((v) => v !== bvid));
		}
	};

	const handleSelectAll = (checked) => {
		if (checked) {
			setSelectedVideos(videoInfo.video_list.map((v) => v.bvid));
		} else {
			setSelectedVideos([]);
		}
	};

	const handlePageChange = (page) => {
		setCurrentPage(page);
	};

	// 计算当前页的视频列表（带搜索过滤）
	const getCurrentPageVideos = () => {
		if (!videoInfo?.video_list) return [];
		// 先过滤掉第一个视频（当前视频）
		let videos = videoInfo.video_list.slice(1);

		// 如果有搜索关键词，进行过滤
		if (searchKeyword.trim()) {
			const keyword = searchKeyword.trim().toLowerCase();
			videos = videos.filter(
				(video) => video.title.toLowerCase().includes(keyword) || video.bvid.toLowerCase().includes(keyword)
			);
		}

		// 分页
		const startIndex = (currentPage - 1) * pageSize;
		return videos.slice(startIndex, startIndex + pageSize);
	};

	// 获取过滤后的总数
	const getFilteredTotal = () => {
		if (!videoInfo?.video_list) return 0;
		let videos = videoInfo.video_list.slice(1);

		if (searchKeyword.trim()) {
			const keyword = searchKeyword.trim().toLowerCase();
			videos = videos.filter(
				(video) => video.title.toLowerCase().includes(keyword) || video.bvid.toLowerCase().includes(keyword)
			);
		}

		return videos.length;
	};

	const handleUpload = async (playlist) => {
		try {
			if (!selectedVideos || selectedVideos.length === 0) {
				message.error("请先选择要上传的视频");
				return;
			}
			setUploading(true);
			setUploadResult(null); // 清空之前的上传结果

			const requestData = {
				bvid: selectedVideos,
				splaylist: !!playlist.pid,
				pid: playlist.pid || undefined,
				// 仅当用户选择自定义时，才传递覆盖标题
				...(useOriginalTitle ? {} : { titleOverride: titleOverride }),
			};

			// 1. 发起任务创建请求，获取task_id
			const response = await axiosInstance.post("bilibili/createtask", requestData);
			if (response.data.code === 200 && response.data.data?.task_id) {
				const tid = response.data.data.task_id;
				setProgress(0);
				setTaskStatus("pending");
				setProgressError("");
				setProgressModalVisible(true);

				// 2. 启动轮询
				const timer = setInterval(() => {
					axiosInstance
						.get(`bilibili/checktask/${tid}`)
						.then((res) => {
							if (res.data.code === 200 && res.data.data) {
								const task = res.data.data;
								setProgress(task.progress || 0);
								setTaskStatus(task.status);
								if (task.status === "completed" || task.status === "failed") {
									clearInterval(timer);
									setProgressModalVisible(false);
									setUploading(false);

									// 保存结果
									setUploadResult({
										success: task.success || [],
										failed: task.failed || [],
									});

									if (task.status === "completed") {
										if (task.failed && task.failed.length > 0) {
											message.warning("部分视频上传失败，详情见下方");
										} else {
											message.success("所有音乐上传成功 (≧▽≦)");
										}
									} else {
										setProgressError(task.error || "任务失败");
										message.error(task.error || "任务失败");
									}
								}
							} else {
								clearInterval(timer);
								setProgressModalVisible(false);
								setUploading(false);
								message.error("查询任务状态失败");
							}
						})
						.catch(() => {
							clearInterval(timer);
							setProgressModalVisible(false);
							setUploading(false);
							message.error("查询任务状态失败");
						});
				}, 2000); // 每2秒轮询
				setProgressTimer(timer);
			} else {
				setUploading(false);
				message.error(`上传失败: ${response.data.msg || "未知错误"}`);
			}
		} catch (error) {
			setUploading(false);
			message.error(`上传失败: ${error.response?.data?.msg || error.message}`);
		}
	};

	useEffect(() => {
		// 进入页面先检查登录状态，未登录则跳转登录页
		const ensureLogin = async () => {
			try {
				const res = await checkLoginStatus();
				if (res.code !== 200) {
					navigate("/login");
				} else {
					// 已登录
					// 获取用户头像（失败则忽略，使用占位）
					try {
						// 后端现在直接返回图片数据，所以直接使用接口URL作为图片src
						const avatarUrl = `${axiosInstance.defaults.baseURL}/netcloud/useravatar`;
						// console.log("设置头像URL:", avatarUrl);
						setAvatarUrl(avatarUrl);
					} catch (error) {
						// console.error("获取头像失败:", error);
					}
				}
			} catch (e) {
				navigate("/login");
			}
		};
		ensureLogin();

		return () => {
			if (progressTimer) clearInterval(progressTimer);
		};
	}, [navigate]); // 移除progressTimer依赖，避免重复调用

	const handleLogout = async () => {
		try {
			await axiosInstance.post("netcloud/logout");
			message.success("已退出登录");
			navigate("/login");
		} catch (e) {
			message.error("退出失败，请重试");
		}
	};

	return (
		<App>
			<div
				style={{
					maxWidth: "800px",
					margin: "0 auto",
					padding: "20px",
					backgroundColor: "#fff",
					borderRadius: "8px",
					boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
				}}
			>
				<div
					style={{
						display: "flex",
						alignItems: "center",
						justifyContent: "center",
						marginBottom: 16,
						position: "relative",
					}}
				>
					<h1 style={{ margin: 0 }}>BVTC</h1>
					<div style={{ position: "absolute", right: 0, top: "50%", transform: "translateY(-50%)" }}>
						<div
							style={{
								width: 48,
								height: 48,
								borderRadius: "50%",
								overflow: "hidden",
								cursor: "pointer",
								boxShadow: "0 2px 6px rgba(0,0,0,0.12)",
								transition: "all 0.3s ease",
								display: "flex",
								alignItems: "center",
								justifyContent: "center",
								background: "#fafafa",
								position: "relative",
							}}
							onMouseEnter={() => setIsHovering(true)}
							onMouseLeave={() => setIsHovering(false)}
							onClick={handleLogout}
						>
							{/* 头像图片 */}
							{avatarUrl ? (
								<>
									{/* console.log("渲染头像，URL:", avatarUrl) */}
									<img
										src={avatarUrl}
										alt="avatar"
										style={{
											width: "100%",
											height: "100%",
											objectFit: "cover",
											transition: "all 0.3s ease",
											opacity: isHovering ? 0 : 1,
											transform: isHovering ? "scale(0.8)" : "scale(1)",
										}}
										onLoad={() => {
											/* console.log("头像加载成功") */
										}}
										onError={(e) => {
											/* console.error("头像加载失败:", e) */
										}}
									/>
								</>
							) : (
								<>
									{/* console.log("头像URL为空，显示默认图标") */}
									<UserOutlined
										style={{
											transition: "all 0.3s ease",
											opacity: isHovering ? 0 : 1,
											transform: isHovering ? "scale(0.8)" : "scale(1)",
											fontSize: "24px",
											color: "#666",
										}}
									/>
								</>
							)}
						</div>

						{/* 退出登录文字 - 条形样式，不受圆形限制 */}
						<div
							style={{
								position: "absolute",
								top: "50%",
								left: "50%",
								transform: "translate(-50%, -50%)",
								padding: "8px 16px",
								background: "#87ceeb",
								color: "white",
								borderRadius: "20px",
								fontSize: "12px",
								fontWeight: "bold",
								transition: "all 0.3s ease",
								opacity: isHovering ? 1 : 0,
								transform: isHovering ? "translate(-50%, -50%) scale(1)" : "translate(-50%, -50%) scale(0.8)",
								whiteSpace: "nowrap",
								userSelect: "none",
								zIndex: 10,
								minWidth: "70px",
								textAlign: "center",
								boxShadow: "0 2px 8px rgba(135, 206, 235, 0.4)",
								pointerEvents: "none",
							}}
						>
							退出登录
						</div>
					</div>
				</div>
				<div
					style={{
						display: "flex",
						gap: "10px",
						marginBottom: "20px",
						justifyContent: "center",
					}}
				>
					<Input
						placeholder="请输入B站视频ID (bvid)"
						value={videoId}
						onChange={(e) => setVideoId(e.target.value)}
						style={{ width: "300px", marginRight: "10px" }}
					/>
					<Button type="primary" onClick={handleSearch} loading={loading}>
						搜索
					</Button>
				</div>

				{videoInfo && (
					<div>
						<h2>作品信息</h2>
						<p>UP主：{videoInfo.author}</p>
						{videoInfo.video_list && videoInfo.video_list.length > 0 && (
							<div style={{ marginBottom: "20px" }}>
								<h3>当前视频</h3>
								<List
									dataSource={[videoInfo.video_list[0]]}
									renderItem={(video) => (
										<List.Item>
											<List.Item.Meta
												title={
													<Space>
														<Checkbox
															checked={selectedVideos.includes(video.bvid)}
															onChange={(e) => handleVideoSelect(video.bvid, e.target.checked)}
														/>
														<span>{video.title}</span>
													</Space>
												}
												description={
													<Space>
														<span>BV号：</span>
														<a href={video.url} target="_blank" rel="noopener noreferrer">
															{video.bvid}
														</a>
													</Space>
												}
											/>
										</List.Item>
									)}
								/>
							</div>
						)}

						{videoInfo.video_list && videoInfo.video_list.length > 1 && (
							<>
								<h2>合集列表：{videoInfo.list_title?.replace("合集·", "")}</h2>

								{/* 搜索框 */}
								<div style={{ marginBottom: "16px" }}>
									<Input.Search
										placeholder="搜索视频标题或BV号..."
										value={searchKeyword}
										onChange={(e) => {
											setSearchKeyword(e.target.value);
											setCurrentPage(1); // 搜索时重置到第一页
										}}
										onSearch={() => {}} // 实时搜索，不需要点击按钮
										allowClear
										style={{ width: "100%" }}
									/>
									{searchKeyword.trim() && (
										<div style={{ marginTop: "8px", color: "#666", fontSize: "14px" }}>
											找到 {getFilteredTotal()} 个结果
										</div>
									)}
								</div>

								<div style={{ marginBottom: "10px" }}>
									<Checkbox
										onChange={(e) => handleSelectAll(e.target.checked)}
										checked={selectedVideos.length === videoInfo.video_list.length}
										indeterminate={selectedVideos.length > 0 && selectedVideos.length < videoInfo.video_list.length}
									>
										全选
									</Checkbox>
								</div>
								<List
									dataSource={getCurrentPageVideos()}
									renderItem={(video) => (
										<List.Item>
											<List.Item.Meta
												title={
													<Space>
														<Checkbox
															checked={selectedVideos.includes(video.bvid)}
															onChange={(e) => handleVideoSelect(video.bvid, e.target.checked)}
														/>
														<span>{video.title}</span>
													</Space>
												}
												description={
													<Space>
														<span>BV号：</span>
														<a href={video.url} target="_blank" rel="noopener noreferrer">
															{video.bvid}
														</a>
													</Space>
												}
											/>
										</List.Item>
									)}
								/>
								<div style={{ marginTop: "16px", textAlign: "right" }}>
									<Pagination
										current={currentPage}
										pageSize={pageSize}
										total={getFilteredTotal()}
										onChange={handlePageChange}
										showSizeChanger={false}
									/>
								</div>
							</>
						)}

						<Button
							type="primary"
							onClick={handleSave}
							style={{ marginTop: "20px" }}
							disabled={selectedVideos.length === 0 || uploading || progressModalVisible}
							loading={uploading || progressModalVisible}
						>
							{uploading || progressModalVisible ? "上传中..." : `保存到网易云歌单 (${selectedVideos.length}个视频)`}
						</Button>
					</div>
				)}

				<Modal title="选择歌单" open={isModalVisible} onCancel={() => setIsModalVisible(false)} footer={null}>
					<List
						dataSource={[{ pname: "不加入歌单，仅添加到云盘" }, ...playlists]}
						style={{ maxHeight: "400px", overflowY: "auto" }}
						renderItem={(playlist) => (
							<List.Item>
								<List.Item.Meta title={playlist.pname} />
								<Button
									onClick={() => {
										// console.log("选择歌单:", playlist);
										setIsModalVisible(false);
										setSelectedPlaylist(playlist);
										// 先询问是否保留原标题
										setKeepTitleModalVisible(true);
									}}
								>
									选择
								</Button>
							</List.Item>
						)}
					/>
				</Modal>

				{/* 是否保留原标题弹框 */}
				<Modal
					title="是否保留原视频名称？"
					open={keepTitleModalVisible}
					onCancel={() => {
						setKeepTitleModalVisible(false);
						setSelectedPlaylist(null);
						setIsModalVisible(true);
					}}
					footer={[
						<Button
							key="yes"
							onClick={() => {
								setUseOriginalTitle(true);
								setKeepTitleModalVisible(false);
								setConfirmModalVisible(true);
							}}
						>
							是，保留原标题
						</Button>,
						<Button
							key="no"
							type="primary"
							loading={titleSuggesting}
							onClick={async () => {
								setUseOriginalTitle(false);
								setKeepTitleModalVisible(false);
								// 先打开编辑弹窗，再流式填充
								setTitleEditModalVisible(true);
								try {
									await prepareTitleSuggestions();
								} catch (e) {
									message.error("AI 标题建议失败，请稍后重试");
									setSelectedPlaylist(null);
								}
							}}
						>
							否，我要自定义
						</Button>,
					]}
					centered
					width={480}
				>
					<div style={{ textAlign: "center", padding: "10px 0" }}>
						<p style={{ fontSize: "14px", color: "#666", margin: 0 }}>可先给出 AI 建议名，再自行修改</p>
					</div>
				</Modal>

				{/* 标题编辑弹框 */}
				<Modal
					title="编辑上传标题（可基于 AI 建议）"
					open={titleEditModalVisible}
					onCancel={() => {
						setTitleEditModalVisible(false);
						setSelectedPlaylist(null);
						setIsModalVisible(true);
					}}
					footer={[
						<Button
							key="back"
							onClick={() => {
								setTitleEditModalVisible(false);
								setKeepTitleModalVisible(true);
							}}
						>
							上一步
						</Button>,
						<Button
							key="confirm"
							type="primary"
							disabled={titleSuggesting}
							onClick={() => {
								setTitleEditModalVisible(false);
								setConfirmModalVisible(true);
							}}
						>
							确定
						</Button>,
					]}
					centered
					width={640}
				>
					<div style={{ maxHeight: 360, overflowY: "auto" }}>
						{selectedVideos.map((bvid) => {
							const original = getOriginalTitleByBvid(bvid);
							return (
								<div key={bvid} style={{ marginBottom: 12 }}>
									<div style={{ fontSize: 12, color: "#999", marginBottom: 6 }}>BV号：{bvid}</div>
									<div style={{ fontSize: 12, color: "#666", marginBottom: 6 }}>原标题：{original}</div>
									<div style={{ display: "flex", gap: 8 }}>
										<Input
											placeholder="请输入歌曲名"
											value={titleOverride[bvid] ?? ""}
											onChange={(e) => setTitleOverride((prev) => ({ ...prev, [bvid]: e.target.value }))}
											maxLength={60}
										/>
										<Button size="small" loading={titleSuggesting} onClick={() => regenerateSuggestion(bvid)}>
											重新生成
										</Button>
									</div>
								</div>
							);
						})}
					</div>
				</Modal>
				<Modal
					title="登录提示"
					open={isLoginModalVisible}
					onCancel={() => setIsLoginModalVisible(false)}
					footer={[
						<Button key="cancel" onClick={() => setIsLoginModalVisible(false)}>
							取消
						</Button>,
						<Button
							key="login"
							type="primary"
							onClick={() => {
								setIsLoginModalVisible(false);
								navigate("/login");
							}}
						>
							去登录
						</Button>,
					]}
				>
					<p>请先登录网易云音乐账号</p>
				</Modal>

				<Modal open={progressModalVisible} footer={null} closable={false} centered title="上传进度">
					<div style={{ marginBottom: 16 }}>
						{progressError ? (
							<span style={{ color: "red" }}>{progressError}</span>
						) : (
							<>
								<p>正在上传，请稍候...</p>
								<Progress
									percent={progress}
									status={taskStatus === "failed" ? "exception" : taskStatus === "completed" ? "success" : "active"}
								/>
							</>
						)}
					</div>
				</Modal>

				<Modal
					open={!!uploadResult}
					onCancel={() => setUploadResult(null)}
					footer={[
						<Button key="close" type="primary" onClick={() => setUploadResult(null)}>
							关闭
						</Button>,
					]}
					title="上传结果"
				>
					{uploadResult && (
						<>
							{uploadResult.success.length > 0 && (
								<div style={{ marginBottom: 12 }}>
									<b style={{ color: "green" }}>上传成功：</b>
									<ul>
										{uploadResult.success.map((title, idx) => (
											<li key={idx}>{title}</li>
										))}
									</ul>
								</div>
							)}
							{uploadResult.failed.length > 0 && (
								<div>
									<b style={{ color: "red" }}>上传失败：</b>
									<ul>
										{uploadResult.failed.map((item, idx) => (
											<li key={idx}>
												{item.title}：{item.error}
											</li>
										))}
									</ul>
								</div>
							)}
						</>
					)}
				</Modal>

				{/* 上传确认弹框 */}
				<Modal
					title="确认上传"
					open={confirmModalVisible}
					onCancel={() => {
						setConfirmModalVisible(false);
						setSelectedPlaylist(null);
						setIsModalVisible(true); // 重新打开歌单选择弹框
					}}
					footer={[
						<Button
							key="cancel"
							onClick={() => {
								setConfirmModalVisible(false);
								setSelectedPlaylist(null);
								setIsModalVisible(true); // 重新打开歌单选择弹框
							}}
						>
							取消
						</Button>,
						<Button
							key="confirm"
							type="primary"
							onClick={() => {
								setConfirmModalVisible(false);
								handleUpload(selectedPlaylist);
								setSelectedPlaylist(null);
							}}
						>
							确认上传
						</Button>,
					]}
					centered
					width={400}
				>
					<div style={{ textAlign: "center", padding: "20px 0" }}>
						<p style={{ fontSize: "16px", color: "#666", margin: 0 }}>
							是否确认将选中的 {selectedVideos.length} 首歌曲上传到
							<span style={{ color: "#1890ff", fontWeight: "bold" }}>{selectedPlaylist?.pname || "云盘"}</span>？
						</p>
					</div>
				</Modal>
			</div>
		</App>
	);
};

export default BilibiliPage;
