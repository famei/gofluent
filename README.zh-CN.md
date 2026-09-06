# go-Fluent

**go-Fluent** 是基于 [miqt](https://github.com/mappu/miqt)（Go 的 Qt5 绑定）对 [PyQt-Fluent-Widgets](https://github.com/zhiyiYo/PyQt-Fluent-Widgets) 的 Go 移植。它把这套现代、Windows 11 风格的 Fluent 组件库带入 Go，让你无需 Python 就能用相同的观感编写桌面应用。

> ⚙️ **由 DeepSeek 生成。** 本仓库中的大部分源代码、文档与迁移工作均由 DeepSeek 模型产出。

> 📄 **For English see [README.md](README.md).**

> ⚠ 注意此项目处于初期阶段,请不要在生产环境上运行!

## 关于

go-Fluent 将 PyQt-Fluent-Widgets 的 Fluent 组件集移植到 Go + Qt5，涵盖按钮、卡片、输入框、列表、菜单、导航面板、设置卡片、日期/时间选择器、对话框和布局助手。每个控件尽可能贴近原始 Python API，同时适配 Go 的惯用写法（内嵌 `*qt.QWidget` 字段、回调字段和 `NewXxx` 构造函数）。

### 引用项目

| 项目 | 作用 | 许可证 |
| --- | --- | --- |
| [zhiyiYo/PyQt-Fluent-Widgets](https://github.com/zhiyiYo/PyQt-Fluent-Widgets) | 本项目移植的原始组件库（API 与视觉参考） | [GPL-3.0](https://www.gnu.org/licenses/gpl-3.0.html) |
| [mappu/miqt](https://github.com/mappu/miqt) | 本项目所依赖的 Qt Go 绑定 | [MIT](https://opensource.org/licenses/MIT) |

## 特性

- **Fluent 设计** — 圆角卡片、主题色按钮、带动画的开关、胶囊开关、Windows 11 风格导航。
- **组件覆盖广** — 100+ 控件，分布在 `widgets`、`navigation`、`dialog_box`、`settings`、`date_time`、`layout` 包中。

## 目录结构

```
gofluent/
├── common/          # 主题、样式表、字体、图标、配置等公共设施
├── components/      # 组件库
│   ├── widgets/     # 基础控件
│   ├── navigation/  # 导航
│   ├── dialog_box/  # 对话框
│   ├── settings/    # 设置卡片
│   ├── date_time/   # 日期/时间选择
│   └── layout/      # 布局
├── examples/        # 80+ 可运行示例
├── resources/       # QSS 主题、图标、字体
└── docs/            # 使用文档
```

## 编译

本项目构建在 [miqt](https://github.com/mappu/miqt) 之上，Qt 工具链与交叉编译环境的配置请参照 miqt 官方指南，见 **[miqt › Building](https://github.com/mappu/miqt#building)**（Windows 交叉编译：MXE / Docker / MSYS2）。

工具链就绪后，使用本仓库脚本：

```bash
cd gofluent

# 编译 gallery 演示
bash build_exes.sh

# 编译全部示例，每个 .exe 输出到各自的 main.go 同目录
bash build_all_examples.sh
```

## 文档

- [控件使用文档（中文）](docs/控件使用文档.md) — 每个控件的功能、参数、调用方法与效果。
- [Controls Usage Guide (English)](docs/controls_usage.md) — function, parameters, usage and effect for every control.

## 许可证

**GPL-3.0** — 见 [LICENSE](LICENSE)。

本项目是 PyQt-Fluent-Widgets 的移植，后者采用 **GPL-3.0**。GPL-3.0 是与两个上游项目都兼容的许可证：GPL 要求衍生作品保持同一许可证，而 miqt 的宽松 MIT 许可证允许其代码被并入 GPL-3.0 项目。
