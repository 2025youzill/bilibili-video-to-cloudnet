import React, { useState } from "react";
import { Input, Button, List, message, Modal, Checkbox, Space, Pagination, App } from "antd";
import { useNavigate } from "react-router-dom";
import axios from "../axiosInstance";

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
	const pageSize = 10;
	const navigate = useNavigate();
	const { message: messageApi } = App.useApp();

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
			let url = "/bilibili/list";
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

			console.log("请求URL:", url);
			const response = await axios.get(url);
			console.log("响应数据:", response.data);
			setVideoInfo(response.data.data);
			// 重置选中的视频
			setSelectedVideos([]);
		} catch (error) {
			console.error("请求错误:", error);
			// 打印完整的错误对象，帮助调试
			console.log("错误详情:", {
				response: error.response,
				status: error.response?.status,
				data: error.response?.data,
			});

			// 检查错误响应
			if (error.response) {
				console.log("错误状态码:", error.response.status);
				console.log("错误信息:", error.response.data);
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
			const cloudCheck = await axios.get("/netcloud/login/check");
			if (!cloudCheck.data.data) {
				setIsLoginModalVisible(true);
				return;
			}

			const response = await axios.get("/netcloud/playlist");
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

	// 计算当前页的视频列表
	const getCurrentPageVideos = () => {
		if (!videoInfo?.video_list) return [];
		const startIndex = (currentPage - 1) * pageSize;
		return videoInfo.video_list.slice(1).slice(startIndex, startIndex + pageSize);
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
				<h1 style={{ textAlign: "center", marginBottom: "24px" }}>BVTC (bilibili-video-to-cloudnet)</h1>
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
										total={videoInfo.video_list.length - 1}
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
							disabled={selectedVideos.length === 0 || uploading}
							loading={uploading}
						>
							{uploading ? "上传中..." : `保存到网易云歌单 (${selectedVideos.length}个视频)`}
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
										console.log("选择歌单:", playlist);
										setIsModalVisible(false);
										const confirmModal = message.confirm({
											title: "确认上传",
											content: `是否确认上传到${playlist.pname || "云盘"}？`,
											okText: "确认",
											cancelText: "取消",
											onOk: async () => {
												console.log("开始上传，选中的视频:", selectedVideos);
												setUploading(true);
												// 关闭确认弹框
												confirmModal.destroy();
												// 显示上传中的Modal
												const uploadModal = message.info({
													title: "上传中，请稍等...(๑´ڡ`๑)",
													icon: null,
													okButtonProps: { style: { display: "none" } },
													cancelButtonProps: { style: { display: "none" } },
													closable: false,
													maskClosable: false,
												});
												try {
													const requestData = {
														bvid: selectedVideos,
														splaylist: !!playlist.pid,
														pid: playlist.pid || undefined,
													};
													console.log("发送请求数据:", requestData);
													const response = await axios.post("/bilibili/load", requestData);
													console.log("上传响应:", response.data);
													// 关闭上传中的Modal
													uploadModal.destroy();
													if (response.data.code === 200) {
														const data = response.data.data;
														if (data.failed && data.failed.length > 0) {
															const failedMsgs = data.failed
																.map((f) => `《${f.title}》处理失败: ${f.error}`)
																.join("\n");
															message.error(`部分视频上传失败 (ŏ﹏ŏ、)\n${failedMsgs}`);
														} else if (data.success && data.success.length > 0) {
															message.success("所有音乐上传成功 (≧▽≦)");
														}
													} else {
														message.error(`上传失败 (ŏ﹏ŏ、)۶: ${response.data.msg || "未知错误"}`);
													}
												} catch (error) {
													console.error("上传错误:", error);
													// 关闭上传中的Modal
													uploadModal.destroy();
													if (error.code === "ERR_NETWORK") {
														message.error("网络连接失败，请检查网络连接或稍后重试 (ŏ﹏ŏ、)۶");
													} else if (error.code === "ECONNABORTED") {
														message.error("上传超时，请检查网络连接或稍后重试 (ŏ﹏ŏ、)۶");
													} else {
														message.error(`上传失败 (ŏ﹏ŏ、)۶: ${error.response?.data?.msg || error.message}`);
													}
												} finally {
													setUploading(false);
												}
											},
											onCancel: () => {
												console.log("取消上传，返回歌单选择");
												setIsModalVisible(true);
											},
										});
										console.log("确认弹框已创建");
									}}
								>
									选择
								</Button>
							</List.Item>
						)}
					/>
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
			</div>
		</App>
	);
};

export default BilibiliPage;
