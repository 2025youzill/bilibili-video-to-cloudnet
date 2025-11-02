import { useEffect, useState } from "react";
import "../styles/background.css";

// 背景图片列表（移到组件外避免重新创建）
const backgrounds = [
	"/picture/wallhaven-z8o88j.jpg",
	"/picture/wallhaven-k7wor1.jpg",
	"/picture/wallhaven-281d5y.png",
	"/picture/wallhaven-rqrv21.png",
];

// 获取随机背景图（避免重复）
const getRandomBackground = (currentBg) => {
	const availableBackgrounds = backgrounds.filter((bg) => bg !== currentBg);
	const randomIndex = Math.floor(Math.random() * availableBackgrounds.length);
	return availableBackgrounds[randomIndex];
};

// 预加载图片
const preloadImage = (src) => {
	return new Promise((resolve, reject) => {
		const img = new Image();
		img.onload = () => resolve(src);
		img.onerror = reject;
		img.src = src;
	});
};

const BackgroundImage = () => {
	const [layer1Image, setLayer1Image] = useState("");
	const [layer2Image, setLayer2Image] = useState("");
	const [activeLayer, setActiveLayer] = useState(1); // 1 或 2
	const [layer1State, setLayer1State] = useState("hidden"); // hidden, zooming, visible, fading
	const [layer2State, setLayer2State] = useState("hidden");

	useEffect(() => {
		// 初始化：随机选择第一张背景图并预加载
		const initialBg = backgrounds[Math.floor(Math.random() * backgrounds.length)];

		preloadImage(initialBg).then(() => {
			setLayer1Image(initialBg);
			setLayer1State("zooming");
			setActiveLayer(1);
			// 1.5秒后动画完成
			setTimeout(() => {
				setLayer1State("visible");
			}, 1500);
		});

		// 预加载所有背景图片
		backgrounds.forEach((bg) => {
			if (bg !== initialBg) {
				preloadImage(bg);
			}
		});

		// 使用ref来保存最新的状态，避免闭包问题
		let currentActiveLayer = 1;
		let currentLayer1Image = initialBg;
		let currentLayer2Image = "";

		// 每5秒切换背景（2s过渡 + 3s显示）
		const interval = setInterval(() => {
			const currentBg = currentActiveLayer === 1 ? currentLayer1Image : currentLayer2Image;
			const nextBg = getRandomBackground(currentBg || backgrounds[0]);

			// 预加载下一张图片后开始过渡
			preloadImage(nextBg).then(() => {
				if (currentActiveLayer === 1) {
					// 当前是layer1，切换到layer2
					currentLayer2Image = nextBg;

					// 先确保layer2是hidden状态，然后设置新图片
					setLayer2State("hidden");
					setLayer2Image(nextBg);

					// layer1开始淡出
					setLayer1State("fading");

					// layer2开始从小变大
					setTimeout(() => {
						setLayer2State("zooming");
					}, 500);

					// 2秒后，layer2完全显示，layer1立即隐藏
					setTimeout(() => {
						setLayer2State("visible");
						setLayer1State("hidden");
						setActiveLayer(2);
						currentActiveLayer = 2;
					}, 2000);
				} else {
					// 当前是layer2，切换到layer1
					currentLayer1Image = nextBg;

					// 先确保layer1是hidden状态，然后设置新图片
					setLayer1State("hidden");
					setLayer1Image(nextBg);

					// layer2开始淡出（1.5秒）
					setLayer2State("fading");

					// 0.5秒后，layer1开始从小变大
					setTimeout(() => {
						setLayer1State("zooming");
					}, 500);

					// 2秒后，layer1完全显示，layer2立即隐藏
					setTimeout(() => {
						setLayer1State("visible");
						setLayer2State("hidden");
						setActiveLayer(1);
						currentActiveLayer = 1;
					}, 2000);
				}
			});
		}, 7000); // 7秒周期：2s过渡 + 5s完整显示

		return () => clearInterval(interval);
	}, []); // 空依赖数组，只在组件挂载时执行一次

	// 获取图层的类名
	const getLayerClassName = (state) => {
		if (state === "zooming") return "background-image background-image-next zoom-in";
		if (state === "fading") return "background-image fade-out";
		if (state === "visible") return "background-image";
		return "background-image-hidden";
	};

	return (
		<>
			{/* 图层1 - 使用图片URL作为key */}
			{layer1Image && layer1State !== "hidden" && (
				<div
					key={`layer1-${layer1Image}`}
					className={getLayerClassName(layer1State)}
					style={{
						backgroundImage: `url(${layer1Image})`,
					}}
				/>
			)}
			{/* 图层2 - 使用图片URL作为key */}
			{layer2Image && layer2State !== "hidden" && (
				<div
					key={`layer2-${layer2Image}`}
					className={getLayerClassName(layer2State)}
					style={{
						backgroundImage: `url(${layer2Image})`,
					}}
				/>
			)}
			{/* 半透明渐变遮罩层 */}
			<div className="background-overlay" />
		</>
	);
};

export default BackgroundImage;
