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

#### 方式一：从源码运行

```bash
go run ./cmd/histprune version
```

#### 方式二：安装到本机

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

分析 zsh history：

```bash
histprune analyze --file ~/.zsh_history
```

预览去重结果，不修改文件：

```bash
histprune prune --file ~/.zsh_history --dedupe
```

确认后写入：

```bash
histprune prune --file ~/.zsh_history --dedupe --write
```

删除包含指定文本的历史：

```bash
histprune prune --file ~/.zsh_history --contains 'gti status'
histprune prune --file ~/.zsh_history --contains 'gti status' --write
```

删除匹配正则的历史：

```bash
histprune prune --file ~/.zsh_history --regex 'token=[^ ]+'
```

按行号删除：

```bash
histprune prune --file ~/.zsh_history --line 1287
```

按时间删除：

```bash
histprune prune --file ~/.zsh_history --before 2024-01-01
histprune prune --file ~/.zsh_history --between 2024-01-01 2024-06-30
```

输出 JSON 报告：

```bash
histprune prune --file ~/.zsh_history --dedupe --json
```

## 安全模型

`prune` 默认是 dry-run，不会修改历史文件。只有显式传入 `--write` 时才会落盘：

```bash
histprune prune --file ~/.zsh_history --dedupe --write
```

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

## 备份与恢复

列出备份：

```bash
histprune backups --file ~/.zsh_history
```

恢复最近一次备份：

```bash
histprune restore latest --file ~/.zsh_history
```

恢复指定备份：

```bash
histprune restore /path/to/.zsh_history.histprune-backup-20260504T120000 --file ~/.zsh_history
```

恢复前会先备份当前 history 文件。

## 配置

当前版本暂不读取配置文件，所有行为都通过 CLI flags 配置。

因此本仓库现在不放 `configs/config.example.yaml`，避免出现“看起来支持配置、实际还没接入”的误导。后续如果加入结构化规则配置，会采用：

- `configs/config.example.yaml`：提交到版本库的模板配置
- `configs/config.yaml`：本地真实配置，不提交到版本库

敏感信息不应写入 history，也不需要放在本项目配置中；如果未来确实需要敏感配置，会使用 `.example.env` + 本地忽略的 `.env`。

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
