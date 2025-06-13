<div align="center">

# BVTC(Bilibili-Video-To-CloudNet)

![](https://img.shields.io/github/go-mod/go-version/2025youzill/bilibili-video-to-mp4?filename=banked%2Fgo.mod) ![](https://img.shields.io/badge/npm-10.9.0-blue)

</div>

## ⚠️ 声明

**切勿用作商业用途、非法用途使用！！！**

**本项目解析得到的所有内容均来自B站UP主上传、分享，其版权均归原作者所有，请尊重up主的努力。**

**本项目存储在网易云当中，最后所属权归网易云所有，请勿用作资源分享。**

**本项目仅供个人学习使用,利用本项目造成不良影响及后果与本人无关。**

## :book:使用说明

- 使用内嵌ffmpeg，下载链接：[ffmpeg-master-latest-win64-gpl.zip](https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip)，解压，并将ffmpeg.exe添加到banked/tool/ffmpeg
- go程序运行使用air运行（可热重载），相关信息查看相关git仓库：[air-verse/air: ☁️ Live reload for Go apps](https://github.com/air-verse/air)

## :gear:运行

### 后端

- 进入banked文件夹
  ```bash
  cd banked
  ```
- 运行程序
  ```bash
  air
  ````

### 前端

- 进入fronted文件夹
  ```
  cd fronted
  ```
- 安装依赖
  ```
  npm install
  ```
- 运行程序
  ```
  npm start
  ```
- 端口将在 http://localhost:3000 开放

## :hammer_and_wrench:TODO

- 保存的歌曲没有歌词，对歌词功能的完善（没有什么想法，只能等大佬发现方法了）

## ❤️ 鸣谢

- [✨ 网易云音乐 Golang 🎵](https://github.com/chaunsin/netease-cloud-music)
- [CuteReimu/bilibili: 哔哩哔哩bilibili的API的Go SDK](https://github.com/CuteReimu/bilibili)
- [FFmpeg](https://ffmpeg.org/)

以及本项目所依赖的所有优秀的库。
