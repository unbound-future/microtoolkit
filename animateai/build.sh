#!/bin/bash

# 支持的服务列表
SERVICES=("gateway" "billing" "controller" "patrol")

# 显示帮助信息
show_help() {
    echo "使用方法: $0 [服务名]"
    echo ""
    echo "支持的服务:"
    for service in "${SERVICES[@]}"; do
        echo "  - $service"
    done
    echo "  - all (构建所有服务)"
    echo ""
    echo "功能:"
    echo "  - 自动构建 Docker 镜像"
    echo "  - 自动打标签到腾讯云容器镜像仓库"
    echo "  - 自动推送到远程仓库"
    echo ""
    echo "示例:"
    echo "  $0 gateway          # 构建并推送 gateway 服务"
    echo "  $0 all              # 构建并推送所有服务"
    echo "  $0                  # 默认构建并推送 gateway 服务"
}

# 检查服务名是否有效
is_valid_service() {
    local service=$1
    for valid_service in "${SERVICES[@]}"; do
        if [ "$service" = "$valid_service" ]; then
            return 0
        fi
    done
    return 1
}

# 构建单个服务
build_service() {
    local service=$1
    echo "构建服务: $service"
    
    # 设置构建时间戳环境变量
    export BUILD_TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    echo "构建时间戳: $BUILD_TIMESTAMP"
    
    python3.11 -m podman_compose build $service
    
    if [ $? -eq 0 ]; then
        echo "✅ $service 构建成功！"
        
        # 打标签
        local original_tag="tunnel-api-${service}:${BUILD_TIMESTAMP}"
        local registry_tag="furion-sh.tencentcloudcr.com/furion/tunnel-api-${service}:${BUILD_TIMESTAMP}"
        
        echo "正在打标签: $original_tag -> $registry_tag"
        docker tag $original_tag $registry_tag
        
        if [ $? -eq 0 ]; then
            echo "✅ 标签创建成功！"
            
            # 推送到仓库
            echo "正在推送到仓库: $registry_tag"
            podman push $registry_tag
            
            if [ $? -eq 0 ]; then
                echo "✅ $service 推送成功！"
            else
                echo "❌ $service 推送失败！"
                return 1
            fi
        else
            echo "❌ 标签创建失败！"
            return 1
        fi
    else
        echo "❌ $service 构建失败！"
        return 1
    fi
}

# 构建所有服务
build_all() {
    echo "开始构建所有服务..."
    echo ""
    
    local failed_services=()
    
    for service in "${SERVICES[@]}"; do
        echo "正在构建 $service..."
        if ! build_service $service; then
            failed_services+=($service)
        fi
        echo ""
    done
    
    if [ ${#failed_services[@]} -eq 0 ]; then
        echo "🎉 所有服务构建成功！"
    else
        echo "⚠️  以下服务构建失败: ${failed_services[*]}"
        exit 1
    fi
}

# 主逻辑
case "${1:-gateway}" in
    "help"|"-h"|"--help")
        show_help
        ;;
    "all")
        build_all
        ;;
    *)
        if is_valid_service "$1"; then
            build_service "$1"
        else
            echo "❌ 错误: 不支持的服务 '$1'"
            echo ""
            show_help
            exit 1
        fi
        ;;
esac