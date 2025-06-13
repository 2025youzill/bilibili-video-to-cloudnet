import React, { useState, useEffect } from "react";
import { getPlaylists } from "../api/netease";
import { List, Card, message } from "antd";

const PlaylistList = ({ onSelect }) => {
	const [playlists, setPlaylists] = useState([]);
	const [loading, setLoading] = useState(false);

	useEffect(() => {
		fetchPlaylists();
	}, []);

	const fetchPlaylists = async () => {
		try {
			setLoading(true);
			const response = await getPlaylists();
			if (response.code === 200 && Array.isArray(response.data)) {
				setPlaylists(response.data);
			} else {
				message.error("获取歌单列表失败");
			}
		} catch (error) {
			message.error("获取歌单列表失败");
		} finally {
			setLoading(false);
		}
	};

	const handlePlaylistSelect = (playlist) => {
		if (onSelect) {
			onSelect(playlist);
		}
	};

	return (
		<div style={{ padding: "20px" }}>
			<h2>选择歌单</h2>
			<List
				grid={{ gutter: 16, column: 4 }}
				dataSource={playlists}
				loading={loading}
				renderItem={(item) => (
					<List.Item>
						<Card hoverable title={item.pname} onClick={() => handlePlaylistSelect(item)}>
							<p>歌单ID: {item.pid}</p>
						</Card>
					</List.Item>
				)}
			/>
		</div>
	);
};

export default PlaylistList;
