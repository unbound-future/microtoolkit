#!/bin/sh

# 读取服务名称
SERVICE_NAME=$(cat /app/service_name)

echo "Starting ${SERVICE_NAME} service..."

# 确保二进制文件有执行权限
chmod +x /app/service

# 检查文件是否存在
if [ ! -f "/app/service" ]; then
    echo "Error: Service binary /app/service not found!"
    exit 1
fi

# 检查文件是否可执行
if [ ! -x "/app/service" ]; then
    echo "Error: Service binary /app/service is not executable!"
    exit 1
fi

echo "Executing /app/service..."

# 启动服务
exec /app/service
