import { useEffect, useState } from "react";
import "../styles/background.css";

// 背景图片列表（移到组件外避免重新创建）
const backgrounds = {
	desktop: [
		"/picture/desktop/wallhaven-z8o88j.jpg",
		"/picture/desktop/wallhaven-k7wor1.jpg",
		"/picture/desktop/wallhaven-281d5y.png",
		"/picture/desktop/wallhaven-rqrv21.png",
	],
	mobile: ["/picture/mobile/wallhaven-d8x71m.jpg", "/picture/mobile/wallhaven-mlwdvk.png", ],
};

// 获取随机背景图（避免重复）
const getRandomBackground = (currentBg, isMobile) => {
	const list = isMobile ? backgrounds.mobile : backgrounds.desktop;
	const availableBackgrounds = list.filter((bg) => bg !== currentBg);
	if (availableBackgrounds.length === 0) return list[0];
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
	const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);

	const [isImageLoadFailed, setIsImageLoadFailed] = useState(false);

	useEffect(() => {
		const handleResize = () => {
			setIsMobile(window.innerWidth <= 768);
		};
		window.addEventListener("resize", handleResize);

		// 初始化：随机选择第一张背景图并预加载
		const list = isMobile ? backgrounds.mobile : backgrounds.desktop;
		const initialBg = list[Math.floor(Math.random() * list.length)];

		preloadImage(initialBg)
			.then(() => {
				setLayer1Image(initialBg);
				setLayer1State("zooming");
				setActiveLayer(1);
				// 1.5秒后动画完成
				setTimeout(() => {
					setLayer1State("visible");
				}, 1500);
			})
			.catch(() => {
				// 如果第一张图片预加载失败（比如文件夹不存在），记录状态
				setIsImageLoadFailed(true);
			});

		// 预加载所有背景图片
		const allBgs = [...backgrounds.desktop, ...backgrounds.mobile];
		allBgs.forEach((bg) => {
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
			const nextBg = getRandomBackground(
				currentBg || (isMobile ? backgrounds.mobile[0] : backgrounds.desktop[0]),
				isMobile,
			);

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
			{/* 如果图片加载失败或没有图片，显示纯白色背景 */}
			{isImageLoadFailed ? (
				<div
					style={{
						position: "fixed",
						top: 0,
						left: 0,
						width: "100vw",
						height: "100vh",
						background: "#ffffff",
						zIndex: -5,
					}}
				/>
			) : (
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
			)}
		</>
	);
};

export default BackgroundImage;
