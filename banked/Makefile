# Makefile for cross-platform FFmpeg setup
FFMPEG_WIN_URL := https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip
FFMPEG_LINUX_URL := https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz

# 使用环境变量设置 FFMPEG_DIR，默认为 tool/ffmpeg
FFMPEG_DIR ?= tool/ffmpeg

# 平台检测
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    PLATFORM := linux
else ifeq ($(UNAME_S),Darwin)
    PLATFORM := macos
else
    PLATFORM := windows
endif

.PHONY: setup-ffmpeg setup-windows-ffmpeg setup-linux-ffmpeg clean-ffmpeg

# 默认设置（根据当前平台）
setup-ffmpeg: setup-$(PLATFORM)-ffmpeg

# Windows设置
setup-windows-ffmpeg:
	@echo Setting up FFmpeg for Windows...
	
	@rem 创建目录（如果不存在）
	@if not exist "$(FFMPEG_DIR)" mkdir "$(FFMPEG_DIR)"
	
	@rem 下载ZIP文件
	@echo Downloading FFmpeg for Windows...
	@powershell -Command "Invoke-WebRequest -Uri '$(FFMPEG_WIN_URL)' -OutFile '$(FFMPEG_DIR)\ffmpeg-windows.zip'"
	
	@rem 解压文件
	@echo Extracting FFmpeg...
	@powershell -Command "Expand-Archive -Path '$(FFMPEG_DIR)\ffmpeg-windows.zip' -DestinationPath '$(FFMPEG_DIR)'"
	
	@rem 复制ffmpeg.exe
	@echo Copying ffmpeg.exe...
	@copy "$(FFMPEG_DIR)\ffmpeg-master-latest-win64-gpl\bin\ffmpeg.exe" "$(FFMPEG_DIR)\" > NUL
	
	@rem 清理临时文件
	@echo Cleaning up...
	@del /Q "$(FFMPEG_DIR)\ffmpeg-windows.zip" > NUL
	@rmdir /S /Q "$(FFMPEG_DIR)\ffmpeg-master-latest-win64-gpl" > NUL
	
	@echo Windows FFmpeg setup completed successfully!

# Linux设置（Docker 容器内使用）
setup-linux-ffmpeg:
	@echo "Setting up FFmpeg for Linux in Docker..."
	@mkdir -p $(FFMPEG_DIR)
	
	@echo "Downloading FFmpeg for Linux..."
	@curl -L -o $(FFMPEG_DIR)/ffmpeg-linux.tar.xz $(FFMPEG_LINUX_URL)
	
	@echo "Extracting FFmpeg..."
	@tar -xf $(FFMPEG_DIR)/ffmpeg-linux.tar.xz -C $(FFMPEG_DIR)
	
	@echo "Moving ffmpeg binary..."
	@mv $(FFMPEG_DIR)/ffmpeg-master-latest-linux64-gpl/bin/ffmpeg $(FFMPEG_DIR)/
	
	@echo "Cleaning up..."
	@rm -f $(FFMPEG_DIR)/ffmpeg-linux.tar.xz
	@rm -rf $(FFMPEG_DIR)/ffmpeg-master-latest-linux64-gpl
	
	@chmod +x $(FFMPEG_DIR)/ffmpeg
	@echo "Linux FFmpeg setup completed successfully in Docker!"

# 清理 FFmpeg 文件
clean-ffmpeg:
	@echo Cleaning FFmpeg files...
ifeq ($(PLATFORM),windows)
	@if exist "$(FFMPEG_DIR)\ffmpeg.exe" del /Q "$(FFMPEG_DIR)\ffmpeg.exe" > NUL
else
	@rm -f $(FFMPEG_DIR)/ffmpeg
endif
	@rm -f $(FFMPEG_DIR)/ffmpeg-*.zip
	@rm -f $(FFMPEG_DIR)/ffmpeg-*.tar.xz
	@rm -rf $(FFMPEG_DIR)/ffmpeg-master-*
	@echo FFmpeg files cleaned!