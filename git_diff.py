import subprocess
from pathlib import Path

def export_unstaged_changes(output_file="unstaged.diff"):
    """
    将 Git 仓库中所有未暂存的更改导出到指定文件中（兼容 Windows 编码问题）。
    """
    try:
        # 检查是否在 Git 仓库中
        subprocess.run(
            ["git", "rev-parse", "--is-inside-work-tree"],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )

        # 强制用 UTF-8 编码读取输出
        result = subprocess.run(
            ["git", "diff"],
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )

        # 解码 stdout 字节流为 UTF-8 字符串
        diff_text = result.stdout.decode("utf-8", errors="replace")

        # 写入文件（也用 UTF-8 编码）
        Path(output_file).write_text(diff_text, encoding="utf-8")

        print(f"✅ 未暂存的更改已导出到 {output_file}")

    except subprocess.CalledProcessError as e:
        print("❌ Git 命令执行失败。")
        print(e.stderr.decode("utf-8", errors="replace"))
    except Exception as e:
        print(f"❌ 发生未知错误：{e}")

if __name__ == "__main__":
    export_unstaged_changes()
