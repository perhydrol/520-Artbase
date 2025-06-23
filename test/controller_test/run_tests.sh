#!/bin/bash

# Controller集成测试运行脚本
# 用于运行所有controller层的集成测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否在正确的目录
if [ ! -f "../../go.mod" ]; then
    print_error "请在项目根目录下运行此脚本"
    exit 1
fi

# 切换到项目根目录
cd ../..

print_info "开始运行Controller集成测试..."

# 设置测试环境变量
export GO_ENV=test
export GIN_MODE=test

# 检查测试依赖
print_info "检查测试依赖..."
if ! go mod verify; then
    print_error "Go模块验证失败"
    exit 1
fi

# 运行测试前的准备
print_info "准备测试环境..."

# 确保测试数据库配置正确
if [ ! -f "configs/demo520.yaml" ]; then
    print_warning "配置文件不存在，请确保configs/demo520.yaml存在"
fi

# 创建测试临时目录
mkdir -p test/controller_test/temp_image

# 运行具体的测试
print_info "运行用户Controller测试..."
if go test -v ./test/controller_test -run TestUserController; then
    print_success "用户Controller测试通过"
else
    print_error "用户Controller测试失败"
    exit 1
fi

print_info "运行图片Controller测试..."
if go test -v ./test/controller_test -run TestImageController; then
    print_success "图片Controller测试通过"
else
    print_error "图片Controller测试失败"
    exit 1
fi

print_info "运行完整Controller集成测试..."
if go test -v ./test/controller_test -run TestControllerIntegration; then
    print_success "完整Controller集成测试通过"
else
    print_error "完整Controller集成测试失败"
    exit 1
fi

# 生成测试覆盖率报告
print_info "生成测试覆盖率报告..."
if go test -v ./test/controller_test -coverprofile=test/controller_test/coverage.out; then
    go tool cover -html=test/controller_test/coverage.out -o test/controller_test/coverage.html
    print_success "测试覆盖率报告已生成: test/controller_test/coverage.html"
else
    print_warning "测试覆盖率报告生成失败"
fi

# 清理测试临时文件
print_info "清理测试临时文件..."
rm -rf test/controller_test/temp_image/*

print_success "所有Controller集成测试完成！"

# 显示测试结果摘要
print_info "测试结果摘要:"
echo "  ✓ 用户Controller测试"
echo "  ✓ 图片Controller测试"
echo "  ✓ 完整Controller集成测试"
echo "  ✓ 测试覆盖率报告"

print_info "要查看详细的测试覆盖率，请打开: test/controller_test/coverage.html"