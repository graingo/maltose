#!/bin/bash

# 脚本会在遇到任何错误时立即退出
set -e

# 检查是否提供了版本号参数
if [ -z "$1" ]; then
    echo "错误：请提供版本号作为参数。"
    echo "用法: ./release.sh v0.1.0"
    exit 1
fi

VERSION=$1

echo "将为所有模块发布版本: $VERSION"
echo ""

# 1. 根据规则更新所有子模块的 go.mod 文件
echo "===== [1/3] 更新子模块依赖 ====="
find . -name go.mod | while read -r mod_file; do
    dir=$(dirname "$mod_file")
    if [ "$dir" == "." ]; then
        continue
    fi

    module_path=${dir#./}
    echo "-> 正在处理模块: $module_path"

    # 使用子 shell, 避免影响当前目录
    (
        cd "$dir"

        # 规则：cmd/maltose 模块移除 replace 指令
        if [[ "$module_path" == "cmd/maltose" ]]; then
            echo "   - [CMD 规则] 正在移除 replace 指令..."
            if grep -q "github.com/graingo/maltose =>" go.mod; then
                go mod edit -dropreplace=github.com/graingo/maltose
            fi
        # 规则：contrib/* 模块保留 replace 指令
        elif [[ "$module_path" == contrib/* ]]; then
            echo "   - [CONTRIB 规则] 保留 replace 指令。"
            # 此处无需操作
        fi

        echo "   - 设置 maltose 依赖版本为 $VERSION..."
        go mod edit -require="github.com/graingo/maltose@$VERSION"
        echo "   - 整理 go.mod..."
        go mod tidy
    )
done
echo "================================"
echo ""

# 2. 提交 go.mod 的变更
echo "===== [2/3] 提交变更 ====="
# 检查是否有文件变更需要提交
if [[ -z $(git status -s --untracked-files=no) ]]; then
    echo "-> 没有文件变更，无需提交。"
else
    echo "-> 正在提交所有 go.mod 和 go.sum 文件的变更..."
    git add .
    git commit -m "chore(release): align dependencies for $VERSION"
fi
echo "========================"
echo ""

# 3. 为所有模块打标签
echo "===== [3/3] 创建标签 ====="
# 为根模块打标签
echo "-> 正在为根模块打标签: $VERSION"
git tag "$VERSION"

# 查找所有子模块的 go.mod 并打标签
find . -name go.mod | while read -r mod_file; do
    dir=$(dirname "$mod_file")
    if [ "$dir" == "." ]; then
        continue
    fi

    module_path=${dir#./}
    TAG_NAME="${module_path}/${VERSION}"

    echo "-> 正在为子模块 '$module_path' 打标签: $TAG_NAME"
    git tag "$TAG_NAME"
done
echo "======================="
echo ""

echo "✅ 所有模块版本更新和标签创建成功！"
echo ""
echo "下一步，请运行以下命令将【提交】和【标签】都推送到远程仓库:"
echo "  git push && git push --tags"
