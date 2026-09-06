# gofluent/examples — 迁移约定与构建说明

本目录把 `PyQt-Fluent-Widgets-master/examples` 的 110 个 Python 源文件（约 70 个独立 demo、
gallery 展示应用、多个 window 应用含 `.ui`/`.qrc`）迁移为 Go/miqt 可执行示例，复用已迁移的
`gofluent` 库。每个示例都是一个自包含的 `package main`，共享统一入口与资源嵌入约定。

## 1. 目录布局（1:1 镜像 Python examples）

```
gofluent/examples/
├── README.md                 # 本文档
├── build_windows.sh          # WSL+MXE 静态交叉编译脚本
├── hello/                    # 规范模板（唯一真实可编译的参考示例）
│   ├── main.go
│   └── resource/hello.qss
├── internal/
│   ├── demo/                 # 共享 demo 入口（package demo）
│   └── asset/                # 共享资源读取助手（package asset）
├── basic_input/  button/ check_box/ combo_box/ model_combo_box/
│                 radio_button/ slider/ switch_button/
├── date_time/    calendar_picker/ fast_calendar_picker/ time_picker/
├── dialog_flyout/ color_dialog/ custom_message_box/ dialog/ flyout/
│                 folder_list_dialog/ message_dialog/ teaching_tip/
├── gallery/      （展示应用，含 app/{common,components,resource,view}）
├── layout/       adaptive_flow_layout/ flow_layout/
├── material/     acrylic_brush/ acrylic_combo_box/ acrylic_flyout/ acrylic_label/
│                 acrylic_line_edit/ acrylic_menu/ acrylic_tool_tip/ acrylic_widget_menu/
├── media/        avatar_widget/ media_player/
├── menu/         command_bar/ menu/ system_tray_menu/ widget_menu/
├── navigation/   breadcrumb_bar/ navigation_bar/ navigation_header/ navigation_user_card/
│                 navigation1/ navigation2/ navigation3/ pivot/ segmented_tool_widget/
│                 segmented_widget/ stacked_widget/ tab_view/ tab_widget/
├── scroll/       pips_pager/ scroll_area/
├── status_info/  info_badge/ info_bar/ progress_bar/ progress_ring/ state_tool_tip/ tool_tip/
├── text/         font_icon/ image_label/ label/ line_edit/ spin_box/ text_browser/
├── view/         card_widget/ flip_view/ list_view/ table_view/ tree_view/ tree_widget/
└── window/       clock/ fluent_widget/ fluent_window/ login/ ms_fluent_window/
                  settings/ splash_screen/ split_fluent_window/ web_engine/
```

命名规则：每个 demo 目录 = `examples/<category>/<name>/`，Go 包名为 `main`，对应 `.exe` 名为
`<category>_<name>.exe`（gallery 特例为 `gallery.exe`，见 `build_windows.sh`）。目录里若有
`resource/`（或 `.ui`/`.qrc` 对应的生成代码）随 demo 一起迁移。

## 2. 共享 demo 入口（`internal/demo`）

统一入口位于 `gofluent/examples/internal/demo`，把 PyQt 每个 demo 结尾重复的
`if __name__ == '__main__'` 样板（High-DPI 设置 + `QApplication` + `show` + `exec_`）收敛为一处：

| 函数 | 作用 |
|---|---|
| `demo.Run(newWindows ...func() *qt.QWidget)` | 一站式：High-DPI 设置 → 建 `QApplication` → 调用工厂创建并 `Show` 所有窗口 → 跑事件循环（永不返回） |
| `demo.SetupHighDPI()` | 仅应用 High-DPI 属性（`PassThrough`、`AA_EnableHighDpiScaling`、`AA_UseHighDpiPixmaps`） |
| `demo.NewApp() *qt.QApplication` | 建并持有唯一 `QApplication`（供需要装翻译器等场景） |
| `demo.Exec()` | 跑事件循环并 `os.Exit` 返回码（永不返回） |

简单 demo 的 `main.go` 模板：

```go
package main

import (
    "gofluent/examples/internal/demo"
)

func main() {
    demo.Run(newDemoWidget)
}
```

其中 `newDemoWidget` 是 `func newDemoWidget() *qt.QWidget` 工厂函数：`demo.Run` 会先建好
`QApplication` 再调用它，因此窗口对象只在 `QApplication` 之后创建。

需要“建 app 之后、跑循环之前”做额外配置（如 gallery 装翻译器）时，拆开写：

```go
func main() {
    demo.SetupHighDPI()
    app := demo.NewApp()
    // install translators / set window icon / ...
    w := mainwindow.NewMainWindow()
    w.Show()
    demo.Exec()
}
```

要点：

- miqt 在首次 `NewQApplication` 时对 Go 运行时 `LockOSThread`，因此 `demo.*` 全部必须在
  **main goroutine** 调用。
- `demo.Run`/`demo.Exec` 内部 `os.Exit`，demo 自身的 `defer` 不会执行（与 Python
  `sys.exit(app.exec_())` 行为一致）；进程退出即释放，无泄漏问题。

## 3. 资源嵌入约定（`internal/asset` + `go:embed`）

遵循 MIGRATION_GUIDE §11 的 `go:embed` 方案（优先于 `miqt-rcc`）。约束：`go:embed` 只能引用
声明它的 `.go` 文件所在目录树内，因此每个 demo 的资源必须**复制到该 demo 自己的目录**下，
约定放在 `resource/` 子目录（结构对齐 Python 原 `resource/`，如 `resource/dark/demo.qss`）。

每个需要资源的 demo 在包内用 `go:embed` 声明，再用 `internal/asset` 转成 miqt 对象：

```go
import (
    _ "embed"
    "embed"

    "gofluent/examples/internal/asset"
)

// 单个文件：直接嵌入为 []byte / string
//go:embed resource/logo.png
var logoPNG []byte

// 多个/目录：嵌入为 embed.FS，用 asset 读取
//go:embed resource/dark/*.qss resource/light/*.qss
var demoQSS embed.FS

func loadQSS(dark bool) string {
    dir := "resource/light/"
    if dark { dir = "resource/dark/" }
    return asset.String(demoQSS, dir+"demo.qss")
}
```

`internal/asset` 提供的助手（返回的 Qt 对象由调用方负责 `Delete()`，见 MIGRATION_GUIDE §6）：

| 函数 | 说明 |
|---|---|
| `asset.Bytes(fsys, name) []byte` | 读嵌入文件，缺失即 panic |
| `asset.String(fsys, name) string` | 读嵌入文本（QSS/翻译/JSON） |
| `asset.QImage(fsys, name) *qt.QImage` | 解码图片（`QImage_FromDataWithData`） |
| `asset.QPixmap(fsys, name) *qt.QPixmap` | 解码图片（`QPixmap.LoadFromDataWithData`） |
| `asset.QIcon(fsys, name) *qt.QIcon` | 解码为图标（`NewQIcon2`） |
| `asset.FontID(data []byte) int` | 加载字体（`QFontDatabase_AddApplicationFontFromData`） |

> Python 里 `QIcon('resource/logo.png')` / `open('resource/dark/demo.qss')` 这类**运行时路径**访问，
> Go 侧一律改为“`go:embed` + `asset` 取字节/对象”，不要保留相对路径读取（静态链接的 .exe 里
> 没有独立资源文件）。`.ui` 文件按 MIGRATION_GUIDE §4/§11 用 `miqt-rcc` 或等价方式生成 Go 代码，
> 资源文件仍走 `go:embed`。

## 4. 构建（WSL + MXE 静态交叉编译）

脚本：`gofluent/examples/build_windows.sh`（须在 WSL Ubuntu 24.04 内执行）。相比库根的
`gofluent/build_windows.sh`，本脚本**补上了 `PKG_CONFIG_PATH`**（后者缺失，导致 pkg-config 找不到
`Qt5Widgets`）。完整封装 TASK.md §Constraints 的环境变量 + `--tags=windowsqtstatic` + `-ldflags "-s -w"`。

```bash
cd /path/to/gofluent
bash examples/build_windows.sh              # 快检：go vet ./... + 全量链接 gallery/settings（若已实现）
FULL_BUILD=1 bash examples/build_windows.sh # 完整：把每个 main 包逐个静态链接成 dist/<category>_<name>.exe
```

- 默认模式先 `go vet --tags=windowsqtstatic ./...` 做全模块编译级检查（覆盖库 + 所有 demo），
  再静态链接代表性应用 `gallery`、`window/settings`（目录尚未实现时自动跳过）。
- `FULL_BUILD=1` 通过 `go list -f '{{if eq .Name "main"}}...' ./examples/...` 枚举所有 main 包逐个链接；
  目录里尚未添加 `main.go` 时会被自动跳过，随迁移推进自动纳入。
- 产物落在 `gofluent/dist/`。

## 5. 迁移自检清单（写入各自任务 output）

- 每个 demo 是 `package main`，`main()` 只做“控件构建封装为工厂函数 + `demo.Run`”（或拆开的 `demo.*` 三步）。
- `gofmt -l .` 为空；`go vet --tags=windowsqtstatic ./...` 无错误（在 WSL 内）。
- 引用的 miqt 符号均在 `miqt-master/qt/gen_*.go` 中实际存在（含数字后缀重载号）。
- QSS 字符串与 Python `resource/**/*.qss` 逐字一致，`SetStyleSheet` 调用位置与 Python 版对齐。
- 不再使用的 Qt 对象 `DeleteLater()`，临时值类型 `defer x.Delete()`。

## 6. 构建结果（t6）

- **类型级全量校验**：`gopls check` 覆盖 gofluent 库 + examples 共 213 个 `.go` 文件，**0 类型错误**。
- **格式**：`gofmt -l`（全部 `.go`）为空。
- **纯 Go 冒烟**：`go build ./resources`、`go test ./resources` 均通过（exit 0）。
- **编译错误修复**：window 系列 4 处类型错误已修复——`settings/main.go` 的 `app.SetAttribute`
  → `qt.QCoreApplication_SetAttribute`；`clock/view/focus_interface.go` 的 `widgets.TimePicker/NewTimePicker`
  → `date_time.TimePicker/NewTimePicker`（补 import `gofluent/components/date_time`）；`web_engine/main.go`
  的 `NewQLabel3(text, parent)` → `NewQLabel5(text, parent)`。另有 10 处 window main.go 未导出字段引用
  已改用 `NavigationInterface()` / `NavigationBar()` 访问器（访问器由 t3 提供）。
- **main 包枚举**：`go list` 发现 examples 下 **77 个 `package main`**（含 gallery、window/settings 等代表性应用）。
- **环境限制（非代码问题）**：`windows/amd64` 静态交叉编译须在 WSL Ubuntu 24.04 + MXE 内执行；
  当前构建沙箱禁止 WSL 服务访问（`wsl --status` → `Wsl/EnumerateDistros/Service/E_ACCESSDENIED`），
  且 Windows 宿主无 MXE/Qt5（`go build ./...` → `Package Qt5Widgets was not found in the pkg-config search path`），
  故本次会话未能实际产出 `.exe`。`build_windows.sh` 命令与环境变量已就绪，在真实 WSL+MXE 环境直接
  `bash examples/build_windows.sh`（或 `FULL_BUILD=1`）即可完成静态链接。
