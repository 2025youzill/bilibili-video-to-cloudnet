<div align="center">

# BVTC(Bilibili-Video-To-CloudNet)

![](https://img.shields.io/github/go-mod/go-version/2025youzill/bilibili-video-to-mp4?filename=banked%2Fgo.mod) ![](https://img.shields.io/badge/npm-10.9.0-blue)

</div>

## :blue_book:项目介绍

BVTC（Bilibili-Video-To-CloudNet）是一个将 B 站视频转换为 MP4 格式并上传到网易云的网站，输入 Bvid 和登录网易云后，后端采用 API 接口抓取视频，FFmpeg 提取音频，然后通过网易云网盘上传到歌单。
如果你在 b 站有喜欢的音乐但是网易云没有，欢迎使用 BVTC，如果你有喜欢的 AMSR 但是网易云没有，欢迎使用 BVTC，~~如果你有喜欢的美女视频，欢迎分享给我:point_up:~~
如果有帮助到你或者你很喜欢的话，给鼠鼠一个 star:star2:再走吧
如果你发现了什么问题或者有任何改进的建议，不要害羞，无需吝啬你的 [issue](https://github.com/2025youzill/bilibili-video-to-cloudnet/issues/new) 和 pr ,如果你不清楚如何提交，可以参考 [Github Docs](https://docs.github.com/en/pull-requests)(建议是先 fork 再提交会方便些)。

## :open_book:使用说明

- 使用内嵌 ffmpeg，下载链接：[ffmpeg-master-latest-win64-gpl.zip](https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip)，解压，并将 ffmpeg.exe 添加到 banked/tool/ffmpeg
- go 程序运行使用 air 运行（可热重载），相关信息可查看 git 仓库：[☁️ Live reload for Go apps](https://github.com/air-verse/air)

## :gear:运行

### 后端

- 进入 banked 文件夹
  ```bash
  cd banked
  ```
- 运行程序
  ```bash
  air
  ```

### 前端

- 进入 fronted 文件夹
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

- [ ]  保存的歌曲没有歌词，对歌词功能的完善（现在不支持读取lrc文件，没有什么想法，只能等大佬发现方法了）

## ❤️ 鸣谢

- [✨ 网易云音乐 Golang 🎵](https://github.com/chaunsin/netease-cloud-music)
- [bilibili 的 API 的 Go SDK](https://github.com/CuteReimu/bilibili)
- [FFmpeg](https://ffmpeg.org/)

以及本项目所依赖的所有优秀的库。

## ⚠️ 声明

**切勿用作商业用途、非法用途使用！！！**

**本项目解析得到的所有内容均来自 B 站 UP 主上传、分享，其版权均归原作者所有，请尊重 up 主的努力。**

**本项目存储在网易云当中，最后所属权归网易云所有，请勿用作资源分享。**

**本项目仅供个人学习使用,利用本项目造成不良影响及后果与本人无关。**
