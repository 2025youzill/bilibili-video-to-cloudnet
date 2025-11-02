import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import { App as AntdApp, ConfigProvider, message } from "antd";
import LoginPage from "./pages/LoginPage";
import BilibiliPage from "./pages/BilibiliPage";
import PlaylistList from "./components/PlaylistList";
import BackgroundImage from "./components/BackgroundImage";

// 配置全局message
message.config({
	duration: 2,
	maxCount: 1,
	top: 50,
});

function App() {
	return (
		<ConfigProvider>
			<AntdApp>
				{/* 全局背景组件 */}
				<BackgroundImage />
				<Router>
					<Routes>
						{/* 根路径重定向到bvtc */}
						<Route path="/" element={<Navigate to="/bvtc" replace />} />

						{/* 主要路由，都添加bvtc前缀 */}
						<Route path="/bvtc" element={<BilibiliPage />} />
						<Route path="/bvtc/login" element={<LoginPage />} />
						<Route path="/bvtc/bilibili" element={<BilibiliPage />} />
						<Route path="/bvtc/playlists" element={<PlaylistList />} />

						{/* 兼容旧路径，重定向到新路径 */}
						<Route path="/login" element={<Navigate to="/bvtc/login" replace />} />
						<Route path="/bilibili" element={<Navigate to="/bvtc/bilibili" replace />} />
						<Route path="/playlists" element={<Navigate to="/bvtc/playlists" replace />} />
					</Routes>
				</Router>
			</AntdApp>
		</ConfigProvider>
	);
}

export default App;
