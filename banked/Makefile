# Makefile for FFmpeg setup - Independent targets
FFMPEG_WIN_URL := https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip
FFMPEG_LINUX_URL := https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz

# 使用环境变量设置 FFMPEG_DIR，默认为 tool/ffmpeg
FFMPEG_DIR ?= tool/ffmpeg

.PHONY: setup-windows-ffmpeg setup-linux-ffmpeg clean-ffmpeg help

# 帮助信息
help:
	@echo "FFmpeg Setup Makefile"
	@echo "===================="
	@echo "Available targets:"
	@echo "  setup-windows-ffmpeg  - Setup FFmpeg for Windows (Windows only)"
	@echo "  setup-linux-ffmpeg    - Setup FFmpeg for Linux (Cross-platform: Windows & Linux)"
	@echo "  clean-ffmpeg          - Clean FFmpeg files (Cross-platform)"
	@echo "  help                  - Show this help message"
	@echo ""
	@echo "Usage: make <target>"
	@echo "Examples:"
	@echo "  make setup-windows-ffmpeg  # Download Windows FFmpeg"
	@echo "  make setup-linux-ffmpeg    # Download Linux FFmpeg (works on both Windows and Linux)"

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

# Linux设置
setup-linux-ffmpeg:
	@echo "Setting up FFmpeg for Linux..."
ifeq ($(OS),Windows_NT)
	@rem Windows版本
	@if not exist "$(FFMPEG_DIR)" mkdir "$(FFMPEG_DIR)"
	@echo "Downloading FFmpeg for Linux..."
	@powershell -Command "Invoke-WebRequest -Uri '$(FFMPEG_LINUX_URL)' -OutFile '$(FFMPEG_DIR)\ffmpeg-linux.tar.xz'"
	@echo "Extracting FFmpeg..."
	@tar -xf "$(FFMPEG_DIR)\ffmpeg-linux.tar.xz" -C "$(FFMPEG_DIR)" 2>NUL || echo "Note: tar command not available, you may need to manually extract the file"
	@echo "Copying ffmpeg binary..."
	@copy "$(FFMPEG_DIR)\ffmpeg-master-latest-linux64-gpl\bin\ffmpeg" "$(FFMPEG_DIR)\" > NUL
	@echo "Cleaning up..."
	@del /Q "$(FFMPEG_DIR)\ffmpeg-linux.tar.xz" > NUL 2>&1
	@rmdir /S /Q "$(FFMPEG_DIR)\ffmpeg-master-latest-linux64-gpl" > NUL 2>&1
	@echo "Linux FFmpeg setup completed successfully on Windows!"
else
	@rem Linux/Unix版本
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
	@echo "Linux FFmpeg setup completed successfully on Linux!"
endif

# 清理 FFmpeg 文件
clean-ffmpeg:
	@echo Cleaning FFmpeg files...
	@echo "Note: This will clean all FFmpeg files regardless of platform"
	@if exist "$(FFMPEG_DIR)\ffmpeg.exe" del /Q "$(FFMPEG_DIR)\ffmpeg.exe" > NUL 2>&1 || echo "No Windows ffmpeg.exe found"
	@if exist "$(FFMPEG_DIR)\ffmpeg" del /Q "$(FFMPEG_DIR)\ffmpeg" > NUL 2>&1 || echo "No Linux ffmpeg found"
	@if exist "$(FFMPEG_DIR)\ffmpeg-*.zip" del /Q "$(FFMPEG_DIR)\ffmpeg-*.zip" > NUL 2>&1 || echo "No zip files found"
	@if exist "$(FFMPEG_DIR)\ffmpeg-*.tar.xz" del /Q "$(FFMPEG_DIR)\ffmpeg-*.tar.xz" > NUL 2>&1 || echo "No tar.xz files found"
	@if exist "$(FFMPEG_DIR)\ffmpeg-master-*" rmdir /S /Q "$(FFMPEG_DIR)\ffmpeg-master-*" > NUL 2>&1 || echo "No extracted directories found"
	@echo FFmpeg files cleaned!
