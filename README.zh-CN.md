# histprune

[入口页](./README.md) | [English](./README.en.md)

`histprune` 是一个用 Go 写的 shell history 清理工具，第一版重点支持 zsh。它的目标不是“聪明地替你猜哪些命令错了”，而是安全、可解释、可恢复地清理历史记录。

## 特性

- first-class 支持 zsh history，包括 extended history 格式
- 默认 dry-run，只预览不写入
- 每条删除记录都有原因说明
- 去重时保留最后一次出现的命令
- 支持按文本、正则、行号、时间范围删除
- 写入前自动创建备份
- 支持列出备份和恢复备份
- 支持文本报告和 JSON 报告

## 安装

安装 CLI：

```bash
go install github.com/walker1211/histprune/cmd/histprune@latest
histprune version
```

仓库本地开发时可以用脚本构建二进制：

```bash
./build.sh
./histprune version
```

## 快速开始

分析默认 zsh history 文件：

```bash
histprune analyze
```

先预览清理结果：

```bash
histprune prune --dedupe
```

确认预览后，用同一条命令加 `--write` 实际写入：

```bash
histprune prune --dedupe --write
```

默认情况下，`histprune` 会优先使用 `$HISTFILE`，否则回退到 `~/.zsh_history`。只有清理其他 history 文件时才需要 `--file PATH`：

```bash
histprune analyze --file /path/to/history
```

查看所有命令和参数：

```bash
histprune --help
```

## 清理规则

所有 `prune` 命令默认只预览，不会修改文件。先看报告，确认后再给同一条命令追加 `--write`。

去重并保留最后一次出现：

```bash
histprune prune --dedupe
```

删除包含指定文本的历史：

```bash
histprune prune --contains 'gti status'
```

删除匹配正则的历史：

```bash
histprune prune --regex 'token=[^ ]+'
```

按行号删除：

```bash
histprune prune --line 1287
```

按时间删除：

```bash
histprune prune --before 2024-01-01
histprune prune --between 2024-01-01 2024-06-30
```

输出 JSON 报告：

```bash
histprune prune --dedupe --json
```

## 安全模型

`prune` 默认是 dry-run，不会修改历史文件。只有显式传入 `--write` 时才会落盘。

写入时会：

1. 读取并解析 history 文件
2. 生成删除决策和原因
3. 先创建 `*.histprune-backup-*` 备份
4. 通过临时文件和 rename 替换原文件

修改当前 zsh 正在使用的 history 后，手动重新加载：

```bash
fc -R ~/.zsh_history
```

如果你打开了多个 zsh 终端，建议清理前先在当前 shell 中写出最新历史：

```bash
fc -W ~/.zsh_history
```

## 可选：推荐的 zsh history 配置

`histprune` 不要求修改 zsh 配置，但下面这些设置可以让 history 文件更适合后续清理：

```zsh
# zsh history
HISTFILE="$HOME/.zsh_history"
HISTSIZE=50000
SAVEHIST=50000

setopt EXTENDED_HISTORY        # 保存时间戳
setopt APPEND_HISTORY          # 追加写入历史文件
setopt INC_APPEND_HISTORY      # 命令执行后立即写入
```

其中 `EXTENDED_HISTORY` 会保存时间戳，让 `--before` / `--between` 这类时间范围规则更有用；`APPEND_HISTORY` 和 `INC_APPEND_HISTORY` 可以减少多个终端覆盖 history 的风险，让文件更接近当前执行状态。

如果你希望 zsh 自己减少之后产生的重复历史，也可以额外开启：

```zsh
setopt HIST_IGNORE_DUPS        # 连续重复命令不记录
setopt HIST_IGNORE_ALL_DUPS    # 新命令重复时删除旧记录
setopt HIST_SAVE_NO_DUPS       # 保存到文件时不写重复项
setopt HIST_EXPIRE_DUPS_FIRST  # 历史满了优先删除重复项
setopt HIST_FIND_NO_DUPS       # 搜索历史时不显示重复项
setopt HIST_REDUCE_BLANKS      # 压缩多余空格
setopt HIST_IGNORE_SPACE       # 以空格开头的命令不记录
```

这些选项不能保证 `~/.zsh_history` 文件里永远没有重复项。已有重复项、配置生效前的历史、多终端并发写入，仍然可能被 `histprune prune --dedupe` 扫描出来。建议把 zsh 配置理解为减少新重复，`histprune` 负责对磁盘 history 做可预览、可备份、可恢复的清理。

## 备份与恢复

列出备份：

```bash
histprune backups
```

恢复最近一次备份：

```bash
histprune restore latest
```

恢复指定备份：

```bash
histprune restore /path/to/.zsh_history.histprune-backup-20260504T120000
```

恢复前会先备份当前 history 文件。

## 开发 / 测试

```bash
go test ./...
go vet ./...
```

## Roadmap

- bash / fish history 支持
- 配置文件规则
- 备份轮转
- secret redaction
- typo / failed command 分析
- Homebrew / release automation

## License

本项目使用 MIT License，详见 [LICENSE](./LICENSE)。
