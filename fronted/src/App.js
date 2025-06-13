import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { App as AntdApp, ConfigProvider, message } from "antd";
import LoginPage from "./pages/LoginPage";
import BilibiliPage from "./pages/BilibiliPage";
import PlaylistList from "./components/PlaylistList";

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
				<Router>
					<Routes>
						<Route path="/" element={<BilibiliPage />} />
						<Route path="/login" element={<LoginPage />} />
						<Route path="/bilibili" element={<BilibiliPage />} />
						<Route path="/playlists" element={<PlaylistList />} />
					</Routes>
				</Router>
			</AntdApp>
		</ConfigProvider>
	);
}

export default App;
