<div align="center">

# BVTC(Bilibili-Video-To-CloudNet)

![](https://img.shields.io/github/go-mod/go-version/2025youzill/bilibili-video-to-mp4?filename=banked%2Fgo.mod) ![](https://img.shields.io/badge/npm-10.9.0-blue)

</div>

## :blue_book:项目介绍

- BVTC（Bilibili-Video-To-CloudNet）是一个将 B 站视频转换为 MP4 格式并上传到网易云的网站，输入 Bvid 和登录网易云后，后端采用 **API** 接口抓取视频，FFmpeg 提取音频，然后通过网易云网盘上传到歌单。
- 如果你在 b 站有喜欢的音乐但是网易云没有，欢迎使用 BVTC，如果你有喜欢的 AMSR 但是网易云没有，欢迎使用 BVTC，~~如果你有喜欢的美女视频，欢迎分享给我 ☝️~~
- 如果有帮助到你或者你很喜欢的话，给鼠鼠一个 🌟 再走吧
- 如果你发现了什么问题或者有任何改进的建议以及想要新增的功能，不要害羞，无需吝啬你的 [issue](https://github.com/2025youzill/bilibili-video-to-cloudnet/issues/new) 和 pr ,如果你不清楚如何提交，可以参考 [Github Docs](https://docs.github.com/en/pull-requests)。

## :open_book:使用说明

- 后端使用 go 版本为 1.24.5，前端使用 npm 版本为 10.9.0，node 版本为 22.12.0，其余具体库版本见 go.mod 和 package.json
- 使用内嵌 ffmpeg，可自行下载（会快一些）：[ffmpeg-master-latest-win64-gpl.zip](https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip)，解压，并将 ffmpeg.exe 添加到 banked/tool/ffmpeg ，或者用下边运行步骤中的 Makefile 下载
- go 程序推荐使用 air 运行（可热重载），相关信息可查看 git 仓库：[☁️ Live reload for Go apps](https://github.com/air-verse/air)
- 项目在线部署在 [https://youzill.top/bvtc](https://youzill.top/bvtc)

## :gear:运行

### 环境配置

- 首先先完成.env 文件的设置

  ```bash
  cp env.template .env
  ```
- 然后完善.env 文件

### Docker Compose 部署

- 首先拉取 Linux 所用的 FFMPEG

  ```bash
  cd banked
  make setup-linux-ffmpeg
  ```
- 在根目录运行 docker-compose-local.yml

  ```bash
  cd ..
  docker compose -f docker-compose.local.yml up -d
  ```
- 如果拉取不到基础镜像可以先拉镜像

  ```bash
  docker pull golang:1.24-alpine
  docker pull ubuntu:22.04
  docker pull node:22-alpine
  docker pull nginx:alpine
  docker pull redis:alpine
  ```

### Windows 部署

#### 后端部署

- 进入 banked 文件夹
  ```bash
  cd banked
  ```
- 运行 Makefile 安装 ffmpeg.exe
  ```bash
  make setup-windows-ffmpeg
  ```
- go install 安装 air
  ```bash
  go install github.com/air-verse/air@latest
  ```
- 运行程序
  ```bash
  air
  ```

#### 前端部署

- 进入 fronted 文件夹
  ```bash
  cd fronted
  ```
- 安装依赖
  ```bash
  npm install
  ```
- 运行程序
  ```bash
  npm start
  ```

## 🎉 启动成功

端口将在本地 http://localhost:8000 开放

## 📷 运行截图

<p align="center">
  <img src="https://github.com/user-attachments/assets/098213f1-021d-4a92-8b47-741f0216edd5" alt="照片" width=50%>
</p>

---

<p align="center">
  <img src="https://github.com/user-attachments/assets/1331c1b6-af35-4378-b5e3-f3a980508486" alt="照片" width=50%>
</p>

---

<p align="center">
  <img src="https://github.com/user-attachments/assets/31e730c4-5a2f-47b3-87b7-b909ee80e993" alt="照片" width=50%>
</p>

## :hammer_and_wrench:TODO

- [ ]  保存的歌曲没有歌词，对歌词功能的完善（现在不支持读取 lrc 文件，没有什么想法，只能等大佬发现方法了）
- [X]  采用 qwen2.5 7B 模型通过标题和简介智能生成歌名
- [ ]  批量识别歌名
- [ ]  修改前端背景

## ❤️ 鸣谢

- [✨ 网易云音乐 Golang 🎵](https://github.com/chaunsin/netease-cloud-music)
- [bilibili 的 API 的 Go SDK](https://github.com/CuteReimu/bilibili)
- [FFmpeg](https://ffmpeg.org/)
- [FFmpeg Static Auto-Builds](https://github.com/BtbN/FFmpeg-Builds)

  以及本项目所依赖的所有优秀的库。

## ⚠️ 声明

**切勿用作商业用途、非法用途使用！！！**

**本项目解析得到的所有内容均来自 B 站 UP 主上传、分享，其版权均归原作者所有，请尊重 up 主的努力。**

**本项目存储在网易云当中，最后所属权归网易云所有，请勿用作资源分享。**

**本项目仅供个人学习使用,利用本项目造成不良影响及后果与本人无关。**
